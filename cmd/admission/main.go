package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/openshift/generic-admission-server/pkg/cmd/server"
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/util/logs"

	"github.com/phoracek/kubetron/pkg/admission"
)

func main() {
	flagset := pflag.NewFlagSet("kubetron-admission", pflag.ExitOnError)

	ah := &admission.AdmissionHook{}

	flagset.StringVarP(&ah.ProviderURL, "provider-url", "p", "", "URL of OVN manager (e.g. Neutron) API server")
	flagset.StringVarP(&ah.ResourceName, "resource-name", "r", "", "Name of resource exposed by Kubetron's Device Plugin")

	if ah.ProviderURL == "" {
		glog.Fatal(fmt.Errorf("provider-url was not set"))
	}
	if ah.ResourceName == "" {
		glog.Fatal(fmt.Errorf("resource-name was not set"))
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	stopCh := genericapiserver.SetupSignalHandler()

	cmd := server.NewCommandStartAdmissionServer(os.Stdout, os.Stderr, stopCh, ah)
	cmd.Short = "Launch Kubetron Admission Server"
	cmd.Long = "Launch Kubetron Admission Server"

	// Add admission hook flags
	cmd.PersistentFlags().AddFlagSet(flagset)

	// Flags for glog
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// Fix glog printing "Error: logging before flag.Parse"
	flag.CommandLine.Parse([]string{})

	if err := cmd.Execute(); err != nil {
		glog.Fatal(err)
	}
}
