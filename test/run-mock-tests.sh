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
if [ ! -d test/mock ]
then
    echo "directory test/mock not found"
    exit 1
fi

EXIT_ON_FAILURE=1
RUN_CONCURRENTLY=1
for F in test/mock/*.sh
do
    args=""
    name=$(basename $F)
    if [ "$name" == "port-clash.sh" ]
    then
        args=sparse
    fi
    if [ "$name" != "set-mock.sh" ]
    then
        echo "### ./$F $args"
        ./$F $args
        if [ "$?" != "0" ]
        then
            exit 1
        fi
    fi
done

