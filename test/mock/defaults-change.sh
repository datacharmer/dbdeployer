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
export results_log=$PWD/defaults-change.log
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

function check_deployment {
    sandbox_dir=$1
    node_dir=$2
    master_abbr=$3
    slave_abbr=$4
    slave_name=$5
    ok_dir_exists $sandbox_dir
    ok_dir_exists $sandbox_dir/${node_dir}1
    ok_dir_exists $sandbox_dir/${node_dir}2
    ok_executable_exists $sandbox_dir/$master_abbr
    ok_executable_exists $sandbox_dir/${slave_abbr}1
    ok_executable_exists $sandbox_dir/${slave_abbr}2
    ok_executable_exists $sandbox_dir/check_${slave_name}s
    ok_executable_exists $sandbox_dir/initialize_${slave_name}s
}

create_mock_version 5.5.66
create_mock_version 5.6.66
create_mock_version 5.7.66
create_mock_version 8.0.66

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
check_deployment $sandbox_dir branch p r replica

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
check_deployment $sandbox_dir bat b rob robin

# We make sure that the defaults stay the same, and they
# were not affected by the dynamic changes
run dbdeployer deploy replication 5.5.66
sandbox_dir=$SANDBOX_HOME/ms_replication_5_5_66
check_deployment $sandbox_dir branch p r replica

# Restore the original defaults
run dbdeployer defaults reset
run dbdeployer deploy replication 8.0.66

sandbox_dir=$SANDBOX_HOME/rsandbox_8_0_66
check_deployment $sandbox_dir node m s slave

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

