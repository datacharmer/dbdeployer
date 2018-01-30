package sandbox

import (
	"github.com/datacharmer/dbdeployer/common"
	"fmt"
	"os"
)

const (
	GroupReplOptions string = `
binlog_checksum=NONE
log_slave_updates=ON
plugin-load=group_replication.so
group_replication=FORCE_PLUS_PERMANENT
group_replication_start_on_boot=OFF
group_replication_bootstrap_group=OFF
transaction_write_set_extraction=XXHASH64
report-host=127.0.0.1
loose-group_replication_group_name="aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
loose-group-replication-single-primary-mode=off
`
)

func CreateGroupReplication(sdef SandboxDef, origin string, nodes int) {
	// fmt.Println("Group replication not implemented yet")
	vList := VersionToList(sdef.Version)
	rev := vList[2]
	base_port := sdef.Port + GroupReplicationBasePort + (rev * 100)
	base_server_id := 0
	if nodes < 3 {
		fmt.Println("Can't run group replication with less than 3 nodes")
		os.Exit(1)
	}
	var data common.Smap = common.Smap{
		"Copyright":  Copyright,
		"SandboxDir": sdef.SandboxDir,
		"Nodes":     []common.Smap{},
	}
	base_group_port := base_port + 100
	connection_string := ""
	for i := 0; i < nodes; i++ {
		group_port := base_group_port + i + 1
		if connection_string != "" {
			connection_string += ","
		}
		connection_string += fmt.Sprintf("127.0.0.1:%d", group_port)
	}
	for i := 1; i <= nodes; i++ {
		group_port := base_group_port + i
		data["Nodes"] = append(data["Nodes"].([]common.Smap), common.Smap{
			"Node": i,
			"SandboxDir":  sdef.SandboxDir,
			"RplUser":     sdef.RplUser,
			"RplPassword": sdef.RplPassword})

		sdef.DirName = fmt.Sprintf("node%d", i)
		sdef.Port = base_port + i
		sdef.ServerId = (base_server_id + i) * 100

		fmt.Printf("Installing and starting node %d\n", i)
		sdef.ReplOptions = ReplOptions + fmt.Sprintf("\n%s\n", GroupReplOptions)
		sdef.ReplOptions += fmt.Sprintf("\n%s\n", GtidOptions)
		sdef.ReplOptions += fmt.Sprintf("\nloose-group-replication-local-address=127.0.0.1:%d\n", group_port)
		sdef.ReplOptions += fmt.Sprintf("\nloose-group-replication-group-seeds=%s\n", connection_string)
		sdef.Multi = true
		CreateSingleSandbox(sdef, origin)
		var data_node  common.Smap = common.Smap{
			"Node" : i,
			"SandboxDir" : sdef.SandboxDir,
			"Copyright" : Copyright,
		}
		write_script(fmt.Sprintf("n%d",i), node_template, sdef.SandboxDir, data_node, true)
	}

	sb_desc := common.SandboxDescription{
		Basedir : sdef.Basedir + "/" + sdef.Version,
		SBType	: "group",
		Port	: 0,
		Nodes 	: nodes,
	}
	common.WriteSandboxDescription(sdef.SandboxDir, sb_desc)

	write_script("start_all", start_multi_template, sdef.SandboxDir, data, true)
	write_script("restart_all", restart_multi_template, sdef.SandboxDir, data, true)
	write_script("status_all", status_multi_template, sdef.SandboxDir, data, true)
	write_script("stop_all", stop_multi_template, sdef.SandboxDir, data, true)
	write_script("send_kill_all", send_kill_multi_template, sdef.SandboxDir, data, true)
	write_script("use_all", use_multi_template, sdef.SandboxDir, data, true)
	write_script("initialize_nodes", init_nodes_template, sdef.SandboxDir, data, true)
	write_script("check_nodes", check_nodes_template, sdef.SandboxDir, data, true)

	fmt.Println(sdef.SandboxDir + "/initialize_nodes")
	common.Run_cmd(sdef.SandboxDir + "/initialize_nodes")
	fmt.Printf("Replication directory installed in %s\n", sdef.SandboxDir)
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}

