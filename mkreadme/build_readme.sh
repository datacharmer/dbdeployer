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
./make_readme < readme_template.md > README.md
echo "# $PWD"
ls -lhotr 

