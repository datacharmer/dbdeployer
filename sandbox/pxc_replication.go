// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/dustin/go-humanize/english"
	"github.com/pkg/errors"
)

var pxcReplicationOptions string = `
innodb_file_per_table
innodb_autoinc_lock_mode=2
wsrep-provider=__BASEDIR__/lib/libgalera_smm.so
wsrep_cluster_address=__GROUP_COMMUNICATION__
wsrep_node_incoming_address=127.0.0.1
wsrep_provider_options=gmcast.listen_addr=tcp://127.0.0.1:__GROUP_PORT__
wsrep_sst_method=rsync
wsrep_sst_auth=root:
wsrep_node_address=127.0.0.1
innodb_flush_method=O_DIRECT
core-file
secure-file-priv=
loose-innodb-status-file=1
log-output=none
wsrep_slave_threads=2
wsrep_sst_receive_address=127.0.0.1:__RSYNC_PORT__
`

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

	baseServerId := 0
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
	changeMasterExtra := ""

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

	var groupCommunication string = ""
	var auxGroupCommunication string = ""

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
			"RplPassword":       sandboxDef.RplPassword})

		sandboxDef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.MorePorts = []int{
			groupPort,
			groupPort + 1, // IST port
			rsyncPort}
		sandboxDef.ServerId = (baseServerId + i) * 100
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)

		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)

		sbItem.Port = append(sbItem.Port, groupPort)
		sbDesc.Port = append(sbDesc.Port, groupPort)

		sbItem.Port = append(sbItem.Port, groupPort+1) // IST port
		sbDesc.Port = append(sbDesc.Port, groupPort+1) // IST port

		sbItem.Port = append(sbItem.Port, rsyncPort)
		sbDesc.Port = append(sbDesc.Port, rsyncPort)

		if i == 1 {
			groupCommunication = "gcomm://"
			auxGroupCommunication = fmt.Sprintf("gcomm://%s:%d", masterIp, groupPort)
		} else {
			auxGroupCommunication += ","
			auxGroupCommunication += fmt.Sprintf("gcomm://%s:%d", masterIp, groupPort)
			groupCommunication = auxGroupCommunication
		}

		if !sandboxDef.RunConcurrently {
			installationMessage := "Installing and starting %s %d\n"
			if sandboxDef.SkipStart {
				installationMessage = "Installing %s %d\n"
			}
			common.CondPrintf(installationMessage, nodeLabel, i)
			logger.Printf(installationMessage, nodeLabel, i)
		}
		sandboxDef.ReplOptions = SingleTemplates["replication_options"].Contents + fmt.Sprintf("\n%s\n", pxcReplicationOptions)
		reMasterIp := regexp.MustCompile(`127\.0\.0\.1`)
		reGroupPort := regexp.MustCompile(`__GROUP_PORT__`)
		reRSyncPort := regexp.MustCompile(`__RSYNC_PORT__`)
		reGroupCommunication := regexp.MustCompile(`__GROUP_COMMUNICATION__`)
		reBasedir := regexp.MustCompile(`__BASEDIR__`)
		sandboxDef.ReplOptions = reMasterIp.ReplaceAllString(sandboxDef.ReplOptions, masterIp)
		sandboxDef.ReplOptions = reGroupCommunication.ReplaceAllString(sandboxDef.ReplOptions, groupCommunication)
		sandboxDef.ReplOptions = reBasedir.ReplaceAllString(sandboxDef.ReplOptions, sandboxDef.Basedir)
		sandboxDef.ReplOptions = reGroupPort.ReplaceAllString(sandboxDef.ReplOptions, fmt.Sprintf("%d", groupPort))
		sandboxDef.ReplOptions = reRSyncPort.ReplaceAllString(sandboxDef.ReplOptions, fmt.Sprintf("%d", rsyncPort))

		sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates["gtid_options_57"].Contents)
		sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates["repl_crash_safe_options"].Contents)
		// 8.0.11
		// isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
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
			sandboxDef.LoadGrants = true
		} else {
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
		err = writeScript(logger, MultipleTemplates, fmt.Sprintf("n%d", i), "node_template", sandboxDef.SandboxDir, dataNode, true)
		if err != nil {
			return err
		}
		if sandboxDef.EnableAdminAddress {
			logger.Printf("Create admin script for node %d\n", i)
			err = writeScript(logger, MultipleTemplates, fmt.Sprintf("na%d", i),
				"node_admin_template", sandboxDef.SandboxDir, dataNode, true)
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
			{globals.ScriptStartAll, "start_multi_template", true},
			{globals.ScriptRestartAll, "restart_multi_template", true},
			{globals.ScriptStatusAll, "status_multi_template", true},
			{globals.ScriptTestSbAll, "test_sb_multi_template", true},
			{globals.ScriptStopAll, "stop_multi_template", true},
			{globals.ScriptClearAll, "clear_multi_template", true},
			{globals.ScriptSendKillAll, "send_kill_multi_template", true},
			{globals.ScriptUseAll, "use_multi_template", true},
			{globals.ScriptReplicateFrom, "replicate_from_multi_template", true},
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
			{useAllSlaves, "multi_source_use_slaves_template", true},
			{useAllMasters, "multi_source_use_masters_template", true},
			{globals.ScriptTestReplication, "multi_source_test_template", true},
		},
	}
	sbPxc := ScriptBatch{
		tc:         PxcTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptCheckNodes, "check_pxc_nodes_template", true},
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
		err = writeScript(logger, MultipleTemplates, "use_all_admin",
			"use_multi_admin_template", sandboxDef.SandboxDir, data, true)
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
