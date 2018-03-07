#!/bin/bash
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
export results_log=$PWD/port-clash.log
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
# {
#  	"version": "0.2.2",
#  	"sandbox-home": "$HOME/sandboxes",
#  	"sandbox-binary": "$HOME/opt/mysql",
#  	"master-slave-base-port": 11000,
#  	"group-replication-base-port": 12000,
#  	"group-replication-sp-base-port": 13000,
#  	"fan-in-replication-base-port": 14000,
#  	"all-masters-replication-base-port": 15000,
#  	"multiple-base-port": 16000,
#  	"group-port-delta": 125,
#  	"master-name": "master",
#  	"master-abbr": "m",
#  	"node-prefix": "node",
#  	"slave-prefix": "slave",
#  	"slave-abbr": "s",
#  	"sandbox-prefix": "msb_",
#  	"master-slave-prefix": "rsandbox_",
#  	"group-prefix": "group_msb_",
#  	"group-sp-prefix": "group_sp_msb_",
#  	"multiple-prefix": "multi_msb_",
#  	"fan-in-prefix": "fan_in_msb_",
#  	"all-masters-prefix": "all_masters_msb_"
# }

function ok_generic_exists {
    wanted=$1
    label=$2
    op=$3
    if [ $op "$wanted" ]
    then
        echo "ok - $label $wanted exists"
        pass=$((pass+1))
    else
        echo "NOT OK - $label $wanted does not  exist"
        fail=$((fail+1))
    fi
}

function ok_dir_exists {
    dir=$1
    ok_generic_exists $dir directory -d
}

function ok_file_exists {
    filename=$1
    ok_generic_exists $filename "file" -f
}

function ok_executable_exists {
    filename=$1
    ok_generic_exists $filename "file" -x
}

create_mock_version 5.6.66
create_mock_version 5.7.66
create_mock_version 8.0.66

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
ok_dir_exists $sandbox_dir
ok_dir_exists $sandbox_dir/branch1
ok_dir_exists $sandbox_dir/branch2
ok_executable_exists $sandbox_dir/p
ok_executable_exists $sandbox_dir/r1
ok_executable_exists $sandbox_dir/r2
ok_executable_exists $sandbox_dir/check_replicas
ok_executable_exists $sandbox_dir/initialize_replicas

run dbdeployer deploy replication 5.7.66 \
    --defaults=master-slave-prefix:masterslave_ \
    --defaults=master-name:batman \
    --defaults=master-abbr:b \
    --defaults=slave-prefix:robin \
    --defaults=slave-abbr:rob \
    --defaults=node-prefix:bat 

sandbox_dir=$SANDBOX_HOME/masterslave_5_7_66
ok_dir_exists $sandbox_dir
ok_dir_exists $sandbox_dir/bat1
ok_dir_exists $sandbox_dir/bat2
ok_executable_exists $sandbox_dir/b
ok_executable_exists $sandbox_dir/rob1
ok_executable_exists $sandbox_dir/rob2
ok_executable_exists $sandbox_dir/check_robins
ok_executable_exists $sandbox_dir/initialize_robins

run dbdeployer defaults reset
run dbdeployer deploy replication 8.0.66 

sandbox_dir=$SANDBOX_HOME/rsandbox_8_0_66
ok_dir_exists $sandbox_dir
ok_dir_exists $sandbox_dir/node1
ok_dir_exists $sandbox_dir/node2
ok_executable_exists $sandbox_dir/m
ok_executable_exists $sandbox_dir/s1
ok_executable_exists $sandbox_dir/s2
ok_executable_exists $sandbox_dir/check_slaves
ok_executable_exists $sandbox_dir/initialize_slaves

echo "#Total sandboxes: $(count_catalog)"
echo "#Total sandboxes: $(count_catalog)" >> $results_log
if [ "$fail" != "0" ]
then
    exit 1
fi
run dbdeployer delete ALL --skip-confirm

results "After deletion"
cd $test_dir 

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer
tests=$((pass+fail))
echo "Tests:  $tests"
echo "Pass :  $pass"
echo "Fail :  $fail"

