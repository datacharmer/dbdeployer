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

execdir=$(dirname "$0")

cd "$execdir" || exit 1

if [ ! -f common.sh ]
then
    echo "common.sh not found"
    exit 1
fi

source common.sh

start_timer
pass=0
fail=0
tests=0

(which dbdeployer ; dbdeployer --version ; uname -a ) >> "$results_log"

function user_input {
    answer=""
    while [ "$answer" != "continue" ]
    do
        echo "Press ENTER to continue or choose among { s c q i o r u  h }"
        read answer
        case $answer in
            [cC])
                unset INTERACTIVE
                echo "Now running unattended"
                return
                ;;
            [qQ])
                echo "Interrupted at user's request"
                exit 0
                ;;
            [iI])
                echo inspecting
                show_catalog
                ;;
            [oO])
                echo counting
                count_catalog
                ;;
            [sS])
                echo show sandboxes
                dbdeployer sandboxes --catalog
                ;;
            [rR])
                echo "Enter global command to run"
                echo "Choose among : start restart stop status test test-replication"
                read cmd
                dbdeployer global $cmd
                if [ "$?" != "0" ]
                then
                    exit 1
                fi
                ;;
            [uU])
                echo "Enter query to run"
                read cmd
                dbdeployer global use "$cmd"
                if [ "$?" != "0" ]
                then
                    exit 1
                fi
                ;;
            [hH])
                echo "Commands:"
                echo "c : continue (end interactivity)"
                echo "i : inspect sandbox catalog"
                echo "o : count sandbox instances"
                echo "q : quit the test immediately"
                echo "r : run 'dbdeployer global' command"
                echo "u : run 'dbdeployer global use' query"
                echo "s : show sandboxes"
                echo "h : display this help"
                ;;
            *)
                answer="continue"
        esac
    done
}

[ -z "$BINARY_DIR" ] && BINARY_DIR=$HOME/opt/mysql
[ -z "$SANDBOX_HOME" ] && SANDBOX_HOME=$HOME/sandboxes
if [ ! -d "$BINARY_DIR" ]
then
    echo "Directory (\$BINARY_DIR) "$BINARY_DIR" not found"
    exit 1
fi

if [ ! -d "$SANDBOX_HOME" ]
then
    mkdir "$SANDBOX_HOME"
fi
if [ ! -d "$SANDBOX_HOME" ]
then
    echo "Directory (\$SANDBOX_HOME) "$SANDBOX_HOME" could not be created"
    exit 1
fi

function test_uuid {
    running_version=$1
    group_dir_name=$2
    must_exist=$3
    version_path=$(echo $running_version| tr '.' '_')
    count=0
    if [ -d $SANDBOX_HOME/$group_dir_name$version_path/master ]
    then
        nodes="master node1 node2"
    else
        nodes="node1 node2 node3"
    fi
    for dir in $nodes
    do
        uuid_file=$SANDBOX_HOME/$group_dir_name$version_path/$dir/data/auto.cnf
        count=$((count+1))
        if [ ! -f $uuid_file ]
        then
            if [ -n "$must_exist" ]
            then
                ok "File $uuid_file not found"
            fi
            return
        fi
        uuid=$(tail -n 1 $uuid_file | sed -e 's/server-uuid=//')
        ok "UUID found in $uuid_file"  $uuid
        repeated_count=$count$count$count$count
        expected=$repeated_count-$repeated_count-$repeated_count-${repeated_count}${repeated_count}${repeated_count}
        port=$($SANDBOX_HOME/$group_dir_name$version_path/$dir/use -BN -e 'select @@port')
        uuid_sql=$($SANDBOX_HOME/$group_dir_name$version_path/$dir/use -BN -e 'select @@server_uuid')
        echo "# $uuid"
        ok_contains "UUID" "$uuid" "$expected"
        ok_contains "UUID" "$uuid" "$port"
        ok_equal "UUID in file and SQL" $uuid $uuid_sql
    done
}

function test_deletion {
    del_version=$1
    expected_items=$2
    # test lock: sandboxes become locked against deletion
    num_sandboxes_before=$(dbdeployer sandboxes | wc -l)
    run dbdeployer admin lock ALL
    run dbdeployer delete ALL --skip-confirm
    # deletion of locked sandboxes should be ineffective.
    # We expect to get the same number of sandboxes before and after the deletion
    num_sandboxes_after=$(dbdeployer sandboxes | wc -l)
    ok_equal "num_sandboxes" $num_sandboxes_before $num_sandboxes_after

    how_many=$(count_catalog)
    ok_equal "sandboxes_in_catalog" $how_many $expected_items

    run dbdeployer admin unlock ALL
    run dbdeployer delete ALL --skip-confirm
    results "#$del_version - after deletion"
    num_sandboxes_final=$(dbdeployer sandboxes --catalog | wc -l)
    # After unlocking, deletion must work, and we should see that
    # there are no sandboxes left
    ok_equal "num_sandboxes" $num_sandboxes_final 0

    how_many=$(count_catalog)
    ok_equal "sandboxes_in_catalog" $how_many 0
    if [ "$fail" != "0" ]
    then
        exit 1
    fi
}

function capture_test {
    cmd="$@"
    output=/tmp/capture_test$$
    $cmd > $output 2>&1
    tmp_pass=$(grep '^ok' $output | wc -l | tr -d ' ')
    pass=$((pass+tmp_pass))
    tmp_fail=$(grep -i '^not ok' $output | wc -l | tr -d ' ')
    fail=$((fail+tmp_fail))
    tests=$((tests+tmp_fail+tmp_pass))
    cat $output
    if [ "$tmp_fail" != "0" ]
    then
        echo "CMD: $cmd"
        echo "results in $output ($tmp_fail)"
        exit 1
    fi
    rm -f $output
}

running_mysql=$(ps auxw |grep mysqld | grep $BINARY_DIR)
if [ -n "$running_mysql" ]
then
    ps auxw | grep mysqld
    echo "One or more instances of mysql are running already from $BINARY_DIR."
    echo "This test requires that no mysqld processes are running."
    exit 1
fi

installed_sandboxes=$(dbdeployer sandboxes --catalog)
if [ -n "$installed_sandboxes" ]
then
    (set -x
    dbdeployer sandboxes --catalog
    )
    echo "One or more sandboxes are already deployed. "
    echo "Please remove (or move) the sandboxes and try again"
    exit 1
fi

catalog_items=$(count_catalog)
if [ "$catalog_items" != "0" ]
then
    (set -x
    dbdeployer sandboxes --catalog
    )
    echo "Found $catalog_items items in the catalog. Expected: 0"
    exit 1
fi

# Finding the latest release of every major version

[ -z "$short_versions" ] && short_versions=(5.0 5.1 5.5 5.6 5.7 8.0)

if [ "$(hostname)" == "dbtestmac" ]
then
    # There is a strange bug in docker for Mac, which fails
    # mysteriously when running several instances of MySQL 5.6
    # So we're skipping it if we know that we're running inside
    # a docker for Mac container.
    short_versions=(5.0 5.1 5.5 5.7 8.0)
fi
[ -z "$group_short_versions" ] && group_short_versions=(5.7 8.0)
count=0
all_versions=()
group_versions=()

OS=$(uname)
if [ -x "sort_versions.$OS" ]
then
    cp "sort_versions.$OS" sort_versions
fi

if [ ! -x ./sort_versions ]
then
    if [ -f ./sort_versions.go ]
    then
        ENV GOOS=linux GOARCH=386 go build -o sort_versions.linux sort_versions.go
        ENV GOOS=darwin GOARCH=386 go build -o sort_versions.Darwin sort_versions.go
        ls -l sort_versions*
        cp "sort_versions.$OS" sort_versions
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
    latest=$(ls "$BINARY_DIR" | grep "^$v" | ./sort_versions | tail -n 1)
    if [ -n "$latest" ]
    then
        all_versions[$count]=$latest
        count=$((count+1))
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
        count=$((count+1))
    fi
done

unset will_fail
for V in ${all_versions[*]}
do
    if [ ! -d "$BINARY_DIR/$V" ]
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

# -----------------------------
# Deployment tests start here
# -----------------------------
for V in ${all_versions[*]}
do
    # We test the main deployment methods
    for stype in single multiple replication
    do
        echo "#$V"
        run dbdeployer deploy $stype $V
        # For each type, we display basic info
        results "$stype"
    done
    # Test server UUID. It will skipped for versions
    # that don't support it.
    test_uuid $V multi_msb_
    test_uuid $V rsandbox_
    how_many=$(count_catalog)
    ok_equal "sandboxes_in_catalog" $how_many 3
    sleep 2
    # Runs basic tests
    run dbdeployer global status
    capture_test run dbdeployer global test
    capture_test run dbdeployer global test-replication
    test_deletion $V 3
done

for V in ${group_versions[*]}
do
    echo "# Group operations $V"
    run dbdeployer deploy replication $V --topology=group
    run dbdeployer deploy replication $V --topology=group \
        --single-primary
    results "group"

    capture_test run dbdeployer global test
    capture_test run dbdeployer global test-replication
    test_uuid $V group_msb_ 1
    test_uuid $V group_sp_msb_ 1
    test_deletion $V 2
    results "group - after deletion"
done

for V in ${group_versions[*]}
do
    echo "# Multi-source operations $V"
    run dbdeployer deploy replication $V --topology=fan-in
    run dbdeployer deploy replication $V --topology=all-masters
    results "multi-source"

    capture_test run dbdeployer global test
    capture_test run dbdeployer global test-replication
    test_uuid $V fan_in_msb_ 1
    test_uuid $V all_masters_msb_ 1
    test_deletion $V 2
    results "multi-source - after deletion"
done

stop_timer

echo "Passed subtests: $pass"
echo "Passed subtests: $pass" >> "$results_log"
echo "Failed subtests: $fail"
echo "Failed subtests: $fail" >> "$results_log"
echo "Total  subtests: $tests"
echo "Total  subtests: $tests" >> "$results_log"
exit_code=0
if [ "$fail" != "0" ]
then
    echo "*** FAILURES DETECTED ***"
    exit_code=1
fi
echo "Exit code: $exit_code"
echo "Exit code: $exit_code" >> "$results_log"
exit $exit_code
