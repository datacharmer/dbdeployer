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

exec_dir=$(dirname $0)

if [ ! -f $exec_dir/common.sh ]
then
    echo "$exec_dir/common.sh not found"
    exit 1
fi

source $exec_dir/common.sh
cd $exec_dir
alt_dbdeployer "$@"

