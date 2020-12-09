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

# This script tests dbdeployer with all available versions inside a docker 
# container.
# This setup is useful for testing also template and defaults export/import
# which could be intrusive in a regular environment.

version=$1
test_command=$2

if [ -z "$version" ]
then
    echo "version required"
    exit 1
fi

executable="dbdeployer-${version}.linux"
if [ ! -x $executable ]
then
    echo "executable not found"
    exit 1
fi

if [ ! -d "./test" ]
then
    echo "directory ./test not found"
    exit 1
fi

docker version
go version

container_name=dbtest
#if [ "$(uname)" == "Darwin" ]
#then
#    # This name identifies the container as running in Docker for mac,
#    # and will allow the test script to tune operations accordingly.
#    if [ -z "$INCLUDE_56" ]
#    then
#        container_name=dbtestmac
#    fi
#fi

exists=$(docker ps -a | grep -w $container_name )
if [ -n "$exists" ]
then
    docker rm -v -f $container_name
fi

TARGET_DIR=test

# $GITHUB_ACTIONS is defined when the test is running in Github actions environment
if [ -n "$GITHUB_ACTIONS" ]
then
    export RUN_CONCURRENTLY=1
    export EXIT_ON_FAILURE=1
    export ONLY_MAIN=1
    TARGET_DIR=original_test
fi

[ -n "$INTERACTIVE" ] && DOCKER_OPTIONS="$DOCKER_OPTIONS -e INTERACTIVE=1"
[ -n "$RUN_CONCURRENTLY" ] && DOCKER_OPTIONS="$DOCKER_OPTIONS -e RUN_CONCURRENTLY=1"
[ -n "$VERBOSE_CONCURRENCY" ] && DOCKER_OPTIONS="$DOCKER_OPTIONS -e VERBOSE_CONCURRENCY=1"
[ -n "$EXIT_ON_FAILURE" ] && DOCKER_OPTIONS="$DOCKER_OPTIONS -e EXIT_ON_FAILURE=1"
[ -n "$ONLY_MAIN" ] && DOCKER_OPTIONS="$DOCKER_OPTIONS -e ONLY_MAIN=1"
[ -n "$DBDEPLOYER_LOGGING" ] && DOCKER_OPTIONS="$DOCKER_OPTIONS -e DBDEPLOYER_LOGGING=1"

[ -z "$test_command" ] && test_command="./$TARGET_DIR/functional-test.sh"

interactive="-ti"
if [ "$test_command" != "bash" ]
then
    interactive=""
    test_command="bash -c $test_command"
fi

(set -x
  docker run $interactive \
    -v $PWD/$executable:/usr/bin/dbdeployer \
    -v $PWD/test:/home/msandbox/$TARGET_DIR \
    --name $container_name \
    --hostname $container_name $DOCKER_OPTIONS \
    datacharmer/mysql-sb-full $test_command
)

#    datacharmer/mysql-sb-full bash -c "./test/functional-test.sh"
#    datacharmer/mysql-sb-full bash

