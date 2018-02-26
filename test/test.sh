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

version=$(dbdeployer --version)
if [ -z "$version" ]
then
    echo "dbdeployer not found"
    exit 1
fi

start=$(date)
pass=0
fail=0
tests=0
start_sec=$(date +%s)
date > results.txt
which dbdeployer >> results.txt
dbdeployer --version >> results.txt

function results {
    echo "#$1 $2 $3" 
    echo "#$1 $2 $3" >> results.txt
    echo "dbdeployer sandboxes"
    echo "dbdeployer sandboxes" >> results.txt
    dbdeployer sandboxes 
    dbdeployer sandboxes >> results.txt
    echo ""
    echo "" >> results.txt
}

function ok_equal {
    label=$1
    value1=$2
    value2=$3
    if [ $value1 == $value2 ]
    then
        echo "ok - $label found '$value1' - expected: '$value2' "
        pass=$(($pass+1))
    else
        echo "not ok - $label found '$value1' - expected: '$value2' "
        fail=$(($fail+1))
    fi
    tests=$(($tests+1))
}


function run {
    echo "#$@" >> results.txt
    (set -x
    $@
    )
    exit_code=$?
    echo $exit_code
    if [ "$exit_code" != "0" ]
    then
        exit $exit_code
    fi
}

BINARY_DIR=$HOME/opt/mysql
SANDBOX_HOME=$HOME/sandboxes
if [ ! -d $BINARY_DIR ]
then
    echo "Directory $BINARY_DIR not found"
    exit 1
fi

if [ ! -d $HOME/sandboxes ]
then
    mkdir $HOME/sandboxes
fi

running_mysql=$(ps auxw |grep mysqld | grep $BINARY_DIR)
if [ -n "$running_mysql" ]
then
    ps auxw | grep mysqld
    echo "One or more instances of mysql are running already from $BINARY_DIR."
    echo "This test requires that no mysqld processes are running."
    exit 1
fi

installed_sandboxes=$(dbdeployer sandboxes)
if [ -n "$installed_sandboxes" ]
then
    dbdeployer sandboxes
    echo "One or more sandboxes are already deployed. "
    echo "Please remove (or move) the sandboxes and try again"
    exit 1
fi

# Finding the latest release of every major version
short_versions=(5.0 5.1 5.5 5.6 5.7 8.0)
if [ "$(hostname)" == "dbtestmac" ]
then
    # There is a strange bug in docker for Mac, which fails
    # mysteriously when running several instances of MySQL 5.6
    # So we're skipping it if we know that we're running inside
    # a docker for Mac container.
    short_versions=(5.0 5.1 5.5 5.7 8.0)
fi
group_short_versions=(5.7 8.0)
count=0
all_versions=()
group_versions=()

OS=$(uname)
if [ -x sort_versions.$OS ]
then
    cp sort_versions.$OS sort_versions
fi

if [ ! -x ./sort_versions ]
then
    if [ -f ./sort_versions.go ]
    then
        ENV GOOS=linux GOARCH=386 go build -o sort_versions.linux sort_versions.go 
        ENV GOOS=darwin GOARCH=386 go build -o sort_versions.Darwin sort_versions.go 
        ls -l sort_versions*
        cp sort_versions.$OS sort_versions
    fi
    if [ ! -x ./sort_versions ]
    then
        echo "./sort_versions not found"
        exit 1
    fi
fi

for v in ${short_versions[*]}
do
    #ls $BINARY_DIR | grep "^$v" | ./sort_versions | tail -n 1
    latest=$(ls $BINARY_DIR | grep "^$v" | ./sort_versions | tail -n 1)
    if [ -n "$latest" ]
    then
        all_versions[$count]=$latest
        count=$(($count+1))
    else
        echo "No versions found for $v"
    fi
done

count=0
for v in ${group_short_versions[*]}
do
    latest=$(ls $BINARY_DIR | grep "^$v" | ./sort_versions | tail -n 1)
    if [ -n "$latest" ]
    then
        group_versions[$count]=$latest
        count=$(($count+1))
    fi
done

unset will_fail
for V in ${all_versions[*]} 
do
    if [ ! -d $HOME/opt/mysql/$V ]
    then
        echo "Directory \$HOME/opt/mysql/$V not found"
        will_fail=1
    fi
done

if [ -n "$will_fail" ]
then
    exit 1
fi

#echo  ${all_versions[*]}
searched=${#short_versions[@]}
how_many_versions=${#all_versions[@]}
if [ "$how_many_versions" == "0" ]
then
    echo "Nothing to test. No available versions found in $BINARY_DIR"
    exit 1
fi
echo "Versions to test: $how_many_versions of $searched"
echo "Will test: [${all_versions[*]}]"

for V in ${all_versions[*]} 
do
    for stype in single multiple replication
    do
        echo "#$V"
        run dbdeployer $stype $V

        results "$stype"
    done
    sleep 3
    run dbdeployer global test
    run dbdeployer global test-replication
    # test lock
    num_sandboxes_before=$(dbdeployer sandboxes | wc -l)
    run dbdeployer admin lock ALL
    run dbdeployer delete ALL --skip-confirm
    # deletion of locked sandboxes should be ineffective.
    # We expect to get the same number of sandboxes before and after the deletion
    num_sandboxes_after=$(dbdeployer sandboxes | wc -l)
    ok_equal "num_sandboxes" $num_sandboxes_before $num_sandboxes_after

    run dbdeployer admin unlock ALL
    run dbdeployer delete ALL --skip-confirm
    results "#$V - after deletion"
    num_sandboxes_final=$(dbdeployer sandboxes | wc -l)
    # After unlocking, deletion must work, and we should see that
    # there are no sandboxes left
    ok_equal "num_sandboxes" $num_sandboxes_final 0
 
done

for V in ${group_versions[*]} 
do
    echo "#$V"
    run dbdeployer replication $V --topology=group
    # VF=$(echo $V | tr '.' '_')
    # port=$(~/sandboxes/group_msb_$VF/n1 -BN -e "select @@port")
    run dbdeployer replication $V --topology=group \
        --single-primary
    results "group"

    run dbdeployer global test
    run dbdeployer global test-replication
    run dbdeployer delete ALL --skip-confirm
    results "group - after deletion"

done

stop=$(date)
stop_sec=$(date +%s)
elapsed=$(($stop_sec-$start_sec))
echo "Started: $start"
echo "Started: $start" >> results.txt
echo "Ended  : $stop"
echo "Ended  : $stop" >> results.txt
echo "Elapsed: $elapsed seconds"
echo "Elapsed: $elapsed seconds" >> results.txt
echo "Passed subtests: $pass"
echo "Failed subtests: $fail"
exit_code=0
if [ "$fail" != "0" ]
then
    echo "*** FAILURES DETECTED ***"
    exit_code=1
fi
echo "Exit code: $exit_code"
exit $exit_code
