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
#

""" unit tests for nmstate_linkagg.py """

try:
    from unittest import mock
except ImportError:  # py2
    import mock

import sys
sys.modules['libnmstate'] = mock.Mock()
sys.modules['ansible'] = mock.Mock()
sys.modules['ansible.module_utils.basic'] = mock.Mock()
sys.modules['ansible.module_utils'] = mock.Mock()
sys.modules['ansible.module_utils.network.common'] = mock.Mock()
sys.modules['ansible.module_utils.network.common.utils'] = mock.Mock()
sys.modules['ansible.module_utils.network'] = mock.Mock()

sys.modules['ansible.module_utils.ansible_nmstate'] = mock.Mock()

import nmstate_linkagg as nla  # noqa: E402


def test_import_succeeded():
    assert nla
