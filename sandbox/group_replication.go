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

func getBaseMysqlxPort(basePort int, sdef SandboxDef, nodes int) (int, error) {
	baseMysqlxPort := basePort + defaults.Defaults().MysqlXPortDelta
	// 8.0.11
	isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sdef.Version, globals.MinimumMysqlxDefaultVersion)
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
			err := checkPortAvailability("get_base_mysqlx_port", sdef.SandboxDir, sdef.InstalledPorts, checkPort)
			if err != nil {
				return 0, err
			}
		}
	}
	return baseMysqlxPort, nil
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

	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
	rev := vList[2]
	basePort := sandboxDef.Port + defaults.Defaults().GroupReplicationBasePort + (rev * 100)
	if sandboxDef.SinglePrimary {
		basePort = sandboxDef.Port + defaults.Defaults().GroupReplicationSpBasePort + (rev * 100)
	}
	if sandboxDef.BasePort > 0 {
		basePort = sandboxDef.BasePort
	}

	baseServerId := 0
	if nodes < 3 {
		return fmt.Errorf("Can't run group replication with less than 3 nodes")
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
	if sandboxDef.SinglePrimary {
		sbType = "group-single-primary"
		singleMultiPrimary = GroupReplSinglePrimary
	}
	logger.Printf("Defining group type %s\n", sbType)

	sbDesc := common.SandboxDescription{
		Basedir: sandboxDef.Basedir,
		SBType:  sbType,
		Version: sandboxDef.Version,
		Port:    []int{},
		Nodes:   nodes,
		NodeNum: 0,
		LogFile: sandboxDef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sandboxDef.Version,
		Port:        []int{},
		Nodes:       []string{},
		Destination: sandboxDef.SandboxDir,
	}

	if sandboxDef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sandboxDef.LogFileName)
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
			"SandboxDir":        sandboxDef.SandboxDir,
			"RplUser":           sandboxDef.RplUser,
			"RplPassword":       sandboxDef.RplPassword})

		sandboxDef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.Port = basePort + i
		sandboxDef.MorePorts = []int{groupPort}
		sandboxDef.ServerId = (baseServerId + i) * 100
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)
		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)
		sbItem.Port = append(sbItem.Port, sandboxDef.Port+defaults.Defaults().GroupPortDelta)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port+defaults.Defaults().GroupPortDelta)

		if !sandboxDef.RunConcurrently {
			installationMessage := "Installing and starting %s %d\n"
			if sandboxDef.SkipStart {
				installationMessage = "Installing %s %d\n"
			}
			common.CondPrintf(installationMessage, nodeLabel, i)
			logger.Printf(installationMessage, nodeLabel, i)
		}
		sandboxDef.ReplOptions = SingleTemplates["replication_options"].Contents + fmt.Sprintf("\n%s\n%s\n", GroupReplOptions, singleMultiPrimary)
		reMasterIp := regexp.MustCompile(`127\.0\.0\.1`)
		sandboxDef.ReplOptions = reMasterIp.ReplaceAllString(sandboxDef.ReplOptions, masterIp)
		sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates["gtid_options_57"].Contents)
		sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates["repl_crash_safe_options"].Contents)
		sandboxDef.ReplOptions += fmt.Sprintf("\nloose-group-replication-local-address=%s:%d\n", masterIp, groupPort)
		sandboxDef.ReplOptions += fmt.Sprintf("\nloose-group-replication-group-seeds=%s\n", connectionString)
		// 8.0.11
		isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
		if err != nil {
			return err
		}
		if isMinimumMySQLXDefault {
			sandboxDef.MysqlXPort = baseMysqlxPort + i
			if !sandboxDef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i)
				logger.Printf("adding port %d to node %d\n", baseMysqlxPort+i, i)
			}
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
			"SandboxDir":        sandboxDef.SandboxDir,
		}
		logger.Printf("Create node script for node %d\n", i)
		err = writeScript(logger, MultipleTemplates, fmt.Sprintf("n%d", i), "node_template", sandboxDef.SandboxDir, dataNode, true)
		if err != nil {
			return err
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

	logger.Printf("Writing group replication scripts\n")
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
	sbGroup := ScriptBatch{
		tc:         GroupTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptInitializeNodes, "init_nodes_template", true},
			{globals.ScriptCheckNodes, "check_nodes_template", true},
		},
	}

	for _, sb := range []ScriptBatch{sbMultiple, sbRepl, sbGroup} {
		err := writeScripts(sb)
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
	common.CondPrintf("Replication directory installed in %s\n", common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	common.CondPrintf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return nil
}
