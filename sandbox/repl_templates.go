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
{{.SandboxDir}}/node{{.Node}}/use -BN -e "select CONCAT('port: ', @@port) AS port"
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
ReplicationTemplates  = TemplateCollection{
	"init_slaves_template" : TemplateDesc{
			Description: "Initialize slaves after deployment",
			Notes: "Can also be run after calling './clear_all'",
			Contents : init_slaves_template,
		},
	"start_all_template" : TemplateDesc{
		Description: "Starts nodes in replication order (with optional mysqld arguments)",
			Notes: "",
			Contents : start_all_template,
		},
	"restart_all_template" : TemplateDesc{
			Description: "stops all nodes and restarts them (with optional mysqld arguments)",
			Notes: "",
			Contents : restart_all_template,
		},
	"use_all_template" : TemplateDesc{
			Description: "Execute a query for all nodes",
			Notes: "",
			Contents : use_all_template,
		},
	"stop_all_template" : TemplateDesc{
			Description: "Stops all nodes in reverse replication order",
			Notes: "",
			Contents : stop_all_template,
		},
	"send_kill_all_template" : TemplateDesc{
			Description: "Send kill signal to all nodes",
			Notes: "",
			Contents : send_kill_all_template,
		},
	"clear_all_template" : TemplateDesc{
			Description: "Remove data from all nodes",
			Notes: "",
			Contents : clear_all_template,
		},
	"status_all_template" : TemplateDesc{
			Description: "Show status of all nodes",
			Notes: "",
			Contents : status_all_template,
		},
	"check_slaves_template" : TemplateDesc{
			Description: "Checks replication status in master and slaves",
			Notes: "",
			Contents : check_slaves_template,
		},
	"master_template" : TemplateDesc{
			Description: "Runs the MySQL client for the master",
			Notes: "",
			Contents : master_template,
		},
	"slave_template" : TemplateDesc{
			Description: "Runs the MySQL client for a slave",
			Notes: "",
			Contents : slave_template,
		},
	}
)
