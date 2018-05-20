// TODO: remove orhpan interfaces in another thread
// TODO: use prestart container request, no need to wait
// TODO: cleanup if a step fails
// TODO: load networkSpec annotation name from common module
package deviceplugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockercli "github.com/docker/docker/client"
	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"github.com/phoracek/kubetron/pkg/spec"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	nicsPoolSize               = 100
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

// TODO: check if br-int is available
func (l Lister) Discover(pluginListCh chan dpm.PluginNameList) {
	pluginListCh <- dpm.PluginNameList{l.ResourceName}
}

func (l Lister) NewPlugin(bridge string) dpm.PluginInterface {
	return DevicePlugin{}
}

type DevicePlugin struct{}

func (dp DevicePlugin) GetDevicePluginOptions(ctx context.Context, in *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}, nil
}

func (dp DevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	for {
		var bridgeDevs []*pluginapi.Device
		for i := 0; i < nicsPoolSize; i++ {
			bridgeDevs = append(bridgeDevs, &pluginapi.Device{
				ID:     fmt.Sprintf("nic-%02d", i),
				Health: pluginapi.Healthy,
			})
		}
		s.Send(&pluginapi.ListAndWatchResponse{Devices: bridgeDevs})
		time.Sleep(10 * time.Second)
	}
	return nil
}

// TODO: cleanup if fails
func (dp DevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}

	// TODO: is this needed?
	for _, _ = range r.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{}
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	if len(r.ContainerRequests) != 1 {
		return nil, fmt.Errorf("Allocate request must contain exactly one container request")
	}

	if len(r.ContainerRequests[0].DevicesIDs) != 1 {
		return nil, fmt.Errorf("Allocate request must contain exactly one device")
	}

	allocatedDeviceID := r.ContainerRequests[0].DevicesIDs[0]

	go func() {
		time.Sleep(10 * time.Second)

		podUID, err := findPodUID(allocatedDeviceID)
		if err != nil {
			glog.Errorf("Failed to find pod UID: %v", err)
			return
		}

		pod, err := findPod(podUID)
		if err != nil {
			glog.Errorf("Failed to find pod with given PodUID: %v", err)
			return
		}

		networksSpec, err := buildNetworksSpec(pod)
		if err != nil {
			glog.Errorf("Failed to read networks spec: %v", err)
			return
		}

		containerName := fmt.Sprintf("k8s_POD_%s_%s", pod.Name, pod.Namespace)

		containerPid, err := findContainerPid(containerName)
		if err != nil {
			glog.Errorf("Failed to find container PID: %v", err)
			return
		}

		// TODO: run in parallel, make sure to precreate netns (colission)
		for _, spec := range *networksSpec {
			if err := exec.Command("attach-pod", containerName, spec.PortName, spec.PortID, spec.MacAddress, strconv.Itoa(containerPid)).Run(); err != nil {
				// TODO: include logs here
				glog.Errorf("attach-pod failed, please check logs in Daemon Set /var/log/attach-pod.err.log")
			}
		}
	}()

	return &responses, nil
}

func findPodUID(deviceID string) (string, error) {
	checkpointRaw, err := ioutil.ReadFile(devicepluginCheckpointPath)
	if err != nil {
		return "", fmt.Errorf("failed to read device plugin checkpoint file: %v", err)
	}

	var checkpoint map[string]interface{}
	err = json.Unmarshal(checkpointRaw, &checkpoint)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal device plugin checkpoint file: %v", err)
	}

	for _, entry := range checkpoint["PodDeviceEntries"].([]interface{}) {
		for _, deviceID := range entry.(map[string]interface{})["DeviceIDs"].([]interface{}) {
			if deviceID.(string) == deviceID {
				podUID = entry.(map[string]interface{})["PodUID"].(string)
				return podUID, nil
			}
		}
	}

	return "", fmt.Errorf("failed to find a pod with matching device ID")

}

func findPod(podUID string) (*v1.Pod, error) {
	// TODO: keep client in DP struct
	kubeClientConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain kubernetes client config: %v", err)
	}

	kubeclient, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to intialize kubernetes client: %v", err)
	}

	pods, err := kubeclient.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range pods.Items {
		if string(pod.UID) == podUID {
			return &pod, nil
		}
	}

	return nil, fmt.Errorf("failed to find a pod with matching ID")
}

func buildNetworksSpec(pod *v1.Pod) (*spec.NetworksSpec, error) {
	var networksSpec *spec.NetworksSpec

	annotations := pod.ObjectMeta.GetAnnotations()
	networksSpecAnnotation, _ := annotations[networksSpecAnnotationName]

	err := json.Unmarshal([]byte(networksSpecAnnotation), &networksSpec)

	return networksSpec, err
}

func findContainerPid(containerName string) (int, error) {
	// TODO: keep client in DP struct
	dockerclient, err := dockercli.NewEnvClient()
	if err != nil {
		return 0, fmt.Errorf("failed to intialize docker client: %v", err)
	}

	containers, err := dockerclient.ContainerList(context.Background(), dockertypes.ContainerListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list docker containers: %v", err)
	}

	for i := 0; i <= 10; i++ {
		for _, container := range containers {
			config, err := dockerclient.ContainerInspect(context.Background(), container.ID)
			if err != nil {
				return 0, fmt.Errorf("failed to inspect docker container: %v", err)
			}

			if strings.Contains(config.Name, containerName) {
				return config.State.Pid, nil
			}
		}
		time.Sleep(10 * time.Second)
	}

	return 0, fmt.Errorf("failed to find container PID")

}

// TODO: use this instead of separate thread during Allocate
func (dp DevicePlugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	var response pluginapi.PreStartContainerResponse
	return &response, nil
}
