package main

import "github.com/phoracek/kubetron/pkg/cniplugin"

func main() {
	err := cniplugin.InstallBinaries()
	if err != nil {
		panic(err)
	}
}
