#!/usr/bin/env bash
# Bash script generating JSON entry for DBDeployer's downloads list
# Copyright © 2021 lefred <lefred.descamps@gmail.com>

# v0.1

# 2021-04-30 - lefred - v0.1

html2text_run=$(which html2text 2>/dev/null)
if [[ $? -ne 0 ]]
then
    echo "ERROR: please install html2text (>=2.0)"
    exit 5
fi
html2text_ver=$($html2text_run --version | awk '{print $NF}' | cut -d'.' -f1)
if [[ $html2text_ver -lt 2 ]]
then
    echo "ERROR: please install html2text (>=2.0) - current version is $html2text_ver"
    exit 5
fi

categories=(mysql shell cluster)
os="Linux"
agent="Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:59.0) Gecko/20100101 Firefox/59.0" 

if [[ $# -lt 2 ]]
then
    echo "ERROR: two arguments are requires and the 3rd is optional !"
    echo "       $(basename $0) <mysql|cluster|shell> <version> <Linux|Darwin>"
    exit 1
fi

if ! [[ ${categories[*]} =~ $1 ]]
then
    echo "ERROR: $1 is not a correct category (mysql|cluster|shell) !!"
    exit 2
fi
if [[ ${3,,} == "darwin" || ${3,,} == "macos" || ${3,,} == "mac" ]]
then
    os="Darwin"
    agent="User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (K HTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"
fi


url="https://dev.mysql.com/downloads/$1/"
url_file="https://dev.mysql.com/get/Downloads"
case $1 in
    mysql)
        flavor="mysql"
        to_grep="minimal.tar.xz"
        minimal="true"
        if [[ "$os" == "Darwin" ]]
        then
            to_grep="tar.gz"
            minimal="false"
        fi
        if ! [[ $2 == 8* || $2 == 5.7* ]]
        then
                echo "ERROR: invalid version !"
                exit 3
        fi
        if [[ $2 == 5* ]];
        then
            if [[ "$os" == "Darwin" ]]
            then
                echo "ERROR: no MySQL Server for MacOS for this version !"
                exit 3
            fi
            url="https://dev.mysql.com/downloads/mysql/5.7.html"
            to_grep="_64.tar.gz"
            minimal="false"
        fi
        # shellcheck disable=SC2066
        for line in "$(curl -s -A "$agent" $url | html2text | grep $to_grep -A 1| grep mysql-[0-9] -A 1)"
        do
            file=$(echo $line | awk '{print substr($1, 2, length($1) - 2)}')
            if [[ $file == *$2*   ]]
            then
                checksum=$(echo $line | awk '{print substr($4, 2, length($4) - 2)}')
                url_file_2="MySQL-${2:0:3}"
                size=$(curl -s -I -L  ${url_file}/${url_file_2}/$file | grep Content-Len | cut -d' ' -f 2)
            fi
        done
        version=$(echo $file| cut -d'-' -f2)
        ;;
    shell)
        flavor="shell"
        minimal="false"
        to_grep="64bit.tar.gz"
        if ! [[ $2 == 8* ]]
        then
                echo "ERROR: invalid version, only 8 is supported!"
                exit 3
        fi
        if [[ "$os" == "Darwin" ]]
        then
            to_grep="tar.gz"
        fi
        # shellcheck disable=SC2066
        for line in "$(curl -s -A "$agent" $url | html2text | grep $to_grep -A 1)"
        do
            file=$(echo $line | awk '{print substr($1, 2, length($1) - 2)}')
            if [[ $file == *$2*   ]]
            then
                checksum=$(echo $line | awk '{print substr($4, 2, length($4) - 2)}')
                url_file_2="MySQL-Shell"
                size=$(curl -s -I -L  ${url_file}/${url_file_2}/$file | grep Content-Len | cut -d' ' -f 2)
            fi
        done
        version=$(echo $file| cut -d'-' -f3)
        ;;
    cluster)
        flavor="ndb"
        to_grep="x86_64.tar.gz)"
        minimal="false"
        if [[ "$os" == "Darwin" ]]
        then
            to_grep="tar.gz"
        fi
        if ! [[ $2 == 8* || $2 == 7* ]]
        then
                echo "ERROR: invalid version !"
                exit 3
        fi

        if [[ $2 == 7* ]];
        then
            if [[ "$os" == "Darwin" ]]
            then
                echo "ERROR: no MySQL Cluster for MacOS for this version !"
                exit 3
            fi
            url="https://dev.mysql.com/downloads/cluster/7.6.html"
        fi
        # shellcheck disable=SC2066
        for line in "$(curl -s -A "$agent" $url | html2text | grep $to_grep -A 1| grep mysql-cluster -A 1)"
        do
            file=$(echo $line | awk '{print substr($1, 2, length($1) - 2)}')
            if [[ $file == *$2*   ]]
            then
                checksum=$(echo $line | awk '{print substr($4, 2, length($4) - 2)}')
                url_file_2="MySQL-Cluster-${2:0:3}"
                size=$(curl -s -I -L  ${url_file}/${url_file_2}/$file | grep Content-Len | cut -d' ' -f 2)
            fi
        done
        version=$(echo $file| cut -d'-' -f3)
        if [[ $2 == 7* ]];
        then
            version=$(echo $file| cut -d'-' -f4)
        fi
        ;;
    *)
        echo "Not yet implemented"
        ;;
esac

date_added=$(date +"%F %H:%M")
cat << EOL
     {
       "name": "$file",
       "checksum": "MD5:$checksum",
       "OS": "$os",
       "url": "${url_file}/${url_file_2}/$file",
       "flavor": "$flavor",
       "minimal": $minimal,
       "size": ${size::-1},
       "short_version": "${version:0:3}",
       "version": "$version",
       "updated_by": "get_dbdeployer_mysql_json.sh",
       "date_added": "$date_added"
      }
EOL
