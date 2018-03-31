#!/bin/bash

cd $(dirname $0)
if [ ! -x make_readme ]
then
	go build make_readme.go
fi
if [ ! -x make_readme ]
then
	echo "make_readme not found"
	exit 1
fi
if [ ! -f readme_template.md ]
then
	echo "readme_template.md not found"
	exit 1
fi
if [ -f $HOME/.dbdeployer/config.json ]
then
    dbdeployer defaults reset
fi
./make_readme < readme_template.md > README.md
#./make_readme < api_template.md > API.md
dbdeployer tree --api | ./make_readme > API.md
dbdeployer defaults reset
echo "# $PWD"
ls -lhotr 

