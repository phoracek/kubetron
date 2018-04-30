package cniplugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"sort"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyz"

	multusBinSource      = "/opt/multus"
	multusBinDestination = "/opt/cni/bin/multus"

	kubetronBinSource      = "/opt/kubetron"
	kubetronBinDestination = "/opt/cni/bin/kubetron"

	cniNetworkConfigsDir = "/etc/cni/net.d/"

	kubetronNetworkConfigFile = cniNetworkConfigsDir + "00-kubetron.conf"
)

func InstallBinaries() error {
	fmt.Printf("1 \n")
	err := atomicCopy(multusBinSource, multusBinDestination)
	if err != nil {
		return err
	}

	fmt.Printf("2 \n")
	err = atomicCopy(kubetronBinSource, kubetronBinDestination)
	if err != nil {
		return err
	}

	fmt.Printf("3 \n")
	networkConfigFilesInfo, err := ioutil.ReadDir(cniNetworkConfigsDir)
	if err != nil {
		return err
	}
	if len(networkConfigFilesInfo) == 0 {
		return fmt.Errorf("no network config files were found")
	}

	fmt.Printf("4 \n")
	networkConfigFiles := make([]string, 0)
	for _, networkConfigFileInfo := range networkConfigFilesInfo {
		// TODO: with absolute path
		networkConfigFiles = append(networkConfigFiles, cniNetworkConfigsDir+networkConfigFileInfo.Name())
	}

	fmt.Printf("5 \n")
	sort.Strings(networkConfigFiles)
	networkConfigFile := networkConfigFiles[0]

	fmt.Printf("6 \n")
	networkConfigRaw, err := ioutil.ReadFile(networkConfigFile)
	if err != nil {
		return err
	}

	fmt.Printf("7 \n")
	var networkConfig map[string]interface{}
	json.Unmarshal(networkConfigRaw, &networkConfig)

	configType := networkConfig["type"].(string)
	if configType == "multus" {
		fmt.Printf("mutlus\n")
		// TODO: check if already installed, if not, fail
		for _, subNetworkConfig := range networkConfig["delegates"].([]map[string]interface{}) {
			if subNetworkConfig["type"].(string) == "kubetron" {
				fmt.Printf("kubetron found in multus\n")
			}
		}
	} else {
		fmt.Printf("other\n")
		// TODO: add to multus template, mark as main

		newNetworkConfig := map[string]interface{}{
			"name": "multus-network",
			"type": "multus",
			"delegates": []interface{}{
				map[string]interface{}{
					"type": "kubetron",
				},
				// Existing configuration should be here
			},
		}
		networkConfig["masterplugin"] = true
		newNetworkConfig["delegates"].([]map[string]interface{}) = append(newNetworkConfig["delegates"].([]map[string]interface{}), networkConfig)

		fmt.Printf("original moved to multus:\n%s\n", newNetworkConfig)
	}

	fmt.Printf("8 \n")

	return nil
}

func saveNewCNIConfiguration() error {
	// TODO: read current config
	// TODO: if config not found, fail
	// TODO: if installed, ignore
	// TODO: if multus, just append
	// TODO: if not multus, copy original, add it to multus as main, add our
	return nil
}

func readCNIConfiguration() interface{} {
	// TODO: use cni lib for that?
	return nil
}

func atomicCopy(sourcePath, destinationPath string) error {
	fmt.Printf("c 1 \n")
	destinationTmpPath := fmt.Sprintf("%s.%s", destinationPath, createRandomString(5))

	fmt.Printf("c 2 \n")
	err := exec.Command("cp", sourcePath, destinationTmpPath).Run()
	if err != nil {
		return fmt.Errorf("failed to copy %s to %s", sourcePath, destinationTmpPath)
	}

	fmt.Printf("c 3 \n")
	err = os.Rename(destinationTmpPath, destinationPath)
	if err != nil {
		os.Remove(destinationTmpPath)
		return err
	}

	fmt.Printf("c 4 \n")
	os.Remove(destinationTmpPath)
	return nil
}

func createRandomString(length int) string {
	str := make([]byte, length)
	for i := range str {
		str[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(str)
}
