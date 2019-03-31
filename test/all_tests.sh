#!/bin/bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2019 Giuseppe Maxia
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
export RUN_CONCURRENTLY=1
export EXIT_ON_FAILURE=1
[ -n "$SEQUENTIAL" ] && unset RUN_CONCURRENTLY
[ -n "$CONTINUE_ON_FAILURE" ] && unset EXIT_ON_FAILURE

if [ ! -d "./test" ]
then
    echo "directory ./test not found"
    exit 1
fi

version=$1
if [ -z $version ]
then
    echo "version needed"
    exit 1
fi

executable=dbdeployer-${version}.linux
if [ ! -x $executable  ]
then
    echo "executable '$executable' not found"
    exit 1
fi

if [ ! -f ./test/common.sh ]
then
    echo "script './test/common.sh' not found"
    exit 1
fi

source ./test/common.sh

start_timer
timestamp=$(date +%Y-%m-%d-%H.%M)

if [ ! -d ./test/logs ]
then
    mkdir ./test/logs
fi

mkdir ./test/logs/$timestamp
log_summary=./test/logs/$timestamp/all_tests-summary.log

function summary {
    exit_code=$1
    cat $log_summary
    stop_timer $log_summary
    #rm -f $log_summary
    echo "# Exit code: $exit_code"
    concurrency=no
    if [ -n "$RUN_CONCURRENTLY" ]
    then
        concurrency=yes
    fi
    echo "Runs concurrently: $concurrency"
    echo "Runs concurrently: $concurrency" >> $log_summary
    exit $exit_code
}

function run_test {
    test_name="$1"
    test_arg="$2"
    start_test=$(date +%s)
    test_base_name=$(basename $test_name .sh)
    test_log=./test/logs/$timestamp/${test_base_name}.log
    echo "# Running $test_name $test_arg"
    $test_name $test_arg > $test_log
    exit_code=$?
    fail_count=$(grep -i -c '^not ok' $test_log)
    end_test=$(date +%s)
    elapsed=$((end_test-start_test))
    test_arg="[$test_arg]"
    printf "%-30s %-9s - time: %4ds (%10s) - exit code: %d\n" $test_base_name $test_arg ${elapsed} $(minutes_seconds $elapsed) $exit_code >> $log_summary
    if [ "$test_base_name" == "port-clash" ]
    then
        sandboxes=$(grep -c catalog $test_log)
        ports=$(grep "Total ports installed" $test_log | awk '{print $NF}')
        changed=$(grep -c changed $test_log)
        echo "# Deployed: $sandboxes sandboxes ($ports total ports) - Changed: $changed" >> $log_summary
    fi
    if [ "$exit_code" != "0" -o "$fail_count"  != "0" ]
    then
        echo $dash_line
        echo "# Error detected: $test_base_name "
        echo "# exit_code     : $exit_code"
        echo "# fail count    : $fail_count"
        echo $dash_line
        tail -n 20 $test_log
        echo $dash_line
        summary $exit_code
    fi
}

function all_tests {
    if [ -z "$ONLY_MOCK" ]
    then
        run_test ./scripts/sanity_check.sh
        run_test ./test/go-unit-tests.sh
        run_test ./test/functional-test.sh
        run_test ./test/docker-test.sh $version
    fi
    run_test ./test/mock/defaults-change.sh
    run_test ./test/mock/short-versions.sh
    run_test ./test/mock/direct-paths.sh
    run_test ./test/mock/expected_ports.sh
    run_test ./test/mock/read-only-replication.sh
    run_test ./test/mock/ndb_test.sh
    run_test ./test/mock/pxc_test.sh
    run_test ./test/mock/cookbook.sh
    if [ -n "$COMPLETE_PORT_TEST" ]
    then
        run_test ./test/mock/port-clash.sh
    else
        run_test ./test/mock/port-clash.sh sparse
    fi
}

all_tests

echo $dash_line
echo $dash_line >> $log_summary
for logfile in ./test/logs/$timestamp/*.log
do
    fname=$(basename $logfile .log)
    if [ "$fname" != "all_tests-summary" ]
    then
        lf_pass=$(grep -c '^ok' $logfile)
        lfg_pass=$(grep -c ': ok' $logfile)
        lfg_fail=$(grep -c ': not ok' $logfile)
        lf_fail=$(grep -i -c '^not ok' $logfile)
        lf_fail=$((lf_fail+lfg_fail))
        lf_pass=$((lf_pass+lfg_pass))
        lf_tests=$((lf_pass+lf_fail))
        printf "# %-25s - tests: %4d - pass: %4d - fail: %4d\n" $fname $lf_tests $lf_pass $lf_fail
        printf "# %-25s - tests: %4d - pass: %4d - fail: %4d\n" $fname $lf_tests $lf_pass $lf_fail >> $log_summary
    fi
done
echo $dash_line
echo $dash_line >> $log_summary
pass=$(grep '^ok' ./test/logs/$timestamp/*.log| wc -l | tr -d ' ' )
fail=$(grep -i '^not ok' ./test/logs/$timestamp/*.log| wc -l | tr -d ' ' )
gpass=$(grep ': ok' ./test/logs/$timestamp/*.log| wc -l | tr -d ' ' )
gfail=$(grep -i ': not ok' ./test/logs/$timestamp/*.log| wc -l | tr -d ' ' )
pass=$((pass+gpass))
fail=$((fail+gfail))
tests=$((pass+fail))
exit_code=0
if [ "$fail"  != "0" ]
then
    exit_code=1
fi
echo $dash_line >> $log_summary
echo "# Total tests: $tests"              >> $log_summary
echo "#       pass : $pass"               >> $log_summary
echo "#       fail : $fail"               >> $log_summary
echo $dash_line >> $log_summary
summary $exit_code
