# Ansible Support for nmstate
[![Build Status](https://travis-ci.org/nmstate/ansible-nmstate.svg?branch=master)](https://travis-ci.org/nmstate/ansible-nmstate)
[![Coverage Status](https://coveralls.io/repos/github/nmstate/ansible-nmstate/badge.svg?branch=master)](https://coveralls.io/github/nmstate/ansible-nmstate?branch=master)

Ansible-NMState allows to configure the network state with
[NMState](https://nmstate.github.io/) through Ansible.

## Development Environment

Run unit tests:
```shell
tox
```

## Installation

To install the modules, run:

```shell
make install
```

To install them system-wide, run this command as root. Alternatively the path
to the module utils and the library can be specified on the command line:

```shell
ANSIBLE_MODULE_UTILS=$PWD/module_utils ansible -M $PWD/library ...
```

or for playbooks:

```shell
ANSIBLE_MODULE_UTILS=$PWD/module_utils ansible-playbook -M $PWD/library ...
```

Using aliases keeps the command-line shorter:

```shell
alias ansible="ANSIBLE_MODULE_UTILS=$PWD/module_utils ansible -M $PWD/library"
alias ansible-playbook="ANSIBLE_MODULE_UTILS=$PWD/module_utils ansible-playbook -M $PWD/library"
```

Another possiblity for testing is to install [https://direnv.net/](direnv) to
automatically activate the necessary environment when entering the repository.

## Basic Operations

Enable link aggregation with the interface `web-bond` using the members `eth1` and `eth2` on the host `rhel7-cloud`:

```shell
ansible -m net_linkagg -a 'name=web-bond state=up members=eth1,eth2' -e ansible_network_os=nmstate -i rhel7-cloud, all

```

Set an IP address for the interface `web-bond` on the host `rhel7-cloud`:

```shell
ansible -m net_l3_interface -a 'name=web-bond state=present ipv4=192.0.2.7/24' -e ansible_network_os=nmstate -i rhel7-cloud, all
```

For example playbooks, see the `examples/` directory. Run a playbook:

```shell
ansible-playbook examples/web-bond.yml -i rhel7-cloud,
```
