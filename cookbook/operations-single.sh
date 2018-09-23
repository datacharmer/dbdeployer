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

version=$1
[ -z "$version" ] && version=5.7.23
check_version $version

if [ -z "$(dbdeployer sandboxes | grep 'single\s*'$version)" ]
then
    echo "single version $version is not installed"
    echo "Run './single.sh $version' before trying again"
    exit 1
fi

sandbox_dir=$SANDBOX_HOME/msb_$(echo $version | tr '.' '_' )

header "Running a simple command in the sandbox." \
   "Notice the usage of the '-e', as if we were using the 'mysql' client" 

(set -x
$sandbox_dir/use -e 'SHOW SCHEMAS'
)

header "Creating a table"
(set -x
$sandbox_dir/use test -e 'DROP TABLE IF EXISTS t1'
$sandbox_dir/use test -e 'CREATE TABLE t1(id int not null primary key)'
)

header "Inserting 3 lines into the new table"
for N in 1 2 3
do
    (set -x
    $sandbox_dir/use test -e "INSERT INTO t1 VALUES($N)"
    )
done

header "Getting the table contents"
(set -x
$sandbox_dir/use test -e "SELECT * FROM t1"
)

header "Getting a value from the sandbox into a shell variable"
echo "TABLE_COUNT=\$($sandbox_dir/use test -BN -e \"SELECT COUNT(*) FROM information_schema.tables where table_schema='mysql'\")"
TABLE_COUNT=$($sandbox_dir/use test -BN -e "SELECT COUNT(*) FROM information_schema.tables where table_schema='mysql'")
echo "echo \"The database 'mysql' has \$TABLE_COUNT tables\""
echo "The database 'mysql' has $TABLE_COUNT tables"

