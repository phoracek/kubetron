package main

import (
	"flag"
	"strings"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"github.com/phoracek/kubetron/pkg/deviceplugin"
)

func main() {
	resourceName := flag.String("resource-name", "", "TODO")
	flag.Parse()

	resourceSplit := strings.Split(*resourceName, "/")

	manager := dpm.NewManager(deviceplugin.Lister{
		ResourceName:      resourceSplit[1],
		ResourceNamespace: resourceSplit[0],
	})
	manager.Run()
}
