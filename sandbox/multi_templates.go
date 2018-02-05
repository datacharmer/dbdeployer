package sandbox

// Templates for multiple sandboxes

var (
	start_multi_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "start"' on {{.SandboxDir}}
{{ range .Nodes }}
echo 'executing "start" on node {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/start "$@"
{{end}}
`
	restart_multi_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
{{.SandboxDir}}/stop_all
{{.SandboxDir}}/start_all "$@"
`
	use_multi_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
if [ "$1" = "" ]
then
  echo "syntax: $0 command"
  exit 1
fi

{{range .Nodes}}
echo "# server: {{.Node}} " 
echo "$@" | {{.SandboxDir}}/node{{.Node}}/use $MYCLIENT_OPTIONS 
{{end}} 
`
	stop_multi_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "stop"' on {{.SandboxDir}}
{{ range .Nodes }}
echo 'executing "stop" on node {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/stop "$@"
{{end}}
`
	send_kill_multi_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "send_kill"' on {{.SandboxDir}}
{{ range .Nodes }}
echo 'executing "send_kill" on node {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/send_kill "$@"
{{end}}
`
	clear_multi_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo '# executing "clear"' on {{.SandboxDir}}
{{range .Nodes}}
echo 'executing "clear" on node {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/clear "$@"
{{end}}
`
	status_multi_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}
echo "MULTIPLE  {{.SandboxDir}}"
{{ range .Nodes }}
{{.SandboxDir}}/node{{.Node}}/status 
{{.SandboxDir}}/node{{.Node}}/use -BN -e "select CONCAT('port: ', @@port) AS port"
{{end}}
`
	node_template string = `#!/bin/sh
{{.Copyright}}
# Template : {{.TemplateName}}

{{.SandboxDir}}/node{{.Node}}/use "$@"
`

MultipleTemplates  = TemplateCollection{
	"start_multi_template" : TemplateDesc{
			Description: "Starts all nodes (with optional mysqld arguments)",
			Notes: "",
			Contents : start_multi_template,
		},
	"restart_multi_template" : TemplateDesc{
			Description: "Restarts all nodes (with optional mysqld arguments)",
			Notes: "",
			Contents : restart_multi_template,
		},
	"use_multi_template" : TemplateDesc{
			Description: "Runs the same SQL query in all nodes",
			Notes: "",
			Contents : use_multi_template,
		},
	"stop_multi_template" : TemplateDesc{
			Description: "Stops all nodes",
			Notes: "",
			Contents : stop_multi_template,
		},
	"send_kill_multi_template" : TemplateDesc{
			Description: "Sends kill signal to all nodes",
			Notes: "",
			Contents : send_kill_multi_template,
		},
	"clear_multi_template" : TemplateDesc{
			Description: "Removes data from all nodes",
			Notes: "",
			Contents : clear_multi_template,
		},
	"status_multi_template" : TemplateDesc{
			Description: "Shows status for all nodes",
			Notes: "",
			Contents : status_multi_template,
		},
	"node_template" : TemplateDesc{
			Description: "Runs the MySQL client for a given node",
			Notes: "",
			Contents : node_template,
		},
}
)
