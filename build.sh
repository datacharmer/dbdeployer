#!/bin/bash
target=$1
version=$2

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
