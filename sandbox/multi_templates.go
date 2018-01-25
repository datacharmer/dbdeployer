package sandbox

// Templates for multiple sandboxes

const (
	start_multi_template string = `#!/bin/sh
{{.Copyright}}
echo '# executing "start"' on {{.SandboxDir}}
{{ range .Nodes }}
echo 'executing "start" on node {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/start "$@"
{{end}}
`
	restart_multi_template string = `#!/bin/sh
{{.Copyright}}
{{.SandboxDir}}/stop_all
{{.SandboxDir}}/start_all "$@"
`
	use_multi_template string = `#!/bin/sh
{{.Copyright}}
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
echo '# executing "stop"' on {{.SandboxDir}}
{{ range .Nodes }}
echo 'executing "stop" on node {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/stop "$@"
{{end}}
`
	send_kill_multi_template string = `#!/bin/sh
{{.Copyright}}
echo '# executing "send_kill"' on {{.SandboxDir}}
{{ range .Nodes }}
echo 'executing "send_kill" on node {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/send_kill "$@"
{{end}}
`
	clear_multi_template string = `#!/bin/sh
{{.Copyright}}
echo '# executing "clear"' on {{.SandboxDir}}
{{range .Nodes}}
echo 'executing "clear" on node {{.Node}}'
{{.SandboxDir}}/node{{.Node}}/clear "$@"
{{end}}
`
	status_multi_template string = `#!/bin/sh
{{.Copyright}}
echo "MULTIPLE  {{.SandboxDir}}"
{{ range .Nodes }}
{{.SandboxDir}}/node{{.Node}}/status 
{{.SandboxDir}}/node{{.Node}}/use -BN -e "select CONCAT('port: ', @@port) AS port"
{{end}}
`
	node_template string = `#!/bin/sh
{{.Copyright}}

{{.SandboxDir}}/node{{.Node}}/use "$@"
`
)
