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

fail=0
if [ -d $HOME/opt ]
then
    mv $HOME/opt $HOME/noopt
    if [ -d $HOME/opt ]
    then
        echo "not ok - error renaming $HOME/opt"
        exit 1
    fi
fi

function remove_completion_file {
    if [  -f dbdeployer_completion.sh ]
    then
        rm -f dbdeployer_completion.sh
    fi
}

function cleanup {
    bin=$(dbdeployer info defaults sandbox-binary)
    smb=$(dbdeployer info defaults sandbox-home)
    dbdeployer defaults reset
    if [ -d $bin ]
    then
        rm -rf $bin
    fi
    if [ -d $smb ]
    then
        rm -rf $smb
    fi
    remove_completion_file
}

function check_exit_code {

    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo "not ok - non-zero exit code detected ($exit_code)"
        exit $exit_code
    fi
}

function exists {
    file_name=$1
    mode=$2

    def=""
    case $mode in
        -f)
            def=file
            ;;
        -d)
            def=directory
            ;;
        -e)
            def=executable
            ;;
    esac

    if [ $mode $file_name ]
    then
        echo "ok - $def $file_name exists"
    else
        echo "not ok - $def $file_name does not exist"
        fail=$((fail+1))
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            exit 1
        fi
    fi
}

function not_exists {
    fname=$1
    if [ -e $fname ]
    then
        echo "not ok - $fname exists"
        fail=$((fail+1))
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            exit 1
        fi
    else
        echo "ok - $fname does not exist"
    fi
}

function ok_equal {
    label="$1"
    a="$2"
    b="$3"
    if [ "$a" == "$b" ]
    then
        echo "ok - $label - '$a'"
    else
        echo "not ok - $label - '$a' and '$b' are different"
        fail=$((fail+1))
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            exit 1
        fi
    fi
}

completion_file=/etc/bash_completion.d/dbdeployer_completion.sh

if [ -f $completion_file ]
then
    rm -f $completion_file
fi

dashline="# ---------------------------------------------------------"

echo $dashline
echo "# default initialization: will create binary directory and populate it"
echo $dashline
dbdeployer init 
check_exit_code

exists $HOME/opt/mysql -d
exists $HOME/sandboxes -d
exists $HOME/opt/mysql/*/bin/mysqld -x
exists $completion_file -f

ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$HOME/sandboxes"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$HOME/opt/mysql"


echo $dashline
echo "# idempotent initialization 1: will check existence and do nothing"
echo $dashline
dbdeployer init 
check_exit_code

export mybin=$HOME/mybin
export mysmb=$HOME/mysmb

echo $dashline
echo "# initialization with custom parameters"
echo $dashline

dbdeployer init --sandbox-home=$mysmb --sandbox-binary=$mybin
check_exit_code

exists $mybin -d
exists $mysmb -d
exists $mybin/*/bin/mysqld -x

ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$mysmb"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$mybin"

remove_completion_file
echo $dashline
echo "# idempotent initialization 2: will check existence and do nothing"
echo $dashline
dbdeployer init 
check_exit_code
ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$mysmb"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$mybin"

echo $dashline
echo "# initialization with custom environmengt variables"
echo $dashline
export SANDBOX_BINARY=$HOME/mybinaries
export SANDBOX_HOME=$HOME/mysandboxes

dbdeployer init 
check_exit_code

exists $SANDBOX_HOME -d
exists $SANDBOX_BINARY -d
exists $SANDBOX_BINARY/*/bin/mysqld -x
ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$SANDBOX_HOME"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$SANDBOX_BINARY"


remove_completion_file
echo $dashline
echo "# idempotent initialization 3: will check existence and do nothing"
echo $dashline
dbdeployer init 
check_exit_code
ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$SANDBOX_HOME"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$SANDBOX_BINARY"


remove_completion_file
echo $dashline
echo "# dry-run initialization: will show actions and do nothing"
echo $dashline
dbdeployer init --dry-run
check_exit_code
ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$SANDBOX_HOME"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$SANDBOX_BINARY"


remove_completion_file
echo $dashline
echo "# dry-run initialization with options: will show actions and do nothing"
echo $dashline
dbdeployer init --dry-run --sandbox-home=dummy1 --sandbox-binary=dummy2
check_exit_code
exists $SANDBOX_HOME -d
exists $SANDBOX_BINARY -d
exists $SANDBOX_BINARY/*/bin/mysqld -x
ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$SANDBOX_HOME"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$SANDBOX_BINARY"

mysqld=$SANDBOX_BINARY/*/bin/mysqld

echo $dashline
echo "# initialization without downloads"
echo $dashline
cleanup
not_exists $SANDBOX_HOME
not_exists $SANDBOX_BINARY

dbdeployer init --skip-all-downloads
check_exit_code
ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$SANDBOX_HOME"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$SANDBOX_BINARY"
not_exists $mysqld


remove_completion_file
echo $dashline
echo "# initialization without tarball download"
echo $dashline
cleanup
not_exists $SANDBOX_HOME
not_exists $SANDBOX_BINARY

dbdeployer init --skip-tarball-download
check_exit_code
ok_equal "sandbox home" "$(dbdeployer info defaults sandbox-home)" "$SANDBOX_HOME"
ok_equal "sandbox binary" "$(dbdeployer info defaults sandbox-binary)" "$SANDBOX_BINARY"
not_exists $mysqld


if [ "$fail" != "0" ]
then
    exit 1
fi
