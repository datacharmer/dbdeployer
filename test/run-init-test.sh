#!/usr/bin/env bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2022 Giuseppe Maxia
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

version=$1
if [ -z "$version" ]
then
    echo "version required"
    exit 1
fi

extra=$2

executable=dbdeployer-${version}.linux
if [ ! -x $executable  ]
then
    echo "executable '$executable' not found"
    exit 1
fi


exists=$(docker ps -a | grep initdbtest )
if [ -n "$exists" ]
then
    docker rm -v -f initdbtest
fi

interactive=""
cmd=/home/msandbox/test-init.sh
if [ -n "$extra" ]
then
    interactive="-ti"
    cmd=bash
fi

docker run $interactive \
    -v $PWD/$executable:/usr/bin/dbdeployer \
    -v $PWD/test/test-init.sh:/home/msandbox/test-init.sh \
    -e EXIT_ON_FAILURE=1 \
    --name=initdbtest \
    --hostname=initdbtest \
    datacharmer/mysql-sb-base $cmd

exit_code=$?

echo "exit code: $exit_code"

exit $exit_code
