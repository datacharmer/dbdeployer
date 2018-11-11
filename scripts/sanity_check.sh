#!/bin/bash
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


exec_dir=$(dirname $0)
cd $exec_dir
cd ..

local_items=(.build cmd defaults common unpack abbreviations concurrent sandbox compare)
exit_code=0
spaces="        "
function run {
    cmd="$@"
    output=$($@)
    if [ -n "$output" ]
    then
        exit_code=1
        echo "$spaces $output"
    fi
}

echo "## gofmt"
for dir in ${local_items[*]} docs/coding
do
    cd $dir
    echo "# $dir/"
    run "gofmt -l *.go"
    cd -    > /dev/null
done

echo ""
echo "## go vet"
for dir in ${local_items[*]}
do
    cd $dir
    echo "# $dir/"
    run "go vet"
    cd -    > /dev/null
done
echo "# docs/coding/"
for gf in docs/coding/*.go
do
    run "go vet $gf"
done

echo "# copyright"
for dir in ${local_items[*]}
do
    cd $dir
    for F in *.go
    do
        has_copyright1=$(head -n 1 $F | grep DBDeployer )
        has_copyright2=$(head -n 2 $F | tail -n 1 | grep Copyright )
        if [ -z "$has_copyright1" -o -z "$has_copyright2" ]
        then
            exit_code=1
            echo "File $dir/$F has no copyright"
        fi
    done
    cd - > /dev/null
done
for SF in $(grep -v '^#' ./scripts/sh.txt)
do
    has_copyright1=$(head -n 2 $SF | tail -n 1 | grep DBDeployer )
    has_copyright2=$(head -n 3 $SF | tail -n 1 | grep Copyright )
    if [ -z "$has_copyright1" -o -z "$has_copyright2" ]
    then
        exit_code=1
        echo "File $SF has no copyright"
    fi
done

if [ "$exit_code" == "0" ]
then
    echo "# Sanity check passed"
else
    echo "### SANITY CHECK ($0) FAILED ###"
fi
exit $exit_code
