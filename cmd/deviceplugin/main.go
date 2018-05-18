package main

import (
	"flag"
	"strings"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"github.com/phoracek/kubetron/pkg/deviceplugin"
)

func main() {
	resourceName := flag.String("resource-name", "", "Name of resource exposed by Kubetron's Device Plugin")
	flag.Parse()

	// We keep full name of the resource in Kubetron config. Here we split it to resource namespace and actual name
	resourceSplit := strings.Split(*resourceName, "/")

	manager := dpm.NewManager(deviceplugin.Lister{
		ResourceName:      resourceSplit[1],
		ResourceNamespace: resourceSplit[0],
	})
	manager.Run()
}
