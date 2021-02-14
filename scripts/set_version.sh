#!/bin/bash
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
version=$1
vtype=$2

version_dir=./common
if [  ! -d $version_dir ]
then
    echo "Directory $version_dir not found"
    exit 1
fi

cd $version_dir

if [ -z "$version" ]
then
    echo "Syntax: $0 version [compatible]"
    echo "Current main version : $(cat VERSION)"
    echo "Compatible version   : $(cat COMPATIBLE_VERSION)"
    exit 1
fi

if [ -n "$vtype" ]
then
    if [ "$vtype" != "compatible" ]
    then
        echo "The only supported version type is 'compatible'"
        exit 1
    fi
fi

function set_version {
    file_name=$1
    version=$2
    existing_version=$(cat $file_name | tr -d '\n')
    if [[ $existing_version > $version ]]
    then
        echo "ERROR: existing version '$existing_version' from file '$file_name' is greater than version '$version'"
        exit 1
    fi
    echo "Setting version from '$existing_version' to '$version'"
    echo -n $version > $file_name
}

if [ "$vtype" == "compatible" ]
then
    set_version COMPATIBLE_VERSION $version
else
    set_version VERSION $version
fi 
