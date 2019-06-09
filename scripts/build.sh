#!/bin/bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2019 Giuseppe Maxia
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
executable=$(basename $0)
cd $build_dir
executable=$PWD/$executable
cd ..
build_dir=$PWD


if [ -z "$GOPATH" ]
then
    GOPATH=$HOME/go
fi
if [ ! -d "$GOPATH" ]
then
    echo "\$GOPATH directory ($GOPATH) not found"
    exit 1
fi

in_go_path=$(echo $PWD | grep "^$GOPATH")
if [ -z "$in_go_path" ]
then
    echo "This directory ($PWD) is not in \$GOPATH ($GOPATH)"
    exit 1
fi
dependencies=(github.com/spf13/cobra github.com/spf13/pflag)
if [ -n "$MKDOCS" ]
then
    dependencies=(github.com/spf13/cobra github.com/spf13/pflag github.com/spf13/cobra/doc)
fi

goversion=$(go version)
go_major=$(echo $goversion | awk '{print $3}' | sed -e 's/^go//' | tr '.' ' ' | awk '{print $1}')
go_minor=$(echo $goversion | awk '{print $3}' | sed -e 's/^go//' | tr '.' ' ' | awk '{print $2}')

if [[ $go_major -gt 1 ]]
then
    echo "This application has only been tested with go 1.10+ - Detected ${go_major}.${go_minor}"
    exit 1
fi

if [[ $go_major -eq 1 && $go_minor -lt 10 ]]
then
    echo "Minimum Go version should be 1.10 - Detected ${go_major}.${go_minor}"
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
all_ok=yes
for dep in ${dependencies[*]}
do
    if [ ! -d ./vendor/$dep ]
    then
        echo $dash_line
        echo "Needed package $dep not installed"
        echo "run 'go get $dep'"
        echo $dash_line
        all_ok=no
    fi
done

for item in ${local_items[*]}
do
    if [ ! -e ./$item ]
    then
        echo "item $item not found"
        all_ok=no
    fi
done

if [ "$all_ok" == "no" ]
then
    echo "Missing dependencies or essential code"
    echo "Use the above 'go get' commands to gather the needed dependencies"
    exit 1
fi

docs_flags=""
docs_tag=""
if [ -n "$DBDEPLOYER_DOCS" -o -n "$MKDOCS" ]
then
    docs_flags="--tags docs"
    docs_tag="-docs"
fi

if [ -z "$version" ]
then
    version=$(cat .build/VERSION)
fi

code_generation=.build/code-generation.go
version_builder=version
tarball_builder=tarball
if [ ! -f $code_generation ]
then
    echo "File $code_generation not found - aborting"
    exit 1
fi

function build_sort {
    OS=$1
    target_os=$(echo $OS | tr 'A-Z' 'a-z')
    cd test
    (set -x
    env GOOS=$target_os GOARCH=amd64 go build -o sort_versions.$OS sort_versions.go
    )
    cd -
}

if [ -z "$target" ]
then
    echo "Syntax: target [version]"
    echo "      target: (linux | OSX) "
    echo "Set the variable MKDOCS to build the docs-enabled dbdeployer (see README.md)"
    exit 1
fi

# Checks whether the regular version and the compatible versions are already in the Go source file
current_version=$(cat .build/VERSION)
current_compatible_version=$(cat .build/COMPATIBLE_VERSION)
is_version=$(grep "VersionDef.*$current_version" common/version.go)
is_comp_version=$(grep "CompatibleVersion.*$current_compatible_version" common/version.go)
# if either version is missing from the build file, the source file is created again
if [ -z "$is_version" -o -z "$is_comp_version" ]
then
    go run $code_generation $version_builder
fi
go run $code_generation $tarball_builder
if [ "$?" != "0" ]
then
    echo "Error while building tarball registry source file"
    exit 1
fi

function shrink {
    executable=$1
    if [ -z "$SHRINK_EXECUTABLES" ]
    then
        return
    fi
    upx_cmd=$(find_in_path upx)
    if [ -z "$upx_cmd" ]
    then
        return
    fi
    upx -9 $executable
}

case $target in
    all)
        $executable OSX $version
        $executable linux $version
        ;;
    OSX)
        executable=dbdeployer-${version}${docs_tag}.osx
	    (set -x
        env GOOS=darwin GOARCH=amd64 go build $docs_flags -o $executable .
        )
        if [ "$?" != "0" ]
        then
            echo "ERROR during build!"
            exit 1
        fi
        tar -c $executable | gzip -c > ${executable}.tar.gz
        shrink $executable
        build_sort Darwin
    ;;
    linux)
        executable=dbdeployer-${version}${docs_tag}.linux
        (set -x
	    env GOOS=linux GOARCH=386 go build $docs_flags -o $executable .
        )
        tar -c $executable | gzip -c > ${executable}.tar.gz
        shrink $executable
        build_sort linux
    ;;
    *)
        echo unrecognized target.
        exit 1
        ;;
esac

ls -lh dbdeployer-${version}*
