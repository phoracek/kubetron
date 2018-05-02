// TODO: remove orhpan interfaces in another thread
package deviceplugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockercli "github.com/docker/docker/client"
	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"github.com/phoracek/kubetron/pkg/spec"
	"github.com/vishvananda/netlink"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
)

const (
	nicsPoolSize               = 100
	interfaceNamePrefix        = "nic_"
	letterBytes                = "abcdefghijklmnopqrstuvwxyz0123456789"
	fakeDeviceHostPath         = "/var/run/kubetron-fakedev"
	fakeDeviceGuestPath        = "/tmp/kubetron-fakedev"
	devicepluginCheckpointPath = "/var/lib/kubelet/device-plugins/kubelet_internal_checkpoint"
	networksSpecAnnotationName = "kubetron.network.kubevirt.io/networksSpec"
)

type Lister struct {
	ResourceName      string
	ResourceNamespace string
}

func (l Lister) GetResourceNamespace() string {
	return l.ResourceNamespace
}

func (l Lister) Discover(pluginListCh chan dpm.PluginNameList) {
	pluginListCh <- dpm.PluginNameList{l.ResourceNamespace + "/" + l.ResourceName}
}

func (l Lister) NewPlugin(bridge string) dpm.PluginInterface {
	return DevicePlugin{}
}

type DevicePlugin struct {
	kubeclient   *kubernetes.Clientset
	dockerclient *dockercli.Client
}

func (dp *DevicePlugin) Start() error {
	kubeClientConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to obtain kubernetes client config: %v", err)
	}

	kubeclient, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return fmt.Errorf("failed to intialize kubernetes client: %v", err)
	}
	dp.kubeclient = kubeclient

	dockerclient, err := dockercli.NewEnvClient()
	if err != nil {
		return fmt.Errorf("failed to intialize docker client: %v", err)
	}
	dp.dockerclient = dockerclient

	err = createFakeDevice()

	return err
}

func (dp *DevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	var bridgeDevs []*pluginapi.Device
	for i := 0; i < nicsPoolSize; i++ {
		bridgeDevs = append(bridgeDevs, &pluginapi.Device{
			ID:     fmt.Sprintf("nic-%02d", i),
			Health: pluginapi.Healthy,
		})
	}
	s.Send(&pluginapi.ListAndWatchResponse{Devices: bridgeDevs})
	for {
		time.Sleep(10 * time.Second)
	}
	return nil
}

func (dp *DevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	var response pluginapi.AllocateResponse

	if len(r.DevicesIDs) != 1 {
		return nil, fmt.Errorf("allocate request must contain exactly one device")
	}

	nic := r.DevicesIDs[0]
	dev := new(pluginapi.DeviceSpec)
	dev.HostPath = fakeDeviceHostPath
	dev.ContainerPath = fakeDeviceGuestPath
	dev.Permissions = "r"
	response.Devices = append(response.Devices, dev)

	// TODO: dont return error, just log it and exit
	go func() {

		time.Sleep(time.Second)

		checkpointRaw, err := ioutil.ReadFile(devicepluginCheckpointPath)
		if err != nil {
			panic(fmt.Errorf("failed to read device plugin checkpoint file: %v", err))
		}

		var checkpoint map[string]interface{}
		err = json.Unmarshal(checkpointRaw, &checkpoint)
		if err != nil {
			panic(fmt.Errorf("failed to unmarshal device plugin checkpoint file: %v", err))
		}

		podUID := ""
		entries := checkpoint["Entries"].([]map[string]interface{})
	EntriesLoop:
		for _, entry := range entries {
			for _, deviceID := range entry["DeviceIDs"].([]string) {
				if deviceID == nic {
					podUID = entry["PodUID"].(string)
					break EntriesLoop
				}
			}
		}
		if podUID == "" {
			panic(fmt.Errorf("failed to find PodUID"))
		}

		var thePod v1.Pod
		podFound := false
		pods, err := dp.kubeclient.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(fmt.Errorf("failed to list pods: %v", err))
		}
		for _, pod := range pods.Items {
			fmt.Println(pod.Name, pod.Status.PodIP)
			if string(pod.UID) == podUID {
				thePod = pod
				podFound = true
				break
			}
		}
		if !podFound {
			panic(fmt.Errorf("failed to find pod with given PodUID"))
		}

		podName := thePod.Name
		podNamespace := thePod.Namespace
		var networksSpec spec.NetworksSpec
		err = json.Unmarshal([]byte(thePod.Labels[networksSpecAnnotationName]), &networksSpec)
		if err != nil {
			panic(fmt.Errorf("failed to read networks spec: %v", err))
		}

		containerName := fmt.Sprintf("k8s_POD_%s_%s", podName, podNamespace)

		containers, err := dp.dockerclient.ContainerList(context.Background(), dockertypes.ContainerListOptions{})
		if err != nil {
			panic(fmt.Errorf("failed to list docker containers: %v", err))
		}

		containerPid := -1
		for i := 0; i <= 10; i++ {
			for _, container := range containers {
				config, err := dp.dockerclient.ContainerInspect(context.Background(), container.ID)
				if err != nil {
					panic(fmt.Errorf("failed to inspect docker container: %v", err))
				}

				if config.Name == containerName {
					containerPid = config.State.Pid
					break
				}
			}
			time.Sleep(10 * time.Second)
		}
		if containerPid == -1 {
			panic(fmt.Errorf("Failed to find container PID"))
		}

		for _, spec := range networksSpec {
			err = exec.Command(
				"ovs-vsctl", "--",
				"add-port", "br-int", spec.PortName, "--",
				"set", "Interface", spec.PortName, "type=internal",
			).Run()
			if err != nil {
				panic(err)
			}

			port, err := netlink.LinkByName(spec.PortName)
			if err != nil {
				panic(err)
			}

			err = netlink.LinkSetNsPid(port, containerPid)
			if err != nil {
				panic(err)
			}

			hwaddr, err := net.ParseMAC(spec.MacAddress)
			if err != nil {
				panic(err)
			}

			err = netlink.LinkSetHardwareAddr(port, hwaddr)
			if err != nil {
				panic(err)
			}

			err = netlink.LinkSetUp(port)
			if err != nil {
				panic(err)
			}

			err = exec.Command(
				"ovs-vsctl", "set", "Interface", spec.PortName, fmt.Sprintf("external_ids:iface-id=%s", spec.PortID),
			).Run()
			if err != nil {
				panic(err)
			}
			// TODO: call dhclient ipam
		}

	}()

	return &response, nil
}

func createFakeDevice() error {
	_, stat_err := os.Stat(fakeDeviceHostPath)
	if stat_err == nil {
		glog.V(3).Info("Fake block device already exists")
		return nil
	} else if os.IsNotExist(stat_err) {
		glog.V(3).Info("Creating fake block device")
		cmd := exec.Command("mknod", fakeDeviceHostPath, "b", "1", "1")
		err := cmd.Run()
		return err
	} else {
		panic(stat_err)
	}
}
