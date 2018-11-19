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
	"path"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/pkg/errors"
)

type Slave struct {
	Node       int
	Port       int
	ServerId   int
	Name       string
	MasterPort int
}

func CreateMasterSlaveReplication(sandboxDef SandboxDef, origin string, nodes int, masterIp string) error {

	var execLists []concurrent.ExecutionList

	fileName, logger := defaults.NewLogger(common.LogDirName(), "master-slave-replication")
	sandboxDef.LogFileName = fileName
	sandboxDef.ReplOptions = SingleTemplates["replication_options"].Contents
	vList := common.VersionToList(sandboxDef.Version)
	rev := vList[2]
	basePort := sandboxDef.Port + defaults.Defaults().MasterSlaveBasePort + (rev * 100)
	if sandboxDef.BasePort > 0 {
		basePort = sandboxDef.BasePort
	}
	baseServerId := 0
	sandboxDef.DirName = defaults.Defaults().MasterName
	// FindFreePort returns the first free port, but base_port will be used
	// with a counter. Thus the availability will be checked using
	// "base_port + 1"
	firstPort := common.FindFreePort(basePort+1, sandboxDef.InstalledPorts, nodes)
	basePort = firstPort - 1
	baseMysqlxPort := getBaseMysqlxPort(basePort, sandboxDef, nodes)
	for checkPort := basePort + 1; checkPort < basePort+nodes+1; checkPort++ {
		CheckPort("CreateMasterSlaveReplication", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
	}

	if nodes < 2 {
		common.Exit(1, "can't run replication with less than 2 nodes")
	}
	common.Mkdir(sandboxDef.SandboxDir)
	logger.Printf("Created directory %s\n", sandboxDef.SandboxDir)
	logger.Printf("Replication Sandbox Definition: %s\n", SandboxDefToJson(sandboxDef))
	common.AddToCleanupStack(common.Rmdir, "Rmdir", sandboxDef.SandboxDir)
	sandboxDef.Port = basePort + 1
	sandboxDef.ServerId = (baseServerId + 1) * 100
	sandboxDef.LoadGrants = false
	masterPort := sandboxDef.Port
	changeMasterExtra := ""
	masterAutoPosition := ""
	if sandboxDef.GtidOptions != "" {
		masterAutoPosition += ", MASTER_AUTO_POSITION=1"
		logger.Printf("Adding MASTER_AUTO_POSITION to slaves setup\n")
	}
	// 8.0.11
	if common.GreaterOrEqualVersion(sandboxDef.Version, defaults.MinimumNativeAuthPluginVersion) {
		if !sandboxDef.NativeAuthPlugin {
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
		"SandboxDir":         sandboxDef.SandboxDir,
		"MasterLabel":        masterLabel,
		"MasterPort":         sandboxDef.Port,
		"SlaveLabel":         slaveLabel,
		"MasterAbbr":         masterAbbr,
		"MasterIp":           masterIp,
		"RplUser":            sandboxDef.RplUser,
		"RplPassword":        sandboxDef.RplPassword,
		"SlaveAbbr":          slaveAbbr,
		"ChangeMasterExtra":  changeMasterExtra,
		"MasterAutoPosition": masterAutoPosition,
		"Slaves":             []common.StringMap{},
	}

	logger.Printf("Defining replication data: %v\n", StringMapToJson(data))
	installationMessage := "Installing and starting %s\n"
	if sandboxDef.SkipStart {
		installationMessage = "Installing %s\n"
	}
	if !sandboxDef.RunConcurrently {
		fmt.Printf(installationMessage, masterLabel)
		logger.Printf(installationMessage, masterLabel)
	}
	sandboxDef.LoadGrants = true
	sandboxDef.Multi = true
	sandboxDef.Prompt = masterLabel
	sandboxDef.NodeNum = 1
	sandboxDef.SBType = "replication-node"
	logger.Printf("Creating single sandbox for master\n")
	execList, err := CreateSingleConcurrentSandbox(sandboxDef)
	if err != nil {
		return errors.Wrap(err, "cannot create a single sandbox for master")
	}

	for _, list := range execList {
		execLists = append(execLists, list)
	}

	sbDesc := common.SandboxDescription{
		Basedir: sandboxDef.Basedir,
		SBType:  defaults.MasterSlaveLabel,
		Version: sandboxDef.Version,
		Port:    []int{sandboxDef.Port},
		Nodes:   slaves,
		NodeNum: 0,
		LogFile: sandboxDef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sandboxDef.Version,
		Port:        []int{sandboxDef.Port},
		Nodes:       []string{defaults.Defaults().MasterName},
		Destination: sandboxDef.SandboxDir,
	}

	if sandboxDef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sandboxDef.LogFileName)
	}

	// 8.0.11
	if common.GreaterOrEqualVersion(sandboxDef.Version, defaults.MinimumMysqlxDefaultVersion) {
		sandboxDef.MysqlXPort = baseMysqlxPort + 1
		if !sandboxDef.DisableMysqlX {
			sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+1)
			sbItem.Port = append(sbItem.Port, baseMysqlxPort+1)
			logger.Printf("Adding mysqlx port %d to master\n", baseMysqlxPort+1)
		}
	}

	nodeLabel := defaults.Defaults().NodePrefix
	for i := 1; i <= slaves; i++ {
		sandboxDef.Port = basePort + i + 1
		data["Slaves"] = append(data["Slaves"].([]common.StringMap), common.StringMap{
			"Copyright":          Copyright,
			"AppVersion":         common.VersionDef,
			"DateTime":           timestamp.Format(time.UnixDate),
			"Node":               i,
			"NodeLabel":          nodeLabel,
			"NodePort":           sandboxDef.Port,
			"SlaveLabel":         slaveLabel,
			"MasterAbbr":         masterAbbr,
			"SlaveAbbr":          slaveAbbr,
			"SandboxDir":         sandboxDef.SandboxDir,
			"MasterPort":         masterPort,
			"MasterIp":           masterIp,
			"ChangeMasterExtra":  changeMasterExtra,
			"MasterAutoPosition": masterAutoPosition,
			"RplUser":            sandboxDef.RplUser,
			"RplPassword":        sandboxDef.RplPassword})
		sandboxDef.LoadGrants = false
		sandboxDef.Prompt = fmt.Sprintf("%s%d", slaveLabel, i)
		sandboxDef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.ServerId = (baseServerId + i + 1) * 100
		sandboxDef.NodeNum = i + 1
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)
		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)
		// 8.0.11
		if common.GreaterOrEqualVersion(sandboxDef.Version, defaults.MinimumMysqlxDefaultVersion) {
			sandboxDef.MysqlXPort = baseMysqlxPort + i + 1
			if !sandboxDef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i+1)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i+1)
				logger.Printf("Adding mysqlx port %d to slave %d\n", baseMysqlxPort+i+1, i)
			}
		}

		installationMessage = "Installing and starting %s%d\n"
		if sandboxDef.SkipStart {
			installationMessage = "Installing %s%d\n"
		}
		if !sandboxDef.RunConcurrently {
			fmt.Printf(installationMessage, slaveLabel, i)
			logger.Printf(installationMessage, slaveLabel, i)
		}
		if sandboxDef.SemiSyncOptions != "" {
			sandboxDef.SemiSyncOptions = SingleTemplates["semisync_slave_options"].Contents
		}
		logger.Printf("Creating single sandbox for slave %d\n", i)
		execListNode := CreateSingleSandbox(sandboxDef)
		for _, list := range execListNode {
			execLists = append(execLists, list)
		}
		var dataSlave = common.StringMap{
			"Copyright":          Copyright,
			"AppVersion":         common.VersionDef,
			"DateTime":           timestamp.Format(time.UnixDate),
			"Node":               i,
			"NodeLabel":          nodeLabel,
			"NodePort":           sandboxDef.Port,
			"SlaveLabel":         slaveLabel,
			"MasterAbbr":         masterAbbr,
			"ChangeMasterExtra":  changeMasterExtra,
			"MasterAutoPosition": masterAutoPosition,
			"SlaveAbbr":          slaveAbbr,
			"SandboxDir":         sandboxDef.SandboxDir,
		}
		logger.Printf("Defining replication node data: %v\n", StringMapToJson(dataSlave))
		logger.Printf("Create slave script %d\n", i)
		writeScript(logger, ReplicationTemplates, fmt.Sprintf("%s%d", slaveAbbr, i), "slave_template", sandboxDef.SandboxDir, dataSlave, true)
		writeScript(logger, ReplicationTemplates, fmt.Sprintf("n%d", i+1), "slave_template", sandboxDef.SandboxDir, dataSlave, true)
	}
	common.WriteSandboxDescription(sandboxDef.SandboxDir, sbDesc)
	logger.Printf("Create sandbox description\n")
	defaults.UpdateCatalog(sandboxDef.SandboxDir, sbItem)

	initializeSlaves := "initialize_" + slaveLabel + "s"
	checkSlaves := "check_" + slaveLabel + "s"

	if sandboxDef.SemiSyncOptions != "" {
		writeScript(logger, ReplicationTemplates, "post_initialization", "semi_sync_start_template", sandboxDef.SandboxDir, data, true)
	}
	logger.Printf("Create replication scripts\n")
	writeScript(logger, ReplicationTemplates, defaults.ScriptStartAll, "start_all_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptRestartAll, "restart_all_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptStatusAll, "status_all_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptTestSbAll, "test_sb_all_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptStopAll, "stop_all_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptClearAll, "clear_all_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptSendKillAll, "send_kill_all_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptUseAll, "use_all_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptUseAllSlaves, "use_all_slaves_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptUseAllMasters, "use_all_masters_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, initializeSlaves, "init_slaves_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, checkSlaves, "check_slaves_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, masterAbbr, "master_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "n1", "master_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, "test_replication", "test_replication_template", sandboxDef.SandboxDir, data, true)
	logger.Printf("Run concurrent sandbox scripts \n")
	concurrent.RunParallelTasksByPriority(execLists)
	if !sandboxDef.SkipStart {
		fmt.Println(path.Join(common.ReplaceLiteralHome(sandboxDef.SandboxDir), initializeSlaves))
		logger.Printf("Run replication initialization script \n")
		common.RunCmd(path.Join(sandboxDef.SandboxDir, initializeSlaves))
	}
	// TODO: Improve logging
	//fmt.Printf("Replication directory installed in %s\n", common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	//fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return nil
}

func CreateReplicationSandbox(sdef SandboxDef, origin string, topology string, nodes int, masterIp, masterList, slaveList string) {

	Basedir := sdef.Basedir
	if !common.DirExists(Basedir) {
		common.Exitf(1, defaults.ErrBaseDirectoryNotFound, Basedir)
	}

	sandboxDir := sdef.SandboxDir
	switch topology {
	case defaults.MasterSlaveLabel:
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().MasterSlavePrefix+common.VersionToName(origin))
	case defaults.GroupLabel:
		if sdef.SinglePrimary {
			sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().GroupSpPrefix+common.VersionToName(origin))
		} else {
			sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().GroupPrefix+common.VersionToName(origin))
		}
		// 5.7.17
		if !common.GreaterOrEqualVersion(sdef.Version, defaults.MinimumGroupReplVersion) {
			common.Exitf(1, defaults.ErrFeatureRequiresVersion, "group replication", common.IntSliceToDottedString(defaults.MinimumGroupReplVersion))
		}
	case defaults.FanInLabel:
		// 5.7.9
		if !common.GreaterOrEqualVersion(sdef.Version, defaults.MinimumMultiSourceReplVersion) {
			common.Exitf(1, defaults.ErrFeatureRequiresVersion, "multi-source replication", common.IntSliceToDottedString(defaults.MinimumMultiSourceReplVersion))
		}
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().FanInPrefix+common.VersionToName(origin))
	case defaults.AllMastersLabel:
		// 5.7.9
		if !common.GreaterOrEqualVersion(sdef.Version, defaults.MinimumMultiSourceReplVersion) {
			common.Exitf(1, defaults.ErrFeatureRequiresVersion, "multi-source replication", common.IntSliceToDottedString(defaults.MinimumMultiSourceReplVersion))
		}
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().AllMastersPrefix+common.VersionToName(origin))
	default:
		common.Exitf(1, "unrecognized topology. Accepted: '%s', '%s', '%s', '%s'",
			defaults.MasterSlaveLabel,
			defaults.GroupLabel,
			defaults.FanInLabel,
			defaults.AllMastersLabel)
	}
	if sdef.DirName != "" {
		sdef.SandboxDir = path.Join(sandboxDir, sdef.DirName)
	}

	if common.DirExists(sdef.SandboxDir) {
		sdef = CheckDirectory(sdef)
	}

	if sdef.HistoryDir == "REPL_DIR" {
		sdef.HistoryDir = sdef.SandboxDir
	}
	switch topology {
	case defaults.MasterSlaveLabel:
		CreateMasterSlaveReplication(sdef, origin, nodes, masterIp)
	case defaults.GroupLabel:
		CreateGroupReplication(sdef, origin, nodes, masterIp)
	case defaults.FanInLabel:
		CreateFanInReplication(sdef, origin, nodes, masterIp, masterList, slaveList)
	case defaults.AllMastersLabel:
		CreateAllMastersReplication(sdef, origin, nodes, masterIp)
	}
}
