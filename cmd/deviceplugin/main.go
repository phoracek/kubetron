package main

import (
	"flag"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"github.com/phoracek/kubetron/pkg/deviceplugin"
)

func main() {
	resourceNamespace := flag.String("resource-namespace", "", "Namespace for resources by Kubetron's Device Plugin")
	reservedOverlayResourceName := flag.String("reserved-overlay-resource-name", "", "Name of resource used for exposing available overlay network, cannot be used for physnet names")
	flag.Parse()

	manager := dpm.NewManager(deviceplugin.Lister{
		ResourceNamespace:           *resourceNamespace,
		ReservedOverlayResourceName: *reservedOverlayResourceName,
	})
	manager.Run()
}
