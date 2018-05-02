package admission

import (
	resty "gopkg.in/resty.v1"
)

// TODO: only supports plain text with no auth, add security later

type providerClient struct {
	client *resty.Client
}

func NewProviderClient(url string) *providerClient {
	client := resty.
		New().
		SetHostURL(url).
		SetHeader("Accept", "application/json")
	return &providerClient{client}
}

func (c *providerClient) ListNetworkIDsByNames() (map[string]string, error) {
	var result map[string][]map[string]interface{}

	// TODO: get id
	// TODO: check provider error output
	_, err := c.client.R().
		SetResult(&result).
		Get("v2.0/networks")
	if err != nil {
		return nil, err
	}

	networkIDsByNames := make(map[string]string)
	for _, network := range result["networks"] {
		networkIDsByNames[network["name"].(string)] = network["id"].(string)
	}

	return networkIDsByNames, nil
}

// TODO: return port id
func (c *providerClient) CreateNetworkPort(network, port, macAddress string) (string, bool, error) {
	var result map[string]map[string]interface{}

	// TODO: check provider error output
	_, err := c.client.R().
		SetResult(&result).
		SetBody(
			map[string]map[string]interface{}{
				"port": map[string]interface{}{
					"network_id":     network,
					"name":           port,
					"mac_address":    macAddress,
					"admin_state_up": true,
				},
			},
		).
		Post("v2.0/ports")

	portID := result["port"]["id"].(string)
	hasFixedIPs := len(result["port"]["fixed_ips"].([]interface{})) != 0

	return portID, hasFixedIPs, err
}

func (c *providerClient) DeleteNetworkPort(portID string) error {
	_, err := c.client.R().
		SetPathParams(map[string]string{"portID": portID}).
		Delete("/v2.0/ports/{portID}")

	return err
}
