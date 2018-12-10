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
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/pkg/errors"
	"os"
	"path"
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

func CreateMasterSlaveReplication(sandboxDef SandboxDef, origin string, nodes int, masterIp string) error {

	var execLists []concurrent.ExecutionList

	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		var fileName string
		var err error
		logger, fileName, err = defaults.NewLogger(common.LogDirName(), "master-slave-replication")
		if err != nil {
			return err
		}
		sandboxDef.LogFileName = common.ReplaceLiteralHome(fileName)
	}

	sandboxDef.ReplOptions = SingleTemplates["replication_options"].Contents
	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
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
	firstPort, err := common.FindFreePort(basePort+1, sandboxDef.InstalledPorts, nodes)
	if err != nil {
		return errors.Wrapf(err, "error detecting free port for replication")
	}
	basePort = firstPort - 1
	baseMysqlxPort, err := getBaseMysqlxPort(basePort, sandboxDef, nodes)
	if err != nil {
		return err
	}
	for checkPort := basePort + 1; checkPort < basePort+nodes+1; checkPort++ {
		err := checkPortAvailability("CreateMasterSlaveReplication", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
		if err != nil {
			return err
		}
	}

	if nodes < 2 {
		return fmt.Errorf("can't run replication with less than 2 nodes")
	}
	err = os.Mkdir(sandboxDef.SandboxDir, globals.PublicDirectoryAttr)
	if err != nil {
		return err
	}
	logger.Printf("Created directory %s\n", sandboxDef.SandboxDir)
	logger.Printf("Replication Sandbox Definition: %s\n", sandboxDefToJson(sandboxDef))
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
	isMinimumNativeAuthPlugin, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumNativeAuthPluginVersion)
	if err != nil {
		return err
	}
	if isMinimumNativeAuthPlugin {
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

	logger.Printf("Defining replication data: %v\n", stringMapToJson(data))
	installationMessage := "Installing and starting %s\n"
	if sandboxDef.SkipStart {
		installationMessage = "Installing %s\n"
	}
	if !sandboxDef.RunConcurrently {
		common.CondPrintf(installationMessage, masterLabel)
		logger.Printf(installationMessage, masterLabel)
	}
	sandboxDef.LoadGrants = true
	sandboxDef.Multi = true
	sandboxDef.Prompt = masterLabel
	sandboxDef.NodeNum = 1
	sandboxDef.SBType = "replication-node"
	logger.Printf("Creating single sandbox for master\n")
	execList, err := CreateChildSandbox(sandboxDef)
	if err != nil {
		return fmt.Errorf(globals.ErrCreatingSandbox, err)
	}
	for _, list := range execList {
		execLists = append(execLists, list)
	}

	sbDesc := common.SandboxDescription{
		Basedir: sandboxDef.Basedir,
		SBType:  globals.MasterSlaveLabel,
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
	isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
	if err != nil {
		return err
	}
	if isMinimumMySQLXDefault {
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
		isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
		if err != nil {
			return err
		}
		if isMinimumMySQLXDefault {
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
			common.CondPrintf(installationMessage, slaveLabel, i)
			logger.Printf(installationMessage, slaveLabel, i)
		}
		if sandboxDef.SemiSyncOptions != "" {
			sandboxDef.SemiSyncOptions = SingleTemplates["semisync_slave_options"].Contents
		}
		logger.Printf("Creating single sandbox for slave %d\n", i)
		execListNode, err := CreateChildSandbox(sandboxDef)
		if err != nil {
			return fmt.Errorf(globals.ErrCreatingSandbox, err)
		}
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
		logger.Printf("Defining replication node data: %v\n", stringMapToJson(dataSlave))
		logger.Printf("Create slave script %d\n", i)
		err = writeScripts(ScriptBatch{ReplicationTemplates, logger, sandboxDef.SandboxDir, dataSlave,
			[]ScriptDef{
				{fmt.Sprintf("%s%d", slaveAbbr, i), "slave_template", true},
				{fmt.Sprintf("n%d", i), "slave_template", true},
			}})
		if err != nil {
			return err
		}
		// writeScript(logger, ReplicationTemplates, fmt.Sprintf("%s%d", slaveAbbr, i), "slave_template", sandboxDef.SandboxDir, dataSlave, true)
		// writeScript(logger, ReplicationTemplates, fmt.Sprintf("n%d", i+1), "slave_template", sandboxDef.SandboxDir, dataSlave, true)
	}
	err = common.WriteSandboxDescription(sandboxDef.SandboxDir, sbDesc)
	if err != nil {
		return errors.Wrapf(err, "unable to write sandbox description")
	}
	logger.Printf("Create sandbox description\n")
	err = defaults.UpdateCatalog(sandboxDef.SandboxDir, sbItem)
	if err != nil {
		return errors.Wrapf(err, "unable to update catalog")
	}

	initializeSlaves := "initialize_" + slaveLabel + "s"
	checkSlaves := "check_" + slaveLabel + "s"

	sb := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		sandboxDir: sandboxDef.SandboxDir,
		data:       data,
		scripts: []ScriptDef{
			{globals.ScriptStartAll, "start_all_template", true},
			{globals.ScriptRestartAll, "restart_all_template", true},
			{globals.ScriptStatusAll, "status_all_template", true},
			{globals.ScriptTestSbAll, "test_sb_all_template", true},
			{globals.ScriptStopAll, "stop_all_template", true},
			{globals.ScriptClearAll, "clear_all_template", true},
			{globals.ScriptSendKillAll, "send_kill_all_template", true},
			{globals.ScriptUseAll, "use_all_template", true},
			{globals.ScriptUseAllSlaves, "use_all_slaves_template", true},
			{globals.ScriptUseAllMasters, "use_all_masters_template", true},
			{initializeSlaves, "init_slaves_template", true},
			{checkSlaves, "check_slaves_template", true},
			{masterAbbr, "master_template", true},
			{"n1", "master_template", true},
			{"test_replication", "test_replication_template", true},
		},
	}
	if sandboxDef.SemiSyncOptions != "" {
		// writeScript(logger, ReplicationTemplates, "post_initialization", "semi_sync_start_template", sandboxDef.SandboxDir, data, true)
		sb.scripts = append(sb.scripts, ScriptDef{"post_initialization", "semi_sync_start_template", true})
	}
	logger.Printf("Create replication scripts\n")
	err = writeScripts(sb)
	if err != nil {
		return err
	}
	logger.Printf("Run concurrent sandbox scripts \n")
	concurrent.RunParallelTasksByPriority(execLists)
	if !sandboxDef.SkipStart {
		common.CondPrintln(path.Join(common.ReplaceLiteralHome(sandboxDef.SandboxDir), initializeSlaves))
		logger.Printf("Run replication initialization script \n")
		_, err = common.RunCmd(path.Join(sandboxDef.SandboxDir, initializeSlaves))
		if err != nil {
			return err
		}
	}
	common.CondPrintf("Replication directory installed in %s\n", common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	common.CondPrintf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return nil
}

func CreateReplicationSandbox(sdef SandboxDef, origin string, topology string, nodes int, masterIp, masterList, slaveList string) error {

	Basedir := sdef.Basedir
	if !common.DirExists(Basedir) {
		return fmt.Errorf(globals.ErrBaseDirectoryNotFound, Basedir)
	}

	sandboxDir := sdef.SandboxDir
	switch topology {
	case globals.MasterSlaveLabel:
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().MasterSlavePrefix+common.VersionToName(origin))
	case globals.GroupLabel:
		if sdef.SinglePrimary {
			sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().GroupSpPrefix+common.VersionToName(origin))
		} else {
			sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().GroupPrefix+common.VersionToName(origin))
		}
		// 5.7.17
		isMinimumGroupRepl, err := common.GreaterOrEqualVersion(sdef.Version, globals.MinimumGroupReplVersion)
		if err != nil {
			return err
		}
		if !isMinimumGroupRepl {
			return fmt.Errorf(globals.ErrFeatureRequiresVersion, "group replication", common.IntSliceToDottedString(globals.MinimumGroupReplVersion))
		}
	case globals.FanInLabel:
		// 5.7.9
		isMinimumMultiSource, err := common.GreaterOrEqualVersion(sdef.Version, globals.MinimumMultiSourceReplVersion)
		if err != nil {
			return err
		}
		if !isMinimumMultiSource {
			return fmt.Errorf(globals.ErrFeatureRequiresVersion, "multi-source replication", common.IntSliceToDottedString(globals.MinimumMultiSourceReplVersion))
		}
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().FanInPrefix+common.VersionToName(origin))
	case globals.AllMastersLabel:
		// 5.7.9

		isMinimumMultiSource, err := common.GreaterOrEqualVersion(sdef.Version, globals.MinimumMultiSourceReplVersion)
		if err != nil {
			return err
		}
		if !isMinimumMultiSource {
			return fmt.Errorf(globals.ErrFeatureRequiresVersion, "multi-source replication", common.IntSliceToDottedString(globals.MinimumMultiSourceReplVersion))
		}
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().AllMastersPrefix+common.VersionToName(origin))
	default:
		return fmt.Errorf("unrecognized topology. Accepted: '%s', '%s', '%s', '%s'",
			globals.MasterSlaveLabel,
			globals.GroupLabel,
			globals.FanInLabel,
			globals.AllMastersLabel)
	}
	if sdef.DirName != "" {
		sdef.SandboxDir = path.Join(sandboxDir, sdef.DirName)
	}

	if common.DirExists(sdef.SandboxDir) {
		var err error
		sdef, err = checkDirectory(sdef)
		if err != nil {
			return err
		}
	}

	if sdef.HistoryDir == "REPL_DIR" {
		sdef.HistoryDir = sdef.SandboxDir
	}
	var err error
	switch topology {
	case globals.MasterSlaveLabel:
		err = CreateMasterSlaveReplication(sdef, origin, nodes, masterIp)
	case globals.GroupLabel:
		err = CreateGroupReplication(sdef, origin, nodes, masterIp)
	case globals.FanInLabel:
		err = CreateFanInReplication(sdef, origin, nodes, masterIp, masterList, slaveList)
	case globals.AllMastersLabel:
		err = CreateAllMastersReplication(sdef, origin, nodes, masterIp)
	}
	return err
}
