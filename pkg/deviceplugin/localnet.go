package deviceplugin

import (
	"os/exec"
	"strings"

	"github.com/golang/glog"
)

func listLocalNetworks() []string {
	networks := make([]string, 0)

	mappingsBytes, err := exec.Command("ovs-vsctl", "get", "open", ".", "external-ids:ovn-bridge-mappings").Output()
	if err != nil {
		glog.V(3).Infof("Failed to list ovn-bridge-mappings: %v", err)
		return networks
	}

	mappingsString := string(mappingsBytes)
	mappingsStringWithoutQuotes := mappingsString[1:len(mappingsString)-1]
	mappings := strings.Split(mappingsStringWithoutQuotes, ",")

	for _, mappingRaw := range mappings {
		mapping := strings.Split(mappingRaw, ":")
		networks = append(networks, mapping[0])
	}

	return networks
}
