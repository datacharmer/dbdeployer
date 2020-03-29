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

#unset DBDEPLOYER_LOGGING
test_dir=$(dirname $0)
cd $test_dir || (echo "error changing directory to $test_dir" ; exit 1)
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
source set-mock.sh
start_timer

# Creates a zero-length catalog file.
# Sandbox creation should not fail
mkdir -p $mock_dir/home/.dbdeployer
touch $mock_dir/home/.dbdeployer/sandboxes.json

cd $mock_dir
NEW_BASH=$PWD/bash
ln -s /bin/bash $NEW_BASH

versions=(5.0 5.1 5.5 5.6 5.7 8.0)
ndb_versions=(ndb7.6 ndb8.0)
pxc_versions=(pxc5.7)
tidb_versions=(tidb3.0)
rev_list="21"

for rev in $rev_list
do
    for vers in ${versions[*]}
    do
        version=${vers}.${rev}
        create_mock_version $version 
    done
    for vers in ${ndb_versions[*]}
    do
        version=${vers}.${rev}
        create_mock_ndb_version $version 
    done
    for vers in ${pxc_versions[*]}
    do
        version=${vers}.${rev}
        create_mock_pxc_version $version 
    done
    for vers in ${tidb_versions[*]}
    do
        version=${vers}.${rev}
        create_mock_tidb_version $version
    done
done

run dbdeployer available

dbdeployer --shell-path=$NEW_BASH cookbook create all

# We exclude from the search the prerequisites file, because there
# is no version replacement in there, and the include file,
# because, in addition to not having any replacement,  it uses 
# the NOTFOUND string for its own ourposes.
exclude_files=(recipes/prerequisites.sh recipes/cookbook_include.sh)

# All the recipes (except the excluded ones) have a placeholder
# for a required version, which dbdeployer fills with the one that was found
# in $SANDBOX_BINARY. If none was found, the string "NOTFOUND_flavorName" is used
for F in recipes/*
do
    unset exclude
    for  EF in ${exclude_files[*]}
    do
        if [ "$F" == "$EF" ]
        then
            exclude=1
        fi
    done
    if [ -z "$exclude" ]
    then
        # If the file does not contain the string "NOTFOUND", it means
        # that it was built using a suitable version for the flavor
        # required on that recipe.
        missed_version=$(grep NOTFOUND $F)
        ok_empty "File $F has a version" "$missed_version"
        missed_value=$(grep 'no value' $F)
        ok_empty "File $F has all the required values" "$missed_value"
        has_new_bash=$(grep "$NEW_BASH" $F)
        ok  "File $F has the new bash" "$has_new_bash"
    fi
done

# Add test for custom bash interpreter

dbdeployer --shell-path=$NEW_BASH deploy single 5.0 --sandbox-directory=newbash

test_completeness 5.0.21 newbash single $NEW_BASH

dbdeployer delete all --concurrent --skip-confirm

results "After deletion"
cd $test_dir || (echo "error changing directory to $mock_dir" ; exit 1)

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

