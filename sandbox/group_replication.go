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

func getBaseMysqlxPort(basePort int, sdef SandboxDef, nodes int) int {
	baseMysqlxPort := basePort + defaults.Defaults().MysqlXPortDelta
	if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
		// FindFreePort returns the first free port, but base_port will be used
		// with a counter. Thus the availability will be checked using
		// "base_port + 1"
		firstGroupPort := common.FindFreePort(baseMysqlxPort+1, sdef.InstalledPorts, nodes)
		baseMysqlxPort = firstGroupPort - 1
		for N := 1; N <= nodes; N++ {
			checkPort := baseMysqlxPort + N
			CheckPort("get_base_mysqlx_port", sdef.SandboxDir, sdef.InstalledPorts, checkPort)
		}
	}
	return baseMysqlxPort
}

func CreateGroupReplication(sdef SandboxDef, origin string, nodes int, masterIp string) {
	var execLists []concurrent.ExecutionList

	fname, logger := defaults.NewLogger(common.LogDirName(), "group-replication")
	sdef.LogFileName = common.ReplaceLiteralHome(fname)
	vList := common.VersionToList(sdef.Version)
	rev := vList[2]
	basePort := sdef.Port + defaults.Defaults().GroupReplicationBasePort + (rev * 100)
	if sdef.SinglePrimary {
		basePort = sdef.Port + defaults.Defaults().GroupReplicationSpBasePort + (rev * 100)
	}
	if sdef.BasePort > 0 {
		basePort = sdef.BasePort
	}

	baseServerId := 0
	if nodes < 3 {
		fmt.Println("Can't run group replication with less than 3 nodes")
		os.Exit(1)
	}
	if common.DirExists(sdef.SandboxDir) {
		sdef = CheckDirectory(sdef)
	}
	// FindFreePort returns the first free port, but base_port will be used
	// with a counter. Thus the availability will be checked using
	// "base_port + 1"
	firstGroupPort := common.FindFreePort(basePort+1, sdef.InstalledPorts, nodes)
	basePort = firstGroupPort - 1
	baseGroupPort := basePort + defaults.Defaults().GroupPortDelta
	firstGroupPort = common.FindFreePort(baseGroupPort+1, sdef.InstalledPorts, nodes)
	baseGroupPort = firstGroupPort - 1
	for checkPort := basePort + 1; checkPort < basePort+nodes+1; checkPort++ {
		CheckPort("CreateGroupReplication", sdef.SandboxDir, sdef.InstalledPorts, checkPort)
	}
	for checkPort := baseGroupPort + 1; checkPort < baseGroupPort+nodes+1; checkPort++ {
		CheckPort("CreateGroupReplication-group", sdef.SandboxDir, sdef.InstalledPorts, checkPort)
	}
	baseMysqlxPort := getBaseMysqlxPort(basePort, sdef, nodes)
	common.Mkdir(sdef.SandboxDir)
	common.AddToCleanupStack(common.Rmdir, "Rmdir", sdef.SandboxDir)
	logger.Printf("Creating directory %s\n", sdef.SandboxDir)
	timestamp := time.Now()
	slaveLabel := defaults.Defaults().SlavePrefix
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterAbbr := defaults.Defaults().MasterAbbr
	masterLabel := defaults.Defaults().MasterName
	masterList := makeNodesList(nodes)
	slaveList := masterList
	if sdef.SinglePrimary {
		masterList = "1"
		slaveList = ""
		for N := 2; N <= nodes; N++ {
			if slaveList != "" {
				slaveList += " "
			}
			slaveList += fmt.Sprintf("%d", N)
		}
		mlist := nodesListToIntSlice(masterList, nodes)
		slist := nodesListToIntSlice(slaveList, nodes)
		checkNodeLists(nodes, mlist, slist)
	}
	changeMasterExtra := ""
	nodeLabel := defaults.Defaults().NodePrefix
	//if common.GreaterOrEqualVersion(sdef.Version, []int{8,0,4}) {
	//	if !sdef.NativeAuthPlugin {
	//		change_master_extra = ", GET_MASTER_PUBLIC_KEY=1"
	//	}
	//}
	var data = common.StringMap{
		"Copyright":         Copyright,
		"AppVersion":        common.VersionDef,
		"DateTime":          timestamp.Format(time.UnixDate),
		"SandboxDir":        sdef.SandboxDir,
		"MasterIp":          masterIp,
		"MasterList":        masterList,
		"NodeLabel":         nodeLabel,
		"SlaveList":         slaveList,
		"RplUser":           sdef.RplUser,
		"RplPassword":       sdef.RplPassword,
		"SlaveLabel":        slaveLabel,
		"SlaveAbbr":         slaveAbbr,
		"ChangeMasterExtra": changeMasterExtra,
		"MasterLabel":       masterLabel,
		"MasterAbbr":        masterAbbr,
		"Nodes":             []common.StringMap{},
	}
	connectionString := ""
	for i := 0; i < nodes; i++ {
		groupPort := baseGroupPort + i + 1
		if connectionString != "" {
			connectionString += ","
		}
		connectionString += fmt.Sprintf("127.0.0.1:%d", groupPort)
	}
	logger.Printf("Creating connection string %s\n", connectionString)

	sbType := "group-multi-primary"
	singleMultiPrimary := GroupReplMultiPrimary
	if sdef.SinglePrimary {
		sbType = "group-single-primary"
		singleMultiPrimary = GroupReplSinglePrimary
	}
	logger.Printf("Defining group type %s\n", sbType)

	sbDesc := common.SandboxDescription{
		Basedir: sdef.Basedir,
		SBType:  sbType,
		Version: sdef.Version,
		Port:    []int{},
		Nodes:   nodes,
		NodeNum: 0,
		LogFile: sdef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sdef.Version,
		Port:        []int{},
		Nodes:       []string{},
		Destination: sdef.SandboxDir,
	}

	if sdef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sdef.LogFileName)
	}

	for i := 1; i <= nodes; i++ {
		groupPort := baseGroupPort + i
		data["Nodes"] = append(data["Nodes"].([]common.StringMap), common.StringMap{
			"Copyright":         Copyright,
			"AppVersion":        common.VersionDef,
			"DateTime":          timestamp.Format(time.UnixDate),
			"Node":              i,
			"MasterIp":          masterIp,
			"NodeLabel":         nodeLabel,
			"SlaveLabel":        slaveLabel,
			"SlaveAbbr":         slaveAbbr,
			"ChangeMasterExtra": changeMasterExtra,
			"MasterLabel":       masterLabel,
			"MasterAbbr":        masterAbbr,
			"SandboxDir":        sdef.SandboxDir,
			"RplUser":           sdef.RplUser,
			"RplPassword":       sdef.RplPassword})

		sdef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sdef.Port = basePort + i
		sdef.MorePorts = []int{groupPort}
		sdef.ServerId = (baseServerId + i) * 100
		sbItem.Nodes = append(sbItem.Nodes, sdef.DirName)
		sbItem.Port = append(sbItem.Port, sdef.Port)
		sbDesc.Port = append(sbDesc.Port, sdef.Port)
		sbItem.Port = append(sbItem.Port, sdef.Port+defaults.Defaults().GroupPortDelta)
		sbDesc.Port = append(sbDesc.Port, sdef.Port+defaults.Defaults().GroupPortDelta)

		if !sdef.RunConcurrently {
			installationMessage := "Installing and starting %s %d\n"
			if sdef.SkipStart {
				installationMessage = "Installing %s %d\n"
			}
			fmt.Printf(installationMessage, nodeLabel, i)
			logger.Printf(installationMessage, nodeLabel, i)
		}
		sdef.ReplOptions = SingleTemplates["replication_options"].Contents + fmt.Sprintf("\n%s\n%s\n", GroupReplOptions, singleMultiPrimary)
		reMasterIp := regexp.MustCompile(`127\.0\.0\.1`)
		sdef.ReplOptions = reMasterIp.ReplaceAllString(sdef.ReplOptions, masterIp)
		sdef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates["gtid_options_57"].Contents)
		sdef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates["repl_crash_safe_options"].Contents)
		sdef.ReplOptions += fmt.Sprintf("\nloose-group-replication-local-address=%s:%d\n", masterIp, groupPort)
		sdef.ReplOptions += fmt.Sprintf("\nloose-group-replication-group-seeds=%s\n", connectionString)
		if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
			sdef.MysqlXPort = baseMysqlxPort + i
			if !sdef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i)
				logger.Printf("adding port %d to node %d\n", baseMysqlxPort+i, i)
			}
		}
		sdef.Multi = true
		sdef.LoadGrants = true
		sdef.Prompt = fmt.Sprintf("%s%d", nodeLabel, i)
		sdef.SBType = "group-node"
		sdef.NodeNum = i
		// fmt.Printf("%#v\n",sdef)
		logger.Printf("Create single sandbox for node %d\n", i)
		execList := CreateSingleSandbox(sdef)
		for _, list := range execList {
			execLists = append(execLists, list)
		}
		var dataNode = common.StringMap{
			"Copyright":         Copyright,
			"AppVersion":        common.VersionDef,
			"DateTime":          timestamp.Format(time.UnixDate),
			"Node":              i,
			"NodeLabel":         nodeLabel,
			"MasterLabel":       masterLabel,
			"MasterAbbr":        masterAbbr,
			"ChangeMasterExtra": changeMasterExtra,
			"SlaveLabel":        slaveLabel,
			"SlaveAbbr":         slaveAbbr,
			"SandboxDir":        sdef.SandboxDir,
		}
		logger.Printf("Create node script for node %d\n", i)
		writeScript(logger, MultipleTemplates, fmt.Sprintf("n%d", i), "node_template", sdef.SandboxDir, dataNode, true)
	}
	logger.Printf("Writing sandbox description in %s\n", sdef.SandboxDir)
	common.WriteSandboxDescription(sdef.SandboxDir, sbDesc)
	defaults.UpdateCatalog(sdef.SandboxDir, sbItem)

	logger.Printf("Writing group replication scripts\n")
	writeScript(logger, MultipleTemplates, "start_all", "start_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "restart_all", "restart_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "status_all", "status_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "test_sb_all", "test_sb_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "stop_all", "stop_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "clear_all", "clear_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "send_kill_all", "send_kill_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "use_all", "use_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "use_all_slaves", "multi_source_use_slaves_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "use_all_masters", "multi_source_use_masters_template", sdef.SandboxDir, data, true)
	writeScript(logger, GroupTemplates, "initialize_nodes", "init_nodes_template", sdef.SandboxDir, data, true)
	writeScript(logger, GroupTemplates, "check_nodes", "check_nodes_template", sdef.SandboxDir, data, true)
	//writeScript(logger, ReplicationTemplates, "test_replication", "test_replication_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "test_replication", "multi_source_test_template", sdef.SandboxDir, data, true)

	logger.Printf("Running parallel tasks\n")
	concurrent.RunParallelTasksByPriority(execLists)
	if !sdef.SkipStart {
		fmt.Println(common.ReplaceLiteralHome(sdef.SandboxDir) + "/initialize_nodes")
		logger.Printf("Running group replication initialization script\n")
		common.RunCmd(sdef.SandboxDir + "/initialize_nodes")
	}
	fmt.Printf("Replication directory installed in %s\n", common.ReplaceLiteralHome(sdef.SandboxDir))
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}
