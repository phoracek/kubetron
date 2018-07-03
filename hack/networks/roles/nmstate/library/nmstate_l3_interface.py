#!/usr/bin/python
#
# Copyright 2018 Red Hat, Inc.
#
# This file is part of ansible-nmstate.
#
# ansible-nmstate is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# ansible-nmstate is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with ansible-nmstate.  If not, see <https://www.gnu.org/licenses/>.

from copy import deepcopy

from libnmstate import netapplier
from libnmstate import netinfo

from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.network.common.utils import remove_default_spec

from ansible.module_utils.ansible_nmstate import get_interface_state
from ansible.module_utils.ansible_nmstate import write_debug_state


MODULE_NAME = "nmstate_l3_interface"

ANSIBLE_METADATA = {
    'metadata_version': '1.1',
    'status': ['preview'],
    'supported_by': 'community'
}

DOCUMENTATION = '''
---
module: nmstate_l3_interface
version_added: "2.6"
author: "Till Maas (@tyll)"
short_description: Configure IP addresses (layer 3) with nmstate
description:
    - "This module allows to configure IP addresses (layer 3) with nmstate
       https://github.com/nmstate/nmstate"
options:
  name:
    description:
      - Name of the L3 interface.
  ipv4:
    description:
      - IPv4 of the L3 interface.
  ipv6:
    description:
      - IPv6 of the L3 interface.
  aggregate:
    description: List of L3 interfaces definitions
  purge:
    description:
      - Purge L3 interfaces not defined in the I(aggregate) parameter.
    default: no
  state:
    description:
      - State of the L3 interface configuration.
    default: present
    choices: ['present', 'absent']
'''

EXAMPLES = '''
- name: Set eth0 IPv4 address
  net_l3_interface:
    name: eth0
    ipv4: 192.168.0.1/24

- name: Remove eth0 IPv4 address
  net_l3_interface:
    name: eth0
    state: absent

- name: Set IP addresses on aggregate
  net_l3_interface:
    aggregate:
      - { name: eth1, ipv4: 192.168.2.10/24 }
      - { name: eth2, ipv4: 192.168.3.10/24, ipv6: "fd5d:12c9:2201:1::1/64" }

- name: Remove IP addresses on aggregate
  net_l3_interface:
    aggregate:
      - { name: eth1, ipv4: 192.168.2.10/24 }
      - { name: eth2, ipv4: 192.168.3.10/24, ipv6: "fd5d:12c9:2201:1::1/64" }
    state: absent
'''

RETURN = '''
state:
    description: Network state after running the module
    type: dict
'''


def create_ip_dict(ciddr_addr):
    ip, prefix = ciddr_addr.split('/')
    addr = {'ip': ip, 'prefix-length': int(prefix)}
    return addr


def set_ipv4_addresses(interface_state, ipv4, purge=False):
    ipconfig = interface_state.setdefault('ipv4', {})
    ipconfig['enabled'] = True
    if purge:
        addresses = []
        ipconfig['addresses'] = addresses
    else:
        addresses = ipconfig.setdefault('addresses', [])

    addr = create_ip_dict(ipv4)
    if addr not in addresses:
        addresses.append(addr)

    return interface_state


def remove_ipv4_address(interface_state, ipv4):
    ipconfig = interface_state.get('ipv4')
    if not ipconfig:
        return interface_state

    addresses = ipconfig.get('addresses')

    if not addresses:
        return interface_state

    addr = create_ip_dict(ipv4)
    try:
        addresses.remove(addr)
    except ValueError:
        pass
    return interface_state


def run_module():
    element_spec = dict(
        name=dict(),
        ipv4=dict(),
        ipv6=dict(),
        state=dict(default='present',
                   choices=['present', 'absent'])
    )

    aggregate_spec = deepcopy(element_spec)
    aggregate_spec['name'] = dict(required=True)

    # remove default in aggregate spec, to handle common arguments
    remove_default_spec(aggregate_spec)

    argument_spec = dict(
        aggregate=dict(type='list', elements='dict', options=aggregate_spec),
        purge=dict(default=False, type='bool'),
        # not in net_* specification
        debug=dict(default=False, type='bool'),
    )

    argument_spec.update(element_spec)

    required_one_of = [['name', 'aggregate']]
    mutually_exclusive = [['name', 'aggregate']]

    result = dict(changed=False)

    module = AnsibleModule(
        argument_spec=argument_spec,
        required_one_of=required_one_of,
        mutually_exclusive=mutually_exclusive,
        supports_check_mode=True
    )

    if module.params['aggregate']:
        # FIXME impelement aggregate
        module.fail_json(msg='Aggregate not yet supported', **result)

    previous_state = netinfo.show()
    interfaces = deepcopy(previous_state['interfaces'])
    name = module.params['name']

    interface_state = get_interface_state(interfaces, name)

    if module.params['state'] == 'present':
        if not interface_state:
            module.fail_json(msg='Interface "%s" not found' % (name,),
                             **result)

        interface_state = set_ipv4_addresses(interface_state,
                                             module.params['ipv4'],
                                             purge=module.params['purge'])

    elif module.params['state'] == 'absent':
        if interface_state:
            if module.params['ipv4']:
                interface_state = remove_ipv4_address(interface_state,
                                                      module.params['ipv4'])
            else:
                ipconfig = interface_state.setdefault('ipv4', {})
                ipconfig['enabled'] = False

        # assume success when interface to configure is not present

    interfaces = [interface_state]
    new_partial_state = {'interfaces': interfaces}

    if module.params.get('debug'):
        result['previous_state'] = previous_state
        result['new_partial_state'] = new_partial_state
        result['debugfile'] = write_debug_state(MODULE_NAME, new_partial_state)

    if module.check_mode:
        new_full_state = deepcopy(previous_state)
        new_full_state.update(new_partial_state)
        result['state'] = new_full_state

        # TODO: maybe compare only the state of the defined interfaces
        if previous_state != new_full_state:
            result['changed'] = True

        module.exit_json(**result)
    else:
        netapplier.apply(new_partial_state)
    current_state = netinfo.show()
    if current_state != previous_state:
        result['changed'] = True
    result['state'] = current_state

    module.exit_json(**result)


def main():
    run_module()


if __name__ == '__main__':
    main()
