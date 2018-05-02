package main

import (
	"flag"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"github.com/phoracek/kubetron/pkg/deviceplugin"
)

func main() {
	flag.Parse()

	manager := dpm.NewManager(deviceplugin.Lister{})
	manager.Run()
}
