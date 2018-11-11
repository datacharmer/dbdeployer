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
)

type Node struct {
	Node     int
	Port     int
	ServerId int
	Name     string
}

func CreateMultipleSandbox(sandboxDef SandboxDef, origin string, nodes int) common.StringMap {

	var execLists []concurrent.ExecutionList

	sbType := sandboxDef.SBType
	if sbType == "" {
		sbType = "multiple"
	}
	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		sandboxDef.LogFileName, logger = defaults.NewLogger(common.LogDirName(), sbType)
		sandboxDef.LogFileName = common.ReplaceLiteralHome(sandboxDef.LogFileName)
	}
	Basedir := sandboxDef.Basedir
	if !common.DirExists(Basedir) {
		common.Exitf(1, defaults.ErrBaseDirectoryNotFound, Basedir)
	}
	if sandboxDef.DirName == "" {
		sandboxDef.SandboxDir = path.Join(sandboxDef.SandboxDir, defaults.Defaults().MultiplePrefix+common.VersionToName(origin))
	} else {
		sandboxDef.SandboxDir = path.Join(sandboxDef.SandboxDir, sandboxDef.DirName)
	}
	if common.DirExists(sandboxDef.SandboxDir) {
		sandboxDef = CheckDirectory(sandboxDef)
	}

	vList := common.VersionToList(sandboxDef.Version)
	rev := vList[2]
	basePort := sandboxDef.Port + defaults.Defaults().MultipleBasePort + (rev * 100)
	if sandboxDef.BasePort > 0 {
		basePort = sandboxDef.BasePort
	}
	// FindFreePort returns the first free port, but base_port will be used
	// with a counter. Thus the availability will be checked using
	// "base_port + 1"
	firstPort := common.FindFreePort(basePort+1, sandboxDef.InstalledPorts, nodes)
	basePort = firstPort - 1
	for checkPort := basePort + 1; checkPort < basePort+nodes; checkPort++ {
		CheckPort("CreateMultipleSandbox", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
	}
	baseMysqlxPort := getBaseMysqlxPort(basePort, sandboxDef, nodes)
	common.Mkdir(sandboxDef.SandboxDir)
	logger.Printf("Created directory %s\n", sandboxDef.SandboxDir)
	logger.Printf("Multiple Sandbox Definition: %s\n", SandboxDefToJson(sandboxDef))

	common.AddToCleanupStack(common.Rmdir, "Rmdir", sandboxDef.SandboxDir)

	sandboxDef.ReplOptions = SingleTemplates["replication_options"].Contents
	baseServerId := 0
	if nodes < 2 {
		common.Exit(1, "only one node requested. For single sandbox deployment, use the 'single' command")
	}
	timestamp := time.Now()
	var data = common.StringMap{
		"Copyright":  Copyright,
		"AppVersion": common.VersionDef,
		"DateTime":   timestamp.Format(time.UnixDate),
		"SandboxDir": sandboxDef.SandboxDir,
		"Nodes":      []common.StringMap{},
	}

	sbDesc := common.SandboxDescription{
		Basedir: Basedir,
		SBType:  sandboxDef.SBType,
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

	logger.Printf("Defining multiple sandbox data: %v\n", StringMapToJson(data))
	nodeLabel := defaults.Defaults().NodePrefix
	for i := 1; i <= nodes; i++ {
		sandboxDef.Port = basePort + i
		data["Nodes"] = append(data["Nodes"].([]common.StringMap), common.StringMap{
			"Copyright":  Copyright,
			"AppVersion": common.VersionDef,
			"DateTime":   timestamp.Format(time.UnixDate),
			"Node":       i,
			"NodePort":   sandboxDef.Port,
			"NodeLabel":  nodeLabel,
			"SandboxDir": sandboxDef.SandboxDir,
		})
		sandboxDef.LoadGrants = true
		sandboxDef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.ServerId = (baseServerId + i) * 100
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)
		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)
		if common.GreaterOrEqualVersion(sandboxDef.Version, defaults.MinimumMysqlxDefaultVersion) {
			sandboxDef.MysqlXPort = baseMysqlxPort + i
			if !sandboxDef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i)
				logger.Printf("Adding mysqlx port %d to node %d\n", baseMysqlxPort+i, i)
			}
		}
		sandboxDef.Multi = true
		sandboxDef.NodeNum = i
		sandboxDef.Prompt = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.SBType = sbType + "-node"
		if !sandboxDef.RunConcurrently {
			fmt.Printf("Installing and starting %s %d\n", nodeLabel, i)
			logger.Printf("installing and starting %s %d", nodeLabel, i)
		}
		logger.Printf("Creating single sandbox for node %d\n", i)
		execList := CreateSingleSandbox(sandboxDef)
		for _, list := range execList {
			execLists = append(execLists, list)
		}

		var dataNode = common.StringMap{
			"Node":       i,
			"NodePort":   sandboxDef.Port,
			"NodeLabel":  nodeLabel,
			"SandboxDir": sandboxDef.SandboxDir,
			"Copyright":  Copyright,
		}
		logger.Printf("Creating node script for node %d\n", i)
		logger.Printf("Defining multiple sandbox node inner data: %v\n", StringMapToJson(dataNode))
		writeScript(logger, MultipleTemplates, fmt.Sprintf("n%d", i), "node_template", sandboxDef.SandboxDir, dataNode, true)
	}
	logger.Printf("Write sandbox description\n")
	common.WriteSandboxDescription(sandboxDef.SandboxDir, sbDesc)
	defaults.UpdateCatalog(sandboxDef.SandboxDir, sbItem)

	logger.Printf("Write multiple sandbox scripts\n")
	writeScript(logger, MultipleTemplates, defaults.ScriptStartAll, "start_multi_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, defaults.ScriptRestartAll, "restart_multi_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, defaults.ScriptStatusAll, "status_multi_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, defaults.ScriptTestSbAll, "test_sb_multi_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, defaults.ScriptStopAll, "stop_multi_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, defaults.ScriptClearAll, "clear_multi_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, defaults.ScriptSendKillAll, "send_kill_multi_template", sandboxDef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, defaults.ScriptUseAll, "use_multi_template", sandboxDef.SandboxDir, data, true)

	logger.Printf("Run concurrent tasks\n")
	concurrent.RunParallelTasksByPriority(execLists)

	fmt.Printf("%s directory installed in %s\n", sbType, common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return data
}
