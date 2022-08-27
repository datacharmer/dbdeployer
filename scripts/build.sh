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
target=$1
version=$2

build_dir=$(dirname $0)
build_script=$(basename $0)
cd $build_dir
build_script=$PWD/$build_script
cd ..
build_dir=$PWD

export GO111MODULE=on
goversion=$(go version)
go_major=$(echo $goversion | awk '{print $3}' | sed -e 's/^go//' | tr '.' ' ' | awk '{print $1}')
go_minor=$(echo $goversion | awk '{print $3}' | sed -e 's/^go//' | tr '.' ' ' | awk '{print $2}')
go_rev=$(echo $goversion | awk '{print $3}' | sed -e 's/^go//' | tr '.' ' ' | awk '{print $3}')

echo "# Detected go version: $go_major.$go_minor.$go_rev"

if [[ $go_major -gt 1 ]]
then
    echo "This application needs go 1.11+ - Detected ${go_major}.${go_minor}"
    exit 1
fi

if [[ $go_major -eq 1 && $go_minor -lt 11 ]]
then
    echo "Minimum Go version should be 1.11 - Detected ${go_major}.${go_minor}"
    exit 1
fi

local_items=(cmd defaults main.go common globals unpack abbreviations concurrent sandbox)

function find_in_path {
    wanted=$1
    for dir in $(echo $PATH | tr ':' ' ')
    do
        if [ -x $dir/$wanted ]
        then
            echo "$dir/$wanted"
            return
        fi
    done
}

dash_line="--------------------------------------------------------------------------------"

docs_flags=""
docs_tag=""
if [ -n "$DBDEPLOYER_DOCS" -o -n "$MKDOCS" ]
then
    docs_flags="--tags docs"
    docs_tag="-docs"
fi

if [ -z "$version" ]
then
    version=$(cat common/VERSION)
fi

if [ -z "$target" ]
then
    echo "Syntax: target [version]"
    echo "      target: (linux | OSX) "
    echo "Set the variable MKDOCS to build the docs-enabled dbdeployer (see README.md)"
    exit 1
fi

function shrink {
    binary=$1
    if [ -z "$SHRINK_EXECUTABLES" ]
    then
        return
    fi
    upx_cmd=$(find_in_path upx)
    if [ -z "$upx_cmd" ]
    then
        return
    fi
    upx -9 $binary
}

function make_signature {
    binary_file=$1
    sha_sum_cmd=$(find_in_path shasum)
    if [ -z "$sha_sum_cmd" ]
    then
        echo "shasum not found - signature missing"
        return
    fi
    $sha_sum_cmd -a 256 $binary_file > ${binary_file}.sha256

    $sha_sum_cmd -a 256 -c ${binary_file}.sha256
}

function build_binary {
    temp_binary=$1
    OS=$2
    arch=$3
    echo "env GOOS=$OS GOARCH=$arch go build $docs_flags -o $temp_binary ."
    env GOOS=$OS GOARCH=$arch go build $docs_flags -o $temp_binary .
    if [ "$?" != "0" ]
    then
        echo "ERROR during OSX build! ($temp_binary)"
        exit 1
    fi
}

function pack_binary {
    binary=$1
    tar -c $binary | gzip -c > ${binary}.tar.gz
    shrink $binary
    make_signature $binary
    make_signature ${binary}.tar.gz
}

function build_universal {
    binary=$1
    build_binary ${binary}-amd darwin amd64
    build_binary ${binary}-arm darwin arm64
    echo "lipo -create -output  $binary ${binary}-amd ${binary}-arm"
    lipo -create -output  $binary ${binary}-amd ${binary}-arm
    if [ "$?" != "0" ]
    then
        echo "ERROR during OSX build! ($temp_binary)"
        exit 1
    fi
    if [ ! -x $binary ]
    then
        echo "universal binary $binary not created"
        exit 1
    fi
    rm -f ${binary}-amd
    rm -f ${binary}-arm
}

case $target in
    all)
        $build_script OSX $version
        $build_script linux $version
        ;;
    OSX)
        binary=dbdeployer-${version}${docs_tag}.osx
        has_lipo=$(find_in_path lipo)
        if [ -n "$has_lipo" ]
        then
            echo "'lipo' executable detected: building universal binary"
            build_universal $binary
        else
            build_binary $binary darwin amd64
        fi
        pack_binary $binary
    ;;
    linux)
        binary=dbdeployer-${version}${docs_tag}.linux
        build_binary $binary linux amd64
        pack_binary $binary
    ;;
    *)
        echo "unrecognized target $target"
        exit 1
        ;;
esac

ls -lh dbdeployer-${version}*
