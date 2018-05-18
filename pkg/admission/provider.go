// TODO: only supports plain text with no auth, add security later
// TODO: use keystone to access Neutron API
package admission

import (
	resty "gopkg.in/resty.v1"
)

type providerClient struct {
	client *resty.Client
}

// NewProviderClient creates a REST client to access Neutron API
func NewProviderClient(url string) *providerClient {
	client := resty.
		New().
		SetHostURL(url).
		SetHeader("Accept", "application/json")
	return &providerClient{client}
}

// ListNetworkIDsByNames returns a map where key is name of Neutron Network and key is its ID
func (c *providerClient) ListNetworkIDsByNames() (map[string]string, error) {
	var result map[string][]map[string]interface{}

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

// CreateNetworkPort creates new Neutron Port on Neutron network with given ID
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

	// Read assigned ID of created Pod
	portID := result["port"]["id"].(string)

	// Check whether Port has fixed_ips, if it does, it means that selected Network has assigned subnet
	hasFixedIPs := len(result["port"]["fixed_ips"].([]interface{})) != 0

	return portID, hasFixedIPs, err
}

// DeleteNetworkPort removes Neutron Port with given ID
func (c *providerClient) DeleteNetworkPort(portID string) error {
	_, err := c.client.R().
		SetPathParams(map[string]string{"portID": portID}).
		Delete("/v2.0/ports/{portID}")

	return err
}
