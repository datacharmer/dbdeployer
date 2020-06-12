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

versions=(5.6 5.7 8.0)
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
for vers in ${versions[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy single $SANDBOX_BINARY/$version
        run dbdeployer deploy replication $SANDBOX_BINARY/$version --concurrent

        # Check existence 
        ok_dir_exists "$SANDBOX_HOME/msb_$version_name"
        ok_dir_exists "$SANDBOX_HOME/rsandbox_$version_name"
        
        single_command_line=$(grep "command-line.*dbdeployer deploy single $SANDBOX_BINARY/$version" $CATALOG)
        replication_command_line=$(grep "command-line.*dbdeployer deploy replication $SANDBOX_BINARY/$version" $CATALOG)
        ok "single command line <$single_command_line>" "$single_command_line"
        ok "replication command line <$replication_command_line>" "$replication_command_line"

        test_completeness $version msb_ single
        test_completeness $version rsandbox_ replication

        results "$version"
    done
done

function test_default_sandbox {
    sb_name=$1
    exec_name=$2
    cmd=$3

    if [ -z "$exec_name" ]
    then
        exec_name=default
    fi

    options=""

    if [ "$exec_name" != "default" ]
    then
        options="--default-sandbox-executable=$exec_name"
    fi

    run dbdeployer admin set-default $sb_name $options

    ok_executable_exists $SANDBOX_HOME/$exec_name
    ok_file_exists $SANDBOX_HOME/$sb_name/is_default
    status_text=$( $SANDBOX_HOME/$exec_name $cmd)
    ok_contains "status found" "$status_text" "$sb_name"
    run dbdeployer admin remove-default $options
    ok_file_does_not_exists $SANDBOX_HOME/$sb_name/is_default
    ok_executable_does_not_exist $SANDBOX_HOME/$exec_name

    run dbdeployer admin set-default $sb_name $options
    ok_executable_exists $SANDBOX_HOME/$exec_name
    run dbdeployer delete $sb_name
    ok_executable_does_not_exist $SANDBOX_HOME/$exec_name
}


for vers in ${versions[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        
        test_default_sandbox msb_$version_name default status
        test_default_sandbox rsandbox_$version_name repl status_all
    done

done
 

run dbdeployer delete ALL --skip-confirm

results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

