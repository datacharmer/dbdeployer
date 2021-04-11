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
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/pkg/errors"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

func CreatePxcReplication(sandboxDef SandboxDef, origin string, nodes int, masterIp string) error {
	var execLists []concurrent.ExecutionList

	err := common.CheckPrerequisites("PXC", globals.NeededPxcExecutables)
	if err != nil {
		return err
	}
	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		var fileName string
		var err error
		logger, fileName, err = defaults.NewLogger(common.LogDirName(), "pxc-replication")
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
		return fmt.Errorf("options --read-only and --super-read-only can't be used for PXC topology")
	}

	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
	rev := vList[2]
	basePort := sandboxDef.Port + defaults.Defaults().PxcBasePort + (rev * 100)
	if sandboxDef.BasePort > 0 {
		basePort = sandboxDef.BasePort
	}

	//baseServerId := sandboxDef.BaseServerId
	if nodes < 3 {
		return fmt.Errorf("can't run PXC replication with less than 3 nodes")
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
	pxcPortDelta := defaults.Defaults().GroupPortDelta
	pxcRsyncPortDelta := pxcPortDelta + nodes + 10
	basePort = firstGroupPort - 1
	baseGroupPort := basePort + pxcPortDelta
	baseRsyncPort := basePort + pxcRsyncPortDelta

	firstGroupPort, err = common.FindFreePort(baseGroupPort+1, sandboxDef.InstalledPorts, nodes*2)
	if err != nil {
		return errors.Wrapf(err, "error retrieving PXC replication free port")
	}
	baseGroupPort = firstGroupPort - 1
	for checkPort := basePort + 1; checkPort < basePort+nodes+1; checkPort++ {
		err = checkPortAvailability("CreatePxcReplication", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
		if err != nil {
			return err
		}
	}
	for checkPort := baseGroupPort + 1; checkPort < baseGroupPort+(nodes*2)+1; checkPort++ {
		err = checkPortAvailability("CreatePxcReplication-group", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
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

	firstGroupPort, err = common.FindFreePort(baseRsyncPort+1, sandboxDef.InstalledPorts, nodes)
	if err != nil {
		return errors.Wrapf(err, "error retrieving PXC replication free port")
	}
	baseRsyncPort = firstGroupPort - 1
	for checkPort := baseRsyncPort + 1; checkPort < baseRsyncPort+nodes+1; checkPort++ {
		err = checkPortAvailability("CreatePxcReplication-rsync", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
		if err != nil {
			return err
		}
	}

	err = os.Mkdir(sandboxDef.SandboxDir, globals.PublicDirectoryAttr)
	if err != nil {
		return err
	}
	common.AddToCleanupStack(common.Rmdir, "Rmdir", sandboxDef.SandboxDir)
	logger.Printf("Creating directory %s\n", sandboxDef.SandboxDir)
	timestamp := time.Now()
	slaveLabel := defaults.Defaults().SlavePrefix
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterAbbr := defaults.Defaults().MasterAbbr
	masterLabel := defaults.Defaults().MasterName
	masterList := makeNodesList(nodes)
	slaveList := masterList
	changeMasterExtra := setChangeMasterProperties("", sandboxDef.ChangeMasterOptions, logger)

	var pxcEncryptClusterTraffic = "pxc_encrypt_cluster_traffic=off"

	isMinimumEncryptTraffic, err := common.HasCapability(sandboxDef.Flavor, common.XtradbClusterEncryptCluster, sandboxDef.Version)
	if err != nil {
		return err
	}
	if !isMinimumEncryptTraffic {
		pxcEncryptClusterTraffic = "# pxc_encrypt_cluster_traffic=off # requires PXC 5.7"
	}
	stopNodeList := ""
	for i := nodes; i > 0; i-- {
		stopNodeList += fmt.Sprintf(" %d", i)
	}
	nodeLabel := defaults.Defaults().NodePrefix
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
	// Connection ports are 1, 3, 5, etc
	// IST ports are 2, 4, 6, etc
	groupPorts := []int{0}
	for i := 1; i <= nodes*2; i++ {
		if i%2 == 0 {
			groupPort := baseGroupPort + i
			groupPorts = append(groupPorts, groupPort)
			if connectionString != "" {
				connectionString += ","
			}
			connectionString += fmt.Sprintf("127.0.0.1:%d", groupPort)
		}
	}
	logger.Printf("Creating connection string %s\n", connectionString)

	sbType := "Percona-Xtradb-Cluster"

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

	skipLogSlaveUpdates, err := common.HasCapability(common.PxcFlavor, common.XtradbClusterNoSlaveUpdates, sandboxDef.Version)
	if err != nil {
		return err
	}
	if !skipLogSlaveUpdates {
		sandboxDef.ReplOptions = fmt.Sprintf("%s\nlog_slave_updates=ON\n", sandboxDef.ReplOptions)
	}
	baseReplicationOptions := sandboxDef.ReplOptions
	var groupCommunication string = "gcomm://"
	//var auxGroupCommunication string = ""
	var sstMethod = "rsync"

	// XtraDB 8.0.15+
	isMinimumXtrabackupSupport, err := common.HasCapability(sandboxDef.Flavor, common.XtradbClusterXtrabackup, sandboxDef.Version)
	if err != nil {
		return err
	}
	if isMinimumXtrabackupSupport {
		sstMethod = "xtrabackup-v2"
	}

	for i := 1; i <= nodes; i++ {
		groupPort := groupPorts[i]
		groupCommunication += fmt.Sprintf("%s:%d", masterIp, groupPort)
		if i < nodes {
			groupCommunication += ","
		}
	}

	for i := 1; i <= nodes; i++ {
		groupPort := groupPorts[i]
		rsyncPort := baseRsyncPort + i
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
			"StopNodeList":      stopNodeList,
			"SandboxDir":        sandboxDef.SandboxDir,
			"RplUser":           sandboxDef.RplUser,
			"RplPassword":       sandboxDef.RplPassword,
		})

		sandboxDef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.MorePorts = []int{
			groupPort,
			groupPort + 1, // IST port
			rsyncPort}
		// sandboxDef.ServerId = (baseServerId + i) * 100
		sandboxDef.ServerId = setServerId(sandboxDef, i)
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)

		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)

		sbItem.Port = append(sbItem.Port, groupPort)
		sbDesc.Port = append(sbDesc.Port, groupPort)

		sbItem.Port = append(sbItem.Port, groupPort+1) // IST port
		sbDesc.Port = append(sbDesc.Port, groupPort+1) // IST port

		sbItem.Port = append(sbItem.Port, rsyncPort)
		sbDesc.Port = append(sbDesc.Port, rsyncPort)

		if !sandboxDef.RunConcurrently {
			installationMessage := "Installing and starting %s %d\n"
			if sandboxDef.SkipStart {
				installationMessage = "Installing %s %d\n"
			}
			common.CondPrintf(installationMessage, nodeLabel, i)
			logger.Printf(installationMessage, nodeLabel, i)
		}

		pxcReplicationText := PxcTemplates[globals.TmplPxcReplication].Contents

		pxcReplicationData := common.StringMap{
			"NodeIp":                   masterIp,
			"GroupCommunication":       groupCommunication,
			"Basedir":                  sandboxDef.Basedir,
			"RsyncPort":                rsyncPort,
			"GroupPort":                groupPort,
			"SstMethod":                sstMethod,
			"PxcEncryptClusterTraffic": pxcEncryptClusterTraffic,
		}
		pxcFilledTemplate, err := common.SafeTemplateFill(globals.TmplPxcReplication, pxcReplicationText, pxcReplicationData)
		if err != nil {
			return fmt.Errorf("error filling pxc replication template %s", err)
		}

		sandboxDef.ReplOptions = baseReplicationOptions + fmt.Sprintf("\n%s\n", pxcFilledTemplate)

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
		if i == 1 {
			sandboxDef.StartArgs = []string{"--wsrep-new-cluster"}
			sandboxDef.LoadGrants = true
		} else {
			sandboxDef.StartArgs = []string{}
			sandboxDef.LoadGrants = false
		}
		sandboxDef.Prompt = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.SBType = "pxc-node"
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
			logger.Printf("Create admin script for node %d\n", i)
			err = writeScript(logger, MultipleTemplates, fmt.Sprintf("na%d", i),
				globals.TmplNodeAdmin, sandboxDef.SandboxDir, dataNode, true)
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

	logger.Printf("Writing PXC replication scripts\n")
	sbMultiple := ScriptBatch{
		tc:         MultipleTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
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
		},
	}

	slavePlural := english.PluralWord(2, slaveLabel, "")
	masterPlural := english.PluralWord(2, masterLabel, "")
	useAllMasters := "use_all_" + masterPlural
	useAllSlaves := "use_all_" + slavePlural

	sbRepl := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{useAllSlaves, globals.TmplMultiSourceUseSlaves, true},
			{useAllMasters, globals.TmplMultiSourceUseMasters, true},
			{globals.ScriptTestReplication, globals.TmplMultiSourceTest, true},
		},
	}
	sbPxc := ScriptBatch{
		tc:         PxcTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptStartAll, globals.TmplPxcStart, true},
			{globals.ScriptCheckNodes, globals.TmplPxcCheckNodes, true},
		},
	}

	for _, sb := range []ScriptBatch{sbMultiple, sbRepl, sbPxc} {
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

	common.CondPrintf("Replication directory installed in %s\n", common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	common.CondPrintf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return nil
}
