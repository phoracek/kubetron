package admission

type NetworksSpec map[string]NetworkSpec

type NetworkSpec struct {
	PortName   string `json:"portName"`
	PortID     string `json:"portID"`
	MacAddress string `json:"macAddress"`
}
