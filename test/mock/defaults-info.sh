#!/bin/bash

echo "# Checks that dbdeployer info defaults and dbdeployer defaults list are in sync"

function check_defaults {
    list="$1"
    for d in $list
    do
        result=$(dbdeployer info defaults $d)
        error=$(echo $result | grep '# ERROR')
        if [ -n "$error" ]
        then 
            echo "not ok - $result"  
        else
            echo "ok - $d $result"
        fi
    done
}

check_defaults "$(dbdeployer defaults list | grep '"' | awk '{print $1}' | tr -d '":')"
check_defaults "$(dbdeployer defaults list --camel-case | awk '{print $1}')"
