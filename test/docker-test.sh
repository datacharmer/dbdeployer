#!/bin/bash
version=$1

if [ -z "$version" ]
then
    echo "version required"
    exit 1
fi

executable="dbdeployer-${version}.linux"
if [ ! -x $executable ]
then
    echo "executable not found"
    exit 1
fi

if [ ! -d "./test" ] 
then
    echo "directory ./test not found"
    exit 1
fi

exists=$(docker ps |grep dbtest)
if [ -n "$exists" ]
then
    docker rm -v -f dbtest
fi

docker run -ti  \
    -v $PWD/$executable:/usr/bin/dbdeployer \
    -v $PWD/test:/home/msandbox/test \
    --name dbtest \
    --hostname dbtest \
    datacharmer/mysql-sb-full bash -c "./test/test.sh"

#    datacharmer/mysql-sb-full bash -c "./test/test.sh"
#    datacharmer/mysql-sb-full bash

