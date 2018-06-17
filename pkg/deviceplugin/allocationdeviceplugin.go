package deviceplugin

import (
	"context"
	"fmt"
	"time"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	// attachmentDevicesPoolSize represents amount of available "devices", i.e. overlay attachments, where one pod, no matter how many networks it requests, takes one overlay attachment
	attachmentDevicesPoolSize = 100
)

// AllocationDevicePlugin implements deviceplugin v1beta1 interface. It reports available (fake) devices and connects Pod to requested networks
type AllocationDevicePlugin struct{}

// GetDevicePluginOptions is used to pass parameters to device manager
func (dp AllocationDevicePlugin) GetDevicePluginOptions(ctx context.Context, in *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}, nil
}

// ListAndWatch returns list of devices available on the host. Since we don't work with limited physical interfaces, we report set of virtual devices that are later used to identify Pod
func (dp AllocationDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	for {
		var devs []*pluginapi.Device
		for i := 0; i < attachmentDevicesPoolSize; i++ {
			devs = append(devs, &pluginapi.Device{
				ID:     fmt.Sprintf("dev-%02d", i),
				Health: pluginapi.Healthy,
			})
		}
		s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
		time.Sleep(10 * time.Second)
	}
	return nil
}

// Allocate is executed when a new Pod appears and it requires requested resource to be allocated and attached. In our case we create new interface on OVS integration bridge (mapped to selected OVN network) and pass it to the Pod
func (dp AllocationDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}

	// TODO: is this needed?
	for _, _ = range r.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{}
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	return &responses, nil
}

// PreStartContainer is currently unused method that should be later used to move OVS interface to Pod and configure it
func (dp AllocationDevicePlugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	var response pluginapi.PreStartContainerResponse
	return &response, nil
}
