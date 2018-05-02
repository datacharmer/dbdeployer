// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2018 Giuseppe Maxia
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
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
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

func get_base_mysqlx_port(base_port int, sdef SandboxDef, nodes int) int {
	base_mysqlx_port := base_port + 10000
	if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
		base_mysqlx_port = FindFreePort(base_mysqlx_port, sdef.InstalledPorts, nodes)
		for check_port := base_mysqlx_port + 1; check_port < base_mysqlx_port+nodes+1; check_port++ {
			CheckPort(sdef.SandboxDir, sdef.InstalledPorts, check_port)
		}
	}
	return base_mysqlx_port
}

func CreateGroupReplication(sdef SandboxDef, origin string, nodes int, master_ip string) {
	var exec_lists []concurrent.ExecutionList
	vList := common.VersionToList(sdef.Version)
	rev := vList[2]
	// base_port := sdef.Port + defaults.Defaults().GroupReplicationBasePort + (rev * 100)
	base_port := sdef.Port + defaults.Defaults().GroupReplicationBasePort + rev
	if sdef.SinglePrimary {
		base_port = sdef.Port + defaults.Defaults().GroupReplicationSpBasePort + (rev * 100)
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
	base_port = FindFreePort(base_port, sdef.InstalledPorts, nodes)
	base_mysqlx_port := get_base_mysqlx_port(base_port, sdef, nodes)
	base_group_port := base_port + defaults.Defaults().GroupPortDelta
	base_group_port = FindFreePort(base_group_port, sdef.InstalledPorts, nodes)
	for check_port := base_port + 1; check_port < base_port+nodes+1; check_port++ {
		CheckPort(sdef.SandboxDir, sdef.InstalledPorts, check_port)
	}
	for check_port := base_group_port + 1; check_port < base_group_port+nodes+1; check_port++ {
		CheckPort(sdef.SandboxDir, sdef.InstalledPorts, check_port)
	}
	common.Mkdir(sdef.SandboxDir)
	timestamp := time.Now()
	slave_label := defaults.Defaults().SlavePrefix
	slave_abbr := defaults.Defaults().SlaveAbbr
	master_abbr := defaults.Defaults().MasterAbbr
	master_label := defaults.Defaults().MasterName
	master_list := make_nodes_list(nodes)
	slave_list := master_list
	if sdef.SinglePrimary {
		master_list = "1"
		slave_list = ""
		for N := 2; N <= nodes; N++ {
			if slave_list != "" {
				slave_list += " "
			}
			slave_list += fmt.Sprintf("%d", N)
		}
		mlist := nodes_list_to_int_slice(master_list, nodes)
		slist := nodes_list_to_int_slice(slave_list, nodes)
		check_node_lists(nodes, mlist, slist)
	}
	change_master_extra := ""
	node_label := defaults.Defaults().NodePrefix
	//if common.GreaterOrEqualVersion(sdef.Version, []int{8,0,4}) {
	//	if !sdef.NativeAuthPlugin {
	//		change_master_extra = ", GET_MASTER_PUBLIC_KEY=1"
	//	}
	//}
	var data common.Smap = common.Smap{
		"Copyright":         Copyright,
		"AppVersion":        common.VersionDef,
		"DateTime":          timestamp.Format(time.UnixDate),
		"SandboxDir":        sdef.SandboxDir,
		"MasterIp":          master_ip,
		"MasterList":        master_list,
		"NodeLabel":         node_label,
		"SlaveList":         slave_list,
		"RplUser":           sdef.RplUser,
		"RplPassword":       sdef.RplPassword,
		"SlaveLabel":        slave_label,
		"SlaveAbbr":         slave_abbr,
		"ChangeMasterExtra": change_master_extra,
		"MasterLabel":       master_label,
		"MasterAbbr":        master_abbr,
		"Nodes":             []common.Smap{},
	}
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

	sb_desc := common.SandboxDescription{
		Basedir: sdef.Basedir,
		SBType:  sb_type,
		Version: sdef.Version,
		Port:    []int{},
		Nodes:   nodes,
		NodeNum: 0,
	}

	sb_item := defaults.SandboxItem{
		Origin:      sb_desc.Basedir,
		SBType:      sb_desc.SBType,
		Version:     sdef.Version,
		Port:        []int{},
		Nodes:       []string{},
		Destination: sdef.SandboxDir,
	}

	for i := 1; i <= nodes; i++ {
		group_port := base_group_port + i
		data["Nodes"] = append(data["Nodes"].([]common.Smap), common.Smap{
			"Copyright":         Copyright,
			"AppVersion":        common.VersionDef,
			"DateTime":          timestamp.Format(time.UnixDate),
			"Node":              i,
			"MasterIp":          master_ip,
			"NodeLabel":         node_label,
			"SlaveLabel":        slave_label,
			"SlaveAbbr":         slave_abbr,
			"ChangeMasterExtra": change_master_extra,
			"MasterLabel":       master_label,
			"MasterAbbr":        master_abbr,
			"SandboxDir":        sdef.SandboxDir,
			"RplUser":           sdef.RplUser,
			"RplPassword":       sdef.RplPassword})

		sdef.DirName = fmt.Sprintf("%s%d", node_label, i)
		sdef.Port = base_port + i
		sdef.MorePorts = []int{group_port}
		sdef.ServerId = (base_server_id + i) * 100
		sb_item.Nodes = append(sb_item.Nodes, sdef.DirName)
		sb_item.Port = append(sb_item.Port, sdef.Port)
		sb_desc.Port = append(sb_desc.Port, sdef.Port)
		sb_item.Port = append(sb_item.Port, sdef.Port+defaults.Defaults().GroupPortDelta)
		sb_desc.Port = append(sb_desc.Port, sdef.Port+defaults.Defaults().GroupPortDelta)

		if !sdef.RunConcurrently {
			installation_message := "Installing and starting %s %d\n"
			if sdef.SkipStart {
				installation_message = "Installing %s %d\n"
			}
			fmt.Printf(installation_message, node_label, i)
		}
		sdef.ReplOptions = SingleTemplates["replication_options"].Contents + fmt.Sprintf("\n%s\n%s\n", GroupReplOptions, single_multi_primary)
		re_master_ip := regexp.MustCompile(`127\.0\.0\.1`)
		sdef.ReplOptions = re_master_ip.ReplaceAllString(sdef.ReplOptions, master_ip)
		sdef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates["gtid_options"].Contents)
		sdef.ReplOptions += fmt.Sprintf("\nloose-group-replication-local-address=%s:%d\n", master_ip, group_port)
		sdef.ReplOptions += fmt.Sprintf("\nloose-group-replication-group-seeds=%s\n", connection_string)
		if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
			sdef.MysqlXPort = base_mysqlx_port + i
			if !sdef.DisableMysqlX {
				sb_desc.Port = append(sb_desc.Port, base_mysqlx_port+i)
				sb_item.Port = append(sb_item.Port, base_mysqlx_port+i)
			}
		}
		sdef.Multi = true
		sdef.LoadGrants = true
		sdef.Prompt = fmt.Sprintf("%s%d", node_label, i)
		sdef.SBType = "group-node"
		sdef.NodeNum = i
		// fmt.Printf("%#v\n",sdef)
		exec_list := CreateSingleSandbox(sdef, origin)
		for _, list := range exec_list {
			exec_lists = append(exec_lists, list)
		}
		var data_node common.Smap = common.Smap{
			"Copyright":         Copyright,
			"AppVersion":        common.VersionDef,
			"DateTime":          timestamp.Format(time.UnixDate),
			"Node":              i,
			"NodeLabel":         node_label,
			"MasterLabel":       master_label,
			"MasterAbbr":        master_abbr,
			"ChangeMasterExtra": change_master_extra,
			"SlaveLabel":        slave_label,
			"SlaveAbbr":         slave_abbr,
			"SandboxDir":        sdef.SandboxDir,
		}
		write_script(MultipleTemplates, fmt.Sprintf("n%d", i), "node_template", sdef.SandboxDir, data_node, true)
	}
	common.WriteSandboxDescription(sdef.SandboxDir, sb_desc)
	defaults.UpdateCatalog(sdef.SandboxDir, sb_item)

	write_script(MultipleTemplates, "start_all", "start_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "restart_all", "restart_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "status_all", "status_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "test_sb_all", "test_sb_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "stop_all", "stop_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "clear_all", "clear_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "send_kill_all", "send_kill_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "use_all", "use_multi_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "use_all_slaves", "multi_source_use_slaves_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "use_all_masters", "multi_source_use_masters_template", sdef.SandboxDir, data, true)
	write_script(GroupTemplates, "initialize_nodes", "init_nodes_template", sdef.SandboxDir, data, true)
	write_script(GroupTemplates, "check_nodes", "check_nodes_template", sdef.SandboxDir, data, true)
	//write_script(ReplicationTemplates, "test_replication", "test_replication_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "test_replication", "multi_source_test_template", sdef.SandboxDir, data, true)

	concurrent.RunParallelTasksByPriority(exec_lists)
	if !sdef.SkipStart {
		fmt.Println(common.ReplaceLiteralHome(sdef.SandboxDir) + "/initialize_nodes")
		common.Run_cmd(sdef.SandboxDir + "/initialize_nodes")
	}
	fmt.Printf("Replication directory installed in %s\n", common.ReplaceLiteralHome(sdef.SandboxDir))
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}
