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
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
)

type Slave struct {
	Node       int
	Port       int
	ServerId   int
	Name       string
	MasterPort int
}

func CreateMasterSlaveReplication(sdef SandboxDef, origin string, nodes int, masterIp string) {

	var execLists []concurrent.ExecutionList

	fname, logger := defaults.NewLogger(common.LogDirName(), "master-slave-replication")
	sdef.LogFileName = fname
	sdef.ReplOptions = SingleTemplates["replication_options"].Contents
	vList := common.VersionToList(sdef.Version)
	rev := vList[2]
	basePort := sdef.Port + defaults.Defaults().MasterSlaveBasePort + (rev * 100)
	if sdef.BasePort > 0 {
		basePort = sdef.BasePort
	}
	baseServerId := 0
	sdef.DirName = defaults.Defaults().MasterName
	// FindFreePort returns the first free port, but base_port will be used
	// with a counter. Thus the availability will be checked using
	// "base_port + 1"
	firstPort := common.FindFreePort(basePort+1, sdef.InstalledPorts, nodes)
	basePort = firstPort - 1
	baseMysqlxPort := getBaseMysqlxPort(basePort, sdef, nodes)
	for checkPort := basePort + 1; checkPort < basePort+nodes+1; checkPort++ {
		CheckPort("CreateMasterSlaveReplication", sdef.SandboxDir, sdef.InstalledPorts, checkPort)
	}

	if nodes < 2 {
		common.Exit(1, "Can't run replication with less than 2 nodes")
	}
	common.Mkdir(sdef.SandboxDir)
	logger.Printf("Created directory %s\n", sdef.SandboxDir)
	logger.Printf("Replication Sandbox Definition: %s\n", SandboxDefToJson(sdef))
	common.AddToCleanupStack(common.Rmdir, "Rmdir", sdef.SandboxDir)
	sdef.Port = basePort + 1
	sdef.ServerId = (baseServerId + 1) * 100
	sdef.LoadGrants = false
	masterPort := sdef.Port
	changeMasterExtra := ""
	masterAutoPosition := ""
	if sdef.GtidOptions != "" {
		masterAutoPosition += ", MASTER_AUTO_POSITION=1"
		logger.Printf("Adding MASTER_AUTO_POSITION to slaves setup\n")
	}
	if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 4}) {
		if !sdef.NativeAuthPlugin {
			changeMasterExtra += ", GET_MASTER_PUBLIC_KEY=1"
			logger.Printf("Adding GET_MASTER_PUBLIC_KEY to slaves setup \n")
		}
	}
	slaves := nodes - 1
	masterAbbr := defaults.Defaults().MasterAbbr
	masterLabel := defaults.Defaults().MasterName
	slaveLabel := defaults.Defaults().SlavePrefix
	slaveAbbr := defaults.Defaults().SlaveAbbr
	timestamp := time.Now()

	var data = common.StringMap{
		"Copyright":          Copyright,
		"AppVersion":         common.VersionDef,
		"DateTime":           timestamp.Format(time.UnixDate),
		"SandboxDir":         sdef.SandboxDir,
		"MasterLabel":        masterLabel,
		"MasterPort":         sdef.Port,
		"SlaveLabel":         slaveLabel,
		"MasterAbbr":         masterAbbr,
		"MasterIp":           masterIp,
		"RplUser":            sdef.RplUser,
		"RplPassword":        sdef.RplPassword,
		"SlaveAbbr":          slaveAbbr,
		"ChangeMasterExtra":  changeMasterExtra,
		"MasterAutoPosition": masterAutoPosition,
		"Slaves":             []common.StringMap{},
	}

	logger.Printf("Defining replication data: %v\n", SmapToJson(data))
	installationMessage := "Installing and starting %s\n"
	if sdef.SkipStart {
		installationMessage = "Installing %s\n"
	}
	if !sdef.RunConcurrently {
		fmt.Printf(installationMessage, masterLabel)
		logger.Printf(installationMessage, masterLabel)
	}
	sdef.LoadGrants = true
	sdef.Multi = true
	sdef.Prompt = masterLabel
	sdef.NodeNum = 1
	sdef.SBType = "replication-node"
	logger.Printf("Creating single sandbox for master\n")
	execList := CreateSingleSandbox(sdef)
	for _, list := range execList {
		execLists = append(execLists, list)
	}

	sbDesc := common.SandboxDescription{
		Basedir: sdef.Basedir,
		SBType:  "master-slave",
		Version: sdef.Version,
		Port:    []int{sdef.Port},
		Nodes:   slaves,
		NodeNum: 0,
		LogFile: sdef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sdef.Version,
		Port:        []int{sdef.Port},
		Nodes:       []string{defaults.Defaults().MasterName},
		Destination: sdef.SandboxDir,
	}

	if sdef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sdef.LogFileName)
	}

	if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
		sdef.MysqlXPort = baseMysqlxPort + 1
		if !sdef.DisableMysqlX {
			sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+1)
			sbItem.Port = append(sbItem.Port, baseMysqlxPort+1)
			logger.Printf("Adding mysqlx port %d to master\n", baseMysqlxPort+1)
		}
	}

	nodeLabel := defaults.Defaults().NodePrefix
	for i := 1; i <= slaves; i++ {
		sdef.Port = basePort + i + 1
		data["Slaves"] = append(data["Slaves"].([]common.StringMap), common.StringMap{
			"Copyright":          Copyright,
			"AppVersion":         common.VersionDef,
			"DateTime":           timestamp.Format(time.UnixDate),
			"Node":               i,
			"NodeLabel":          nodeLabel,
			"NodePort":           sdef.Port,
			"SlaveLabel":         slaveLabel,
			"MasterAbbr":         masterAbbr,
			"SlaveAbbr":          slaveAbbr,
			"SandboxDir":         sdef.SandboxDir,
			"MasterPort":         masterPort,
			"MasterIp":           masterIp,
			"ChangeMasterExtra":  changeMasterExtra,
			"MasterAutoPosition": masterAutoPosition,
			"RplUser":            sdef.RplUser,
			"RplPassword":        sdef.RplPassword})
		sdef.LoadGrants = false
		sdef.Prompt = fmt.Sprintf("%s%d", slaveLabel, i)
		sdef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sdef.ServerId = (baseServerId + i + 1) * 100
		sdef.NodeNum = i + 1
		sbItem.Nodes = append(sbItem.Nodes, sdef.DirName)
		sbItem.Port = append(sbItem.Port, sdef.Port)
		sbDesc.Port = append(sbDesc.Port, sdef.Port)
		if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
			sdef.MysqlXPort = baseMysqlxPort + i + 1
			if !sdef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i+1)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i+1)
				logger.Printf("Adding mysqlx port %d to slave %d\n", baseMysqlxPort+i+1, i)
			}
		}

		installationMessage = "Installing and starting %s%d\n"
		if sdef.SkipStart {
			installationMessage = "Installing %s%d\n"
		}
		if !sdef.RunConcurrently {
			fmt.Printf(installationMessage, slaveLabel, i)
			logger.Printf(installationMessage, slaveLabel, i)
		}
		if sdef.SemiSyncOptions != "" {
			sdef.SemiSyncOptions = SingleTemplates["semisync_slave_options"].Contents
		}
		logger.Printf("Creating single sandbox for slave %d\n", i)
		execListNode := CreateSingleSandbox(sdef)
		for _, list := range execListNode {
			execLists = append(execLists, list)
		}
		var dataSlave = common.StringMap{
			"Copyright":          Copyright,
			"AppVersion":         common.VersionDef,
			"DateTime":           timestamp.Format(time.UnixDate),
			"Node":               i,
			"NodeLabel":          nodeLabel,
			"NodePort":           sdef.Port,
			"SlaveLabel":         slaveLabel,
			"MasterAbbr":         masterAbbr,
			"ChangeMasterExtra":  changeMasterExtra,
			"MasterAutoPosition": masterAutoPosition,
			"SlaveAbbr":          slaveAbbr,
			"SandboxDir":         sdef.SandboxDir,
		}
		logger.Printf("Defining replication node data: %v\n", SmapToJson(dataSlave))
		logger.Printf("Create slave script %d\n", i)
		writeScript(logger, ReplicationTemplates, fmt.Sprintf("%s%d", slaveAbbr, i), "slave_template", sdef.SandboxDir, dataSlave, true)
		writeScript(logger, ReplicationTemplates, fmt.Sprintf("n%d", i+1), "slave_template", sdef.SandboxDir, dataSlave, true)
	}
	common.WriteSandboxDescription(sdef.SandboxDir, sbDesc)
	logger.Printf("Create sandbox description\n")
	defaults.UpdateCatalog(sdef.SandboxDir, sbItem)

	initializeSlaves := "initialize_" + slaveLabel + "s"
	checkSlaves := "check_" + slaveLabel + "s"

	if sdef.SemiSyncOptions != "" {
		writeScript(logger, ReplicationTemplates, "post_initialization", "semi_sync_start_template", sdef.SandboxDir, data, true)
	}
	logger.Printf("Create replication scripts\n")
	writeScript(logger, ReplicationTemplates, "start_all", "start_all_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "restart_all", "restart_all_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "status_all", "status_all_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "test_sb_all", "test_sb_all_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "stop_all", "stop_all_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "clear_all", "clear_all_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "send_kill_all", "send_kill_all_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "use_all", "use_all_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "use_all_slaves", "use_all_slaves_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "use_all_masters", "use_all_masters_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, initializeSlaves, "init_slaves_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, checkSlaves, "check_slaves_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, masterAbbr, "master_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "n1", "master_template", sdef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "test_replication", "test_replication_template", sdef.SandboxDir, data, true)
	logger.Printf("Run concurrent sandbox scripts \n")
	concurrent.RunParallelTasksByPriority(execLists)
	if !sdef.SkipStart {
		fmt.Println(common.ReplaceLiteralHome(sdef.SandboxDir) + "/" + initializeSlaves)
		logger.Printf("Run replication initialization script \n")
		common.RunCmd(sdef.SandboxDir + "/" + initializeSlaves)
	}
	fmt.Printf("Replication directory installed in %s\n", common.ReplaceLiteralHome(sdef.SandboxDir))
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}

func CreateReplicationSandbox(sdef SandboxDef, origin string, topology string, nodes int, masterIp, masterList, slaveList string) {

	Basedir := sdef.Basedir
	if !common.DirExists(Basedir) {
		common.Exitf(1, "Base directory %s does not exist", Basedir)
	}

	sandboxDir := sdef.SandboxDir
	switch topology {
	case "master-slave":
		sdef.SandboxDir += "/" + defaults.Defaults().MasterSlavePrefix + common.VersionToName(origin)
	case "group":
		if sdef.SinglePrimary {
			sdef.SandboxDir += "/" + defaults.Defaults().GroupSpPrefix + common.VersionToName(origin)
		} else {
			sdef.SandboxDir += "/" + defaults.Defaults().GroupPrefix + common.VersionToName(origin)
		}
		if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 17}) {
			common.Exit(1, "Group replication requires MySQL 5.7.17 or greater")
		}
	case "fan-in":
		if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 9}) {
			common.Exit(1, "multi-source replication requires MySQL 5.7.9 or greater")
		}
		sdef.SandboxDir += "/" + defaults.Defaults().FanInPrefix + common.VersionToName(origin)
	case "all-masters":
		if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 9}) {
			common.Exit(1, "multi-source replication requires MySQL 5.7.9 or greater")
		}
		sdef.SandboxDir += "/" + defaults.Defaults().AllMastersPrefix + common.VersionToName(origin)
	default:
		common.Exit(1, "Unrecognized topology. Accepted: 'master-slave', 'group', 'fan-in', 'all-masters'")
	}
	if sdef.DirName != "" {
		sdef.SandboxDir = sandboxDir + "/" + sdef.DirName
	}

	if common.DirExists(sdef.SandboxDir) {
		sdef = CheckDirectory(sdef)
	}

	if sdef.HistoryDir == "REPL_DIR" {
		sdef.HistoryDir = sdef.SandboxDir
	}
	switch topology {
	case "master-slave":
		CreateMasterSlaveReplication(sdef, origin, nodes, masterIp)
	case "group":
		CreateGroupReplication(sdef, origin, nodes, masterIp)
	case "fan-in":
		CreateFanInReplication(sdef, origin, nodes, masterIp, masterList, slaveList)
	case "all-masters":
		CreateAllMastersReplication(sdef, origin, nodes, masterIp)
	}
}
