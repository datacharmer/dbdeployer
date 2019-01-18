#
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
if [ -z "$SANDBOX_HOME" ]
then
    export SANDBOX_HOME=$HOME/sandboxes
fi

if [ -z "$SANDBOX_BINARY" ]
then
    export SANDBOX_BINARY=$HOME/opt/mysql
fi

function check_version {
    wanted_version=$1
    check_upgrade=$2
    if [ ! -d $SANDBOX_BINARY/$wanted_version ]
    then
        echo "Directory $SANDBOX_BINARY/$wanted_version not found"
        echo "To install the binaries, use: "
        echo "    dbdeployer unpack mysql-$version-YOUR-OPERATING-SYSTEM.tar.gz"
        exit 1
    fi
    if [ -z "$check_upgrade" ]
    then
        return
    fi
    if [ ! -x $SANDBOX_BINARY/$wanted_version/bin/mysql_upgrade ]
    then
        echo "mysql_upgrade not found in $wanted_version"
        exit 1
    fi
}

dash_line="# ----------------------------------------------------------------------------"
star_line="# ****************************************************************************"
hash_line="# ############################################################################"

function header {
    msg="$1"
    msg2="$2"
    msg3="$3"
    echo ""
    echo "$star_line"
    echo "# $msg"
    if [ -n "$msg2" ] ; then echo "# $msg2" ; fi
    if [ -n "$msg3" ] ; then echo "# $msg3" ; fi
    echo "$star_line"
}
