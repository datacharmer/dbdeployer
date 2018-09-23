#!/bin/bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2018 Giuseppe Maxia
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
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

