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

import json
import os
import tempfile
import time


def write_debug_state(module_name, state):
    debugfile, debugname = tempfile.mkstemp(
        prefix='{}_debug-{}-'.format(module_name, int(time.time())))
    debugfile = os.fdopen(debugfile, "w")
    debugfile.write(json.dumps(state, indent=4))

    return debugname


def get_interface_state(interfaces, name):
    '''
    Get the state for first interface with the specified name
    '''
    for interface_state in interfaces:
        if interface_state['name'] == name:
            break
    else:
        interface_state = None
    return interface_state
