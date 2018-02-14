package sandbox

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"os"
	"time"
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
`
	GroupReplSinglePrimary string = `
loose-group-replication-single-primary-mode=on
`
	GroupReplMultiPrimary string = `
loose-group-replication-single-primary-mode=off
`
)

func CreateGroupReplication(sdef SandboxDef, origin string, nodes int) {
	// fmt.Println("Group replication not implemented yet")
	vList := VersionToList(sdef.Version)
	rev := vList[2]
	base_port := sdef.Port + GroupReplicationBasePort + (rev * 100)
	if sdef.SinglePrimary {
		base_port = sdef.Port + GroupReplicationSPBasePort + (rev * 100)
	}
	if sdef.BasePort > 0 {
		base_port = sdef.BasePort
	}

	base_server_id := 0
	if nodes < 3 {
		fmt.Println("Can't run group replication with less than 3 nodes")
		os.Exit(1)
	}
	if common.DirExists(sdef.SandboxDir) {
		sdef = CheckDirectory(sdef)
	}
	for check_port := base_port + 1; check_port < base_port+nodes+1; check_port++ {
		CheckPort(sdef.SandboxDir, sdef.InstalledPorts, check_port)
		CheckPort(sdef.SandboxDir, sdef.InstalledPorts, check_port+GroupPortDelta)
	}
	err := os.Mkdir(sdef.SandboxDir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	timestamp := time.Now()
	var data common.Smap = common.Smap{
		"Copyright":  Copyright,
		"AppVersion":   common.VersionDef,
		"DateTime":    	timestamp.Format(time.UnixDate),
		"SandboxDir": sdef.SandboxDir,
		"Nodes":      []common.Smap{},
	}
	base_group_port := base_port + GroupPortDelta
	connection_string := ""
	for i := 0; i < nodes; i++ {
		group_port := base_group_port + i + 1
		if connection_string != "" {
			connection_string += ","
		}
		connection_string += fmt.Sprintf("127.0.0.1:%d", group_port)
	}

	sb_type := "group-multi-primary"
	single_multi_primary := GroupReplMultiPrimary
	if sdef.SinglePrimary {
		sb_type = "group-single-primary"
		single_multi_primary = GroupReplSinglePrimary
	}
	for i := 1; i <= nodes; i++ {
		group_port := base_group_port + i
		data["Nodes"] = append(data["Nodes"].([]common.Smap), common.Smap{
			"Copyright":    Copyright,
			"AppVersion":   common.VersionDef,
			"DateTime":    	timestamp.Format(time.UnixDate),
			"Node":        i,
			"SandboxDir":  sdef.SandboxDir,
			"RplUser":     sdef.RplUser,
			"RplPassword": sdef.RplPassword})

		sdef.DirName = fmt.Sprintf("node%d", i)
		sdef.Port = base_port + i
		sdef.MorePorts = []int{group_port}
		sdef.ServerId = (base_server_id + i) * 100

		fmt.Printf("Installing and starting node %d\n", i)
		sdef.ReplOptions = ReplOptions + fmt.Sprintf("\n%s\n%s\n", GroupReplOptions, single_multi_primary)
		sdef.ReplOptions += fmt.Sprintf("\n%s\n", GtidOptions)
		sdef.ReplOptions += fmt.Sprintf("\nloose-group-replication-local-address=127.0.0.1:%d\n", group_port)
		sdef.ReplOptions += fmt.Sprintf("\nloose-group-replication-group-seeds=%s\n", connection_string)
		sdef.Multi = true
		sdef.LoadGrants = true
		sdef.Prompt = fmt.Sprintf("node%d", i)
		sdef.SBType = "group-node"
		sdef.NodeNum = i
		CreateSingleSandbox(sdef, origin)
		var data_node common.Smap = common.Smap{
			"Copyright":    Copyright,
			"AppVersion":   common.VersionDef,
			"DateTime":    	timestamp.Format(time.UnixDate),
			"Node":       i,
			"SandboxDir": sdef.SandboxDir,
		}
		write_script(MultipleTemplates, fmt.Sprintf("n%d", i), "node_template", sdef.SandboxDir, data_node, true)
	}

	sb_desc := common.SandboxDescription{
		Basedir: sdef.Basedir + "/" + sdef.Version,
		SBType:  sb_type,
		Version: sdef.Version,
		Port:    []int{0},
		Nodes:   nodes,
		NodeNum : 0,
	}
	common.WriteSandboxDescription(sdef.SandboxDir, sb_desc)

	write_script(MultipleTemplates, "start_all", "start_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "restart_all", "restart_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "status_all", "status_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "test_sb_all", "test_sb_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "stop_all", "stop_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "send_kill_all", "send_kill_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "use_all", "use_multi_template", sdef.SandboxDir, data, true)
	write_script(GroupTemplates, "initialize_nodes", "init_nodes_template", sdef.SandboxDir, data, true)
	write_script(GroupTemplates, "check_nodes", "check_nodes_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "test_replication", "test_replication_template", sdef.SandboxDir, data, true)

	fmt.Println(sdef.SandboxDir + "/initialize_nodes")
	common.Run_cmd(sdef.SandboxDir + "/initialize_nodes")
	fmt.Printf("Replication directory installed in %s\n", sdef.SandboxDir)
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}
