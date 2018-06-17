package deviceplugin

import (
	"time"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
)

// Lister implements device-plugin-manager Lister interface. It is used to track and initialize available resources within one resource namespace. In this case it handles only one resource - OVN overlay.
type Lister struct {
	ResourceNamespace           string
	ReservedOverlayResourceName string
}

// GetResourceNamespace returns namespace of the resource
func (l Lister) GetResourceNamespace() string {
	return l.ResourceNamespace
}

// Discover keeps list of currently available resources. In this implementation it is static, one overlay is always exposed
// TODO: Check if br-int is available, if not, don't expose any resource
func (l Lister) Discover(pluginListCh chan dpm.PluginNameList) {
	pluginListCh <- dpm.PluginNameList{l.ReservedOverlayResourceName}

	for {
		localNetworks := listLocalNetworks()
		resources := append(localNetworks, l.ReservedOverlayResourceName)
		pluginListCh <- dpm.PluginNameList(resources)
		time.Sleep(10 * time.Second)
	}
}

// NewPlugin initializes new Device Plugin instance. This is called by device-plugin-manager once a new resource is discovered
func (l Lister) NewPlugin(resourceName string) dpm.PluginInterface {
	return DevicePlugin{}
}
