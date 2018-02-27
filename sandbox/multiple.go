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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"os"
	"time"
)

type Node struct {
	Node     int
	Port     int
	ServerId int
	Name     string
}

func CreateMultipleSandbox(sdef SandboxDef, origin string, nodes int) {

	Basedir := sdef.Basedir + "/" + sdef.Version
	if !common.DirExists(Basedir) {
		fmt.Printf("Base directory %s does not exist\n", Basedir)
		os.Exit(1)
	}
	if sdef.DirName == "" {
		sdef.SandboxDir += "/" + defaults.Defaults().MultiplePrefix + common.VersionToName(origin)
	} else {
		sdef.SandboxDir += "/" + sdef.DirName
	}
	if common.DirExists(sdef.SandboxDir) {
		sdef = CheckDirectory(sdef)
	}
	common.Mkdir(sdef.SandboxDir)

	sdef.ReplOptions = SingleTemplates["replication_options"].Contents
	vList := common.VersionToList(sdef.Version)
	rev := vList[2]
	base_port := sdef.Port + defaults.Defaults().MultipleBasePort + (rev * 100)
	if sdef.BasePort > 0 {
		base_port = sdef.BasePort
	}
	base_server_id := 0
	if nodes < 2 {
		fmt.Println("For single sandbox deployment, use the 'single' command")
		os.Exit(1)
	}
	timestamp := time.Now()
	var data common.Smap = common.Smap{
		"Copyright":  Copyright,
		"AppVersion": common.VersionDef,
		"DateTime":   timestamp.Format(time.UnixDate),
		"SandboxDir": sdef.SandboxDir,
		"Nodes":      []common.Smap{},
	}

	sb_desc := common.SandboxDescription{
		Basedir: Basedir,
		SBType:  "multiple",
		Version: sdef.Version,
		Port:    []int{},
		Nodes:   nodes,
		NodeNum: 0,
	}

	sb_item := defaults.SandboxItem{
		Origin : sb_desc.Basedir,
		SBType : sb_desc.SBType,
		Version: sdef.Version,
		Port:    []int{},
		Nodes:   []string{},
		Destination: sdef.SandboxDir,
	}

	for i := 1; i <= nodes; i++ {
		data["Nodes"] = append(data["Nodes"].([]common.Smap), common.Smap{
			"Copyright":  Copyright,
			"AppVersion": common.VersionDef,
			"DateTime":   timestamp.Format(time.UnixDate),
			"Node":       i,
			"SandboxDir": sdef.SandboxDir,
		})
		sdef.LoadGrants = true
		sdef.DirName = fmt.Sprintf("node%d", i)
		sdef.Port = base_port + i + 1
		sdef.ServerId = (base_server_id + i) * 100
		sb_item.Nodes = append(sb_item.Nodes, sdef.DirName)
		sb_item.Port = append(sb_item.Port, sdef.Port)
		sb_desc.Port = append(sb_desc.Port, sdef.Port)

		fmt.Printf("Installing and starting node %d\n", i)
		sdef.Multi = true
		sdef.NodeNum = i
		sdef.Prompt = fmt.Sprintf("node%d", i)
		sdef.SBType = "multiple-node"
		CreateSingleSandbox(sdef, origin)
		var data_node common.Smap = common.Smap{
			"Node":       i,
			"SandboxDir": sdef.SandboxDir,
			"Copyright":  Copyright,
		}
		write_script(MultipleTemplates, fmt.Sprintf("n%d", i), "node_template", sdef.SandboxDir, data_node, true)
	}
	common.WriteSandboxDescription(sdef.SandboxDir, sb_desc)
	defaults.UpdateCatalog(sdef.SandboxDir, sb_item)

	write_script(MultipleTemplates, "start_all", "start_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "restart_all", "restart_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "status_all", "status_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "test_sb_all", "test_sb_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "stop_all", "stop_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "clear_all", "clear_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "send_kill_all", "send_kill_multi_template", sdef.SandboxDir, data, true)
	write_script(MultipleTemplates, "use_all", "use_multi_template", sdef.SandboxDir, data, true)
	fmt.Printf("Multiple directory installed in %s\n", sdef.SandboxDir)
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}
