#!/usr/bin/env bash
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

mock_dir=mock_dir
if [ -d $mock_dir ]
then
    echo "mock directory "$PWD/$mock_dir" already exists"
    exit 1
fi

mkdir $mock_dir
cd $mock_dir || (echo "error changing directorty to $mock_dir"; exit 1)
mock_dir=$PWD
export HOME=$mock_dir/home
export CATALOG=$HOME/.dbdeployer/sandboxes.json
export SANDBOX_HOME=$HOME/sandboxes
export SANDBOX_BINARY=$HOME/opt/mysql
export SANDBOX_TARBALL=$HOME/downloads
export SLEEP_TIME=0
export SB_MOCKING=1

pwd
ls -l
echo "HOME           : $HOME"
echo "SANDBOX_HOME   : $SANDBOX_HOME"
echo "SANDBOX_BINARY : $SANDBOX_BINARY"
echo "SANDBOX_TARBALL: $SANDBOX_TARBALL"

mkdir $HOME
mkdir -p $SANDBOX_BINARY
mkdir $SANDBOX_HOME
mkdir $SANDBOX_TARBALL

export PS1='[mock \W] $ '

function make_dir {
    dir=$1
    if [ ! -d $dir ]
    then
        mkdir -p $dir
    fi
}

# A mock version is a collection of
# fake MySQL executable that will create
# empty sandboxes with no-op key executables.
# Its purpose is testing the administrative part of
# dbdeployer.
function create_mock_version {
    version_label=$1
    if [ -z "$SANDBOX_BINARY" ]
    then
        echo "SANDBOX_BINARY not set"
        exit 1
    fi
    if [ ! -d "$SANDBOX_BINARY" ]
    then
        echo "$SANDBOX_BINARY not found"
        exit 1
    fi
    OS=$(uname)
    case $OS in
        Darwin)
            OS_extension=dylib
            ;;
        Linux)
            OS_extension=so
            ;;
        *)
            echo "Unhandled operating system $OS"
            exit 1
            ;;
    esac
    make_dir $SANDBOX_BINARY/$version_label
    make_dir $SANDBOX_BINARY/$version_label/bin
    make_dir $SANDBOX_BINARY/$version_label/scripts
    make_dir $SANDBOX_BINARY/$version_label/lib
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/bin/mysqld
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/bin/mysql
    dbdeployer defaults templates show mysqld_safe_mock > $SANDBOX_BINARY/$version_label/bin/mysqld_safe
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/scripts/mysql_install_db
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/lib/libmysqlclient.$OS_extension
    chmod +x $SANDBOX_BINARY/$version_label/bin/*
    chmod +x $SANDBOX_BINARY/$version_label/scripts/*
    echo mysql > $SANDBOX_BINARY/$version_label/FLAVOR
}

function create_mock_ndb_version {
    version_label=$1
    create_mock_version $version_label

    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/bin/ndbd
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/bin/ndb_mgm
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/bin/ndb_mgmd
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/bin/ndbd
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/bin/ndbmtd
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/lib/ndb_engine.so
    chmod +x $SANDBOX_BINARY/$version_label/bin/*
    echo ndb > $SANDBOX_BINARY/$version_label/FLAVOR
}

function create_mock_pxc_version {
    version_label=$1
    create_mock_version $version_label

    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/bin/garbd
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/lib/libgalera_smm.a
    dbdeployer defaults templates show no_op_mock > $SANDBOX_BINARY/$version_label/lib/libperconaserverclient.a
    chmod +x $SANDBOX_BINARY/$version_label/bin/*
    echo pxc > $SANDBOX_BINARY/$version_label/FLAVOR
}

function create_mock_tidb_version {
    version_label=$1
    make_dir $SANDBOX_BINARY/$version_label
    make_dir $SANDBOX_BINARY/$version_label/bin

    dbdeployer defaults templates show tidb_mock > $SANDBOX_BINARY/$version_label/bin/tidb-server
    chmod +x $SANDBOX_BINARY/$version_label/bin/*
    echo tidb > $SANDBOX_BINARY/$version_label/FLAVOR
}

# a mock tarball is a tarball that contains mock MySQL executables
# for the purpose of testing "dbdeployer unpack"
function create_mock_tarball {
    version_label=$1
    tarball_dir=$2
    save_sandbox_binary=$SANDBOX_BINARY
    # Changes SANDBOX_BINARY so that create_mock_version
    # will create the mock directory in the tarball place.
    SANDBOX_BINARY=$tarball_dir
    create_mock_version $version_label
    cd $tarball_dir || (echo "error changing directory to $tarball_dir"; exit 1)
    if [ ! -d $version_label ]
    then
        echo "$version_label not found in $PWD"
        exit 1
    fi
    # Change the name of the directory, so that it's different
    # from the ultimate destination
    mv $version_label mysql-${version_label}
    tar -c mysql-${version_label} | gzip -c > mysql-${version_label}.tar.gz
    rm -rf mysql-$version_label
    cd - || (echo "error returning to previous directory"; exit 1 )
    export SANDBOX_BINARY=$save_sandbox_binary
}

