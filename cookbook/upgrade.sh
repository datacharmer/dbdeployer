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
source cookbook_include.sh

function upgrade_db {
    UPGRADE_FROM=$1
    UPGRADE_TO=$2

    for dir in $UPGRADE_FROM $UPGRADE_TO 
    do
        check_version $dir
    done

    upgrade_from_dir=msb_$(echo $UPGRADE_FROM | tr '.' '_')
    upgrade_to_dir=msb_$(echo $UPGRADE_TO | tr '.' '_')

    if [ ! -d $SANDBOX_HOME/$upgrade_from_dir ]
    then
        ( set -x 
        dbdeployer deploy single $UPGRADE_FROM --master
        )
    fi
    if [ ! -d $SANDBOX_HOME/$upgrade_to_dir ]
    then
        (set -x 
        dbdeployer deploy single $UPGRADE_TO --master
        )
    fi
    (set -x
    $SANDBOX_HOME/$upgrade_from_dir/use -e "CREATE TABLE IF NOT EXISTS test.upgrade_log(id int not null auto_increment primary key, server_id int, vers varchar(50), urole varchar(20), ts timestamp)"
    $SANDBOX_HOME/$upgrade_from_dir/use -e "INSERT INTO test.upgrade_log (server_id, vers, urole) VALUES (@@server_id, @@version, 'original')"
    dbdeployer admin upgrade $upgrade_from_dir $upgrade_to_dir
    )
    if [ ! -f $upgrade_to_dir/no_upgrade ]
    then
        (set -x
        dbdeployer delete $upgrade_from_dir
        $SANDBOX_HOME/$upgrade_to_dir/use -e "INSERT INTO test.upgrade_log (server_id, vers, urole) VALUES (@@server_id, @@version, 'upgraded')"
        $SANDBOX_HOME/$upgrade_to_dir/use -e "SELECT * FROM test.upgrade_log"
        )
    fi
}

ver_55=5.5.53
ver_56=5.6.41
ver_57=5.7.23
ver_80=8.0.12

header "Upgrading from $ver_55 to $ver_56"
upgrade_db $ver_55 $ver_56
header "The upgraded database is now upgrading from $ver_56 to $ver_57 "
upgrade_db $ver_56 $ver_57
header "The further upgraded database is now upgrading from $ver_57 to $ver_80"
upgrade_db $ver_57 $ver_80

