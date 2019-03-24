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

execdir=$(dirname "$0")
execdirname=$(basename $execdir)

# If we are inside a docker container (under TravisCI)
# the test directry was renamed original_test.
# The container will be unable to modify files in the original test directory
# and therefore we make a local copy.
if [ "$execdirname" == "original_test" ]
then
    if [ -d "./test" ]
    then
        echo "Can't create ./test from $execdir. Directory already exists "
        exit 1
    else
        cp -r $execdir ./test
        execdir=./test
        ls -lh
    fi
fi

cd "$execdir" || exit 1

if [ ! -f common.sh ]
then
    echo "common.sh not found"
    exit 1
fi
source common.sh

options=$1

if [ -n "$ONLY_MAIN" ]
then
    options=main
fi

if [ -n "$options" ]
then
    # If there is any option on the command line,
    # disable all tests
    export skip_main_deployment_methods=1
    export skip_tidb_deployment_methods=1
    export skip_skip_start_deployment=1
    export skip_pre_post_operations=1
    export skip_semisync_operations=1
    export skip_group_operations=1
    export skip_dd_operations=1
    export skip_upgrade_operations=1
    export skip_multi_source_operations=1
    export skip_pxc_operations=1
    export skip_ndb_operations=1
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
        exitfail)
            export EXIT_ON_FAILURE=1
            echo "# Enabling EXIT_ON_FAILURE"
            ;;
        all)
            unset skip_main_deployment_methods
            unset skip_tidb_deployment_methods
            unset skip_skip_start_deployment
            unset skip_pre_post_operations
            unset skip_semisync_operations
            unset skip_group_operations
            unset skip_dd_operations
            unset skip_upgrade_operations
            unset skip_multi_source_operations
            unset skip_pxc_operations
            unset skip_ndb_operations
            unset no_tests
            echo "# Enabling all tests"
            ;;
        tidb)
            unset skip_tidb_deployment_methods
            unset no_tests
            echo "# Enabling tidb tests"
            ;;
        main)
            unset skip_main_deployment_methods
            unset no_tests
            echo "# Enabling main tests"
            ;;
        skip)
            unset skip_skip_start_deployment
            unset no_tests
            echo "# Enabling skip-start tests"
            ;;
        semi)
            unset skip_semisync_operations
            unset no_tests
            echo "# Enabling semi_sync tests"
            ;;
        upgrade)
            unset skip_upgrade_operations
            unset no_tests
            echo "# Enabling upgrade tests"
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
        dd)
            unset skip_dd_operations
            unset no_tests
            echo "# Enabling dd operations tests"
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
        pxc)
            unset skip_pxc_operations
            unset no_tests
            echo "# Enabling PXC operations tests"
            ;;
        ndb)
            unset skip_ndb_operations
            unset no_tests
            echo "# Enabling NDB operations tests"
            ;;
        *)
            echo "Allowed tests (you can choose more than one):"
            echo "  main     : main deployment methods"
            echo "  tidb     : tidb deployment methods"
            echo "  skip     : skip-start deployments"
            echo "  pre/post : pre/post grants operations"
            echo "  semi     : semisync operations"
            echo "  group    : group replication operations "
            echo "  dd       : data dictionary operations "
            echo "  upgrade  : upgrade operations "
            echo "  multi    : multi-source operations (fan-in, all-masters)"
            echo "  pxc      : PXC operations"
            echo "  ndb      : NDB operations"
            echo "  all      : enable all the above tests"
            echo ""
            echo "Allowed modifiers:"
            echo "  concurrent  : Enable concurrent operations"
            echo "  sequential  : Disable concurrent operations"
            echo "  interactive : Enable interaction with user"
            echo "  exitfail    : Enable exit on failure"
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
    echo "Directory (\$BINARY_DIR) '$BINARY_DIR' not found"
    exit 1
fi

if [ ! -d "$SANDBOX_HOME" ]
then
    mkdir "$SANDBOX_HOME"
fi
if [ ! -d "$SANDBOX_HOME" ]
then
    echo "Directory (\$SANDBOX_HOME) '$SANDBOX_HOME' could not be created"
    exit 1
fi

function test_ports {
    running_version=$1
    dir_name=$2
    expected_ports=$3
    nodes=$4
    bare_version=$(echo $running_version | sed -e 's/^[^0-9]*//')
    major=$(echo $bare_version | tr '.' ' ' | awk '{print $1}')
    minor=$(echo $bare_version | tr '.' ' ' | awk '{print $2}')
    rev=$(echo $bare_version | tr '.' ' ' | awk '{print $3}')
    if [[ $major -eq 8 && $minor -eq 0  && $rev -ge 11 ]]
    then
        expected_ports=$((expected_ports+nodes))
    fi
    test_header test_ports "$dir_name $running_version"
    how_many_ports=$(sandbox_num_ports $running_version $dir_name)
    ok_equal "Ports in $dir_name $running_version" $how_many_ports $expected_ports
    check_for_exit test_ports
}


function test_slave_hosts {
    running_version=$1
    dir_name=$2
    expected_nodes=$3
    version_path=$(echo $running_version| tr '.' '_')
    test_header test_slave_hosts "$dir_name$version_path"
    sbdir=$SANDBOX_HOME/$dir_name$version_path
    found_nodes=$($sbdir/m -BN -e 'show slave hosts' | wc -l | tr -d ' ')
    ok_equal "slave hosts" $found_nodes $expected_nodes
    check_for_exit test_slave_hosts skip_log_check
}

function test_use_masters_slaves {
    running_version=$1
    dir_name=$2
    expected_masters=$3
    expected_slaves=$4
    version_path=$(echo $running_version| tr '.' '_')
    test_header test_use_masters_slaves "$dir_name$version_path"
    sbdir=$SANDBOX_HOME/$dir_name$version_path
    found_masters=$($sbdir/use_all_masters 'select @@server_id' | grep -c '^[0-9]\+$')
    found_slaves=$($sbdir/use_all_slaves 'select @@server_id' | grep -c '^[0-9]\+$')
    ok_equal "master hosts" $found_masters $expected_masters
    ok_equal "slave hosts" $found_slaves $expected_slaves
    check_for_exit test_use_masters_slaves skip_log_check
}

function test_start_restart {
    running_version=$1
    dir_name=$2
    mode=$3
    test_header test_start_restart "$running_version $mode"
    use_name=use
    start_name=start
    stop_name=stop
    use_name2=""
    if [ "$mode" == "multiple" ]
    then
        use_name="n1"
        use_name2="n2"
        start_name=start_all
        stop_name=stop_all
    fi
    version_path=$(echo $running_version| tr '.' '_')
    sbdir=$SANDBOX_HOME/$dir_name$version_path
    before_connections=$($sbdir/$use_name -BN -e 'select @@max_connections' | tr -d ' ' )
    new_connections=66
    ok_not_equal "Initial max connections" "$before_connections" $new_connections
    #(set -x
    #find $SANDBOX_HOME -name '*.pid' -exec cat {} \;
    #pgrep mysqld
    #)
    $sbdir/$stop_name
    $sbdir/$start_name --max-connections=$new_connections
    after_connections=$($sbdir/$use_name -BN -e 'select @@max_connections' | tr -d ' ' )
    ok_equal "start: changed max connections" "$after_connections" $new_connections
    if [ -n "$use_name2" ]
    then
        after_connections=$($sbdir/$use_name2 -BN -e 'select @@max_connections' | tr -d ' ' )
        ok_equal "start: changed max connections (node2)" "$after_connections" $new_connections
    fi
    new_connections=$before_connections
    $sbdir/re$start_name --max-connections=$new_connections
    after_connections=$($sbdir/$use_name -BN -e 'select @@max_connections' | tr -d ' ' )
    ok_equal "restart: changed max connections" "$after_connections" $new_connections
    if [ -n "$use_name2" ]
    then
        after_connections=$($sbdir/$use_name2 -BN -e 'select @@max_connections' | tr -d ' ' )
        ok_equal "restart: changed max connections (node2)" "$after_connections" $new_connections
    fi
    if [ "$mode" == "single" ]
    then
        new_connections=88
        $sbdir/add_option max-connections=$new_connections
        after_connections=$($sbdir/$use_name -BN -e 'select @@max_connections' | tr -d ' ' )
        ok_equal "add_options: changed max connections" "$after_connections" $new_connections
        new_option=$(grep "max-connections=$new_connections" $sbdir/my.sandbox.cnf)
        ok_equal "add_options: added line to my.sandbox.cnf" "$new_option" "max-connections=$new_connections"
    fi
    #(set -x
    #find $SANDBOX_HOME -name '*.pid' -exec cat {} \;
    #pgrep mysqld
    #)
    check_for_exit test_start_restart skip_log_check
}


function test_gtid {
    running_version=$1
    group_dir_name=$2
    if [ -z "$GTID_TEST" ]
    then
        return
    fi
    test_header test_gtid $running_version
    version_path=$(echo $running_version| tr '.' '_')
    sbdir=$SANDBOX_HOME/$group_dir_name$version_path
    gtid_slave1=$($sbdir/s1 -B -e 'show slave status\G' | grep '1111-1111-1111' )
    gtid_slave2=$($sbdir/s1 -B -e 'show slave status\G' | grep 'Auto_Position: 1' )
    gtid_master=$($sbdir/m -B -e 'show master status\G' | grep '1111-1111-1111' )
    $sbdir/test_replication
    ok "master GTID is enabled" "$gtid_master"
    ok "slave GTID is enabled" "$gtid_slave1"
    ok "slave uses auto position" "$gtid_slave2"
    check_for_exit test_gtid skip_log_check
}

function test_semi_sync {
    running_version=$1
    group_dir_name=$2
    test_header test_semi_sync $running_version
    version_path=$(echo $running_version| tr '.' '_')
    sbdir=$SANDBOX_HOME/$group_dir_name$version_path
    master_enabled=$($sbdir/m -BN -e 'select @@rpl_semi_sync_master_enabled' | tr -d ' ' )
    master_yes_trx_before=$($sbdir/m -BN -e 'show global status like "Rpl_semi_sync_master_yes_tx"' | awk '{print $2}' )
    master_no_trx_before=$($sbdir/m -BN -e 'show global status like "Rpl_semi_sync_master_no_tx"' | awk '{print $2}' )
    slave1_enabled=$($sbdir/s1 -BN -e 'select @@rpl_semi_sync_slave_enabled' | tr -d ' ' )
    slave2_enabled=$($sbdir/s2 -BN -e 'select @@rpl_semi_sync_slave_enabled' | tr -d ' ' )
    ok_equal "Master semisync enabled" "$master_enabled" 1 -1
    ok_equal "Slave 1 semisync enabled" "$slave1_enabled" 1 -1
    ok_equal "Slave 2 semisync enabled" "$slave2_enabled" 1 -1
    $sbdir/test_replication
    master_yes_trx_after=$($sbdir/m -BN -e 'show global status like "Rpl_semi_sync_master_yes_tx"' | awk '{print $2}' )
    master_no_trx_after=$($sbdir/m -BN -e 'show global status like "Rpl_semi_sync_master_no_tx"' | awk '{print $2}' )
    ok_equal "Same number of async trx" $master_no_trx_before $master_no_trx_after
    ok_greater "Bigger number of sync trx" $master_yes_trx_after $master_yes_trx_before
    check_for_exit test_semi_sync skip_log_check
}

function test_expose_dd {
    running_version=$1
    dir_name=msb_
    version_path=$(echo $running_version| tr '.' '_')
    test_header test_expose_dd "${dir_name}$version_path"
    sandbox_dir=$dir_name$version_path
    # run dbdeployer deploy single $running_version --expose-dd-tables

    capture_test $SANDBOX_HOME/$sandbox_dir/test_sb
    using_debug1=$($SANDBOX_HOME/$sandbox_dir/use -BN -e "select version() REGEXP 'debug'" )
    using_debug2=$($SANDBOX_HOME/$sandbox_dir/use -BN -e "select @@debug is not null" )
    tables_found1=$($SANDBOX_HOME/$sandbox_dir/use -BN -e "select count(*) from information_schema.tables where table_name ='tables' and table_schema='mysql'" )
    tables_found2=$($SANDBOX_HOME/$sandbox_dir/use -BN -e "select count(*) from mysql.tables where name ='tables' and schema_id=1" )
    ok_equal "using debug" $using_debug1 1
    ok_equal "debug variable not null" $using_debug2 1
    ok_equal "table 'tables' found in information_schema " $tables_found1 1
    ok_equal "table 'tables' found in mysql " $tables_found2 1
    check_for_exit test_expose_dd
    #run dbdeployer delete ${dir_name}${version_path}
}


function test_force_single {
    running_version=$1
    dir_name=msb_
    version_path=$(echo $running_version| tr '.' '_')
    test_header test_force_single "${dir_name}$version_path"
    sandbox_dir=$dir_name$version_path
    capture_test $SANDBOX_HOME/$sandbox_dir/test_sb
    port_before=$($SANDBOX_HOME/$sandbox_dir/use -BN -e 'show variables like "port"' | awk '{print $2}')
    run dbdeployer deploy single $CUSTOM_OPTIONS $running_version --force
    port_after=$($SANDBOX_HOME/$sandbox_dir/use -BN -e 'show variables like "port"' | awk '{print $2}')
    ok_equal "Port before and after --force redeployment" $port_after $port_before
    check_for_exit test_force_single
}

function test_force_replication {
    running_version=$1
    dir_name=rsandbox_
    version_path=$(echo $running_version| tr '.' '_')
    test_header test_force_replication "${dir_name}$version_path"
    sandbox_dir=$dir_name$version_path
    capture_test $SANDBOX_HOME/$sandbox_dir/test_sb_all
    port_before=$($SANDBOX_HOME/$sandbox_dir/m -BN -e 'show variables like "port"' | awk '{print $2}')
    run dbdeployer deploy $CUSTOM_OPTIONS replication $running_version --force
    port_after=$($SANDBOX_HOME/$sandbox_dir/m -BN -e 'show variables like "port"' | awk '{print $2}')
    ok_equal "Port before and after --force redeployment" $port_after $port_before
    check_for_exit test_force_replication
}

function test_custom_credentials {
    running_version=$1
    mode=$2
    dir_name=$3
    version_path=$(echo $running_version| tr '.' '_')
    test_header test_custom_credentials "${mode} $version_path"
    sandbox_dir=$dir_name$version_path
    # testing correctness before redeploying
    test_sb=test_sb
    test_replication=""
    my_cnf=my.sandbox.cnf
    grants_file=grants.mysql
    if [ "$mode" != "single" ]
    then
        test_sb=test_sb_all
        test_replication=test_replication
        my_cnf=node1/my.sandbox.cnf
        grants_file=node1/grants.mysql
    fi
    capture_test run $SANDBOX_HOME/$sandbox_dir/$test_sb
    if [ -n "$test_replication" ]
    then
        capture_test run $SANDBOX_HOME/$sandbox_dir/$test_replication
    fi
    new_db_user=different
    new_db_password=anotherthing
    new_repl_user=different_rpl
    new_repl_password=anotherthing_rpl
    run dbdeployer deploy $mode $running_version \
        --db-user=$new_db_user --db-password=$new_db_password \
        --force  $CUSTOM_OPTIONS \
        --rpl-user=$new_repl_user --rpl-password=$new_repl_password
    # This deployment will be re-tested later together with the rest of the sandboxes
    user_found=$(grep $new_db_user $SANDBOX_HOME/$sandbox_dir/$my_cnf) 
    password_found=$(grep $new_db_password $SANDBOX_HOME/$sandbox_dir/$my_cnf) 
    repl_user_found=$(grep $new_repl_user $SANDBOX_HOME/$sandbox_dir/$grants_file) 
    repl_password_found=$(grep $new_repl_password $SANDBOX_HOME/$sandbox_dir/$grants_file) 
    ok "custom user found" "$user_found"
    ok "custom password found" "$password_found"
    ok "custom replication user found" "$repl_user_found"
    ok "custom replication password found" "$repl_password_found"
    sleep 1
    check_for_exit test_custom_credentials
}
 
function test_uuid {
    running_version=$1
    group_dir_name=$2
    must_exist=$3
    version_path=$(echo $running_version| tr '.' '_')
    test_header test_uuid "${group_dir_name}${version_path}"
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
    check_for_exit test_uuid
}

function test_deletion {
    del_version=$1
    expected_items=$2
    processes_before=$3
    process_name=$4
    if [ -z "$process_name" ]
    then
        process_name=mysqld
    fi
    test_header test_deletion $del_version

    # test lock: sandboxes become locked against deletion
    num_sandboxes_before=$(dbdeployer sandboxes | wc -l)
    run dbdeployer admin lock ALL
    run dbdeployer delete ALL --skip-confirm
    # deletion of locked sandboxes should be ineffective.
    # We expect to get the same number of sandboxes before and after the deletion
    num_sandboxes_after=$(dbdeployer sandboxes | wc -l)
    ok_equal "num_sandboxes" $num_sandboxes_before $num_sandboxes_after

    if [ -n "$DBDEPLOYER_LOGGING" ]
    then
        num_log_dirs=$(ls $HOME/sandboxes/logs | wc -l)
        ok_equal "num_logs_before" $num_sandboxes_after $num_log_dirs
    fi

    how_many=$(count_catalog)
    ok_equal "sandboxes_in_catalog" $how_many $expected_items

    run dbdeployer admin unlock ALL
    check_for_exit test_deletion skip_log_check
    run dbdeployer delete ALL --skip-confirm
    results "#$del_version - after deletion"
    num_sandboxes_final=$(dbdeployer sandboxes --catalog | wc -l)
    # After unlocking, deletion must work, and we should see that
    # there are no sandboxes left
    ok_equal "num_sandboxes" $num_sandboxes_final 0

    if [ -n "$DBDEPLOYER_LOGGING" ]
    then
        num_log_dirs=$(ls $HOME/sandboxes/logs | wc -l)
        ok_equal "num_logs_after" $num_log_dirs 0
    fi
    how_many=$(count_catalog)
    ok_equal "sandboxes_in_catalog" $how_many 0
    processes_after=$(pgrep $process_name | wc -l | tr -d ' \t')
    ok_equal 'no more '$process_name' processes after deletion' $processes_after $processes_before
    if [ "$fail" != "0" ]
    then
        echo "# detected failures: $fail"
        exit 1
    fi
}

function capture_test {
    cmd="$*"
    output=/tmp/capture_test$$
    # echo "# cmd: <$cmd>"
    $cmd > $output 2>&1
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo $dash_line
        echo "# command    : $cmd"
        echo "# output file: $output"
        echo $dash_line
        cat $output
        echo $dash_line
        rm -f $output
        exit 1
    fi
    tmp_pass=$(grep -c '^ok' $output)
    pass=$((pass+tmp_pass))
    tmp_fail=$(grep -c -i '^not ok' $output)
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
    dbdeployer sandboxes
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
    echo "Check the file ${CATALOG}: it should be empty."
    exit 1
fi

# Finding the latest release of every major version

[ -z "$short_versions" ] && short_versions=(4.1 5.0 5.1 5.5 5.6 5.7 8.0)

[ -z "$group_short_versions" ] && group_short_versions=(5.7 8.0)
[ -z "$dd_short_versions" ] && dd_short_versions=(8.0)
[ -z "$semisync_short_versions" ] && semisync_short_versions=(5.5 5.6 5.7 8.0)
[ -z "$pxc_short_versions" ] && pxc_short_versions=(pxc5.7)
[ -z "$ndb_short_versions" ] && ndb_short_versions=(ndb7.6 ndb8.0)
count=0
all_versions=()
tidb_versions=(tidb3.0.0)
group_versions=()
semisync_versions=()
dd_versions=()
pxc_versions=()
ndb_versions=()

OS=$(uname | tr '[:upper:]' '[:lower:]')
if [ -x "sort_versions.$OS" ]
then
    cp "sort_versions.$OS" sort_versions
fi

if [ ! -x ./sort_versions ]
then
    #if [ -f ./sort_versions.go ]
    #then
    #    env GOOS=linux GOARCH=386 go build -o sort_versions.linux sort_versions.go
    #    env GOOS=darwin GOARCH=386 go build -o sort_versions.Darwin sort_versions.go
    #    ls -l sort_versions*
    #    cp "sort_versions.$OS" sort_versions
    #fi
    #if [ ! -x ./sort_versions ]
    #then
        echo "./sort_versions not found"
        exit 1
    #fi
fi

for v in ${short_versions[*]}
do
    #ls $BINARY_DIR | grep "^$v" | ./sort_versions | tail -n 1
    latest=$(ls "$BINARY_DIR" | grep "^$v" | ./sort_versions | tail -n 1)
    oldest=$(ls "$BINARY_DIR" | grep "^$v" | ./sort_versions | head -n 1)
    if [ -n "$latest" ]
    then
        all_versions[$count]=$latest
        count=$((count+1))
    else
        echo "No versions found for $v"
    fi
    if [ -n "$oldest" ]
    then
        if [ "$oldest" != "$latest" -a "$v" == "5.0" ]
        then
            all_versions[$count]=$oldest
            count=$((count+1))
        fi
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

count=0
for v in ${semisync_short_versions[*]}
do
    latest=$(ls $BINARY_DIR | grep "^$v" | ./sort_versions | tail -n 1)
    if [ -n "$latest" ]
    then
        semisync_versions[$count]=$latest
        count=$((count+1))
    fi
done

count=0
for v in ${dd_short_versions[*]}
do
    latest=$(ls $BINARY_DIR | grep "^$v" | ./sort_versions | tail -n 1)
    if [ -n "$latest" ]
    then
        dd_versions[$count]=$latest
        count=$((count+1))
    fi
done

count=0
for v in ${pxc_short_versions[*]}
do
    latest=$(ls $BINARY_DIR | grep "^$v" | ./sort_versions | tail -n 1)
    if [ -n "$latest" ]
    then
        pxc_versions[$count]=$latest
        count=$((count+1))
    fi
done
count=0
for v in ${ndb_short_versions[*]}
do
    latest=$(ls $BINARY_DIR | grep "^$v" | ./sort_versions | tail -n 1)
    if [ -n "$latest" ]
    then
        ndb_versions[$count]=$latest
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

how_many=$(count_catalog)
ok_equal "sandboxes_in_catalog" $how_many 0

function main_deployment_methods {
    current_test=main_deployment_methods
    test_header main_deployment_methods "" double
    processes_before=$(pgrep mysqld | wc -l | tr -d ' \t')
    for V in ${all_versions[*]}
    do
        # We test the main deployment methods
        echo $dotted_line
        for stype in single multiple replication
        do
            run dbdeployer deploy $stype $CUSTOM_OPTIONS $V 
        done
        echo $dotted_line
        # Runs the post installation check.
        for stype in single multiple replication
        do
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
        test_force_single $V
        test_force_replication $V 
        test_custom_credentials $V single msb_
        test_custom_credentials $V replication rsandbox_
        test_uuid $V multi_msb_
        test_uuid $V rsandbox_
        test_completeness $V msb_ single
        test_completeness $V rsandbox_ replication
        test_completeness $V multi_msb_ multiple
        test_ports $V msb_ 1 1
        test_ports $V rsandbox_ 3 3
        test_ports $V multi_msb_ 3 3
        test_slave_hosts $V rsandbox_ 2
        test_use_masters_slaves $V rsandbox_ 1 2
        capture_test run dbdeployer global test
        capture_test run dbdeployer global test-replication
        test_start_restart $V msb_ single
        test_start_restart $V rsandbox_ multiple
        test_start_restart $V multi_msb_ multiple
        echo "# processes $processes_before"
        test_deletion $V 3 $processes_before
    done
}

function tidb_deployment_methods {
    current_test=tidb_deployment_methods
    test_header tidb_deployment_methods "" double
    version_count=0
    for V in ${tidb_versions[*]}
    do
        if [ -d $SANDBOX_BINARY/$V ]
        then
            version_count=$((version_count+1))
        else
            echo "version $SANDBOX_BINARY/$V not found"
        fi
    done
    if [ $version_count == 0 ]
    then
        echo "no versions found for tidb"
        return
    fi
    latest_5_7=$(ls "$BINARY_DIR" | grep "^5.7" | ./sort_versions | tail -n 1)
    if [ -z "$latest_5_7" ]
    then
        echo "TiDb tests invoked, but no 5.7 client binaries found - aborting"
        exit 1
    fi
    save_custom_options=$CUSTOM_OPTIONS
    CUSTOM_OPTIONS="$CUSTOM_OPTIONS --client-from=$latest_5_7 "
    processes_before=$(pgrep tidb-server | wc -l | tr -d ' \t')
    for V in ${tidb_versions[*]}
    do
        # We test the main deployment methods
        # and also the ability of dbdeployer to handle
        # operations while similar calls occur
        # in the background
        echo $dotted_line
        for stype in single multiple
        do
            echo "# Parallel deployment: $stype $V"
            install_out="/tmp/${stype}-${V}-$$"
            run dbdeployer deploy $stype $CUSTOM_OPTIONS $V > $install_out 2>&1 &
        done
        echo $dotted_line
        # wait for the installation processes to finish
        wait
        # Display the result of each installation
        for stype in single multiple
        do
            install_out="/tmp/${stype}-${V}-$$"
            cat $install_out
            echo $dotted_line
            rm $install_out
        done
        # Runs the post installation check.
        for stype in single multiple
        do
            # For each type, we display basic info
            results "$stype"
        done
        how_many=$(count_catalog)
        ok_equal "sandboxes_in_catalog" $how_many 2
        sleep 2
        # Runs basic tests
        run dbdeployer global status
        test_force_single $V
        test_custom_credentials $V single msb_
        test_completeness $V msb_ single
        test_completeness $V multi_msb_ multiple
        test_ports $V msb_ 1 1
        test_ports $V multi_msb_ 3 2
        capture_test run dbdeployer global test
        echo "# processes $processes_before"
        test_deletion $V 2 $processes_before tidb-server
    done
    CUSTOM_OPTIONS=$save_custom_options
}



function global_status_count_on {
    dbdeployer global status | grep -c 'on  -\|on$'
}

function global_status_count_off {
    dbdeployer global status | grep -c -w off
}

function check_on_off_status {
    expected_on=$1
    expected_off=$2
    status_count_off=$(global_status_count_off)
    status_count_on=$(global_status_count_on)
    ok_equal "idle sandboxes $V" $status_count_off $expected_off
    ok_equal "started sandboxes $V" $status_count_on $expected_on
    check_for_exit check_on_off_status
}

function skip_start_deployment {
    current_test=skip_start_deployment
    test_header skip_start_deployment "" double
    latest=latest$$
    for V in ${all_versions[*]}
    do
        for stype in single multiple replication
        do
            if [ -L $SANDBOX_BINARY/$latest ]
            then
                rm -f $SANDBOX_BINARY/$latest
            fi
            # In addition to skip-start, this test also checks that we 
            # can create sandboxes from basedir name not ending with
            # the version number  x.x.xx
            ln -s $SANDBOX_BINARY/$V $SANDBOX_BINARY/$latest
            run dbdeployer deploy $stype $latest --skip-start --binary-version=$V
        done
        # all sandboxes OFF
        check_on_off_status 0 7
        version_path=$(echo $V| tr '.' '_')
        singledir=msb_$latest
        multidir=multi_msb_$latest
        repldir=rsandbox_$latest

        # One sandbox ON
        $SANDBOX_HOME/$singledir/start
        $SANDBOX_HOME/$singledir/load_grants
        check_on_off_status 1 6
        # Check that the manually started sandbox behaves as expected
        capture_test $SANDBOX_HOME/$singledir/test_sb

        # Three more sandboxes ON
        $SANDBOX_HOME/$repldir/start_all
        $SANDBOX_HOME/$repldir/master/load_grants
        $SANDBOX_HOME/$repldir/initialize_slaves
        check_on_off_status 4 3

        # Three more sandboxes ON
        $SANDBOX_HOME/$multidir/start_all
        $SANDBOX_HOME/$multidir/node1/load_grants
        $SANDBOX_HOME/$multidir/node2/load_grants
        $SANDBOX_HOME/$multidir/node3/load_grants
        check_on_off_status 7 0

        # Check that the manually started replication sandbox behaves as expected
        capture_test $SANDBOX_HOME/$repldir/test_sb_all
        capture_test $SANDBOX_HOME/$repldir/test_replication
        capture_test $SANDBOX_HOME/$multidir/test_sb_all
        check_for_exit skip_start_deployment
        dbdeployer delete all --skip-confirm
        rm -f $SANDBOX_BINARY/$latest
    done

    for V in ${group_versions[*]}
    do
        run dbdeployer deploy replication $V --skip-start --topology=group
        run dbdeployer deploy replication $V --skip-start --topology=group --single-primary
        run dbdeployer deploy replication $V --skip-start --topology=fan-in
        run dbdeployer deploy replication $V --skip-start --topology=all-masters
        check_on_off_status 0 12
        check_for_exit skip_start_deployment2
        dbdeployer delete all --skip-confirm
    done
}

function pre_post_operations {
    test_header pre_post_operations "" double
    current_test=pre_post_operations
    # This test checks the following:
    #   * we can run a SQL command before the grants are loaded
    #   * we match the result of that query with an expected value
    #   * we can run more than one query before and after
    #   * we can compare values before and after the grants were loaded
    #   * we can run queries containing commas without errors.
    for V in ${all_versions[*]}
    do
        echo "#pre-post operations $V"
        outfile=/tmp/pre-post-$V.txt
        is41=$(echo $V | grep '^4\.1')
        if [ -n "$is41" ]
        then
            (set -x
            dbdeployer deploy single $V \
                --pre-grants-sql="select 'preversion' as label, version()" \
                --pre-grants-sql="select 'preusers' as label, count(*) as PRE from mysql.user" \
                --post-grants-sql="select 'postversion' as label, version()" \
                --post-grants-sql="select 'postusers' as label, count(*) as POST from mysql.user" > $outfile 2>&1
            )
        else
            (set -x
            dbdeployer deploy single $V \
                --pre-grants-sql="select 'preversion' as label, version()" \
                --pre-grants-sql="select 'preschema' as label, count(*) as PRE from information_schema.schemata" \
                --pre-grants-sql="select 'preusers' as label, count(*) as PRE from mysql.user" \
                --post-grants-sql="select 'postversion' as label, version()" \
                --post-grants-sql="select 'postschema' as label, count(*) as POST from information_schema.schemata" \
                --post-grants-sql="select 'postusers' as label, count(*) as POST from mysql.user" > $outfile 2>&1
            )
            pre_schema=$(grep preschema $outfile | awk '{print $4}')
            post_schema=$(grep postschema $outfile | awk '{print $4}')
        fi
        # Gets the line with a given label.
        # retrieves the fourth element in the line
        pre_users=$(grep preusers $outfile | awk '{print $4}')
        post_users=$(grep postusers $outfile | awk '{print $4}')
        pre_version=$(grep preversion $outfile | awk '{print $4}')
        post_version=$(grep postversion $outfile | awk '{print $4}')
        # cat $outfile
        ok_greater "post grants users more than pre" $post_users $pre_users
        if [ -z "$is41" ]
        then
            ok_greater_equal "same or more schemas before and after grants" $post_schema $pre_schema
        fi
        ok_contains "Version" $pre_version $V
        ok_contains "Version" $post_version $V
        results "pre/post $V"
        rm $outfile
        check_for_exit pre_post_operations
        dbdeployer delete ALL --skip-confirm
    done
}

function semisync_operations {
    current_test=semisync_operations
    test_header semisync_operations "" double
    for V in ${semisync_versions[*]}
    do
        echo "# semisync operations $V"
        is_5_5=$(echo $V | grep '^5.5')
        GTID_TEST=""
        if [ -z "$is_5_5" ]
        then
            export GTID_TEST=--gtid
        fi
        run dbdeployer deploy replication $V --semi-sync $GTID_TEST
        results "semisync $V"
        #sleep 2
        capture_test run dbdeployer global test
        test_semi_sync $V rsandbox_
        test_gtid $V rsandbox_
        unset GTID_TEST
        check_for_exit semisync_operations skip_log_check
        dbdeployer delete ALL --skip-confirm
        results "semisync $V - after deletion"
    done
}

function dd_operations {
    current_test=dd_operations
    test_header dd_operations "" double
    for V in ${dd_versions[*]}
    do
        if [ -x $SANDBOX_BINARY/$V/bin/mysqld-debug ]
        then
            echo "# data dictionary operations $V"
            run dbdeployer deploy single $V --expose-dd-tables --disable-mysqlx
            results "dd $V"
            capture_test run dbdeployer global test
            test_expose_dd $V msb_
            dbdeployer delete ALL --skip-confirm
            results "dd $V - after deletion"
        else
            echo "Skipping dd operations for ${V}: no mysqld-debug found in ./bin"
        fi
    done
}

function upgrade_operations {
    current_test=upgrade_operations
    test_header upgrade_operations "" double
    latest_5_6=$(ls "$BINARY_DIR" | grep "^5.6" | ./sort_versions | tail -n 1)
    latest_5_7=$(ls "$BINARY_DIR" | grep "^5.7" | ./sort_versions | tail -n 1)
    latest_8_0=$(ls "$BINARY_DIR" | grep "^8.0" | ./sort_versions | tail -n 1)
    if [ -z "$latest_5_7" -o -z "$latest_8_0" ]
    then
        echo "Skipping upgrade test. No suitable version found for 5.7 or 8.0"
        return
    fi
    echo "# upgrade operations $latest_5_7 to $latest_8_0"

    v_path_5_6=$(echo msb_$latest_5_6| tr '.' '_')
    v_path_5_7=$(echo msb_$latest_5_7| tr '.' '_')
    v_path_8_0=$(echo msb_$latest_8_0| tr '.' '_')
    upgrade_from=$v_path_5_7
    upgrade_to=$v_path_8_0
    # run dbdeployer deploy single $latest_5_6
    run dbdeployer deploy single $latest_5_7
    run dbdeployer deploy single $latest_8_0
    results "upgrade $upgrade_from to $upgrade_to"
    capture_test run dbdeployer global test
    echo "dbdeployer admin upgrade $upgrade_from $upgrade_to"
    run dbdeployer admin upgrade $upgrade_from $upgrade_to 
    if [ -f $SANDBOX_HOME/$upgrade_to/no_upgrade ]
    then
        echo "SKIPPING upgrade test. mysql_upgrade not found in destination"
    else
        ok_dir_exists $SANDBOX_HOME/$upgrade_to/data-${upgrade_to}
        ok_dir_exists $SANDBOX_HOME/$upgrade_to/data
        ok_dir_does_not_exist $SANDBOX_HOME/$upgrade_from/data
    fi
    dbdeployer delete ALL --skip-confirm
    results "upgrade $latest_5_7 to $latest_8_0 - after deletion"
}

function group_operations {
    current_test=group_operations
    test_header group_operations "" double
    processes_before=$(pgrep mysqld | wc -l | tr -d ' \t')
    for V in ${group_versions[*]}
    do
        echo "# Group operations $V"
        mysqld_debug=$SANDBOX_BINARY/$V/bin/mysqld-debug
        plugin_debug=$SANDBOX_BINARY/$V/lib/plugin/debug
        if [ -x $mysqld_debug -a -d $plugin_debug ]
        then
            WITH_DEBUG=--custom-mysqld=mysqld-debug
        else
            WITH_DEBUG=""
        fi
        run dbdeployer deploy replication $V --topology=group
        run dbdeployer deploy replication $V --topology=group \
            --single-primary $WITH_DEBUG
        results "group $V"

        capture_test run dbdeployer global test
        capture_test run dbdeployer global test-replication
        test_uuid $V group_msb_ 1
        test_uuid $V group_sp_msb_ 1
        test_use_masters_slaves $V group_msb_ 3 3
        test_use_masters_slaves $V group_sp_msb_ 1 2
        test_ports $V group_msb_ 6 3
        test_ports $V group_sp_msb_ 6 3
        check_for_exit group_operations
        test_deletion $V 2 $processes_before
        results "group $V - after deletion"
    done
}

function multi_source_operations {
    current_test=multi_source_operations
    test_header multi_source_operations "" double
    processes_before=$(pgrep mysqld | wc -l | tr -d ' \t')
    for V in ${group_versions[*]}
    do
        echo "# Multi-source operations $V"
        v_path=$(echo $V| tr '.' '_')
        run dbdeployer deploy replication $V --topology=fan-in
        run dbdeployer deploy replication $V --topology=fan-in \
            --sandbox-directory=fan_in_msb2_$v_path \
            --base-port=31000 \
            --nodes=4 \
            --master-list='1,2' \
            --slave-list='3:4'
        run dbdeployer deploy replication $V --topology=all-masters
        results "multi-source"

        capture_test run dbdeployer global test
        capture_test run dbdeployer global test-replication
        test_uuid $V fan_in_msb_ 1
        test_uuid $V all_masters_msb_ 1
        test_ports $V fan_in_msb_ 3 3
        test_ports $V all_masters_msb_ 3 3
        test_use_masters_slaves $V fan_in_msb_ 2 1
        test_use_masters_slaves $V all_masters_msb_ 3 3
        check_for_exit multi_source_operations
        test_deletion $V 3 $processes_before
        results "multi-source - after deletion"
    done
}

function pxc_operations {
    current_test=pxc_operations
    test_header pxc_operations "" double
    operating_system=$(uname -s | tr 'A-Z' 'a-z' )
    if [ "$operating_system" != "linux" ]
    then
        echo "Skipping PXC tests on non-Linux system"
        return
    fi
    processes_before=$(pgrep mysqld | wc -l | tr -d ' \t')
    for V in ${pxc_versions[*]}
    do
        echo "# PXC operations $V"
        run dbdeployer deploy replication $V --topology=pxc
        results "PXC $V"

        capture_test run dbdeployer global test
        capture_test run dbdeployer global test-replication
        test_use_masters_slaves $V pxc_msb_ 3 3
        test_ports $V pxc_msb_ 12 3
        check_for_exit pxc_operations
        test_deletion $V 1 $processes_before
        results "pxc $V - after deletion"
    done
}

function ndb_operations {
    current_test=ndb_operations
    test_header ndb_operations "" double
    processes_before=$(pgrep mysqld | wc -l | tr -d ' \t')
    for V in ${ndb_versions[*]}
    do
        echo "# NDB operations $V"
        run dbdeployer deploy replication $V --topology=ndb
        results "NDB $V"

        capture_test run dbdeployer global test
        capture_test run dbdeployer global test-replication
        test_use_masters_slaves $V ndb_msb_ 3 3
        test_ports $V ndb_msb_ 4 3
        check_for_exit ndb_operations
        test_deletion $V 1 $processes_before
        results "pxc $V - after deletion"
    done
}


if [ -z "$skip_main_deployment_methods" ]
then
    main_deployment_methods
fi
if [ -z "$skip_tidb_deployment_methods" ]
then
    tidb_deployment_methods
fi
if [ -z "$skip_skip_start_deployment" ]
then
    skip_start_deployment
fi
if [ -z "$skip_pre_post_operations" ]
then
    pre_post_operations
fi
if [ -z "$skip_semisync_operations" ]
then
    semisync_operations
fi
if [ -z "$skip_group_operations" ]
then
    group_operations
fi
if [ -z "$skip_dd_operations" ]
then
    dd_operations
fi
if [ -z "$skip_upgrade_operations" ]
then
    upgrade_operations
fi
if [ -z "$skip_multi_source_operations" ]
then
    multi_source_operations
fi
if [ -z "$skip_pxc_operations" ]
then
    pxc_operations
fi
if [ -z "$skip_ndb_operations" ]
then
    ndb_operations
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
    echo $dash_line
    echo "*** FAILURES DETECTED ***"
    echo $dash_line
    exit_code=1
fi
echo "Exit code: $exit_code"
echo "Exit code: $exit_code" >> "$results_log"
exit $exit_code
