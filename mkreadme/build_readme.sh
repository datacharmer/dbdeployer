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
dbdeployer defaults reset
./make_readme < readme_template.md > README.md
dbdeployer defaults reset
echo "# $PWD"
ls -lhotr 

