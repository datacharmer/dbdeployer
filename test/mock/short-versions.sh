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

# unset DBDEPLOYER_LOGGING
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
#export results_log=$PWD/port-clash.log
source set-mock.sh
export SHOW_CHANGED_PORTS=1
start_timer

versions=(5.0 5.1 5.5 5.6 5.7 8.0)
latest_rev=59
rev_list="26 37 48 $latest_rev"


for rev in $rev_list
do
    for vers in ${versions[*]}
    do
        version=${vers}.${rev}
        create_mock_version $version 
    done
done

# This addition will test that a short version with a single 
# release will provide a last version when invoked.
create_mock_version 6.0.$latest_rev
versions[6]='6.0'

run dbdeployer available
for vers in ${versions[*]}
do
    version=${vers}.${latest_rev}
    version_name=$(echo $version | tr '.' '_')
    # Installs using the short version
    run dbdeployer deploy single $vers
    run dbdeployer deploy multiple $vers
    run dbdeployer deploy replication $vers

    # Check existence using the expected revision number
    ok_dir_exists "$SANDBOX_HOME/msb_$version_name"
    ok_dir_exists "$SANDBOX_HOME/multi_msb_$version_name"
    ok_dir_exists "$SANDBOX_HOME/rsandbox_$version_name"
    
    single_command_line=$(grep "command-line.*dbdeployer deploy single $vers" $CATALOG)
    replication_command_line=$(grep "command-line.*dbdeployer deploy replication $vers" $CATALOG)
    multiple_command_line=$(grep "command-line.*dbdeployer deploy multiple $vers" $CATALOG)
    ok "single command line <$single_command_line>" "$single_command_line"
    ok "replication command line <$replication_command_line>" "$replication_command_line"
    ok "multiple command line <$multiple_command_line>" "$multiple_command_line"
    results "$vers"
done

run dbdeployer delete ALL --skip-confirm

results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

