#!/bin/bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2018 Giuseppe Maxia
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
cd $(dirname $0)
source cookbook_include.sh

version=$1
[ -z "$version" ] && version=5.7.23
check_version $version

if [ -z "$(dbdeployer sandboxes | grep 'master-slave\s*'$version)" ]
then
    echo "master-slave version $version is not installed"
    echo "Run './replication-master-slave.sh $version' before trying again"
    exit 1
fi

sandbox_dir=$SANDBOX_HOME/rsandbox_$(echo $version | tr '.' '_' )

header "Checking the value for max-connections in all nodes"
(set -x
$sandbox_dir/use_all 'select @@max_connections'
)

header "Restarting all nodes with the new value"
(set -x
$sandbox_dir/restart_all --max-connections=66
)

header "Checking the new value for max-connections in all nodes"
(set -x
$sandbox_dir/use_all 'select @@max_connections'
)

header "Restarting slave #2 without specifying any values."
(set -x
$sandbox_dir/node2/restart
)

header "Checking the value for max-connections in all nodes: node #2 has again the default value"
(set -x
$sandbox_dir/use_all 'select @@max_connections'
)

header "Adding a custom number of connection permanently to slave #2 ( NOTE: no dashes! )"
(set -x
$sandbox_dir/node2/add_option max-connections=99
)

header "Restarting slave #2 without specifying any values. We'll see that its own value is preserved"
(set -x
$sandbox_dir/node2/restart
)

header "Checking the value for max-connections in all nodes: node #2 has kept its own value"
(set -x
$sandbox_dir/use_all 'select @@max_connections'
)
