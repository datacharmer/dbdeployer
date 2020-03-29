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

run dbdeployer defaults store
run dbdeployer defaults show

versions=(5.0 5.1 5.5 5.6 5.7 8.0)
versions_mysqld_initialize=(5.7 8.0)
versions_mysql_install_db=(5.0 5.1 5.5 5.6)

rev_sparse="18 27 36 45 54 63 72 81 99"
rev_all="$(seq 1 99)"
rev_list=$rev_all
if [ "$1" == "sparse" ]
then
    rev_list=$rev_sparse
fi

function test_number_of_ports {
    version=$1
    dir_name=$2
    mode=$3
    nodes=$4
    expected_ports=$5
    version_name=$(echo $version | tr '.' '_')
    deploy_command=$mode
    case $mode in 
        groupr)
            deploy_command="replication --topology=group"
            ;;
        groupsp)
            deploy_command="replication --topology=group --single-primary"
            ;;
        allmasters)
            deploy_command="replication --topology=all-masters"
            ;;
        fanin)
            deploy_command="replication --topology=fan-in"
            ;;
    esac
    run dbdeployer deploy $deploy_command $version --disable-mysqlx
    how_many_ports=$(sandbox_num_ports $version $dir_name)
    ok_equal "ports in $dir_name $version (without mysqlx)" $how_many_ports $expected_ports
    run dbdeployer delete $dir_name$version_name
    expected_ports=$((expected_ports+nodes))
    run dbdeployer deploy $deploy_command $version
    how_many_ports=$(sandbox_num_ports $version $dir_name)
    ok_equal "ports in $dir_name $version (with mysqlx)" $how_many_ports $expected_ports
    run dbdeployer delete $dir_name$version_name
}

for rev in $rev_list
do
    for vers in ${versions[*]}
    do
        version=${vers}.${rev}
        create_mock_tarball $version $SANDBOX_TARBALL
        #create_mock_version $version
        run dbdeployer unpack $SANDBOX_TARBALL/mysql-${version}.tar.gz
        # --unpack-version $version
    done

    run dbdeployer available
    for vers in ${versions_mysql_install_db[*]}
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy single $version
        run dbdeployer deploy multiple $version
        run dbdeployer deploy replication $version
        results "$version"
        right_installer=$(grep mysql_install_db $SANDBOX_HOME/msb_${version_name}/init_db)
        if [ -n "$right_installer" ]
        then
            echo "ok installer"
        else
            echo "NOT OK installer"
            exit_code=1
        fi
    done

    for vers in ${versions_mysqld_initialize[*]}
    do
        version=${vers}.${rev}
        if [[ $vers == "8.0" && $rev -ge 11 ]]
        then
            test_number_of_ports $version msb_ single 1 1
            test_number_of_ports $version rsandbox_ replication 3 3
            test_number_of_ports $version multi_msb_ multiple 3 3
            test_number_of_ports $version group_msb_ groupr 3 6
            test_number_of_ports $version group_sp_msb_ groupsp 3 6
            test_number_of_ports $version all_masters_msb_ allmasters 3 3
            test_number_of_ports $version fan_in_msb_ fanin 3 3
        fi
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy single $version
        run dbdeployer deploy multiple $version
        run dbdeployer deploy replication $version
        if [[ "$vers" ==  "5.7" && $rev -lt 18 ]]
        then
            echo "skipping group replication for version $version"
        else
            run dbdeployer deploy replication $version --topology=group
            run dbdeployer deploy replication $version --topology=group --single-primary
            run dbdeployer deploy replication $version --topology=all-masters
            run dbdeployer deploy replication $version --topology=fan-in
            run dbdeployer deploy replication $version --topology=fan-in \
                 --sandbox-directory=fan_in_msb2_$version_name \
                 --base-port=24000 \
                 --nodes=5 \
                 --master-list='1,2' \
                 --slave-list='3:4:5'
            run dbdeployer deploy replication $version --topology=fan-in \
                 --sandbox-directory=fan_in_msb3_$version_name \
                 --base-port=35000 \
                 --nodes=5 \
                 --master-list='1.2.3' \
                 --slave-list='4,5'
        fi
        results "$version"
        right_installer1=$(grep mysqld  $SANDBOX_HOME/msb_${version_name}/init_db )
        right_installer2=$(grep initialize-insecure  $SANDBOX_HOME/msb_${version_name}/init_db )
        if [ -n "$right_installer1" -a -n "$right_installer2" ]
        then
            echo "ok installer"
        else
            echo "NOT OK installer"
            exit_code=1
        fi
    done

    if [ "$exit_code" != "0" ]
    then
        exit $exit_code
    fi
done

echo "#Total sandboxes: $(count_catalog)"
#echo "#Total sandboxes: $(count_catalog)" >> $results_log
num_ports=$(grep -A10 port $CATALOG | grep '^\s*[0-9]\+' | wc -l)
echo "# Total ports installed: $num_ports"
#echo "# Total ports installed: $num_ports" >> $results_log
run dbdeployer delete ALL --skip-confirm

results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

