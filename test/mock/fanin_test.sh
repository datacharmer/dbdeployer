#!/usr/bin/env bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2020 Giuseppe Maxia
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

versions=(7.6 8.0)
rev_list="0 21 99"

for rev in $rev_list
do
    for vers in ${versions[*]}
    do
        version=${vers}.${rev}
        create_mock_version $version 
    done
done

run dbdeployer available

function test_fanin {
    version=$1
    dir_name=$2
    nodes=$3
    master_list=$4
    slave_list=$5
    explicit_lists=$6
    
    lists_options=""
    if [ -n "$explicit_lists" ]
    then
        lists_options="--master-list=$master_list --slave_list=$slave_list"
    fi

    run dbdeployer deploy replication $version \
        --topology=fan-in \
        --sandbox-directory=$dir_name \
        --nodes=$nodes $lists_options

        use_all_masters="$SANDBOX_HOME/$dir_name/use_all_masters"
        use_all_slaves="$SANDBOX_HOME/$dir_name/use_all_slaves"
        ok_dir_exists "$SANDBOX_HOME/$dir_name"
        ok_executable_exists $use_all_masters
        ok_executable_exists $use_all_slaves

        master_nodes=$(echo "$master_list" | tr ',' ' ')
        slave_nodes=$(echo "$slave_list" | tr ',' ' ')
        master_nodes_ok1=$(grep "MASTERS=.$master_nodes" $use_all_masters)
        master_nodes_ok2=$(grep "MASTERS=.$master_nodes" $use_all_slaves)
        slave_nodes_ok=$(grep "SLAVES=.$slave_nodes" $use_all_slaves)
        ok "master nodes 1" "$master_nodes_ok1"
        ok "master nodes 2" "$master_nodes_ok2"
        ok "slave nodes" "$slave_nodes_ok"

        results "$version"
        run dbdeployer delete $dir_name
}

for vers in ${versions[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        dir_name=fanin-$version_name

        test_fanin $version $dir_name 3 '1,2'     '3'
        test_fanin $version $dir_name 4 '1,2,3'   '4'
        test_fanin $version $dir_name 5 '1,2,3,4' '5'
        test_fanin $version $dir_name 3 '2,3'     '1'   explicit_lists
        test_fanin $version $dir_name 4 '1,2,4'   '3'   explicit_nodes
        test_fanin $version $dir_name 5 '1,2,3'   '4,5' explicit_nodes

    done

done


results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

