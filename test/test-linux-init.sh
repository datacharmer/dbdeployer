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

# This script tests dbdeployer initialization within a docker container.

version=$1

if [ -z "$version" ]
then
    echo "version required"
    exit 1
fi

executable="$PWD/dbdeployer-${version}.linux"
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

echo "Using dbdeployer '$executable'"

for name in centos7 centos8 ubuntu ubuntu20 
do
    exists=$(docker ps -a | grep -w $name)
    if [ -n "$exists" ]
    then
        docker rm -v -f $name
    fi
    docker run \
       -v $executable:/usr/bin/dbdeployer \
       -v $PWD/test:/root/test \
       --name=$name \
       --hostname=ostest \
       datacharmer/my-$name \
       /root/test/test-init.sh

    if [ "$?" != "0" ]
    then
        echo "ERROR running OS test for $name"
        exit 1
    fi
    docker rm -f -v $name
done

