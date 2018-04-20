#!/bin/bash

cd $(dirname $0)
# Need to recompile before generating documentation
# because we want the latest version to be included
for tool_name in make_readme
do
    go build $tool_name.go
    if [ "$?" != "0" ]
    then
        exit
    fi
    if [ ! -x $tool_name ]
    then
	    echo "executable $tool_name not found"
	    exit 1
    fi
done

if [ ! -f readme_template.md ]
then
	echo "readme_template.md not found"
	exit 1
fi
if [ -f $HOME/.dbdeployer/config.json ]
then
    dbdeployer defaults reset
fi
# Build README file
./make_readme < readme_template.md > README.md
#./make_readme < api_template.md > API.md

# Build API reference
dbdeployer-docs tree --api | ./make_readme > API.md
dbdeployer defaults reset

# Build completion file
completion_file=dbdeployer_completion.sh
if [ -f $completion_file ]
then
    rm -f $completion_file
fi

dbdeployer-docs tree --bash-completion
if [ ! -f $completion_file ]
then
    echo "# An error occurred: completion file '$completion_file' was not created"
    exit 1
fi

echo "# $PWD"
ls -lhotr 

