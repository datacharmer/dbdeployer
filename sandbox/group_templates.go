package sandbox

// Templates for group replication

const (
	init_nodes_template string = `#!/bin/sh
{{.Copyright}}
multi_sb={{.SandboxDir}}
{{range .Nodes}}
    user_cmd='reset master;'
    user_cmd="$user_cmd CHANGE MASTER TO MASTER_USER='rsandbox', MASTER_PASSWORD='rsandbox' FOR CHANNEL 'group_replication_recovery';"
	echo "# Node {{.Node}} # $user_cmd"
    $multi_sb/node{{.Node}}/use -u root -e "$user_cmd"
{{end}}
echo ""

BEFORE_START_CMD="SET GLOBAL group_replication_bootstrap_group=ON;"
START_CMD="START GROUP_REPLICATION;"
AFTER_START_CMD="SET GLOBAL group_replication_bootstrap_group=OFF;"
echo "# Node 1 # $BEFORE_START_CMD"
$multi_sb/n1 -e "$BEFORE_START_CMD"
{{ range .Nodes}}
	echo "# Node {{.Node}} # $START_CMD"
	$multi_sb/n{{.Node}} -e "$START_CMD"
	sleep 1
{{end}}
echo "# Node 1 # $AFTER_START_CMD"
$multi_sb/n1 -e "$AFTER_START_CMD"
$multi_sb/check_nodes
`
	check_nodes_template string = `#!/bin/sh
{{.Copyright}}
multi_sb={{.SandboxDir}}

CHECK_NODE="select * from performance_schema.replication_group_members"
{{ range .Nodes}}
	echo "# Node {{.Node}} # $CHECK_NODE"
	$multi_sb/node{{.Node}}/use -t -e "$CHECK_NODE"
	sleep 1
{{end}}
`
)

