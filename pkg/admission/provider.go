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

func (c *providerClient) ListNetworkNames() (map[string]bool, error) {
	var result map[string][]map[string]interface{}

	// TODO: check provider error output
	_, err := c.client.R().
		SetResult(&result).
		Get("v2.0/networks")
	if err != nil {
		return nil, err
	}

	networkNames := make(map[string]bool)
	for _, network := range result["networks"] {
		networkNames[network["name"].(string)] = true
	}

	return networkNames, nil
}

// TODO: return port id
func (c *providerClient) CreateNetworkPort(network, port, macAddress string) (string, error) {
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

	return portID, err
}

func (c *providerClient) DeleteNetworkPort(portID string) error {
	_, err := c.client.R().
		SetPathParams(map[string]string{"portID": portID}).
		Delete("/v2.0/ports/{portID}")

	return err
}
