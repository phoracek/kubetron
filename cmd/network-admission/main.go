package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	"github.com/openshift/generic-admission-server/pkg/cmd/server"
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/util/logs"

	admission "github.com/phoracek/kubetron/pkg/network-admission"
)

func main() {
	flagset := pflag.NewFlagSet("kubetron-network-admission", pflag.ExitOnError)

	ah := &admission.AdmissionHook{}

	logs.InitLogs()
	defer logs.FlushLogs()

	stopCh := genericapiserver.SetupSignalHandler()

	cmd := server.NewCommandStartAdmissionServer(os.Stdout, os.Stderr, stopCh, ah)
	cmd.Short = "Launch Kubetron Network Admission Server"
	cmd.Long = "Launch Kubetron Network Admission Server"

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
