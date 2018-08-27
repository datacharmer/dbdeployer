#!/bin/bash
version=$1
vtype=$2

version_dir=$(dirname $0)
cd $version_dir

if [ -z "$version" ]
then
    echo "Syntax: $0 version [compatible]"
    exit 1
fi

if [ -n "$vtype" ]
then
    if [ "$vtype" != "compatible" ]
    then
        echo "The only supported version type is 'compatible'"
        exit 1
    fi
fi

function set_version {
    file_name=$1
    version=$2
    echo $version > $file_name
}

if [ "$vtype" == "compatible" ]
then
    set_version COMPATIBLE_VERSION $version
else
    set_version VERSION $version
fi 
