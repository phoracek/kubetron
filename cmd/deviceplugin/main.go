package main

import (
	"flag"
	"strings"

	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"github.com/phoracek/kubetron/pkg/deviceplugin"
)

func main() {
	resourceName := flag.String("resource-name", "", "TODO")
	flag.Parse()

	resourceSplit := strings.Split(*resourceName, "/")

	glog.V(6).Infof("Starting DP with ns %s, n %s", resourceSplit[0], resourceSplit[1])

	manager := dpm.NewManager(deviceplugin.Lister{
		ResourceName:      resourceSplit[1],
		ResourceNamespace: resourceSplit[0],
	})
	manager.Run()
}
