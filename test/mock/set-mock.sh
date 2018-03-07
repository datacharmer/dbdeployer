mock_dir=mock_dir
if [ -d $mock_dir ]
then
    echo "mock directory "$PWD/$mock_dir" already exists"
    exit 1
fi

mkdir $mock_dir
cd $mock_dir
mock_dir=$PWD
export HOME=$mock_dir/home
export CATALOG=$HOME/.dbdeployer/sandboxes.json
export SANDBOX_HOME=$HOME/sandboxes
export SANDBOX_BINARY=$HOME/opt/mysql
export SLEEP_TIME=0

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
    mkdir $SANDBOX_BINARY/$version_label
    mkdir $SANDBOX_BINARY/$version_label/bin
    mkdir $SANDBOX_BINARY/$version_label/scripts
    dbdeployer defaults templates show no_op_mock_template > $SANDBOX_BINARY/$version_label/bin/mysqld
    dbdeployer defaults templates show no_op_mock_template > $SANDBOX_BINARY/$version_label/bin/mysql
    dbdeployer defaults templates show mysqld_safe_mock_template > $SANDBOX_BINARY/$version_label/bin/mysqld_safe
    dbdeployer defaults templates show no_op_mock_template > $SANDBOX_BINARY/$version_label/scripts/mysql_install_db
    chmod +x $SANDBOX_BINARY/$version_label/bin/*
    chmod +x $SANDBOX_BINARY/$version_label/scripts/*
}

