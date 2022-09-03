// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
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
	"path"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/dustin/go-humanize/english"
	"github.com/pkg/errors"
)

type Slave struct {
	Node       int
	Port       int
	ServerId   int
	Name       string
	MasterPort int
}

type ReplicationData struct {
	Topology   string
	MasterIp   string
	Nodes      int
	NdbNodes   int
	MasterList string
	SlaveList  string
}

func setChangeMasterProperties(currentProperties string, moreProperties []string, logger *defaults.Logger) string {
	for _, property := range moreProperties {
		currentProperties += ", " + property
		logger.Printf("Adding %s to slaves setup \n", property)
	}
	return currentProperties
}

func checkReadOnlyFlags(sandboxDef SandboxDef) (string, error) {
	readOnlyOption := ""
	if sandboxDef.SlavesSuperReadOnly && sandboxDef.SlavesReadOnly {
		return "", fmt.Errorf("only one of --%s or %s should be used", globals.ReadOnlyLabel, globals.SuperReadOnlyLabel)
	}
	if sandboxDef.SlavesSuperReadOnly {
		// readOnlyAllowed, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumSuperReadOnly)
		readOnlyAllowed, err := common.HasCapability(sandboxDef.Flavor, common.SuperReadOnly, sandboxDef.Version)
		if err != nil {
			return "", err
		}
		if !readOnlyAllowed {
			return "", fmt.Errorf(globals.ErrOptionRequiresVersion,
				globals.SuperReadOnlyLabel, common.IntSliceToDottedString(globals.MinimumSuperReadOnly))
		}
		readOnlyOption = "super_read_only=on"
	} else {
		if sandboxDef.SlavesReadOnly {
			// readOnlyAllowed, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumDynVariablesVersion)
			readOnlyAllowed, err := common.HasCapability(sandboxDef.Flavor, common.DynVariables, sandboxDef.Version)
			if err != nil {
				return "", err
			}
			if !readOnlyAllowed {
				return "", fmt.Errorf(globals.ErrOptionRequiresVersion,
					globals.ReadOnlyLabel, common.IntSliceToDottedString(globals.MinimumDynVariablesVersion))
			}
			readOnlyOption = "read_only=on"
		}
	}

	return readOnlyOption, nil
}

func setServerId(sandboxDef SandboxDef, increment int) int {
	if sandboxDef.BaseServerId == 0 {
		return (sandboxDef.BaseServerId + increment) * 100
	}
	return sandboxDef.BaseServerId + increment
}

func computeBaseport(proposed int) int {
	if proposed > globals.MaxAllowedPort {
		return proposed - globals.ReductionOnPortNumberOverflow
	}
	return proposed
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

	sandboxDef.ReplOptions = SingleTemplates[globals.TmplReplicationOptions].Contents
	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
	rev := vList[2]
	basePort := computeBaseport(sandboxDef.Port + defaults.Defaults().MasterSlaveBasePort + (rev * 100))
	if sandboxDef.BasePort > 0 {
		basePort = sandboxDef.BasePort
	}
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
	baseAdminPort, err := getBaseAdminPort(basePort, sandboxDef, nodes)
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

	readOnlyOptions, err := checkReadOnlyFlags(sandboxDef)
	if err != nil {
		return err
	}

	err = os.Mkdir(sandboxDef.SandboxDir, globals.PublicDirectoryAttr)
	if err != nil {
		return err
	}
	logger.Printf("Created directory %s\n", sandboxDef.SandboxDir)
	logger.Printf("Replication Sandbox Definition: %s\n", sandboxDefToJson(sandboxDef))
	common.AddToCleanupStack(common.RmdirAll, "RmdirAll", sandboxDef.SandboxDir)
	sandboxDef.Port = basePort + 1
	//sandboxDef.ServerId = (baseServerId + 1) * 100
	sandboxDef.ServerId = setServerId(sandboxDef, 1)
	sandboxDef.LoadGrants = false
	masterPort := sandboxDef.Port
	changeMasterExtra := ""
	masterAutoPosition := ""
	if sandboxDef.GtidOptions != "" {
		masterAutoPosition += ", MASTER_AUTO_POSITION=1"
		logger.Printf("Adding MASTER_AUTO_POSITION to slaves setup\n")
	}
	// 8.0.11
	// isMinimumNativeAuthPlugin, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumNativeAuthPluginVersion)
	isMinimumNativeAuthPlugin, err := common.HasCapability(sandboxDef.Flavor, common.NativeAuth, sandboxDef.Version)
	if err != nil {
		return err
	}
	if isMinimumNativeAuthPlugin {
		if !sandboxDef.NativeAuthPlugin {
			sandboxDef.ChangeMasterOptions = append(sandboxDef.ChangeMasterOptions, "GET_MASTER_PUBLIC_KEY=1")
		}
	}
	slaves := nodes - 1
	masterAbbr := defaults.Defaults().MasterAbbr
	masterLabel := defaults.Defaults().MasterName
	slaveLabel := defaults.Defaults().SlavePrefix
	slaveAbbr := defaults.Defaults().SlaveAbbr
	timestamp := time.Now()

	changeMasterExtra = setChangeMasterProperties(changeMasterExtra, sandboxDef.ChangeMasterOptions, logger)
	var data = common.StringMap{
		"ShellPath":          sandboxDef.ShellPath,
		"Copyright":          globals.ShellScriptCopyright,
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
	sandboxDef.ReadOnlyOptions = ""
	logger.Printf("Creating single sandbox for master\n")
	execList, err := CreateChildSandbox(sandboxDef)
	if err != nil {
		return fmt.Errorf(globals.ErrCreatingSandbox, err)
	}
	execLists = append(execLists, execList...)

	sbDesc := common.SandboxDescription{
		Basedir: sandboxDef.Basedir,
		SBType:  globals.MasterSlaveLabel,
		Version: sandboxDef.Version,
		Flavor:  sandboxDef.Flavor,
		Port:    []int{sandboxDef.Port},
		Nodes:   slaves,
		NodeNum: 0,
		LogFile: sandboxDef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sandboxDef.Version,
		Flavor:      sandboxDef.Flavor,
		Port:        []int{sandboxDef.Port},
		Nodes:       []string{defaults.Defaults().MasterName},
		Destination: sandboxDef.SandboxDir,
	}

	if sandboxDef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sandboxDef.LogFileName)
	}

	// 8.0.11
	// isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
	isMinimumMySQLXDefault, err := common.HasCapability(sandboxDef.Flavor, common.MySQLXDefault, sandboxDef.Version)
	if err != nil {
		return err
	}
	if isMinimumMySQLXDefault || sandboxDef.EnableMysqlX {
		sandboxDef.MysqlXPort = baseMysqlxPort + 1
		if !sandboxDef.DisableMysqlX {
			sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+1)
			sbItem.Port = append(sbItem.Port, baseMysqlxPort+1)
			logger.Printf("Adding mysqlx port %d to master\n", baseMysqlxPort+1)
		}
	}

	if sandboxDef.EnableAdminAddress {
		sandboxDef.AdminPort = baseAdminPort + 1
		sbDesc.Port = append(sbDesc.Port, baseAdminPort+1)
		sbItem.Port = append(sbItem.Port, baseAdminPort+1)
		logger.Printf("Adding admin port %d to master\n", baseAdminPort+1)
	}

	sandboxDef.ReadOnlyOptions = readOnlyOptions
	nodeLabel := defaults.Defaults().NodePrefix
	for i := 1; i <= slaves; i++ {
		sandboxDef.Port = basePort + i + 1
		data["Slaves"] = append(data["Slaves"].([]common.StringMap), common.StringMap{
			"ShellPath":          sandboxDef.ShellPath,
			"Copyright":          globals.ShellScriptCopyright,
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
		//sandboxDef.ServerId = (baseServerId + i + 1) * 100
		sandboxDef.ServerId = setServerId(sandboxDef, i+1)
		sandboxDef.NodeNum = i + 1
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)
		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)
		// 8.0.11
		// isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
		isMinimumMySQLXDefault, err := common.HasCapability(sandboxDef.Flavor, common.MySQLXDefault, sandboxDef.Version)
		if err != nil {
			return err
		}
		if isMinimumMySQLXDefault || sandboxDef.EnableMysqlX {
			sandboxDef.MysqlXPort = baseMysqlxPort + i + 1
			if !sandboxDef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i+1)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i+1)
				logger.Printf("Adding mysqlx port %d to slave %d\n", baseMysqlxPort+i+1, i)
			}
		}
		if sandboxDef.EnableAdminAddress {
			sandboxDef.AdminPort = baseAdminPort + i + 1
			sbDesc.Port = append(sbDesc.Port, baseAdminPort+i+1)
			sbItem.Port = append(sbItem.Port, baseAdminPort+i+1)
			logger.Printf("Adding admin port %d to slave %d\n", baseAdminPort+i+1, i)
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
			sandboxDef.SemiSyncOptions = SingleTemplates[globals.TmplSemisyncSlaveOptions].Contents
		}
		logger.Printf("Creating single sandbox for slave %d\n", i)
		execListNode, err := CreateChildSandbox(sandboxDef)
		if err != nil {
			return fmt.Errorf(globals.ErrCreatingSandbox, err)
		}
		execLists = append(execLists, execListNode...)
		var dataSlave = common.StringMap{
			"ShellPath":          sandboxDef.ShellPath,
			"Copyright":          globals.ShellScriptCopyright,
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
				{fmt.Sprintf("%s%d", slaveAbbr, i), globals.TmplSlave, true},
				{fmt.Sprintf("n%d", i+1), globals.TmplSlave, true},
			}})
		if err != nil {
			return err
		}
		if sandboxDef.EnableAdminAddress {
			logger.Printf("Create slave admin script %d\n", i)
			err = writeScripts(ScriptBatch{ReplicationTemplates, logger, sandboxDef.SandboxDir, dataSlave,
				[]ScriptDef{
					{fmt.Sprintf("%sa%d", slaveAbbr, i), globals.TmplSlaveAdmin, true},
					{fmt.Sprintf("na%d", i+1), globals.TmplSlaveAdmin, true},
				}})
			if err != nil {
				return err
			}
		}
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

	slavePlural := english.PluralWord(2, slaveLabel, "")
	masterPlural := english.PluralWord(2, masterLabel, "")
	initializeSlaves := "initialize_" + slavePlural
	checkSlaves := "check_" + slavePlural
	useAllMasters := "use_all_" + masterPlural
	useAllSlaves := "use_all_" + slavePlural
	execAllSlaves := "exec_all_" + slavePlural
	execAllMasters := "exec_all_" + masterPlural

	sb := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		sandboxDir: sandboxDef.SandboxDir,
		data:       data,
		scripts: []ScriptDef{
			{globals.ScriptStartAll, globals.TmplStartAll, true},
			{globals.ScriptRestartAll, globals.TmplRestartAll, true},
			{globals.ScriptStatusAll, globals.TmplStatusAll, true},
			{globals.ScriptTestSbAll, globals.TmplTestSbAll, true},
			{globals.ScriptStopAll, globals.TmplStopAll, true},
			{globals.ScriptClearAll, globals.TmplClearAll, true},
			{globals.ScriptSendKillAll, globals.TmplSendKillAll, true},
			{globals.ScriptUseAll, globals.TmplUseAll, true},
			{globals.ScriptExecAll, globals.TmplExecAll, true},
			{globals.ScriptMetadataAll, globals.TmplMetadataAll, true},
			{useAllSlaves, globals.TmplUseAllSlaves, true},
			{useAllMasters, globals.TmplUseAllMasters, true},
			{initializeSlaves, globals.TmplInitSlaves, true},
			{checkSlaves, globals.TmplCheckSlaves, true},
			{masterAbbr, globals.TmplMaster, true},
			{execAllSlaves, globals.TmplExecAllSlaves, true},
			{execAllMasters, globals.TmplExecAllMasters, true},
			{globals.ScriptWipeRestartAll, globals.TmplWipeAndRestartAll, true},
			{"n1", globals.TmplMaster, true},
			{"test_replication", globals.TmplTestReplication, true},
			{globals.ScriptReplicateFrom, globals.TmplReplReplicateFrom, true},
			{globals.ScriptSysbench, globals.TmplReplSysbench, true},
			{globals.ScriptSysbenchReady, globals.TmplReplSysbenchReady, true},
		},
	}
	if sandboxDef.SemiSyncOptions != "" {
		sb.scripts = append(sb.scripts, ScriptDef{"post_initialization", globals.TmplSemiSyncStart, true})
	}
	if sandboxDef.EnableAdminAddress {
		sb.scripts = append(sb.scripts, ScriptDef{masterAbbr + "a", globals.TmplMasterAdmin, true})
		sb.scripts = append(sb.scripts, ScriptDef{"na1", globals.TmplMasterAdmin, true})
		sb.scripts = append(sb.scripts, ScriptDef{globals.ScriptUseAllAdmin, globals.TmplUseAllAdmin, true})
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
		out, err := common.RunCmd(path.Join(sandboxDef.SandboxDir, initializeSlaves))
		if err != nil {
			fmt.Printf("error initializing cluster: %s\n:%s", out, err)
			return err
		}
	}
	common.CondPrintf("Replication directory installed in %s\n", common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	common.CondPrintf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return nil
}

// func CreateReplicationSandbox(sdef SandboxDef, origin string, topology string, nodes int, masterIp, masterList, slaveList string) error {
func CreateReplicationSandbox(sdef SandboxDef, origin string, replData ReplicationData) error {
	if !common.IsIPV4(replData.MasterIp) {
		return fmt.Errorf("IP %s is not a valid IPV4", replData.MasterIp)
	}

	Basedir := sdef.Basedir
	if !common.DirExists(Basedir) {
		return fmt.Errorf(globals.ErrBaseDirectoryNotFound, Basedir)
	}

	sandboxDir := sdef.SandboxDir
	switch replData.Topology {
	case globals.MasterSlaveLabel:
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().MasterSlavePrefix+common.VersionToName(origin))
	case globals.GroupLabel:
		if sdef.SinglePrimary {
			sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().GroupSpPrefix+common.VersionToName(origin))
		} else {
			sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().GroupPrefix+common.VersionToName(origin))
		}
		// 5.7.17
		// isMinimumGroupRepl, err := common.GreaterOrEqualVersion(sdef.Version, globals.MinimumGroupReplVersion)
		isMinimumGroupRepl, err := common.HasCapability(sdef.Flavor, common.GroupReplication, sdef.Version)
		if err != nil {
			return err
		}
		if !isMinimumGroupRepl {
			return fmt.Errorf(globals.ErrFeatureRequiresVersion, "group replication", common.IntSliceToDottedString(globals.MinimumGroupReplVersion))
		}
	case globals.FanInLabel:
		// 5.7.9
		// isMinimumMultiSource, err := common.GreaterOrEqualVersion(sdef.Version, globals.MinimumMultiSourceReplVersion)
		isMinimumMultiSource, err := common.HasCapability(sdef.Flavor, common.MultiSource, sdef.Version)
		if err != nil {
			return err
		}
		if !isMinimumMultiSource {
			return fmt.Errorf(globals.ErrFeatureRequiresVersion, "multi-source replication", common.IntSliceToDottedString(globals.MinimumMultiSourceReplVersion))
		}
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().FanInPrefix+common.VersionToName(origin))
	case globals.AllMastersLabel:
		// 5.7.9

		// isMinimumMultiSource, err := common.GreaterOrEqualVersion(sdef.Version, globals.MinimumMultiSourceReplVersion)
		isMinimumMultiSource, err := common.HasCapability(sdef.Flavor, common.MultiSource, sdef.Version)
		if err != nil {
			return err
		}
		if !isMinimumMultiSource {
			return fmt.Errorf(globals.ErrFeatureRequiresVersion, "multi-source replication", common.IntSliceToDottedString(globals.MinimumMultiSourceReplVersion))
		}
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().AllMastersPrefix+common.VersionToName(origin))
	case globals.PxcLabel:
		isMinimumPxc, err := common.HasCapability(sdef.Flavor, common.XtradbCluster, sdef.Version)
		if err != nil {
			return err
		}
		if !isMinimumPxc {
			return fmt.Errorf(globals.ErrFeatureRequiresCapability, "Xtradb Cluster", common.PxcFlavor, common.IntSliceToDottedString(globals.MinimumXtradbClusterVersion))
		}
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().PxcPrefix+common.VersionToName(origin))
	case globals.NdbLabel:
		isMinimumNdb, err := common.HasCapability(sdef.Flavor, common.NdbCluster, sdef.Version)
		if err != nil {
			return err
		}
		if !isMinimumNdb {
			return fmt.Errorf(globals.ErrFeatureRequiresCapability, "NDB Cluster", common.NdbFlavor,
				common.IntSliceToDottedString(globals.MinimumNdbClusterVersion))
		}
		sdef.SandboxDir = path.Join(sdef.SandboxDir, defaults.Defaults().NdbPrefix+common.VersionToName(origin))
	default:
		return fmt.Errorf("unrecognized topology. Accepted: '%v'", globals.AllowedTopologies)
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
	switch replData.Topology {
	case globals.MasterSlaveLabel:
		err = CreateMasterSlaveReplication(sdef, origin, replData.Nodes, replData.MasterIp)
	case globals.GroupLabel:
		err = CreateGroupReplication(sdef, origin, replData.Nodes, replData.MasterIp)
	case globals.FanInLabel:
		err = CreateFanInReplication(sdef, origin, replData.Nodes, replData.MasterIp, replData.MasterList, replData.SlaveList)
	case globals.AllMastersLabel:
		err = CreateAllMastersReplication(sdef, origin, replData.Nodes, replData.MasterIp)
	case globals.PxcLabel:
		err = CreatePxcReplication(sdef, origin, replData.Nodes, replData.MasterIp)
	case globals.NdbLabel:
		err = CreateNdbReplication(sdef, origin, replData.Nodes, replData.NdbNodes, replData.MasterIp)
	}
	return err
}
