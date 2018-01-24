#!/bin/bash
target=$1
version=$2

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
        executable=dbdeployer-$version.osx
	    env GOOS=darwin GOARCH=386 go build -o $executable .
        tar -c $executable | gzip -c > ${executable}.tar.gz
        ls -lh $executable*
    ;;
    linux)
        executable=dbdeployer-$version.linux
	    env GOOS=linux GOARCH=386 go build -o $executable .
        tar -c $executable | gzip -c > ${executable}.tar.gz
        ls -lh $executable*
    ;;
    *)
        echo unrecognized target.
        exit 1
        ;;
esac

