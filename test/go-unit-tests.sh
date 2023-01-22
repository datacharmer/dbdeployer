#!/usr/bin/env bash
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
testdir=$(dirname $0)
cd $testdir
cd ..

maindir=$PWD

unset DBDEPLOYER_LOGGING

function check_exit_code {
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo "Error during tests"
        exit $exit_code
    fi

}

test_dirs=$(find . -name '*_test.go' -exec dirname {} \; | tr -d './' | sort |uniq)

for dir in $test_dirs
do
    cd $dir
    echo "# Testing $dir"
    go test -v -timeout 30m
    check_exit_code
    cd $maindir
done
