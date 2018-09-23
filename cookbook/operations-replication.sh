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

header "Running a simple command with the master in the sandbox." \
   "Notice the usage of the '-e', as if we were using the 'mysql' client" 

(set -x
$sandbox_dir/m -e 'SHOW MASTER STATUS'
)

header "Creating a table in the master"
(set -x
$sandbox_dir/m -e 'DROP TABLE IF EXISTS test.t1'
$sandbox_dir/m -e 'CREATE TABLE test.t1(id int not null primary key)'
)

header "Inserting 3 lines into the new table"
for N in 1 2 3
do
    (set -x
    $sandbox_dir/m -e "INSERT INTO test.t1 VALUES($N)"
    )
done
sleep 2

header "Getting the table contents from one slave"
(set -x
$sandbox_dir/s1 -e "SELECT * FROM test.t1"
)


header "Getting the table count from all nodes (NOTE: no '-e' is needed)"
(set -x
$sandbox_dir/use_all "SELECT COUNT(*) FROM test.t1"
)

header "Checking the status of all slaves"
(set -x
$sandbox_dir/check_slaves
)

header "Running a multiple query in all slaves"
(set -x
$sandbox_dir/use_all_slaves "STOP SLAVE; SET GLOBAL slave_parallel_workers=3; START SLAVE;show processlist "
)
