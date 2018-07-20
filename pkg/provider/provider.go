// TODO: only supports plain text with no auth, add security later
// TODO: use keystone to access Neutron API
package provider

import (
	resty "gopkg.in/resty.v1"
)

type Network struct {
	Name    string
	ID      string
	Physnet string
}

type ProviderClient struct {
	client *resty.Client
}

// NewProviderClient creates a REST client to access Neutron API
func NewProviderClient(url string) *ProviderClient {
	client := resty.
		New().
		SetHostURL(url).
		SetHeader("Accept", "application/json")
	return &ProviderClient{client}
}

// ListNetworkIDsByNames returns a map where key is name of Neutron Network and key is its ID
func (c *ProviderClient) ListNetworkByName() (map[string]*Network, error) {
	var result map[string][]map[string]interface{}

	// TODO: check provider error output
	_, err := c.client.R().
		SetResult(&result).
		Get("v2.0/networks")
	if err != nil {
		return nil, err
	}

	networkByName := make(map[string]*Network)
	for _, network := range result["networks"] {
		var physnet string
		if physnetRaw := network["provider:physical_network"]; physnetRaw != nil {
			physnet = physnetRaw.(string)
		} else {
			physnet = ""
		}
		networkByName[network["name"].(string)] = &Network{
			Name:    network["name"].(string),
			ID:      network["id"].(string),
			Physnet: physnet,
		}
	}

	return networkByName, nil
}

// CreateNetworkPort creates new Neutron Port on Neutron network with given ID
func (c *ProviderClient) CreateNetworkPort(network, port, macAddress string) (string, bool, error) {
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
func (c *ProviderClient) DeleteNetworkPort(portID string) error {
	_, err := c.client.R().
		SetPathParams(map[string]string{"portID": portID}).
		Delete("/v2.0/ports/{portID}")

	return err
}
