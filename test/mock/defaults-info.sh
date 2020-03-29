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

echo "# Checks that dbdeployer info defaults and dbdeployer defaults list are in sync"

function check_defaults {
    list="$1"
    for d in $list
    do
        result=$(dbdeployer info defaults $d)
        error=$(echo $result | grep '# ERROR')
        if [ -n "$error" ]
        then 
            echo "not ok - $result"  
        else
            echo "ok - $d $result"
        fi
    done
}

check_defaults "$(dbdeployer defaults list | grep '"' | awk '{print $1}' | tr -d '":')"
check_defaults "$(dbdeployer defaults list --camel-case | awk '{print $1}')"
