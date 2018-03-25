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


test_dir=$(dirname $0)
cd $test_dir
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
#export results_log=$PWD/defaults-change.log
source set-mock.sh
export SHOW_CHANGED_PORTS=1
start_timer

pwd
ls -l
echo "HOME:           $HOME"
echo "SANDBOX_HOME :  $SANDBOX_HOME"
echo "SANDBOX_BINARY: $SANDBOX_BINARY"

mkdir $HOME
mkdir -p $SANDBOX_BINARY
mkdir $SANDBOX_HOME
tests=0
fail=0
pass=0


function check_deployment {
    sandbox_dir=$1
    node_dir=$2
    master_name=$3
    master_abbr=$4
    slave_abbr=$5
    slave_name=$6
    ok_dir_exists $sandbox_dir
    ok_dir_exists $sandbox_dir/${master_name}
    ok_dir_exists $sandbox_dir/${node_dir}1
    ok_dir_exists $sandbox_dir/${node_dir}2
    ok_executable_exists $sandbox_dir/$master_abbr
    ok_executable_exists $sandbox_dir/${slave_abbr}1
    ok_executable_exists $sandbox_dir/${slave_abbr}2
    ok_executable_exists $sandbox_dir/check_${slave_name}s
    ok_executable_exists $sandbox_dir/initialize_${slave_name}s
    for sbdir in $master_name ${node_dir}1 ${node_dir}2
    do
        auto_cnf=$sandbox_dir/$sbdir/data/auto.cnf
        if [ -f $auto_cnf ]
        then
            tail -n 1 $auto_cnf | sed -e 's/server-uuid=//'
            #echo $auto_cnf
            #cat $auto_cnf
        fi
    done
}

create_mock_version 5.5.66
create_mock_version 5.6.66
create_mock_version 5.7.66
create_mock_version 8.0.66
create_mock_version 8.0.67

# Changing all defaults statically
run dbdeployer defaults show
run dbdeployer defaults update master-slave-prefix ms_replication_
run dbdeployer defaults update master-name primary
run dbdeployer defaults update master-abbr p
run dbdeployer defaults update slave-prefix replica
run dbdeployer defaults update slave-abbr r
run dbdeployer defaults update node-prefix branch
run dbdeployer defaults show

run dbdeployer deploy replication 5.6.66

sandbox_dir=$SANDBOX_HOME/ms_replication_5_6_66
check_deployment $sandbox_dir branch primary p r replica

# Keeping the changes, we deploy a new replication cluster
# with the defaults changing dynamically.

run dbdeployer deploy replication 5.7.66 \
    --defaults=master-slave-prefix:masterslave_ \
    --defaults=master-name:batman \
    --defaults=master-abbr:b \
    --defaults=slave-prefix:robin \
    --defaults=slave-abbr:rob \
    --defaults=node-prefix:bat

sandbox_dir=$SANDBOX_HOME/masterslave_5_7_66
check_deployment $sandbox_dir bat batman b rob robin

# We make sure that the defaults stay the same, and they
# were not affected by the dynamic changes
run dbdeployer deploy replication 5.5.66
sandbox_dir=$SANDBOX_HOME/ms_replication_5_5_66
check_deployment $sandbox_dir branch primary p r replica

# Restore the original defaults
run dbdeployer defaults reset
run dbdeployer deploy replication 8.0.66

sandbox_dir=$SANDBOX_HOME/rsandbox_8_0_66
check_deployment $sandbox_dir node master m s slave

echo "#Total sandboxes: $(count_catalog)"
#echo "#Total sandboxes: $(count_catalog)" >> $results_log
if [ "$fail" != "0" ]
then
    exit 1
fi

temp_template=t$$.dat
timestamp=$(date +%Y-%m-%d.%H:%M:%S)
echo "#!/bin/bash" > $temp_template
echo "echo 'I AM A CUSTOM_TEMPLATE CREATED ON $timestamp'" >> $temp_template
run dbdeployer deploy --use-template=clear_template:$temp_template single 8.0.67
sandbox_dir=$SANDBOX_HOME/msb_8_0_67
message=$($sandbox_dir/clear)
ok_contains "custom template" "$message" "CUSTOM_TEMPLATE"
ok_contains "custom template" "$message" $timestamp

run dbdeployer delete ALL --skip-confirm

results "After deletion"

run dbdeployer defaults templates export single $mock_dir/templates clear_template
cp $temp_template $mock_dir/templates/single/clear_template
rm -f $temp_template
run dbdeployer defaults templates import single $mock_dir/templates
installed=$(dbdeployer defaults templates list | grep "clear_template" | grep '{F}')
echo "# installed template: <$installed>"
ok "template was installed" "$installed"
run dbdeployer deploy single 8.0.67

sandbox_dir=$SANDBOX_HOME/msb_8_0_67
message=$($sandbox_dir/clear)
ok_contains "installed custom template" "$message" "CUSTOM_TEMPLATE"
ok_contains "installed custom template" "$message" $timestamp

run dbdeployer delete ALL --skip-confirm

cd $test_dir

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer
tests=$((pass+fail))
echo "Tests:  $tests"
echo "Pass :  $pass"
echo "Fail :  $fail"

