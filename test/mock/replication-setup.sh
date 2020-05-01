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

versions_ro=(5.1 5.5 5.6 5.7 8.0)
versions_sro=(5.7 8.0)
#rev_list="0 21 43 65 87 98"
rev_list="20 31 88"

for rev in $rev_list
do
    for vers in ${versions_ro[*]}
    do
        version=${vers}.${rev}
        create_mock_version $version 
    done
done

function ok_var {
    label=$1
    variable=$2
    expected=$3
    value=$4
    if [ "$expected" == "$value" ]
    then
       echo "ok - [$label] var: $variable <$value> "
       pass=$((pass+1))
   else
       echo "not ok - [$label] var: $variable <$value> - expected: $expected" 
       fail=$((fail+1))
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            exit 1
        fi
    fi 
}

function check_variables {
    version=$1
    dir_prefix=$2
    variable=$3
    expected_m=$4
    expected_s1=$5
    expected_s2=$6
    expected_s3=$7
    vpath=$(echo $version | tr '.' '_')
    dir=$SANDBOX_HOME/$dir_prefix$vpath
    if [ -d $dir/master ]
    then
        var_m=$(grep -w $variable $dir/master/my.sandbox.cnf)
    else
        var_m=""
    fi
    var_s1=$(grep -w $variable $dir/node1/my.sandbox.cnf)
    var_s2=$(grep -w $variable $dir/node2/my.sandbox.cnf)
    if [ -d $dir/node3 ]
    then
        var_s3=$(grep -w $variable $dir/node3/my.sandbox.cnf)
    else
        var_s3=""
    fi
    ok_var master $variable "$expected_m" "$var_m"
    ok_var s1 $variable "$expected_s1" "$var_s1"
    ok_var s2 $variable "$expected_s2" "$var_s2"
    ok_var s3 $variable "$expected_s3" "$var_s3"
}

run dbdeployer available

# When no read-only is set, we expect to find empty definitions in the options file
for vers in ${versions_ro[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy replication $SANDBOX_BINARY/$version 

        # Check existence 
        ok_dir_exists "$SANDBOX_HOME/rsandbox_$version_name"
        check_variables $version rsandbox_ read_only "" "" ""
        check_variables $version rsandbox_ super_read_only "" "" ""
        
        test_completeness $version rsandbox_ replication

        results "$version"
    done
done

run dbdeployer delete ALL --skip-confirm

# test read-only option
for vers in ${versions_ro[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy replication $SANDBOX_BINARY/$version --read-only-slaves

        # Check existence 
        ok_dir_exists "$SANDBOX_HOME/rsandbox_$version_name"
        check_variables $version rsandbox_ read_only "" read_only=on read_only=on
        check_variables $version rsandbox_ super_read_only "" "" ""
        
        test_completeness $version rsandbox_ replication

        results "$version"
    done
done

run dbdeployer delete ALL --skip-confirm

# test super-read-only option
for vers in ${versions_sro[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy replication $SANDBOX_BINARY/$version --super-read-only-slaves

        # Check existence 
        ok_dir_exists "$SANDBOX_HOME/rsandbox_$version_name"
        check_variables $version rsandbox_ super_read_only "" super_read_only=on super_read_only=on
        check_variables $version rsandbox_ read_only "" "" ""
        
        test_completeness $version rsandbox_ replication

        results "$version"
    done
done

run dbdeployer delete ALL --skip-confirm

# test super-read-only option in fan-in
for vers in ${versions_sro[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy replication $SANDBOX_BINARY/$version --super-read-only-slaves --topology=fan-in

        # Check existence 
        ok_dir_exists "$SANDBOX_HOME/fan_in_msb_$version_name"
        check_variables $version fan_in_msb_ super_read_only "" "" "" super_read_only=on
        check_variables $version fan_in_msb_ read_only "" "" "" ""
        
        results "$version"
    done
done

run dbdeployer delete ALL --skip-confirm

# test read-only option in fan-in
for vers in ${versions_sro[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy replication $SANDBOX_BINARY/$version --read-only-slaves --topology=fan-in

        # Check existence 
        ok_dir_exists "$SANDBOX_HOME/fan_in_msb_$version_name"
        check_variables $version fan_in_msb_ super_read_only "" "" "" ""
        check_variables $version fan_in_msb_ read_only "" "" "" read_only=on

        results "$version"
    done
done

run dbdeployer delete ALL --skip-confirm

# test change-master-options
for vers in ${versions_ro[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy replication $SANDBOX_BINARY/$version --change-master-options=DUMMY1=ABCDE  --change-master-options=DUMMY2=WXYZ

        # Check existence
        ok_dir_exists "$SANDBOX_HOME/rsandbox_$version_name"

        dummy1_set=$(grep 'DUMMY1=ABCDE' $SANDBOX_HOME/rsandbox_$version_name/initialize_slaves)
        dummy2_set=$(grep 'DUMMY2=WXYZ' $SANDBOX_HOME/rsandbox_$version_name/initialize_slaves)
        ok "master option 1 set" "$dummy1_set"
        ok "master option 2 set" "$dummy2_set"

        results "$version"
    done
done

run dbdeployer delete ALL --skip-confirm

results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

