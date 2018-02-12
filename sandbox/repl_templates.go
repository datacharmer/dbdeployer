package sandbox

// Templates for replication

var (
	init_slaves_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}

# Don't use directly.
# This script is called by 'start_all' when needed

{{ range .Slaves }}
echo "initializing slave {{.Node}}"
if [ ! -f needs_initialization ]
then
	# First run: root is running without password
	export NOPASSWORD=1
fi
echo 'CHANGE MASTER TO  master_host="127.0.0.1",  master_port={{.MasterPort}},  master_user="{{.RplUser}}",  master_password="{{.RplPassword}}" ' | {{.SandboxDir}}/node{{.Node}}/use -u root
{{.SandboxDir}}/node{{.Node}}/use -u root -e 'START SLAVE'

{{end}}
`
	start_all_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "start"' on {{.SandboxDir}}
echo 'executing "start" on master'
{{.SandboxDir}}/master/start "$@"
{{ range .Slaves }}
echo 'executing "start" on slave {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/start "$@"
{{end}}
if [ -f {{.SandboxDir}}/needs_initialization ] 
then
	{{.SandboxDir}}/initialize_slaves
    rm -f {{.SandboxDir}}/needs_initialization
fi
`
	restart_all_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
{{.SandboxDir}}/stop_all
{{.SandboxDir}}/start_all "$@"
`
	use_all_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
if [ "$1" = "" ]
then
  echo "syntax: $0 command"
  exit 1
fi

echo "# master  " 
echo "$@" | {{.SandboxDir}}/master/use  

{{range .Slaves}}
echo "# server: {{.Node}} " 
echo "$@" | {{.SandboxDir}}/node{{.Node}}/use $MYCLIENT_OPTIONS 
{{end}} 
`
	stop_all_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "stop"' on {{.SandboxDir}}
{{ range .Slaves }}
echo 'executing "stop" on slave {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/stop "$@"
{{end}}
echo 'executing "stop" on master'
{{.SandboxDir}}/master/stop "$@"
`
	send_kill_all_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "send_kill"' on {{.SandboxDir}}
{{ range .Slaves }}
echo 'executing "send_kill" on slave {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/send_kill "$@"
{{end}}
echo 'executing "send_kill" on master'
{{.SandboxDir}}/master/send_kill "$@"
`
	clear_all_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "clear"' on {{.SandboxDir}}
{{range .Slaves}}
echo 'executing "clear" on slave {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/clear "$@"
{{end}}
echo 'executing "clear" on master'
{{.SandboxDir}}/master/clear "$@"
date > {{.SandboxDir}}/needs_initialization
`
	status_all_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo "REPLICATION  {{.SandboxDir}}"
{{.SandboxDir}}/master/status
{{.SandboxDir}}/master/use -BN -e "select CONCAT('port: ', @@port) AS port"
{{ range .Slaves }}
{{.SandboxDir}}/node{{.Node}}/status 
# {{.SandboxDir}}/node{{.Node}}/use -BN -e "select CONCAT('port: ', @@port) AS port"
{{end}}
`
	test_sb_all_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "test_sb"' on {{.SandboxDir}}
echo 'executing "test_sb" on master'
{{.SandboxDir}}/master/test_sb "$@"
exit_code=$?
if [ "$exit_code" != "0" ] ; then exit $exit_code ; fi
{{ range .Slaves }}
echo 'executing "test_sb" on slave {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/test_sb "$@"
exit_code=$?
if [ "$exit_code" != "0" ] ; then exit $exit_code ; fi
{{end}}
`

	check_slaves_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo "master"
{{.SandboxDir}}/master/use -BN -e "select CONCAT('port: ', @@port) AS port"
{{.SandboxDir}}/master/use -e 'show master status\G' | grep "File\|Position\|Executed"
{{ range .Slaves }}
echo "Slave{{.Node}}"
{{.SandboxDir}}/node{{.Node}}/use -BN -e "select CONCAT('port: ', @@port) AS port"
{{.SandboxDir}}/node{{.Node}}/use -e 'show slave status\G' | grep "\(Running:\|Master_Log_Pos\|\<Master_Log_File\|Retrieved\|Executed\)"
{{end}}
`
	master_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}

{{.SandboxDir}}/master/use "$@"
`
	slave_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}

{{.SandboxDir}}/node{{.Node}}/use "$@"
`
	test_replication_template string = `#!/bin/bash
{{.Copyright}}
# Template : {{.TemplateName}}
SBDIR={{.SandboxDir}}
cd $SBDIR

if [ -x ./m ]
then
    MASTER=./m
elif [ -x ./n1 ]
then
    MASTER=./n1
else
    echo "# No master found"
    exit 1
fi
$MASTER -e 'create schema if not exists test'
$MASTER test -e 'drop table if exists t1'
$MASTER test -e 'create table t1 (i int not null primary key, msg varchar(50), d date, t time, dt datetime, ts timestamp)'
#$MASTER test -e "insert into t1 values (1, 'test sandbox 1', '2015-07-16', '11:23:40','2015-07-17 12:34:50', null)"
#$MASTER test -e "insert into t1 values (2, 'test sandbox 2', '2015-07-17', '11:23:41','2015-07-17 12:34:51', null)"
for N in $(seq -f '%02.0f' 1 20)
do
    #echo "$MASTER test -e \"insert into t1 values ($N, 'test sandbox $N', '2015-07-$N', '11:23:$N','2015-07-17 12:34:$N', null)\""
    $MASTER test -e "insert into t1 values ($N, 'test sandbox $N', '2015-07-$N', '11:23:$N','2015-07-17 12:34:$N', null)"
done
sleep 0.5
MASTER_RECS=$($MASTER -BN -e 'select count(*) from test.t1')

master_status=master_status$$
slave_status=slave_status$$
$MASTER -e 'show master status\G' > $master_status
master_binlog=$(grep 'File:' $master_status | awk '{print $2}' )
master_pos=$(grep 'Position:' $master_status | awk '{print $2}' )
echo "# Master log: $master_binlog - Position: $master_pos - Rows: $MASTER_RECS"
rm -f $master_status

FAILED=0
PASSED=0

function ok_equal
{
    fact="$1"
    expected="$2"
    msg="$3"
    if [ "$fact" == "$expected" ]
    then
        echo -n "ok"
        PASSED=$(($PASSED+1))
    else
        echo -n "not ok - (expected: <$expected> found: <$fact>) "
        FAILED=$(($FAILED+1))
    fi
    echo " - $msg"
}

function test_summary
{
    TESTS=$(($PASSED+$FAILED))
    if [ -n "$TAP_TEST" ]
    then
        echo "1..$TESTS"
    else
        PERCENT_PASSED=$(($PASSED/$TESTS*100))
        PERCENT_FAILED=$(($FAILED/$TESTS*100))
        printf "# TESTS : %5d\n" $TESTS
        printf "# FAILED: %5d (%5.1f%%)\n" $FAILED $PERCENT_FAILED
        printf "# PASSED: %5d (%5.1f%%)\n" $PASSED $PERCENT_PASSED
    fi
    exit_code=0
    if [ "$FAILED" != "0" ]
    then
        exit_code=1
    fi
    echo "# exit code: $exit_code"
    exit $exit_code
}

for SLAVE_N in 1 2 3 4 5 6 7 8 9
do
    N=$(($SLAVE_N+1)) 
    unset SLAVE
    if [ -x ./s$SLAVE_N ]
    then
        SLAVE=./s$SLAVE_N
    elif [ -x ./n$N ]
    then
        SLAVE=./n$N
    fi
    if [ -n "$SLAVE" ]
    then
        echo "# Testing slave #$SLAVE_N"
        if [ -f initialize_nodes ]
        then
            sleep 3
        else
            S_READY=$($SLAVE -BN -e "select master_pos_wait('$master_binlog', $master_pos,60)")
            # master_pos_wait can return 0 or a positive number for successful replication
            # Any result that is not NULL or -1 is acceptable
            if [ "$S_READY" != "-1" -a "$S_READY" != "NULL" ]
            then
                S_READY=0
            fi
            ok_equal $S_READY 0 "Slave #$SLAVE_N acknowledged reception of transactions from master" 
        fi
		if [ -f initialize_slaves ]
		then
			$SLAVE -e 'show slave status\G' > $slave_status
			IO_RUNNING=$(grep -w Slave_IO_Running $slave_status | awk '{print $2}')
			ok_equal $IO_RUNNING Yes "Slave #$SLAVE_N IO thread is running"
			SQL_RUNNING=$(grep -w Slave_IO_Running $slave_status | awk '{print $2}')
			ok_equal $SQL_RUNNING Yes "Slave #$SLAVE_N SQL thread is running"
			rm -f $slave_status 
		fi
        [ $FAILED == 0 ] || exit 1

        T1_EXISTS=$($SLAVE -BN -e 'show tables from test like "t1"')
        ok_equal $T1_EXISTS t1 "Table t1 found on slave #$SLAVE_N"
        T1_RECS=$($SLAVE -BN -e 'select count(*) from test.t1')
        ok_equal $T1_RECS $MASTER_RECS "Table t1 has $MASTER_RECS rows on #$SLAVE_N"
    fi
done
test_summary

`

	ReplicationTemplates = TemplateCollection{
		"init_slaves_template": TemplateDesc{
			Description: "Initialize slaves after deployment",
			Notes:       "Can also be run after calling './clear_all'",
			Contents:    init_slaves_template,
		},
		"start_all_template": TemplateDesc{
			Description: "Starts nodes in replication order (with optional mysqld arguments)",
			Notes:       "",
			Contents:    start_all_template,
		},
		"restart_all_template": TemplateDesc{
			Description: "stops all nodes and restarts them (with optional mysqld arguments)",
			Notes:       "",
			Contents:    restart_all_template,
		},
		"use_all_template": TemplateDesc{
			Description: "Execute a query for all nodes",
			Notes:       "",
			Contents:    use_all_template,
		},
		"stop_all_template": TemplateDesc{
			Description: "Stops all nodes in reverse replication order",
			Notes:       "",
			Contents:    stop_all_template,
		},
		"send_kill_all_template": TemplateDesc{
			Description: "Send kill signal to all nodes",
			Notes:       "",
			Contents:    send_kill_all_template,
		},
		"clear_all_template": TemplateDesc{
			Description: "Remove data from all nodes",
			Notes:       "",
			Contents:    clear_all_template,
		},
		"status_all_template": TemplateDesc{
			Description: "Show status of all nodes",
			Notes:       "",
			Contents:    status_all_template,
		},
		"test_sb_all_template": TemplateDesc{
			Description: "Run sb test on all nodes",
			Notes:       "",
			Contents:    test_sb_all_template,
		},
		"test_replication_template": TemplateDesc{
			Description: "Tests replication flow",
			Notes:       "",
			Contents:    test_replication_template,
		},
		"check_slaves_template": TemplateDesc{
			Description: "Checks replication status in master and slaves",
			Notes:       "",
			Contents:    check_slaves_template,
		},
		"master_template": TemplateDesc{
			Description: "Runs the MySQL client for the master",
			Notes:       "",
			Contents:    master_template,
		},
		"slave_template": TemplateDesc{
			Description: "Runs the MySQL client for a slave",
			Notes:       "",
			Contents:    slave_template,
		},
	}
)
