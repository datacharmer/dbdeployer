// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sandbox

import (
	"fmt"
	"os"
	"regexp"
)

var (
	tidbInitTemplate string = `#!/bin/bash
{{.Copyright}}
# Generated by dbdeployer {{.AppVersion}} using {{.TemplateName}} on {{.DateTime}}
export SBDIR="{{.SandboxDir}}"
export BASEDIR={{.Basedir}}
export DATADIR=$SBDIR/data

mkdir -p $DATADIR

cat << EOF > $SBDIR/tidb.toml
# TiDB Configuration.
# TiDB server host.
host = "127.0.0.1"
# tidb server advertise IP.
advertise-address = ""
# TiDB server port.
port = {{.Port}} 
# Registered store name, [tikv, mocktikv]
store = "mocktikv"
# TiDB storage path.
path = "$DATADIR"
# The socket file to use for connection.
socket = "{{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock"
# Run ddl worker on this tidb-server.
run-ddl = true
# Schema lease duration, very dangerous to change only if you know what you do.
lease = "45s"
# When create table, split a separated region for it. It is recommended to
# turn off this option if there will be a large number of tables created.
split-table = true
# The limit of concurrent executed sessions.
token-limit = 1000
# Only print a log when out of memory quota.
# Valid options: ["log", "cancel"]
oom-action = "log"
# Set the memory quota for a query in bytes. Default: 32GB
mem-quota-query = 34359738368
# Enable coprocessor streaming.
enable-streaming = false
# Set system variable 'lower_case_table_names'
lower-case-table-names = 2
# Make "kill query" behavior compatible with MySQL. It's not recommend to
# turn on this option when TiDB server is behind a proxy.
compatible-kill-query = false
[log]
# Log level: debug, info, warn, error, fatal.
level = "info"
# Log format, one of json, text, console.
format = "text"
# Disable automatic timestamp in output
disable-timestamp = false
# Stores slow query log into separated files.
slow-query-file = "$DATADIR/slow.log"
# Queries with execution time greater than this value will be logged. (Milliseconds)
slow-threshold = 300
# Queries with internal result greater than this value will be logged.
expensive-threshold = 10000
# Maximum query length recorded in log.
query-log-max-len = 2048
# File logging.
[log.file]
# Log file name.
filename = "$DATADIR/msandbox.err"
# Max log file size in MB (upper limit to 4096MB).
max-size = 300
# Max log file keep days. No clean up by default.
max-days = 0
# Maximum number of old log files to retain. No clean up by default.
max-backups = 0
# Rotate log by day
log-rotate = true
[security]
# Path of file that contains list of trusted SSL CAs for connection with mysql client.
ssl-ca = ""
# Path of file that contains X509 certificate in PEM format for connection with mysql client.
ssl-cert = ""
# Path of file that contains X509 key in PEM format for connection with mysql client.
ssl-key = ""
# Path of file that contains list of trusted SSL CAs for connection with cluster components.
cluster-ssl-ca = ""
# Path of file that contains X509 certificate in PEM format for connection with cluster components.
cluster-ssl-cert = ""
# Path of file that contains X509 key in PEM format for connection with cluster components.
cluster-ssl-key = ""
[status]
# If enable status report HTTP service.
report-status = false 
# TiDB status port.
status-port = 10080
# Prometheus pushgateway address, leaves it empty will disable prometheus push.
metrics-addr = ""
# Prometheus client push interval in second, set \"0\" to disable prometheus push.
metrics-interval = 15
[performance]
# Max CPUs to use, 0 use number of CPUs in the machine.
max-procs = 0
# Max memory size to use, 0 use the total usable memory in the machine.
max-memory = 0
# StmtCountLimit limits the max count of statement inside a transaction.
stmt-count-limit = 5000
# Set keep alive option for tcp connection.
tcp-keep-alive = true
# Whether support cartesian product.
cross-join = true
# Stats lease duration, which influences the time of analyze and stats load.
stats-lease = "3s"
# Run auto analyze worker on this tidb-server.
run-auto-analyze = true
# Probability to use the query feedback to update stats, 0 or 1 for always false/true.
feedback-probability = 0.05
# The max number of query feedback that cache in memory.
query-feedback-limit = 1024
# Pseudo stats will be used if the ratio between the modify count and
# row count in statistics of a table is greater than it.
pseudo-estimate-ratio = 0.8
# Force the priority of all statements in a specified priority.
# The value could be "NO_PRIORITY", "LOW_PRIORITY", "HIGH_PRIORITY" or "DELAYED".
force-priority = "NO_PRIORITY"
[proxy-protocol]
# PROXY protocol acceptable client networks.
# Empty string means disable PROXY protocol, * means all networks.
networks = ""
# PROXY protocol header read timeout, unit is second
header-timeout = 5
[prepared-plan-cache]
enabled = false
capacity = 100
memory-guard-ratio = 0.1
[opentracing]
# Enable opentracing.
enable = false
# Whether to enable the rpc metrics.
rpc-metrics = false
[opentracing.sampler]
# Type specifies the type of the sampler: const, probabilistic, rateLimiting, or remote
type = "const"
# Param is a value passed to the sampler.
# Valid values for Param field are:
# - for "const" sampler, 0 or 1 for always false/true respectively
# - for "probabilistic" sampler, a probability between 0 and 1
# - for "rateLimiting" sampler, the number of spans per second
# - for "remote" sampler, param is the same as for "probabilistic"
# and indicates the initial sampling rate before the actual one
# is received from the mothership
param = 1.0
# SamplingServerURL is the address of jaeger-agent's HTTP sampling server
sampling-server-url = ""
# MaxOperations is the maximum number of operations that the sampler
# will keep track of. If an operation is not tracked, a default probabilistic
# sampler will be used rather than the per operation specific sampler.
max-operations = 0
# SamplingRefreshInterval controls how often the remotely controlled sampler will poll
# jaeger-agent for the appropriate sampling strategy.
sampling-refresh-interval = 0
[opentracing.reporter]
# QueueSize controls how many spans the reporter can keep in memory before it starts dropping
# new spans. The queue is continuously drained by a background go-routine, as fast as spans
# can be sent out of process.
queue-size = 0
# BufferFlushInterval controls how often the buffer is force-flushed, even if it's not full.
# It is generally not useful, as it only matters for very low traffic services.
buffer-flush-interval = 0
# LogSpans, when true, enables LoggingReporter that runs in parallel with the main reporter
# and logs all submitted spans. Main Configuration.Logger must be initialized in the code
# for this option to have any effect.
log-spans = false
#  LocalAgentHostPort instructs reporter to send spans to jaeger-agent at this address
local-agent-host-port = ""
[tikv-client]
# Max gRPC connections that will be established with each tikv-server.
grpc-connection-count = 16
# After a duration of this time in seconds if the client doesn't see any activity it pings
# the server to see if the transport is still alive.
grpc-keepalive-time = 10
# After having pinged for keepalive check, the client waits for a duration of Timeout in seconds
# and if no activity is seen even after that the connection is closed.
grpc-keepalive-timeout = 3
# Max time for commit command, must be twice bigger than raft election timeout.
commit-timeout = "41s"
# The max time a Txn may use (in seconds) from its startTS to commitTS.
# We use it to guarantee GC worker will not influence any active txn. Please make sure that this
# value is less than gc_life_time - 10s.
max-txn-time-use = 590
# Max batch size in gRPC.
max-batch-size = 128
# Overload threshold of TiKV.
overload-threshold = 200
# Max batch wait time in nanosecond to avoid waiting too long. 0 means disable this feature.
max-batch-wait-time = 0
# Batch wait size, to avoid waiting too long.
batch-wait-size = 8
[txn-local-latches]
# Enable local latches for transactions. Enable it when
# there are lots of conflicts between transactions.
enabled = true
capacity = 2048000
[binlog]
# enable to write binlog.
enable = false
# WriteTimeout specifies how long it will wait for writing binlog to pump.
write-timeout = "15s"
# If IgnoreError is true, when writing binlog meets error, TiDB would stop writting binlog,
# but still provide service.
ignore-error = false
# use socket file to write binlog, for compatible with kafka version tidb-binlog.
binlog-socket = ""
# check mb4 value in utf8 is used to control whether to check the mb4 characters when the charset is utf8.
check-mb4-value-in-utf8 = true
EOF

exit_code=$?
if [ "$exit_code" == "0" ]
then
	echo "Database installed in $SBDIR"
else
	echo "Error installing database in $SBDIR"
fi
`
	tidbMyCnfTemplate string = `{{.Copyright}}
# Generated by dbdeployer {{.AppVersion}} using {{.TemplateName}} on {{.DateTime}}
[mysql]
prompt='{{.Prompt}} [\h:{{.Port}}] {\u} (\d) > '
#

[client]
user               = {{.DbUser}} 
password           = {{.DbPassword}} 
port               = {{.Port}}
socket             = {{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock

[mysqld]
# options here will be ignored.  See tidb.toml
`

	tidbStartTemplate string = `#!/bin/bash
{{.Copyright}}
# Generated by dbdeployer {{.AppVersion}} using {{.TemplateName}} on {{.DateTime}}
source {{.SandboxDir}}/sb_include

if [ -e $PIDFILE ]
then
  PID=$(< "$PIDFILE")
  if kill -0 "$PID" > /dev/null 2>&1
  then
    echo "sandbox server already started (found pid file $PIDFILE)"
    echo " sandbox server started"
    exit 0
  else
    # Server exited unclean
    rm -f $PIDFILE
    rm -f {{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock
  fi
fi

echo "Starting server"

$BASEDIR/bin/tidb-server -config $SBDIR/tidb.toml > $SBDIR/data/tidb.log 2>&1 &
echo $! > $PIDFILE

TIMEOUT=180
ATTEMPTS=1
while [ ! -S {{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock ]
do
  ATTEMPTS=$(( $ATTEMPTS + 1 ))
  echo -n "."
  if [ $ATTEMPTS = $TIMEOUT ]
  then
    break
  fi
  sleep $SLEEP_TIME
done

if [ -S {{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock ]
then
  echo " sandbox server started"
else
  echo " sandbox server not started yet"
  exit 1
fi
`

	tidbStopTemplate string = `#!/bin/bash
{{.Copyright}}
# Generated by dbdeployer {{.AppVersion}} using {{.TemplateName}} on {{.DateTime}}
source {{.SandboxDir}}/sb_include
export LD_LIBRARY_PATH=$CLIENT_LD_LIBRARY_PATH

MYSQL_ADMIN="$CLIENT_BASEDIR/bin/mysqladmin"

# Ideally TiDB will support SHUTDOWN soon
# and then the default template can be used!
# https://github.com/pingcap/tidb/issues/5046

$SBDIR/send_kill

`
	tidbSendKillTemplate string = `#!/bin/bash
{{.Copyright}}
# Generated by dbdeployer {{.AppVersion}} using {{.TemplateName}} on {{.DateTime}}
source {{.SandboxDir}}/sb_include

if [ -e $PIDFILE ]
then
  PID=$(< "$PIDFILE")
  if kill -0 "$PID" > /dev/null 2>&1
  then
    echo "stop $SBDIR"
    kill -9 $PID
    SOCKET={{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock
    if [ -e $SOCKET ]
    then
      rm -f $SOCKET
    fi
    rm -f $PIDFILE
  else
    # Server already shutdown, removing stale pid-file
    rm -f $PIDFILE
  fi
fi
`

	tidbGrantsTemplate string = `
# Generated by dbdeployer {{.AppVersion}} using {{.TemplateName}} on {{.DateTime}}
use mysql;
set password='{{.DbPassword}}';

create user {{.DbUser}}@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
grant all on *.* to {{.DbUser}}@'{{.RemoteAccess}}' ;

create user {{.DbUser}}@'localhost' identified by '{{.DbPassword}}';
grant all on *.* to {{.DbUser}}@'localhost';

create user msandbox_rw@'localhost' identified by '{{.DbPassword}}';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE
    on *.* to msandbox_rw@'localhost';

create user msandbox_rw@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE
    on *.* to msandbox_rw@'{{.RemoteAccess}}';

create user msandbox_ro@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
create user msandbox_ro@'localhost' identified by '{{.DbPassword}}';
create user {{.RplUser}}@'{{.RemoteAccess}}' identified by '{{.RplPassword}}';
grant SELECT,EXECUTE on *.* to msandbox_ro@'{{.RemoteAccess}}';
grant SELECT,EXECUTE on *.* to msandbox_ro@'localhost';
grant REPLICATION SLAVE on *.* to {{.RplUser}}@'{{.RemoteAccess}}';
FLUSH PRIVILEGES;
create schema if not exists test;`

	tidbAfterStartTemplate string = `#!/bin/bash
{{.Copyright}}
# Generated by dbdeployer {{.AppVersion}} using {{.TemplateName}} on {{.DateTime}}
source {{.SandboxDir}}/sb_include

# Removes the scripts that we know for sure won't be used with TiDB
#
scripts_to_delete=(add_option show_log show_binlog show_relaylog)
for script in ${scripts_to_delete[*]}
do
    if [ -f $SBDIR/$script ]
    then
        rm -f $SBDIR/$script
    fi
done
exit 0
`
)

const tidbPrefix = "tidb_"

// Every template in this collection will replace the corresponding one in SingleTemplates
// when the flavor is "tidb"
var TidbTemplates = TemplateCollection{
	"tidb_init_db_template": TemplateDesc{
		Description: "Initialization template for the TiDB server",
		Notes:       "This should normally run only once",
		Contents:    tidbInitTemplate,
	},
	"tidb_my_cnf_template": TemplateDesc{
		Description: "Default options file for a TiDB sandbox",
		Notes:       "",
		Contents:    tidbMyCnfTemplate,
	},
	"tidb_start_template": TemplateDesc{
		Description: "Stops a database in a single TiDB sandbox",
		Notes:       "",
		Contents:    tidbStartTemplate,
	},
	"tidb_stop_template": TemplateDesc{
		Description: "Stops a database in a single TiDB sandbox",
		Notes:       "",
		Contents:    tidbStopTemplate,
	},
	"tidb_send_kill_template": TemplateDesc{
		Description: "Sends a kill signal to the TiDB database",
		Notes:       "",
		Contents:    tidbSendKillTemplate,
	},
	"tidb_grants_template5x": TemplateDesc{
		Description: "Grants for TiDB sandboxes",
		Notes:       "",
		Contents:    tidbGrantsTemplate,
	},
	"tidb_after_start_template": TemplateDesc{
		Description: "commands to run after the database started",
		Notes:       "This script does nothing. You can change it and reuse through --use-template",
		Contents:    tidbAfterStartTemplate,
	},
}

func init() {
	// Makes sure that all template names in TidbTemplates start with 'tidb_'
	// This is an important assumption that will be used in sandbox.go
	// to replace templates for "tidb" flavor
	re := regexp.MustCompile(`^` + tidbPrefix)
	for name, _ := range TidbTemplates {
		if !re.MatchString(name) {
			fmt.Printf("found template name '%s' that does not start with '%s'\n", name, tidbPrefix)
			os.Exit(1)
		}
	}
}
