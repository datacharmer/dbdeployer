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

#unset DBDEPLOYER_LOGGING
test_dir=$(dirname $0)
cd $test_dir || (echo "error changing directory to $test_dir"; exit 1)
test_dir=$PWD
#exit_code=0

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

function check_deployment_message {
    version=$1
    shift
    args="$*"
    path_version=$(echo $version | tr '.' '_')
    output_file=deployment$$.txt
    dbdeployer deploy $args $version > $output_file 2>&1
    grep "installed.*\$HOME.*$path_version" $output_file
    installed=$(grep "installed.*\$HOME.*$path_version" $output_file)
    ok "deploy $args $version" "$installed"
    rm $output_file
}

create_mock_version 5.0.66
create_mock_version 5.1.66
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
echo "#!/usr/bin/env bash" > $temp_template
echo "echo 'I AM A CUSTOM_TEMPLATE CREATED ON $timestamp'" >> $temp_template
run dbdeployer deploy --use-template=clear:$temp_template single 8.0.67
sandbox_dir=$SANDBOX_HOME/msb_8_0_67
message=$($sandbox_dir/clear)
ok_contains "custom template" "$message" "CUSTOM_TEMPLATE"
ok_contains "custom template" "$message" $timestamp

run dbdeployer delete ALL --skip-confirm

results "After deletion"

run dbdeployer defaults templates export single $mock_dir/templates clear
cp $temp_template $mock_dir/templates/single/clear
rm -f $temp_template
run dbdeployer defaults templates import single $mock_dir/templates
installed=$(dbdeployer defaults templates list | grep "clear" | grep '{F}')
echo "# installed template: <$installed>"
ok "template was installed" "$installed"
run dbdeployer deploy single 8.0.67

sandbox_dir=$SANDBOX_HOME/msb_8_0_67
message=$($sandbox_dir/clear)
ok_contains "installed custom template" "$message" "CUSTOM_TEMPLATE"
ok_contains "installed custom template" "$message" $timestamp

run dbdeployer delete ALL --skip-confirm

echo "Test installed messages"
check_deployment_message 8.0.67 single
check_deployment_message 8.0.67 multiple
check_deployment_message 8.0.67 replication
check_deployment_message 8.0.67 replication --topology=group
check_deployment_message 8.0.67 replication --topology=group --single-primary
check_deployment_message 8.0.67 replication --topology=fan-in
check_deployment_message 8.0.67 replication --topology=all-masters

run dbdeployer delete ALL --skip-confirm

dbdeployer defaults update reserved-ports '4500,9000,15000,30000'
reserved_ports=$(dbdeployer defaults show| sed -n '/reserved-ports/,/]/p'| grep '^\s*[0-9]'| tr -d ', \t')
how_many=$(echo $reserved_ports | wc -w)
ok_equal "4 reserved ports found" $how_many 4

for port in $reserved_ports 
do
    dbdeployer deploy single 5.7.66 --port=$port
    port_plus_1=$((port+1))
    sandbox_list=$(dbdeployer sandboxes)
    ok_contains "Deployment for port $port contains $port_plus_1" "$sandbox_list" $port_plus_1
    dbdeployer delete msb_5_7_66 
done

function mysqlsh_exists {
    version=$1
    wanted_ex=$2
    option=$3
    path_version=$(echo $version | tr '.' '_')
    client=$SANDBOX_HOME/msb_${path_version}/mysqlsh
    run dbdeployer deploy single $version $option
    if [ "$wanted_ex" == "exists" ]
    then
        ls -l $client
        ok_executable_exists $client
    else
        echo "START TESTING FOR ABSENCE"
        ok_executable_does_not_exist $client
        echo "END TESTING FOR ABSENCE"
    fi

    for log in log binlog relaylog
    do
        viewer=$SANDBOX_HOME/msb_${path_version}/show_$log
        ok_executable_exists $viewer
    done
    run dbdeployer delete msb_${path_version} 
    unset option
}

mysqlsh_exists 5.0.66 does_not ""
mysqlsh_exists 5.1.66 does_not ""
mysqlsh_exists 5.5.66 does_not ""
mysqlsh_exists 5.6.66 does_not ""
mysqlsh_exists 5.7.66 does_not ""
mysqlsh_exists 5.7.66 exists "--enable-mysqlx"
mysqlsh_exists 8.0.66 exists ""

temp_tarball_list=$(cat <<EOF_TARBALL
{
 	"DbdeployerVersion": "1.32.0",
 	"Tarballs": [
 		{
 			"name": "fake-tarball1.tar.gz",
 			"OS": "Darwin",
 			"url": "https://fake.address.org/fake-tarball1.tar.gz",
 			"checksum": "MD5: 7bac88f47e648bf9a38e7886e12d1ec5",
 			"flavor": "mysql",
 			"minimal": false,
 			"size": 26485675,
 			"short_version": "5.0",
 			"version": "5.0.0"
 		},
 		{
 			"name": "fake-tarball2.tar.gz",
 			"OS": "Linux",
 			"checksum": "MD5: 7bac88f47e648bf9a38e7886e12d1ec5",
 			"url": "https://fake.address.org/fake-tarball2.tar.gz",
 			"flavor": "mysql",
 			"minimal": false,
 			"size": 26485675,
 			"short_version": "5.0",
 			"version": "5.0.0"
 		}
   ]
} 

EOF_TARBALL
)

got_mysql_5_0=$( dbdeployer downloads list --OS=all | grep "mysql-5.0.96.tar.xz")
ok "mysql 5.0 found" "$got_mysql_5_0"

temp_tarball=/tmp/temp$$.json
touch $temp_tarball
if [ -f $temp_tarball ]
then

    echo "$temp_tarball_list" > $temp_tarball

    dbdeployer downloads import $temp_tarball
    got_fake_tb1=$(dbdeployer downloads list --OS=all | grep fake-tarball1)
    ok "fake-tarball1 found" "$got_fake_tb1"
    got_fake_tb2=$(dbdeployer downloads list --OS=all | grep fake-tarball2)
    ok "fake-tarball2 found" "$got_fake_tb2"
    got_mysql_5_0=$( dbdeployer downloads list --OS=all | grep "mysql-5.0.96.tar.xz")
    ok_empty "mysql 5.0 empty" "$got_mysql_5_0"
fi

cd $test_dir || (echo "error changing directory to $test_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer
tests=$((pass+fail))
echo "Tests:  $tests"
echo "Pass :  $pass"
echo "Fail :  $fail"

