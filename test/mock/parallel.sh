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
parallel_found=$(exists_in_path parallel)
if [ -z "$parallel_found" ]
then
    echo "Command 'parallel' not found"
    exit 0
fi

source set-mock.sh
mkdir $HOME/.parallel
touch $HOME/.parallel/will-cite
start_timer

versions=(5.0 5.1 5.5 5.6 5.7 8.0)
# versions=(5.0 )
rev_list="26 37 48 59"
# rev_list="26 37"

for rev in $rev_list
do
    for vers in ${versions[*]}
    do
        version=${vers}.${rev}
        create_mock_version $version 
    done
done

run dbdeployer available
for vers in ${versions[*]}
do
    # run parallel dbdeployer deploy {1} $vers.{2} ::: single multiple ::: $rev_list
    parallel --shellquote dbdeployer deploy {1} $vers.{2} ::: single multiple ::: $rev_list
    run parallel  dbdeployer deploy {1} $vers.{2} ::: single multiple ::: $rev_list

    how_many=$(dbdeployer sandboxes | wc -l | tr -d ' \t')
    how_many_catalog=$(count_catalog)
    ok_equal "sandboxes_in_catalog" $how_many_catalog $how_many
    
    results "$vers"
    run dbdeployer delete ALL --skip-confirm
done


results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

