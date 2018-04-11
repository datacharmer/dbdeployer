#!/bin/bash
target=$1
version=$2

build_dir=$(dirname $0)
cd $build_dir

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
local_items=(cmd defaults main.go common unpack abbreviations concurrent pflag sandbox)

all_ok=yes
for dep in ${dependencies[*]}
do
    if [ ! -d $GOPATH/src/$dep ]
    then
        echo "Needed package $dep not installed"
        unset all_ok
    fi
done

for item in ${local_items[*]}
do
    if [ ! -e ./$item ]
    then
        echo "item $item not found"
        unset all_ok
    fi
done

if [ -z "$all_ok" ]
then
    echo "Missing dependencies or essential code"
    echo "Use this command to gather the needed dependencies:"
    echo "   go get github.com/datacharmer/dbdeployer"
    echo "Also be sure to read pflag/README.md"
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
    echo "Syntax: target  version"
    echo "      target: (linux | OSX) "
    echo "Set the variable MKDOCS to build the docs-enabled dbdeployer (see README.md)"
    exit 1
fi

case $target in 
    all)
        $0 OSX $version
        $0 linux $version
        ;;
    OSX)
        executable=dbdeployer-${version}${docs_tag}.osx
	    (set -x
        env GOOS=darwin GOARCH=386 go build $docs_flags -o $executable .
        )
        tar -c $executable | gzip -c > ${executable}.tar.gz
    ;;
    linux)
        executable=dbdeployer-${version}${docs_tag}.linux
        (set -x
	    env GOOS=linux GOARCH=386 go build $docs_flags -o $executable .
        )
        tar -c $executable | gzip -c > ${executable}.tar.gz
    ;;
    *)
        echo unrecognized target.
        exit 1
        ;;
esac

ls -lh dbdeployer-${version}*
