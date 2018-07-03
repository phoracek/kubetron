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

MODULE_NAME = "nmstate_linkagg"

ANSIBLE_METADATA = {
    'metadata_version': '1.1',
    'status': ['preview'],
    'supported_by': 'community'
}

DOCUMENTATION = '''
---
module: nmstate_linkagg
version_added: "2.6"
author: "Till Maas (@tyll)"
short_description: Configure link aggregation with nmstate

description:
    - "This module allows to configure link aggregation with
       https://github.com/nmstate/nmstate"
options:
  name:
    description:
      - Name of the link aggregation group.
    required: true
  mode:
    description:
      - Mode of the link aggregation group. A value of C(on) will enable
      LACP/802.3ad, the same as C(active) configures the link to actively
      information about the state of the link.
    default: on
    choices: ['on', 'active', 'passive', 'balance-rr', 'active-backup',
              'balance-xor', 'broadcast', '802.3ad', 'balance-tlb',
              'balance-alb']
  members:
    description:
      - List of members interfaces of the link aggregation group. The value can
      be single interface or list of interfaces.
    required: true
  min_links:
    description:
      - Minimum members that should be up
        before bringing up the link aggregation group.
  aggregate:
    description: List of link aggregation definitions.
  purge:
    description:
      - Purge link aggregation groups not defined in the I(aggregate)
        parameter.
    default: no
  state:
    description:
      - State of the link aggregation group.
    default: present
    choices: ['present', 'absent', 'up', 'down']
'''

EXAMPLES = '''
- name: Take bond interface down
  nmstate_linkagg:
      name=bond0
      state=down
      members=eth10

- name: configure link aggregation group
  net_linkagg:
    name: bond0
    members:
      - eth0
      - eth1

- name: remove configuration
  net_linkagg:
    name: bond0
    state: absent

- name: Create aggregate of linkagg definitions
  net_linkagg:
    aggregate:
        - { name: bond0, members: [eth1] }
        - { name: bond1, members: [eth2] }

- name: Remove aggregate of linkagg definitions
  net_linkagg:
    aggregate:
      - name: bond0
      - name: bond1
    state: absent
'''

RETURN = '''
state:
    description: Network state after running the module
    type: dict
'''


def run_module():
    element_spec = dict(
        members=dict(type='list'),
        min_links=dict(type='int'),
        # net_linkagg only knows on, active and passive
        # on and active is 802.3ad on Linux, passive is not supported
        mode=dict(choices=['on', 'active', 'passive', 'balance-rr',
                           'active-backup', 'balance-xor', 'broadcast',
                           '802.3ad', 'balance-tlb', 'balance-alb'],
                  default='on'),
        name=dict(),
        state=dict(default='present',
                   choices=['present', 'absent', 'up', 'down'])
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

    result = dict(
        changed=False,
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        required_one_of=required_one_of,
        mutually_exclusive=mutually_exclusive,
        supports_check_mode=True
    )

    if module.params['aggregate']:
        module.fail_json(msg='Aggregate not yet supported', **result)

    previous_state = netinfo.show()
    members = module.params['members']
    if not isinstance(members, list):
        members = [members]

    # Fail when member state is missing
    if module.params['state'] in ['up', 'present']:
        missing = []
        for member in members:
            member_state = get_interface_state(previous_state['interfaces'],
                                               member)
            if not member_state:
                missing.append(member)

        if missing:
            module.fail_json(msg='Did not find specified members in network '
                             'state: ' + ', '.join(missing), **result)

    mode = module.params['mode']
    if mode in ['on', 'active']:
        mode = '802.3ad'
    elif mode in ['passive']:
        # passive mode is not supported on Linux:
        # noqa:
        # https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/7/html/networking_guide/sec-comparison_of_network_teaming_to_bonding
        module.fail_json(msg='passive mode is not supported on Linux',
                         **result)

    link_aggregation = {'mode': mode,
                        'options': {},  # FIXME: add support for options?
                        'slaves': members,
                        }
    interface_state = {'name': module.params['name'],
                       'state': module.params['state'],
                       'type': 'bond',
                       'link-aggregation': link_aggregation,
                       }
    interfaces = []
    interfaces.append(interface_state)
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
