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
	"os"
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

func CreateMultipleSandbox(sdef SandboxDef, origin string, nodes int) common.Smap {

	var exec_lists []concurrent.ExecutionList
	sb_type := sdef.SBType
	if sb_type == "" {
		sb_type = "multiple"
	}
	Basedir := sdef.Basedir
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

	vList := common.VersionToList(sdef.Version)
	rev := vList[2]
	// base_port := sdef.Port + defaults.Defaults().MultipleBasePort + (rev * 100)
	base_port := sdef.Port + defaults.Defaults().MultipleBasePort + rev
	if sdef.BasePort > 0 {
		base_port = sdef.BasePort
	}
	base_mysqlx_port := get_base_mysqlx_port(base_port, sdef, nodes)
	base_port = FindFreePort(base_port, sdef.InstalledPorts, nodes)
	for check_port := base_port + 1; check_port < base_port+nodes; check_port++ {
		CheckPort(sdef.SandboxDir, sdef.InstalledPorts, check_port)
	}
	common.Mkdir(sdef.SandboxDir)

	sdef.ReplOptions = SingleTemplates["replication_options"].Contents
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
		SBType:  sdef.SBType,
		Version: sdef.Version,
		Port:    []int{},
		Nodes:   nodes,
		NodeNum: 0,
	}

	sb_item := defaults.SandboxItem{
		Origin:      sb_desc.Basedir,
		SBType:      sb_desc.SBType,
		Version:     sdef.Version,
		Port:        []int{},
		Nodes:       []string{},
		Destination: sdef.SandboxDir,
	}

	node_label := defaults.Defaults().NodePrefix
	for i := 1; i <= nodes; i++ {
		sdef.Port = base_port + i
		data["Nodes"] = append(data["Nodes"].([]common.Smap), common.Smap{
			"Copyright":  Copyright,
			"AppVersion": common.VersionDef,
			"DateTime":   timestamp.Format(time.UnixDate),
			"Node":       i,
			"NodePort":   sdef.Port,
			"NodeLabel":  node_label,
			"SandboxDir": sdef.SandboxDir,
		})
		sdef.LoadGrants = true
		sdef.DirName = fmt.Sprintf("%s%d", node_label, i)
		sdef.ServerId = (base_server_id + i) * 100
		sb_item.Nodes = append(sb_item.Nodes, sdef.DirName)
		sb_item.Port = append(sb_item.Port, sdef.Port)
		sb_desc.Port = append(sb_desc.Port, sdef.Port)
		if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
			sdef.MysqlXPort = base_mysqlx_port + i
			if !sdef.DisableMysqlX {
				sb_desc.Port = append(sb_desc.Port, base_mysqlx_port+i)
				sb_item.Port = append(sb_item.Port, base_mysqlx_port+i)
			}
		}
		sdef.Multi = true
		sdef.NodeNum = i
		sdef.Prompt = fmt.Sprintf("%s%d", node_label, i)
		sdef.SBType += "node"
		if !sdef.RunConcurrently {
			fmt.Printf("Installing and starting %s %d\n", node_label, i)
		}
		exec_list := CreateSingleSandbox(sdef, origin)
		for _, list := range exec_list {
			exec_lists = append(exec_lists, list)
		}

		var data_node common.Smap = common.Smap{
			"Node":       i,
			"NodePort":   sdef.Port,
			"NodeLabel":  node_label,
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
	concurrent.RunParallelTasksByPriority(exec_lists)

	fmt.Printf("%s directory installed in %s\n", sb_type, common.ReplaceLiteralHome(sdef.SandboxDir))
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return data
}
