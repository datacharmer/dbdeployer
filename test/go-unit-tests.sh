#!/bin/bash
testdir=$(dirname $0)
cd $testdir
cd ..
maindir=$PWD
ls

function check_exit_code {
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo "Error during tests"
        exit $exit_code
    fi

}

for dir in common sandbox 
do
    cd $dir
    echo "# Testing $dir"
    go test -v
    check_exit_code
    cd $maindir
done
