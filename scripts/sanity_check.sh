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


exec_dir=$(dirname $0)
cd $exec_dir
cd ..

export GO111MODULE=on
function check_latest_version {
    git_version=$(git tag | tail -n 1 | tr -d 'v')
    dbdeployer_version=$(cat common/VERSION | tr -d '\n')
    if [ -z "$git_version" ]
    then
        echo "# --------------------------------------- #"
        echo "# WARNING: could not find git tag version #"
        echo "# --------------------------------------- #"
        return
    fi
    is_beta=$(echo "$git_version" |grep -i beta)
    if [ -z "$is_beta" ]
    then
        echo "## version"
        if [[ $git_version > $dbdeployer_version ]]
        then
            echo "Git tag version '$git_version' is bigger than dbdeployer version '$dbdeployer_version'"
            exit 1
        fi
        echo "Current version '$dbdeployer_version' is compatible with git tag version '$git_version'"
    fi
}

function exists_in_path {
    what=$1
    for dir in $(echo $PATH | tr ':' ' ')
    do
        wanted=$dir/$what
        if [ -x $wanted ]
        then
            echo $wanted
            return
        fi
    done
}


local_items=(cmd defaults downloads common globals compare cookbook unpack abbreviations concurrent sandbox compare rest importing ops ts ts_static)
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

function check_fmt {
    echo ""
    echo "## gofmt"
    for dir in ${local_items[*]} docs/coding
    do
        cd $dir
        echo "# $dir/"
        run "gofmt -l *.go"
        cd -    > /dev/null
    done
}

function check_vet {
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
}

function check_copyright {
    echo ""
    echo "## copyright"
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
    for SF in $(git ls-tree -r HEAD --name-only | grep '\.sh' | grep -v dbdeployer_completion | grep -v vendor)
    do
        has_copyright1=$(head -n 2 $SF | tail -n 1 | grep DBDeployer )
        has_copyright2=$(head -n 3 $SF | tail -n 1 | grep Copyright )
        if [ -z "$has_copyright1" -o -z "$has_copyright2" ]
        then
            exit_code=1
            echo "File $SF has no copyright"
        fi
    done
}

function check_static {
    STATIC_CHECK=$(exists_in_path staticcheck)

    if [ -n "$STATIC_CHECK" ]
    then
        echo ""
        echo "## staticcheck"
        for dir in ${local_items[*]}
        do
            echo "# $dir"
            run staticcheck github.com/datacharmer/dbdeployer/$dir
        done
    fi
}

function check_secure {
    SECURE_CHECK=$(exists_in_path gosec)

    if [ -n "$SECURE_CHECK" ]
    then
        echo ""
        echo "## secure check"
        for dir in ${local_items[*]}
        do
            echo "# $dir"
            run gosec -quiet $dir
        done
    fi
}


req=$1
if [ -n "$req" ]
then
    case $req in
        version)
            check_latest_version
            ;;
        fmt)
            check_fmt
            ;;
        vet)
            check_vet
            ;;
        copyright)
            check_copyright
            ;;
        static)
            check_static
            ;;
        secure)
            check_secure
            ;;
        *)
            echo "Syntax $(basename $0) [check_name]"
            echo "Allowed checks: version | fmt | vet | copyright | static | secure"
            exit 0
            ;;
    esac
else
    check_latest_version
    check_fmt
    check_vet
    check_copyright
    check_static
    check_secure
fi


if [ "$exit_code" == "0" ]
then
    echo "# Sanity check passed"
else
    echo "### SANITY CHECK ($0) FAILED ###"
fi

exit $exit_code
