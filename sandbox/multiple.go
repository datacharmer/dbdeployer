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

type Node struct {
	Node     int
	Port     int
	ServerId int
	Name     string
}

func CreateMultipleSandbox(sdef SandboxDef, origin string, nodes int) common.StringMap {

	var execLists []concurrent.ExecutionList

	sbType := sdef.SBType
	if sbType == "" {
		sbType = "multiple"
	}
	var logger *defaults.Logger
	if sdef.Logger != nil {
		logger = sdef.Logger
	} else {
		sdef.LogFileName, logger = defaults.NewLogger(common.LogDirName(), sbType)
		sdef.LogFileName = common.ReplaceLiteralHome(sdef.LogFileName)
	}
	Basedir := sdef.Basedir
	if !common.DirExists(Basedir) {
		common.Exitf(1, "Base directory %s does not exist", Basedir)
	}
	if sdef.DirName == "" {
		sdef.SandboxDir += "/" + defaults.Defaults().MultiplePrefix + common.VersionToName(origin)
	} else {
		sdef.SandboxDir += "/" + sdef.DirName
	}
	if common.DirExists(sdef.SandboxDir) {
		sdef = CheckDirectory(sdef)
	}

	vList := common.VersionToList(sdef.Version)
	rev := vList[2]
	basePort := sdef.Port + defaults.Defaults().MultipleBasePort + (rev * 100)
	if sdef.BasePort > 0 {
		basePort = sdef.BasePort
	}
	// FindFreePort returns the first free port, but base_port will be used
	// with a counter. Thus the availability will be checked using
	// "base_port + 1"
	firstPort := common.FindFreePort(basePort+1, sdef.InstalledPorts, nodes)
	basePort = firstPort - 1
	for checkPort := basePort + 1; checkPort < basePort+nodes; checkPort++ {
		CheckPort("CreateMultipleSandbox", sdef.SandboxDir, sdef.InstalledPorts, checkPort)
	}
	baseMysqlxPort := getBaseMysqlxPort(basePort, sdef, nodes)
	common.Mkdir(sdef.SandboxDir)
	logger.Printf("Created directory %s\n", sdef.SandboxDir)
	logger.Printf("Multiple Sandbox Definition: %s\n", SandboxDefToJson(sdef))

	common.AddToCleanupStack(common.Rmdir, "Rmdir", sdef.SandboxDir)

	sdef.ReplOptions = SingleTemplates["replication_options"].Contents
	baseServerId := 0
	if nodes < 2 {
		common.Exit(1, "Only one node requested. For single sandbox deployment, use the 'single' command")
	}
	timestamp := time.Now()
	var data = common.StringMap{
		"Copyright":  Copyright,
		"AppVersion": common.VersionDef,
		"DateTime":   timestamp.Format(time.UnixDate),
		"SandboxDir": sdef.SandboxDir,
		"Nodes":      []common.StringMap{},
	}

	sbDesc := common.SandboxDescription{
		Basedir: Basedir,
		SBType:  sdef.SBType,
		Version: sdef.Version,
		Port:    []int{},
		Nodes:   nodes,
		NodeNum: 0,
		LogFile: sdef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sdef.Version,
		Port:        []int{},
		Nodes:       []string{},
		Destination: sdef.SandboxDir,
	}

	if sdef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sdef.LogFileName)
	}

	logger.Printf("Defining multiple sandbox data: %v\n", SmapToJson(data))
	nodeLabel := defaults.Defaults().NodePrefix
	for i := 1; i <= nodes; i++ {
		sdef.Port = basePort + i
		data["Nodes"] = append(data["Nodes"].([]common.StringMap), common.StringMap{
			"Copyright":  Copyright,
			"AppVersion": common.VersionDef,
			"DateTime":   timestamp.Format(time.UnixDate),
			"Node":       i,
			"NodePort":   sdef.Port,
			"NodeLabel":  nodeLabel,
			"SandboxDir": sdef.SandboxDir,
		})
		sdef.LoadGrants = true
		sdef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sdef.ServerId = (baseServerId + i) * 100
		sbItem.Nodes = append(sbItem.Nodes, sdef.DirName)
		sbItem.Port = append(sbItem.Port, sdef.Port)
		sbDesc.Port = append(sbDesc.Port, sdef.Port)
		if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
			sdef.MysqlXPort = baseMysqlxPort + i
			if !sdef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i)
				logger.Printf("Adding mysqlx port %d to node %d\n", baseMysqlxPort+i, i)
			}
		}
		sdef.Multi = true
		sdef.NodeNum = i
		sdef.Prompt = fmt.Sprintf("%s%d", nodeLabel, i)
		sdef.SBType = sbType + "-node"
		if !sdef.RunConcurrently {
			fmt.Printf("Installing and starting %s %d\n", nodeLabel, i)
			logger.Printf("installing and starting %s %d", nodeLabel, i)
		}
		logger.Printf("Creating single sandbox for node %d\n", i)
		execList := CreateSingleSandbox(sdef)
		for _, list := range execList {
			execLists = append(execLists, list)
		}

		var dataNode = common.StringMap{
			"Node":       i,
			"NodePort":   sdef.Port,
			"NodeLabel":  nodeLabel,
			"SandboxDir": sdef.SandboxDir,
			"Copyright":  Copyright,
		}
		logger.Printf("Creating node script for node %d\n", i)
		logger.Printf("Defining multiple sandbox node inner data: %v\n", SmapToJson(dataNode))
		writeScript(logger, MultipleTemplates, fmt.Sprintf("n%d", i), "node_template", sdef.SandboxDir, dataNode, true)
	}
	logger.Printf("Write sandbox description\n")
	common.WriteSandboxDescription(sdef.SandboxDir, sbDesc)
	defaults.UpdateCatalog(sdef.SandboxDir, sbItem)

	logger.Printf("Write multiple sandbox scripts\n")
	writeScript(logger, MultipleTemplates, "start_all", "start_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "restart_all", "restart_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "status_all", "status_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "test_sb_all", "test_sb_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "stop_all", "stop_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "clear_all", "clear_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "send_kill_all", "send_kill_multi_template", sdef.SandboxDir, data, true)
	writeScript(logger, MultipleTemplates, "use_all", "use_multi_template", sdef.SandboxDir, data, true)

	logger.Printf("Run concurrent tasks\n")
	concurrent.RunParallelTasksByPriority(execLists)

	fmt.Printf("%s directory installed in %s\n", sbType, common.ReplaceLiteralHome(sdef.SandboxDir))
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return data
}
