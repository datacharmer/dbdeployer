#!/bin/bash

exec_dir=$(dirname $0)
cd $exec_dir
cd ..

local_items=(cmd defaults common unpack abbreviations concurrent sandbox)
exit_code=0
spaces="        "
function run {
    cmd="$@"
    output=$($@)
    if [ -n "$output" ]
    then
        exit_code=1
        echo "$spaces $output"
    fi
}

echo "## gofmt"
for dir in ${local_items[*]} docs/coding
do
    cd $dir
    echo "# $dir/"
    run "gofmt -l *.go"
    cd -    > /dev/null
done

echo ""
echo "## go vet"
for dir in ${local_items[*]}
do
    cd $dir
    echo "# $dir/"
    run "go vet"
    cd -    > /dev/null
done
echo "# docs/coding/"
for gf in docs/coding/*.go
do
    run "go vet $gf"
done

if [ "$exit_code" == "0" ]
then
    echo "# Sanity check passed"
else
    echo "### SANITY CHECK ($0) FAILED ###"
fi
exit $exit_code
