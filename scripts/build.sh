#!/bin/bash
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

local_items=(cmd defaults main.go common unpack abbreviations concurrent sandbox)

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

dashline="--------------------------------------------------------------------------------"
all_ok=yes
for dep in ${dependencies[*]}
do
    #if [ ! -d $GOPATH/src/$dep ]
    if [ ! -d ./vendor/$dep ]
    then
        echo $dashline
        echo "Needed package $dep not installed"
        echo "run 'go get $dep'"
        echo $dashline
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
    #echo "Then run:"
    #echo "   go get -u github.com/datacharmer/dbdeployer"
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

version_builder=.build/create-version-source-file.go
if [ ! -f $version_builder ]
then
    echo "File $version_builder not found - aborting"
    exit 1
fi

function build_sort {
    OS=$1
    target_os=$(echo $OS | tr 'A-Z' 'a-z')
    cd test
    (set -x
    env GOOS=$target_os GOARCH=386 go build -o sort_versions.$OS sort_versions.go
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

go run $version_builder

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
        env GOOS=darwin GOARCH=386 go build $docs_flags -o $executable .
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
