#!/bin/bash

function run {
    echo "$@"
    $@
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        exit $exit_code
    fi
}

run ./scripts/sanity_check.sh
run ./test/go-unit-tests.sh
run ./scripts/build.sh linux
run ./test/mock/defaults-change.sh
run ./test/mock/short-versions.sh
run ./test/mock/direct-paths.sh
run ./test/mock/expected_ports.sh
run ./test/mock/read-only-replication.sh
export RUN_CONCURRENTLY=1
export EXIT_ON_FAILURE=1
run ./test/docker-test.sh $(cat .build/VERSION)

