#!/bin/bash

exec_dir=$(dirname $0)
cd $exec_dir
cd ..

local_items=(cmd defaults common unpack abbreviations concurrent sandbox)
exit_code=0

function run {
    cmd="$@"
    output=$($@)
    if [ -n "$output" ]
    then
        exit_code=1
        echo $output
    fi
}

echo "# go format"
for dir in ${local_items[*]} docs/coding
do
    cd $dir
    run "gofmt -l *.go"
    cd -    > /dev/null
done

echo "# go vet"
for dir in ${local_items[*]}
do
    cd $dir
    run "go vet"
    cd -    > /dev/null
done
run "go vet docs/coding/minimal-sandbox.go"
run "go vet docs/coding/minimal-sandbox2.go"

if [ "$exit_code" == "0" ]
then
    echo "# Sanity check passed"
else
    echo "### SANITY CHECK ($0) FAILED ###"
fi
exit $exit_code
