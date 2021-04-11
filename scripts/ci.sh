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

function run {
    echo "$@"
    "$@"
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        exit $exit_code
    fi
}

export dbdeployer_version=$(cat common/VERSION)
export RUN_CONCURRENTLY=1
export EXIT_ON_FAILURE=1
run ./scripts/sanity_check.sh
run ./test/go-unit-tests.sh
run ./scripts/build.sh linux

executable=dbdeployer-${dbdeployer_version}.linux
if [ ! -f $executable ]
then
    echo "executable $executable not found"
    exit 1
fi
cp $executable dbdeployer
export PATH=$PWD:$PATH

run ./test/run-mock-tests.sh
run ./test/docker-test.sh $dbdeployer_version
run ./test/test-linux-init.sh $dbdeployer_version

