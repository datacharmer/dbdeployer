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

rundir=$(dirname $0)
if [ ! -f $rundir/common.sh ]
then
    echo "file $rundir/common.sh not found"
    exit 1
fi
source $rundir/common.sh

tarball_dir=$1
if [ -z "$tarball_dir" ]
then
    echo "Syntax: $0 tarball-directory"
    exit 1
fi

export SANDBOX_BINARY=/tmp/extract
if [ ! -d $SANDBOX_BINARY ]
then
    mkdir $SANDBOX_BINARY
fi

start_timer
for f in $tarball_dir/mysql*.gz
do
    echo "# $f"
    dbdeployer extract $f
    if [  "$?" != "0" ]
    then
        echo "error unpacking $f"
        exit 1
    fi
    is_linux=$(echo $f | grep linux)
    is_shell=$(echo $f | grep shell)
    if [ -z "$is_shell" -a -z "$is_linux" ]
    then
        dbdeployer deploy single $SANDBOX_BINARY/*
        if [  "$?" != "0" ]
        then
            echo "error deploying $f"
            exit 1
        fi
        dbdeployer delete all --concurrent --skip-confirm
    fi
    rm -rf $SANDBOX_BINARY/*
done
stop_timer

