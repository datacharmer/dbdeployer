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

# unset DBDEPLOYER_LOGGING
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
source set-mock.sh
start_timer

tests=0
fail=0
pass=0

create_mock_version 5.7.22
create_mock_version 5.7.23
create_mock_version 8.0.66
create_mock_version 8.0.67
create_mock_version 8.0.68

run dbdeployer deploy single 5.7.23
run dbdeployer deploy single 5.7.22
run dbdeployer deploy replication 5.7.23
run dbdeployer deploy replication 5.7.22
# dbdeployer sandboxes

expected_port=$(dbdeployer sandboxes | grep 'msb_5_7_23 .* .5723 ')
ok "expected port 5723" "$expected_port"
expected_port=$(dbdeployer sandboxes | grep 'msb_5_7_22 .* .5722 ')
ok "expected port 5722" "$expected_port"

expected_port=$(dbdeployer sandboxes | grep 'rsandbox_5_7_23 .* .19024 19025 19026')
ok "expected port replication 5723" "$expected_port"
expected_port=$(dbdeployer sandboxes | grep 'rsandbox_5_7_22 .* .18923 18924 18925')
ok "expected port replication 5722" "$expected_port"

run dbdeployer deploy single 8.0.67
expected_port=$(dbdeployer sandboxes | grep 'msb_8_0_67 .* .8067 ')
ok "expected port 8067" "$expected_port"

run dbdeployer deploy single 8.0.66
# Expecting server port and MySQLX port
expected_port=$(dbdeployer sandboxes | grep 'msb_8_0_66 .* .8066 18066')
ok "expected port 8066" "$expected_port"
ok_executable_exists $SANDBOX_HOME/msb_8_0_66/use
ok_executable_does_not_exist $SANDBOX_HOME/msb_8_0_66/use_admin


run dbdeployer deploy single 8.0.68 --enable-admin-address
# Expecting server port, admin port, and MySQLX port
expected_port=$(dbdeployer sandboxes | grep 'msb_8_0_68 .* .8068 19068 18068 ')
ok "expected port 8068" "$expected_port"
ok_executable_exists $SANDBOX_HOME/msb_8_0_68/use_admin

run dbdeployer deploy replication 8.0.68 --enable-admin-address
run dbdeployer deploy multiple 8.0.68 --enable-admin-address
expected_port=$(dbdeployer sandboxes | grep 'multi_msb_8_0_68 .* .30869 40869 41869 ')
expected_port=$(dbdeployer sandboxes | grep 'rsandbox_8_0_68 .* .25869 35869 36869 ')
ok_executable_exists $SANDBOX_HOME/rsandbox_8_0_68/use_all_admin
ok_executable_exists $SANDBOX_HOME/rsandbox_8_0_68/master/use_admin
ok_executable_exists $SANDBOX_HOME/rsandbox_8_0_68/node1/use_admin
ok_executable_exists $SANDBOX_HOME/rsandbox_8_0_68/node2/use_admin
ok_executable_exists $SANDBOX_HOME/multi_msb_8_0_68/use_all_admin
ok_executable_exists $SANDBOX_HOME/multi_msb_8_0_68/node1/use_admin
ok_executable_exists $SANDBOX_HOME/multi_msb_8_0_68/node2/use_admin
ok_executable_exists $SANDBOX_HOME/multi_msb_8_0_68/node3/use_admin
dbdeployer sandboxes

run dbdeployer delete ALL --skip-confirm

results "After deletion"

cd $test_dir || (echo "error changing directory to $test_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer
tests=$((pass+fail))
echo "Tests:  $tests"
echo "Pass :  $pass"
echo "Fail :  $fail"

