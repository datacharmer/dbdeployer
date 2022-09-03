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
	"regexp"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/dustin/go-humanize/english"
	"github.com/pkg/errors"
)

func getBaseMysqlxPort(basePort int, sdef SandboxDef, nodes int) (int, error) {
	baseMysqlxPort := basePort + defaults.Defaults().MysqlXPortDelta
	// 8.0.11
	// isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sdef.Version, globals.MinimumMysqlxDefaultVersion)
	isMinimumMySQLXDefault, err := common.HasCapability(sdef.Flavor, common.MySQLX, sdef.Version)
	if err != nil {
		return 0, err
	}
	if isMinimumMySQLXDefault {
		// FindFreePort returns the first free port, but base_port will be used
		// with a counter. Thus the availability will be checked using
		// "base_port + 1"
		firstGroupPort, err := common.FindFreePort(baseMysqlxPort+1, sdef.InstalledPorts, nodes)
		if err != nil {
			return -1, errors.Wrapf(err, "error finding a free port for MySQLX")
		}
		baseMysqlxPort = firstGroupPort - 1
		for N := 1; N <= nodes; N++ {
			checkPort := baseMysqlxPort + N
			err = checkPortAvailability("get_base_mysqlx_port", sdef.SandboxDir, sdef.InstalledPorts, checkPort)
			if err != nil {
				return 0, err
			}
		}
	}
	return baseMysqlxPort, nil
}

func getBaseAdminPort(basePort int, sdef SandboxDef, nodes int) (int, error) {
	baseAdminPort := basePort + defaults.Defaults().AdminPortDelta
	if !sdef.EnableAdminAddress {
		return basePort, nil
	}
	// 8.0.14
	isMinimumAdminAddress, err := common.HasCapability(sdef.Flavor, common.AdminAddress, sdef.Version)
	if err != nil {
		return 0, err
	}
	if !isMinimumAdminAddress {
		return 0, fmt.Errorf(globals.ErrFeatureRequiresCapability,
			globals.EnableAdminAddressLabel,
			common.MySQLFlavor,
			common.IntSliceToDottedString(globals.MinimumAdminAddressVersion))
	}
	if isMinimumAdminAddress {
		// FindFreePort returns the first free port, but base_port will be used
		// with a counter. Thus the availability will be checked using
		// "base_port + 1"
		firstAdminPort, err := common.FindFreePort(baseAdminPort+1, sdef.InstalledPorts, nodes)
		if err != nil {
			return -1, errors.Wrapf(err, "error finding a free admin port")
		}
		baseAdminPort = firstAdminPort - 1
		for N := 1; N <= nodes; N++ {
			checkPort := baseAdminPort + N
			err := checkPortAvailability("get_base_admin_port", sdef.SandboxDir, sdef.InstalledPorts, checkPort)
			if err != nil {
				return 0, err
			}
		}
	}
	return baseAdminPort, nil
}

func CreateGroupReplication(sandboxDef SandboxDef, origin string, nodes int, masterIp string) error {
	var execLists []concurrent.ExecutionList
	var err error

	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		var fileName string
		var err error
		logger, fileName, err = defaults.NewLogger(common.LogDirName(), "group-replication")
		if err != nil {
			return err
		}
		sandboxDef.LogFileName = common.ReplaceLiteralHome(fileName)
	}

	readOnlyOptions, err := checkReadOnlyFlags(sandboxDef)
	if err != nil {
		return err
	}
	if readOnlyOptions != "" {
		return fmt.Errorf("options --read-only and --super-read-only can't be used for group topology\n" +
			"as the group replication software sets it when needed")
	}

	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
	rev := vList[2]
	basePort := computeBaseport(sandboxDef.Port + defaults.Defaults().GroupReplicationBasePort + (rev * 100))
	if sandboxDef.SinglePrimary {
		basePort = sandboxDef.Port + defaults.Defaults().GroupReplicationSpBasePort + (rev * 100)
	}
	if sandboxDef.BasePort > 0 {
		basePort = sandboxDef.BasePort
	}

	// baseServerId := sandboxDef.BaseServerId
	if nodes < 3 {
		return fmt.Errorf("can't run group replication with less than 3 nodes")
	}
	if common.DirExists(sandboxDef.SandboxDir) {
		sandboxDef, err = checkDirectory(sandboxDef)
		if err != nil {
			return err
		}
	}
	// FindFreePort returns the first free port, but base_port will be used
	// with a counter. Thus the availability will be checked using
	// "base_port + 1"
	firstGroupPort, err := common.FindFreePort(basePort+1, sandboxDef.InstalledPorts, nodes)
	if err != nil {
		return errors.Wrapf(err, "error retrieving free port for replication")
	}
	basePort = firstGroupPort - 1
	baseGroupPort := basePort + defaults.Defaults().GroupPortDelta
	firstGroupPort, err = common.FindFreePort(baseGroupPort+1, sandboxDef.InstalledPorts, nodes)
	if err != nil {
		return errors.Wrapf(err, "error retrieving group replication free port")
	}
	baseGroupPort = firstGroupPort - 1
	for checkPort := basePort + 1; checkPort < basePort+nodes+1; checkPort++ {
		err = checkPortAvailability("CreateGroupReplication", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
		if err != nil {
			return err
		}
	}
	for checkPort := baseGroupPort + 1; checkPort < baseGroupPort+nodes+1; checkPort++ {
		err = checkPortAvailability("CreateGroupReplication-group", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
		if err != nil {
			return err
		}
	}
	baseMysqlxPort, err := getBaseMysqlxPort(basePort, sandboxDef, nodes)
	if err != nil {
		return err
	}
	baseAdminPort, err := getBaseAdminPort(basePort, sandboxDef, nodes)
	if err != nil {
		return err
	}
	err = os.Mkdir(sandboxDef.SandboxDir, globals.PublicDirectoryAttr)
	if err != nil {
		return err
	}
	common.AddToCleanupStack(common.RmdirAll, "RmdirAll", sandboxDef.SandboxDir)
	logger.Printf("Creating directory %s\n", sandboxDef.SandboxDir)
	timestamp := time.Now()
	slaveLabel := defaults.Defaults().SlavePrefix
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterAbbr := defaults.Defaults().MasterAbbr
	masterLabel := defaults.Defaults().MasterName
	masterList := makeNodesList(nodes)
	slaveList := masterList
	if sandboxDef.SinglePrimary {
		masterList = "1"
		slaveList = ""
		for N := 2; N <= nodes; N++ {
			if slaveList != "" {
				slaveList += " "
			}
			slaveList += fmt.Sprintf("%d", N)
		}
		mlist, err := nodesListToIntSlice(masterList, nodes)
		if err != nil {
			return err
		}
		slist, err := nodesListToIntSlice(slaveList, nodes)
		if err != nil {
			return err
		}
		err = checkNodeLists(nodes, mlist, slist)
		if err != nil {
			return err
		}
	}
	changeMasterExtra := setChangeMasterProperties("", sandboxDef.ChangeMasterOptions, logger)
	nodeLabel := defaults.Defaults().NodePrefix
	stopNodeList := ""
	for i := nodes; i > 0; i-- {
		stopNodeList += fmt.Sprintf(" %d", i)
	}
	var data = common.StringMap{
		"ShellPath":         sandboxDef.ShellPath,
		"Copyright":         globals.ShellScriptCopyright,
		"AppVersion":        common.VersionDef,
		"DateTime":          timestamp.Format(time.UnixDate),
		"SandboxDir":        sandboxDef.SandboxDir,
		"MasterIp":          masterIp,
		"MasterList":        masterList,
		"NodeLabel":         nodeLabel,
		"SlaveList":         slaveList,
		"RplUser":           sandboxDef.RplUser,
		"RplPassword":       sandboxDef.RplPassword,
		"SlaveLabel":        slaveLabel,
		"SlaveAbbr":         slaveAbbr,
		"ChangeMasterExtra": changeMasterExtra,
		"MasterLabel":       masterLabel,
		"MasterAbbr":        masterAbbr,
		"StopNodeList":      stopNodeList,
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
	singlePrimaryMode := "off"
	if sandboxDef.SinglePrimary {
		sbType = "group-single-primary"
		singlePrimaryMode = "on"
	}
	logger.Printf("Defining group type %s\n", sbType)

	sbDesc := common.SandboxDescription{
		Basedir: sandboxDef.Basedir,
		SBType:  sbType,
		Version: sandboxDef.Version,
		Flavor:  sandboxDef.Flavor,
		Port:    []int{},
		Nodes:   nodes,
		NodeNum: 0,
		LogFile: sandboxDef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sandboxDef.Version,
		Flavor:      sandboxDef.Flavor,
		Port:        []int{},
		Nodes:       []string{},
		Destination: sandboxDef.SandboxDir,
	}

	if sandboxDef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sandboxDef.LogFileName)
	}

	for i := 1; i <= nodes; i++ {
		groupPort := baseGroupPort + i
		sandboxDef.Port = basePort + i
		data["Nodes"] = append(data["Nodes"].([]common.StringMap), common.StringMap{
			"ShellPath":         sandboxDef.ShellPath,
			"Copyright":         globals.ShellScriptCopyright,
			"AppVersion":        common.VersionDef,
			"DateTime":          timestamp.Format(time.UnixDate),
			"Node":              i,
			"NodePort":          sandboxDef.Port,
			"MasterIp":          masterIp,
			"NodeLabel":         nodeLabel,
			"SlaveLabel":        slaveLabel,
			"SlaveAbbr":         slaveAbbr,
			"ChangeMasterExtra": changeMasterExtra,
			"MasterLabel":       masterLabel,
			"MasterAbbr":        masterAbbr,
			"SandboxDir":        sandboxDef.SandboxDir,
			"StopNodeList":      stopNodeList,
			"RplUser":           sandboxDef.RplUser,
			"RplPassword":       sandboxDef.RplPassword})

		sandboxDef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.MorePorts = []int{groupPort}
		// sandboxDef.ServerId = (baseServerId + i) * 100
		sandboxDef.ServerId = setServerId(sandboxDef, i)
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)
		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)
		sbItem.Port = append(sbItem.Port, groupPort)
		sbDesc.Port = append(sbDesc.Port, groupPort)

		if !sandboxDef.RunConcurrently {
			installationMessage := "Installing and starting %s %d\n"
			if sandboxDef.SkipStart {
				installationMessage = "Installing %s %d\n"
			}
			common.CondPrintf(installationMessage, nodeLabel, i)
			logger.Printf(installationMessage, nodeLabel, i)
		}

		basePortText := fmt.Sprintf("%08d", basePort)
		replicationData := common.StringMap{
			"BasePort":       basePortText,
			"GroupSeeds":     connectionString,
			"LocalAddresses": fmt.Sprintf("%s:%d", masterIp, groupPort),
			"PrimaryMode":    singlePrimaryMode,
		}

		replOptionsText, err := common.SafeTemplateFill("group_replication",
			GroupTemplates[globals.TmplGroupReplOptions].Contents, replicationData)
		if err != nil {
			return err
		}
		sandboxDef.ReplOptions = SingleTemplates[globals.TmplReplicationOptions].Contents + "\n" + replOptionsText

		reMasterIp := regexp.MustCompile(`127\.0\.0\.1`)
		sandboxDef.ReplOptions = reMasterIp.ReplaceAllString(sandboxDef.ReplOptions, masterIp)

		sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates[globals.TmplGtidOptions57].Contents)
		sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates[globals.TmplReplCrashSafeOptions].Contents)

		// 8.0.11
		isMinimumMySQLXDefault, err := common.HasCapability(sandboxDef.Flavor, common.MySQLXDefault, sandboxDef.Version)
		if err != nil {
			return err
		}
		if isMinimumMySQLXDefault || sandboxDef.EnableMysqlX {
			sandboxDef.MysqlXPort = baseMysqlxPort + i
			if !sandboxDef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i)
				logger.Printf("adding port %d to node %d\n", baseMysqlxPort+i, i)
			}
		}
		if sandboxDef.EnableAdminAddress {
			sandboxDef.AdminPort = baseAdminPort + i
			sbDesc.Port = append(sbDesc.Port, baseAdminPort+i)
			sbItem.Port = append(sbItem.Port, baseAdminPort+i)
			logger.Printf("adding port %d to node %d\n", baseAdminPort+i, i)
		}
		sandboxDef.Multi = true
		sandboxDef.LoadGrants = true
		sandboxDef.Prompt = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.SBType = "group-node"
		sandboxDef.NodeNum = i
		// common.CondPrintf("%#v\n",sdef)
		logger.Printf("Create single sandbox for node %d\n", i)
		execList, err := CreateChildSandbox(sandboxDef)
		if err != nil {
			return fmt.Errorf(globals.ErrCreatingSandbox, err)
		}
		execLists = append(execLists, execList...)
		var dataNode = common.StringMap{
			"ShellPath":         sandboxDef.ShellPath,
			"Copyright":         globals.ShellScriptCopyright,
			"AppVersion":        common.VersionDef,
			"DateTime":          timestamp.Format(time.UnixDate),
			"Node":              i,
			"NodePort":          sandboxDef.Port,
			"NodeLabel":         nodeLabel,
			"MasterLabel":       masterLabel,
			"MasterAbbr":        masterAbbr,
			"ChangeMasterExtra": changeMasterExtra,
			"SlaveLabel":        slaveLabel,
			"SlaveAbbr":         slaveAbbr,
			"SandboxDir":        sandboxDef.SandboxDir,
		}
		logger.Printf("Create node script for node %d\n", i)
		err = writeScript(logger, MultipleTemplates, fmt.Sprintf("n%d", i), globals.TmplNode, sandboxDef.SandboxDir, dataNode, true)
		if err != nil {
			return err
		}
		if sandboxDef.EnableAdminAddress {
			err = writeScript(logger, MultipleTemplates, fmt.Sprintf("na%d", i), globals.TmplNodeAdmin, sandboxDef.SandboxDir, dataNode, true)
			if err != nil {
				return err
			}

		}
	}
	logger.Printf("Writing sandbox description in %s\n", sandboxDef.SandboxDir)
	err = common.WriteSandboxDescription(sandboxDef.SandboxDir, sbDesc)
	if err != nil {
		return errors.Wrapf(err, "unable to write sandbox description")
	}
	err = defaults.UpdateCatalog(sandboxDef.SandboxDir, sbItem)
	if err != nil {
		return errors.Wrapf(err, "unable to update catalog")
	}

	slavePlural := english.PluralWord(2, slaveLabel, "")
	masterPlural := english.PluralWord(2, masterLabel, "")
	useAllMasters := "use_all_" + masterPlural
	useAllSlaves := "use_all_" + slavePlural
	execAllSlaves := "exec_all_" + slavePlural
	execAllMasters := "exec_all_" + masterPlural

	logger.Printf("Writing group replication scripts\n")
	sbMultiple := ScriptBatch{
		tc:         MultipleTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptStartAll, globals.TmplStartMulti, true},
			{globals.ScriptRestartAll, globals.TmplRestartMulti, true},
			{globals.ScriptStatusAll, globals.TmplStatusMulti, true},
			{globals.ScriptTestSbAll, globals.TmplTestSbMulti, true},
			{globals.ScriptStopAll, globals.TmplStopMulti, true},
			{globals.ScriptClearAll, globals.TmplClearMulti, true},
			{globals.ScriptSendKillAll, globals.TmplSendKillMulti, true},
			{globals.ScriptUseAll, globals.TmplUseMulti, true},
			{globals.ScriptMetadataAll, globals.TmplMetadataMulti, true},
			{globals.ScriptReplicateFrom, globals.TmplReplicateFromMulti, true},
			{globals.ScriptSysbench, globals.TmplSysbenchMulti, true},
			{globals.ScriptSysbenchReady, globals.TmplSysbenchReadyMulti, true},
			{globals.ScriptExecAll, globals.TmplExecMulti, true},
		},
	}
	sbRepl := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{useAllSlaves, globals.TmplMultiSourceUseSlaves, true},
			{useAllMasters, globals.TmplMultiSourceUseMasters, true},
			{execAllMasters, globals.TmplMultiSourceExecMasters, true},
			{execAllSlaves, globals.TmplMultiSourceExecSlaves, true},
			{globals.ScriptTestReplication, globals.TmplMultiSourceTest, true},
			{globals.ScriptWipeRestartAll, globals.TmplWipeAndRestartAll, true},
		},
	}
	sbGroup := ScriptBatch{
		tc:         GroupTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptInitializeNodes, globals.TmplInitNodes, true},
			{globals.ScriptCheckNodes, globals.TmplCheckNodes, true},
		},
	}

	for _, sb := range []ScriptBatch{sbMultiple, sbRepl, sbGroup} {
		err := writeScripts(sb)
		if err != nil {
			return err
		}
	}
	if sandboxDef.EnableAdminAddress {
		logger.Printf("Creating admin script for all nodes\n")
		err = writeScript(logger, MultipleTemplates, globals.ScriptUseAllAdmin,
			globals.TmplUseMultiAdmin, sandboxDef.SandboxDir, data, true)
		if err != nil {
			return err
		}
	}

	logger.Printf("Running parallel tasks\n")
	concurrent.RunParallelTasksByPriority(execLists)
	if !sandboxDef.SkipStart {
		common.CondPrintln(path.Join(common.ReplaceLiteralHome(sandboxDef.SandboxDir), globals.ScriptInitializeNodes))
		logger.Printf("Running group replication initialization script\n")
		_, err := common.RunCmd(path.Join(sandboxDef.SandboxDir, globals.ScriptInitializeNodes))
		if err != nil {
			return fmt.Errorf("error initializing group replication: %s", err)
		}
	}
	common.CondPrintf("Group Replication directory installed in %s\n", common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	common.CondPrintf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return nil
}
