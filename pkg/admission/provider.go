// TODO: only supports plain text with no auth, add security later
// TODO: check keystone expiration date 
package admission

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

type Network struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Physnet string `json:"provider:physical_network"`
}

type PortParams struct {
	PortId      string
	MacAddress  string
	HasFixedIps bool
}

type providerClient struct {
	client *gophercloud.ServiceClient
}

// NewProviderClient creates a REST client to access Neutron API
func NewProviderClient() (*providerClient, error) {
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}

	return &providerClient{client}, nil
}

// ListNetworkIDsByNames returns a map where key is name of Neutron Network and key is its ID
func (c *providerClient) ListNetworkByName() (map[string]*Network, error) {

	var networkList []Network

	pages, err := networks.List(c.client, networks.ListOpts{}).AllPages()
	if err != nil {
		return nil, err
	}

	err = networks.ExtractNetworksInto(pages, &networkList)
	if err != nil {
		return nil, err
	}

	networkByName := make(map[string]*Network)
	for index, network := range networkList {
		networkByName[network.Name] = &networkList[index]
	}

	return networkByName, nil
}

// CreateNetworkPort creates new Neutron Port on Neutron network with given ID
func (c *providerClient) CreateNetworkPort(network, port string) (*PortParams, error) {
	createOpts := ports.CreateOpts{NetworkID: network, Name: port, AdminStateUp: newTrue()}
	responsePort, err := ports.Create(c.client, createOpts).Extract()

	if err != nil {
		return nil, err
	}
	portParams := PortParams{
		PortId:      responsePort.ID,
		MacAddress:  responsePort.MACAddress,
		HasFixedIps: len(responsePort.FixedIPs) != 0,
	}
	return &portParams, nil
}

// DeleteNetworkPort removes Neutron Port with given ID
func (c *providerClient) DeleteNetworkPort(portID string) error {
	err := ports.Delete(c.client, portID).ExtractErr()

	return err
}
