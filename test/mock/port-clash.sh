#!/bin/bash
test_dir=$(dirname $0)
cd $test_dir
test_dir=$PWD
exit_code=0

if [ ! -f set-mock.sh ]
then
    echo "set-mock.sh not found in $PWD"
    exit 1
fi

if [ ! -f ../common.sh ]
then
    echo "../common.sh not found"
    exit 1
fi

source ../common.sh
export results_log=$PWD/port-clash.log
source set-mock.sh
export SHOW_CHANGED_PORTS=1
start_timer

pwd
ls -l
echo "HOME:            $HOME"
echo "SANDBOX_HOME :   $SANDBOX_HOME"
echo "SANDBOX_BINARY:  $SANDBOX_BINARY"
echo "SANDBOX_TARBALL: $SANDBOX_TARBALL"

mkdir $HOME
mkdir -p $SANDBOX_BINARY
mkdir $SANDBOX_HOME
mkdir $SANDBOX_TARBALL

run dbdeployer defaults store
run dbdeployer defaults show

versions=(5.0 5.1 5.5 5.6 5.7 8.0)
versions_mysqld_initialize=(5.7 8.0)
versions_mysql_install_db=(5.0 5.1 5.5 5.6)

rev_sparse="18 27 36 45 54 63 72 81 99"
rev_all="$(seq 1 99)"
rev_list=$rev_all
if [ "$1" == "sparse" ]
then
    rev_list=$rev_sparse
fi

for rev in $rev_list
do
    for vers in ${versions[*]}
    do
        version=${vers}.${rev}
        create_mock_tarball $version $SANDBOX_TARBALL
        #create_mock_version $version
        run dbdeployer unpack $SANDBOX_TARBALL/mysql-${version}.tar.gz --unpack-version $version
    done

    run dbdeployer available
    for vers in ${versions_mysql_install_db[*]}
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy single $version
        run dbdeployer deploy multiple $version
        run dbdeployer deploy replication $version
        results "$version"
        right_installer=$(grep mysql_install_db $SANDBOX_HOME/msb_${version_name}/init_db)
        if [ -n "$right_installer" ]
        then
            echo "ok installer"
        else
            echo "NOT OK installer"
            exit_code=1
        fi
    done

    for vers in ${versions_mysqld_initialize[*]}
    do
        version=${vers}.${rev}
        version_name=$(echo $version | tr '.' '_')
        run dbdeployer deploy single $version
        run dbdeployer deploy multiple $version
        run dbdeployer deploy replication $version
        if [[ "$vers" ==  "5.7" && $rev -lt 18 ]]
        then
            echo "skipping group replication for version $version"
        else
            run dbdeployer deploy replication $version --topology=group
            run dbdeployer deploy replication $version --topology=group --single-primary
            run dbdeployer deploy replication $version --topology=all-masters
            run dbdeployer deploy replication $version --topology=fan-in
        fi
        results "$version"
        right_installer1=$(grep mysqld  $SANDBOX_HOME/msb_${version_name}/init_db )
        right_installer2=$(grep initialize-insecure  $SANDBOX_HOME/msb_${version_name}/init_db )
        if [ -n "$right_installer1" -a -n "$right_installer2" ]
        then
            echo "ok installer"
        else
            echo "NOT OK installer"
            exit_code=1
        fi
    done

    if [ "$exit_code" != "0" ]
    then
        exit $exit_code
    fi
done

echo "#Total sandboxes: $(count_catalog)"
echo "#Total sandboxes: $(count_catalog)" >> $results_log
num_ports=$(grep -A10 port $CATALOG | grep '^\s*[0-9]\+' | wc -l)
echo "# Total ports installed: $num_ports"
echo "# Total ports installed: $num_ports" >> $results_log
run dbdeployer delete ALL --skip-confirm

results "After deletion"
cd $test_dir

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer

