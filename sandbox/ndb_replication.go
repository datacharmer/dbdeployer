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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/pkg/errors"
	"os"
	"path"
	"regexp"
	"time"
)

func CreateNdbReplication(sandboxDef SandboxDef, origin string, nodes int, ndbNodes int, masterIp string) error {
	var execLists []concurrent.ExecutionList
	var err error

	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		var fileName string
		var err error
		logger, fileName, err = defaults.NewLogger(common.LogDirName(), "ndb-replication")
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
		return fmt.Errorf("options --read-only and --super-read-only can't be used for NDB topology\n")
	}

	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
	rev := vList[2]
	basePort := sandboxDef.Port + defaults.Defaults().NdbBasePort + (rev * 100)
	if sandboxDef.BasePort > 0 {
		basePort = sandboxDef.BasePort
	}

	baseServerId := 0
	if ndbNodes < 3 {
		return fmt.Errorf("can't run MySQL Cluster with less than 3 NDB nodes")
	}
	if nodes < 2 {
		return fmt.Errorf("can't run MySQL Cluster with less than 2 SQL nodes")
	}
	if common.DirExists(sandboxDef.SandboxDir) {
		sandboxDef, err = checkDirectory(sandboxDef)
		if err != nil {
			return err
		}
	}

	ndbClusterPort := defaults.Defaults().NdbClusterPort + (rev * 100)

	ndbClusterPort, err = common.FindFreePort(ndbClusterPort, sandboxDef.InstalledPorts, 1)
	if err != nil {
		return errors.Wrapf(err, "error retrieving free port for ndb cluster")
	}

	sandboxDef.InstalledPorts = append(sandboxDef.InstalledPorts, ndbClusterPort)

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
	common.AddToCleanupStack(common.Rmdir, "Rmdir", sandboxDef.SandboxDir)
	logger.Printf("Creating directory %s\n", sandboxDef.SandboxDir)
	for i := 1; i <= ndbNodes; i++ {
		nodeName := fmt.Sprintf("ndb%s%d", defaults.Defaults().NodePrefix, i)
		nodeDir := path.Join(sandboxDef.SandboxDir, nodeName)
		logger.Printf("Creating directory %s\n", nodeDir)
		err = os.Mkdir(nodeDir, globals.PublicDirectoryAttr)
		if err != nil {
			return err
		}
	}
	timestamp := time.Now()
	slaveLabel := defaults.Defaults().SlavePrefix
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterAbbr := defaults.Defaults().MasterAbbr
	masterLabel := defaults.Defaults().MasterName
	masterList := makeNodesList(nodes)
	slaveList := masterList

	nodeLabel := defaults.Defaults().NodePrefix
	stopNodeList := ""
	for i := nodes; i > 0; i-- {
		stopNodeList += fmt.Sprintf(" %d", i)
	}
	if sandboxDef.ClientBasedir == "" {
		sandboxDef.ClientBasedir = sandboxDef.Basedir
	}
	var data = common.StringMap{
		"Basedir":       sandboxDef.Basedir,
		"EngineClause":  "engine=ndbcluster",
		"ClientBasedir": sandboxDef.ClientBasedir,
		"Copyright":     globals.Copyright,
		"ClusterName":   fmt.Sprintf("%s_%s", defaults.Defaults().NdbPrefix, sandboxDef.Version),
		"AppVersion":    common.VersionDef,
		"DateTime":      timestamp.Format(time.UnixDate),
		"SandboxDir":    sandboxDef.SandboxDir,
		"MasterIp":      masterIp,
		"MasterList":    masterList,
		"NodeLabel":     nodeLabel,
		"NumNodes":      nodes,
		"NumNdbNodes":   ndbNodes,
		"LastNode":      nodes + ndbNodes + 1,
		"SlaveList":     slaveList,
		"RplUser":       sandboxDef.RplUser,
		"RplPassword":   sandboxDef.RplPassword,
		"SlaveLabel":    slaveLabel,
		"SlaveAbbr":     slaveAbbr,
		"MasterLabel":   masterLabel,
		"MasterAbbr":    masterAbbr,
		"StopNodeList":  stopNodeList,
		"ClusterPort":   ndbClusterPort,
		"Nodes":         []common.StringMap{},
		"NdbNodes":      []common.StringMap{},
		"SqlNodes":      []common.StringMap{},
	}
	connectionString := fmt.Sprintf("ndb_connectstring=%s:%d", masterIp, ndbClusterPort)
	logger.Printf("Creating connection string %s\n", connectionString)

	sbType := "ndb"
	logger.Printf("Defining replication type %s\n", sbType)

	sbDesc := common.SandboxDescription{
		Basedir: sandboxDef.Basedir,
		SBType:  sbType,
		Version: sandboxDef.Version,
		Flavor:  sandboxDef.Flavor,
		Port:    []int{ndbClusterPort},
		Nodes:   nodes,
		NodeNum: 0,
		LogFile: sandboxDef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sandboxDef.Version,
		Flavor:      sandboxDef.Flavor,
		Port:        []int{ndbClusterPort},
		Nodes:       []string{},
		Destination: sandboxDef.SandboxDir,
	}

	if sandboxDef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sandboxDef.LogFileName)
	}

	skipStart := sandboxDef.SkipStart
	sandboxDef.SkipStart = true
	for i := 1; i <= nodes; i++ {
		sandboxDef.Port = basePort + i
		nodeStringMap :=
			common.StringMap{
				"Copyright":    globals.Copyright,
				"AppVersion":   common.VersionDef,
				"DateTime":     timestamp.Format(time.UnixDate),
				"ClusterPort":  ndbClusterPort,
				"NumNodes":     nodes,
				"Node":         i,
				"NodePort":     sandboxDef.Port,
				"MasterIp":     masterIp,
				"NodeLabel":    nodeLabel,
				"SlaveLabel":   slaveLabel,
				"SlaveAbbr":    slaveAbbr,
				"MasterLabel":  masterLabel,
				"MasterAbbr":   masterAbbr,
				"SandboxDir":   sandboxDef.SandboxDir,
				"StopNodeList": stopNodeList,
				"RplUser":      sandboxDef.RplUser,
				"RplPassword":  sandboxDef.RplPassword}

		data["Nodes"] = append(data["Nodes"].([]common.StringMap), nodeStringMap)
		data["SqlNodes"] = append(data["SqlNodes"].([]common.StringMap), common.StringMap{"Node": i + ndbNodes})
		sandboxDef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.MorePorts = []int{}
		sandboxDef.ServerId = (baseServerId + i) * 100
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)
		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)

		if !sandboxDef.RunConcurrently {
			installationMessage := "Installing and starting %s %d\n"
			if skipStart {
				installationMessage = "Installing %s %d\n"
			}
			common.CondPrintf(installationMessage, nodeLabel, i)
			logger.Printf(installationMessage, nodeLabel, i)
		}
		sandboxDef.ReplOptions = SingleTemplates["replication_options"].Contents +
			fmt.Sprintf("\n%s\n%s\n", "ndbcluster", connectionString)
		reMasterIp := regexp.MustCompile(`127\.0\.0\.1`)
		sandboxDef.ReplOptions = reMasterIp.ReplaceAllString(sandboxDef.ReplOptions, masterIp)
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
		sandboxDef.SBType = "ndb-node"
		sandboxDef.NodeNum = i
		logger.Printf("Create single sandbox for node %d\n", i)
		execList, err := CreateChildSandbox(sandboxDef)
		if err != nil {
			return fmt.Errorf(globals.ErrCreatingSandbox, err)
		}
		for _, list := range execList {
			execLists = append(execLists, list)
		}
		var dataNode = common.StringMap{
			"Copyright":   globals.Copyright,
			"AppVersion":  common.VersionDef,
			"ClusterPort": ndbClusterPort,
			"DateTime":    timestamp.Format(time.UnixDate),
			"Node":        i,
			"NumNodes":    nodes,
			"NodePort":    sandboxDef.Port,
			"NodeLabel":   nodeLabel,
			"MasterLabel": masterLabel,
			"MasterAbbr":  masterAbbr,
			"SlaveLabel":  slaveLabel,
			"SlaveAbbr":   slaveAbbr,
			"SandboxDir":  sandboxDef.SandboxDir,
		}
		logger.Printf("Create node script for node %d\n", i)
		err = writeScript(logger, MultipleTemplates, fmt.Sprintf("n%d", i),
			"node_template", sandboxDef.SandboxDir, dataNode, true)
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
	for i := 2; i <= ndbNodes; i++ {
		data["NdbNodes"] = append(data["NdbNodes"].([]common.StringMap),
			common.StringMap{
				"Node":       i,
				"NodeLabel":  data["NodeLabel"],
				"SandboxDir": data["SandboxDir"],
			})
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

	logger.Printf("Writing group replication scripts\n")
	logger.Printf("##DATA: %s\n", stringMapToJson(data))
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
			// {globals.ScriptStopAll, "stop_multi_template", true},
			{globals.ScriptClearAll, "clear_multi_template", true},
			{globals.ScriptSendKillAll, "send_kill_multi_template", true},
			{globals.ScriptUseAll, "use_multi_template", true},
		},
	}
	sbRepl := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptUseAllSlaves, "multi_source_use_slaves_template", true},
			{globals.ScriptUseAllMasters, "multi_source_use_masters_template", true},
			{globals.ScriptTestReplication, "multi_source_test_template", true},
		},
	}
	sbNdb := ScriptBatch{
		tc:         NdbTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{"config.ini", "ndb_config_template", false},
			{globals.ScriptInitializeNodes, "ndb_start_cluster", true},
			{globals.ScriptCheckNodes, "ndb_check_status", true},
			{"ndb_mgm", "ndb_mgm_template", true},
			{globals.ScriptStopAll, "ndb_stop_cluster", true},
		},
	}

	for _, sb := range []ScriptBatch{sbMultiple, sbRepl, sbNdb} {
		err := writeScripts(sb)
		if err != nil {
			fmt.Printf("%s\n", err)
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
	if !skipStart {
		common.CondPrintln(path.Join(common.ReplaceLiteralHome(sandboxDef.SandboxDir), globals.ScriptInitializeNodes))
		logger.Printf("Running NDB replication initialization script\n")
		_, err := common.RunCmd(path.Join(sandboxDef.SandboxDir, globals.ScriptInitializeNodes))
		if err != nil {
			return fmt.Errorf("error initializing NDB replication: %s", err)
		}
	}
	common.CondPrintf("NDB cluster directory installed in %s\n", common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	common.CondPrintf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return nil
}
