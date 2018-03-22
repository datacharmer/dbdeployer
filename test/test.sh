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

options=$1

if [ -n "$options" ]
then
    # If there is any option on the command line,
    # disable all tests
    export skip_main_deployment_methods=1
    export skip_pre_post_operations=1
    export skip_group_operations=1
    export skip_multi_source_operations=1
    export no_tests=1
fi

while [ -n "$options" ]
do
    # Enable tests based on command line options.
    case $options in
        interactive)
            export INTERACTIVE=1
            echo "# Enabling INTERACTIVE"
            ;;
        concurrent)
            export RUN_CONCURRENTLY=1
            echo "# Enabling CONCURRENCY"
            ;;
        sequential)
            unset RUN_CONCURRENTLY
            echo "# Disabling CONCURRENCY"
            ;;
        all)
            unset skip_main_deployment_methods
            unset skip_pre_post_operations
            unset skip_group_operations
            unset skip_multi_source_operations
            unset no_tests
            echo "# Enabling all tests"
            ;;
        main)
            unset skip_main_deployment_methods
            unset no_tests
            echo "# Enabling main tests"
            ;;
        pre)
            unset skip_pre_post_operations
            unset no_tests
            echo "# Enabling pre/post tests"
            ;;
        post)
            unset skip_pre_post_operations
            unset no_tests
            echo "# Enabling pre/post tests"
            ;;
        group)
            unset skip_group_operations
            unset no_tests
            echo "# Enabling group operations tests"
            ;;
        multi)
            unset skip_multi_source_operations
            unset no_tests
            echo "# Enabling multi-source operations tests"
            ;;
        *)
            echo "Allowed tests (you can choose more than one):"
            echo "  main     : main deployment methods"
            echo "  pre/post : pre/post grants operations"
            echo "  group    : group replication operations "
            echo "  multi    : multi-source operations (fan-in, all-masters)"
            echo "  all      : enable all the above tests"
            echo ""
            echo "Allowed modifiers:"
            echo "  concurrent  : Enable concurrent operations"
            echo "  sequential  : Disable concurrent operations"
            echo "  interactive : Enable interaction with user"
            exit 1
    esac
    shift
    options=$1
done

if [ -n "$no_tests" ]
then
    echo "No tests were defined - aborting"
    echo "Run '$0 help' for the list of available tests"
    exit 1
fi

start_timer
pass=0
fail=0
tests=0

(which dbdeployer ; dbdeployer --version ; uname -a ) >> "$results_log"

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
        echo "# UUID from file: $uuid"
        echo "# UUID from SQL:  $uuid_sql"
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

function main_deployment_methods {
    current_test=main_deployment_methods
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
        # Test server UUID. It will be skipped for versions
        # that don't support it.
        how_many=$(count_catalog)
        ok_equal "sandboxes_in_catalog" $how_many 3
        sleep 2
        # Runs basic tests
        run dbdeployer global status
        test_uuid $V multi_msb_
        test_uuid $V rsandbox_
        capture_test run dbdeployer global test
        capture_test run dbdeployer global test-replication
        test_deletion $V 3
    done
}

function pre_post_operations {
    current_test=pre_post_operations
    # This test checks the following:
    #   * we can run a SQL command before the grants are loaded
    #   * we match the result of that query with an expected value
    #   * we can run more than one query before and after
    #   * we can compare values before and after the grants were loaded
    #   * we can run queries containing commas without errors.
    #     (see ./pflag/README.md)
    for V in ${all_versions[*]}
    do
        echo "#pre-post operations $V"
        outfile=/tmp/pre-post-$V.txt
        (set -x
        dbdeployer deploy single $V \
            --pre-grants-sql="select 'preversion' as label, @@version" \
            --pre-grants-sql="select 'preschema' as label, count(*) as PRE from information_schema.schemata" \
            --pre-grants-sql="select 'preusers' as label, count(*) as PRE from mysql.user" \
            --post-grants-sql="select 'postversion' as label, @@version" \
            --post-grants-sql="select 'postschema' as label, count(*) as POST from information_schema.schemata" \
            --post-grants-sql="select 'postusers' as label, count(*) as POST from mysql.user" > $outfile 2>&1
        )
        # Gets the line with a given label.
        # retrieves the fourth element in the line
        pre_users=$(grep preusers $outfile | awk '{print $4}')
        post_users=$(grep postusers $outfile | awk '{print $4}')
        pre_version=$(grep preversion $outfile | awk '{print $4}')
        post_version=$(grep postversion $outfile | awk '{print $4}')
        pre_schema=$(grep preschema $outfile | awk '{print $4}')
        post_schema=$(grep postschema $outfile | awk '{print $4}')
        # cat $outfile
        ok_greater "post grants users more than pre" $post_users $pre_users
        ok_greater_equal "same or more schemas before and after grants" $post_schema $pre_schema
        ok_contains "Version" $pre_version $V
        ok_contains "Version" $post_version $V
        results "pre/post $V"
        rm $outfile
        dbdeployer delete ALL --skip-confirm
    done
}

function group_operations {
    current_test=group_operations
    for V in ${group_versions[*]}
    do
        echo "# Group operations $V"
        run dbdeployer deploy replication $V --topology=group
        run dbdeployer deploy replication $V --topology=group \
            --single-primary
        results "group $V"

        capture_test run dbdeployer global test
        capture_test run dbdeployer global test-replication
        test_uuid $V group_msb_ 1
        test_uuid $V group_sp_msb_ 1
        test_deletion $V 2
        results "group $V - after deletion"
    done
}

function multi_source_operations {
    current_test=multi_source_operations
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
}

if [ -z "$skip_main_deployment_methods" ]
then
    main_deployment_methods
fi
if [ -z "$skip_pre_post_operations" ]
then
    pre_post_operations
fi
if [ -z "$skip_group_operations" ]
then
    group_operations
fi
if [ -z "$skip_multi_source_operations" ]
then
    multi_source_operations
fi

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
