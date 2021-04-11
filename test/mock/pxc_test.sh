#!/bin/bash
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

# Creates a zero-length catalog file.
# Sandbox creation should not fail
mkdir -p $mock_dir/home/.dbdeployer
touch $mock_dir/home/.dbdeployer/sandboxes.json

# Creates a fake $HOME/bin directory, containing the required executables for PXC
mkdir $mock_dir/home/bin
export PATH=$PATH:$mock_dir/home/bin
dbdeployer defaults templates show no_op_mock > $mock_dir/home/bin/socat
dbdeployer defaults templates show no_op_mock > $mock_dir/home/bin/rsync
dbdeployer defaults templates show no_op_mock > $mock_dir/home/bin/lsof
chmod +x $mock_dir/home/bin/*


versions=(pxc5.6 pxc5.7 pxc8.0)
rev_list="21 99"

for rev in $rev_list
do
    for vers in ${versions[*]}
    do
        version=${vers}.${rev}
        create_mock_pxc_version $version 
    done
done

function check_sst_method {
    dir=$1
    expected=$2

    my_file=$SANDBOX_HOME/$dir/node1/my.sandbox.cnf
    ok_file_exists $my_file
    ok "expected is defined" "$expected"

    found=$(grep "wsrep_sst_method\s*=\s*$expected" $my_file )
    ok "Expected $expected found in $my_file" "$found"
}

run dbdeployer available
for vers in ${versions[*]}
do
    for rev in $rev_list
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy replication $version --topology=pxc

        # Check existence 
        test_completeness $version pxc_msb_ multiple

        # Check SST method
        if [ "$vers" == "pxc8.0" ]
        then
            expected=xtrabackup-v2
        else
            expected=rsync
        fi
        check_sst_method pxc_msb_$version_name $expected

        ok_dir_exists "$SANDBOX_HOME/pxc_msb_$version_name"
        results "$version"
    done
done

run dbdeployer delete ALL --skip-confirm

results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

