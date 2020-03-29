#!/usr/bin/env bash
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

#unset DBDEPLOYER_LOGGING
test_dir=$(dirname $0)
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)
test_dir=$PWD
exit_code=0

if [ ! -f set-mock.sh ]
then
    echo "set-mock.sh not found in $PWD"
    exit 1
fi

if [ ! -f ../common.sh ]
then
    echo "../common.sh not found"
    exit 1
fi

source ../common.sh
source set-mock.sh
export SHOW_CHANGED_PORTS=1
start_timer

# Creates a zero-length catalog file.
# Sandbox creation should not fail
mkdir -p $mock_dir/home/.dbdeployer
touch $mock_dir/home/.dbdeployer/sandboxes.json

versions=(ndb7.6 ndb8.0)
rev_list="0 21 99"

for rev in $rev_list
do
    for vers in ${versions[*]}
    do
        version=${vers}.${rev}
        create_mock_ndb_version $version 
    done
done

run dbdeployer available
for vers in ${versions[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy replication $version --topology=ndb

        # Check existence 
        test_completeness $version ndb_msb_ multiple
        ok_dir_exists "$SANDBOX_HOME/ndb_msb_$version_name"
        ok_dir_exists "$SANDBOX_HOME/ndb_msb_$version_name/ndb_conf"
        ok_file_exists "$SANDBOX_HOME/ndb_msb_$version_name/ndb_conf/config.ini"
        for node_num in 1 2 3
        do
            ok_dir_exists "$SANDBOX_HOME/ndb_msb_$version_name/ndbnode$node_num"
        done
        results "$version"
    done
done

run dbdeployer delete ALL --skip-confirm

results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

