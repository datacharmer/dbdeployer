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

type Slave struct {
	Node       int
	Port       int
	ServerId   int
	Name       string
	MasterPort int
}

func CreateMasterSlaveReplication(sdef SandboxDef, origin string, nodes int, master_ip string) {

	sdef.ReplOptions = SingleTemplates["replication_options"].Contents
	vList := common.VersionToList(sdef.Version)
	rev := vList[2]
	// base_port := sdef.Port + defaults.Defaults().MasterSlaveBasePort + (rev * 100)
	base_port := sdef.Port + defaults.Defaults().MasterSlaveBasePort + rev 
	if sdef.BasePort > 0 {
		base_port = sdef.BasePort
	}
	base_server_id := 0
	sdef.DirName = defaults.Defaults().MasterName
	base_port = FindFreePort(base_port, sdef.InstalledPorts,  nodes)
	for check_port := base_port + 1; check_port < base_port+nodes+1; check_port++ {
		CheckPort(sdef.SandboxDir, sdef.InstalledPorts, check_port)
	}

	common.Mkdir(sdef.SandboxDir)
	sdef.Port = base_port + 1
	sdef.ServerId = (base_server_id + 1) * 100
	sdef.LoadGrants = false
	master_port := sdef.Port
	change_master_extra := ""
	if common.GreaterOrEqualVersion(sdef.Version, []int{8,0,4}) {
		if !sdef.NativeAuthPlugin {
			change_master_extra = ", GET_MASTER_PUBLIC_KEY=1"
		}
	}
	if nodes < 2 {
		fmt.Println("Can't run replication with less than 2 nodes")
		os.Exit(1)
	}
	slaves := nodes - 1
	master_abbr := defaults.Defaults().MasterAbbr
	master_label := defaults.Defaults().MasterName
	slave_label := defaults.Defaults().SlavePrefix
	slave_abbr := defaults.Defaults().SlaveAbbr
	timestamp := time.Now()
	var data common.Smap = common.Smap{
		"Copyright":  Copyright,
		"AppVersion": common.VersionDef,
		"DateTime":   timestamp.Format(time.UnixDate),
		"SandboxDir": sdef.SandboxDir,
		"MasterLabel": master_label,
		"MasterPort": sdef.Port,
		"SlaveLabel": slave_label,
		"MasterAbbr": master_abbr,
		"MasterIp":    master_ip,
		"RplUser":     sdef.RplUser,
		"RplPassword": sdef.RplPassword,
		"SlaveAbbr": slave_abbr,
		"ChangeMasterExtra" : change_master_extra,
		"Slaves":     []common.Smap{},
	}

	fmt.Printf("Installing and starting %s\n", master_label)
	sdef.LoadGrants = true
	sdef.Multi = true
	sdef.Prompt = master_label
	sdef.NodeNum = 1
	sdef.SBType = "replication-node"
	CreateSingleSandbox(sdef, origin)

	sb_desc := common.SandboxDescription{
		Basedir: sdef.Basedir + "/" + sdef.Version,
		SBType:  "master-slave",
		Version: sdef.Version,
		Port:    []int{sdef.Port},
		Nodes:   slaves,
		NodeNum: 0,
	}

	sb_item := defaults.SandboxItem{
		Origin : sb_desc.Basedir,
		SBType : sb_desc.SBType,
		Version: sdef.Version,
		Port:    []int{sdef.Port},
		Nodes:   []string{defaults.Defaults().MasterName},
		Destination: sdef.SandboxDir,
	}

	node_label := defaults.Defaults().NodePrefix
	for i := 1; i <= slaves; i++ {
		sdef.Port = base_port + i + 1
		data["Slaves"] = append(data["Slaves"].([]common.Smap), common.Smap{
			"Copyright":   Copyright,
			"AppVersion":  common.VersionDef,
			"DateTime":    timestamp.Format(time.UnixDate),
			"Node":        i,
			"NodeLabel":   node_label,
			"NodePort":   sdef.Port,
			"SlaveLabel":  slave_label,
			"MasterAbbr": master_abbr,
			"SlaveAbbr": slave_abbr,
			"SandboxDir":  sdef.SandboxDir,
			"MasterPort":  master_port,
			"MasterIp":    master_ip,
			"ChangeMasterExtra" : change_master_extra,
			"RplUser":     sdef.RplUser,
			"RplPassword": sdef.RplPassword})
		sdef.LoadGrants = false
		sdef.Prompt = fmt.Sprintf("%s%d",slave_label, i)
		sdef.DirName = fmt.Sprintf("%s%d", node_label, i)
		sdef.ServerId = (base_server_id + i + 1) * 100
		sdef.NodeNum = i + 1
		sb_item.Nodes = append(sb_item.Nodes, sdef.DirName)
		sb_item.Port = append(sb_item.Port, sdef.Port)
		sb_desc.Port = append(sb_desc.Port, sdef.Port)
		fmt.Printf("Installing and starting %s %d\n", slave_label, i)
		CreateSingleSandbox(sdef, origin)
		var data_slave common.Smap = common.Smap{
			"Copyright":  Copyright,
			"AppVersion": common.VersionDef,
			"DateTime":   timestamp.Format(time.UnixDate),
			"Node":       i,
			"NodeLabel":  node_label,
			"NodePort":  sdef.Port,
			"SlaveLabel":  slave_label,
			"MasterAbbr": master_abbr,
			"ChangeMasterExtra" : change_master_extra,
			"SlaveAbbr": slave_abbr,
			"SandboxDir": sdef.SandboxDir,
		}
		write_script(ReplicationTemplates, fmt.Sprintf("%s%d", slave_abbr, i), "slave_template", sdef.SandboxDir, data_slave, true)
		write_script(ReplicationTemplates, fmt.Sprintf("n%d", i+1), "slave_template", sdef.SandboxDir, data_slave, true)
	}
	common.WriteSandboxDescription(sdef.SandboxDir, sb_desc)
	defaults.UpdateCatalog(sdef.SandboxDir, sb_item)
	
	initialize_slaves := "initialize_" + slave_label + "s"
	check_slaves := "check_" + slave_label + "s"

	write_script(ReplicationTemplates, "start_all", "start_all_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "restart_all", "restart_all_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "status_all", "status_all_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "test_sb_all", "test_sb_all_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "stop_all", "stop_all_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "clear_all", "clear_all_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "send_kill_all", "send_kill_all_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "use_all", "use_all_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, initialize_slaves, "init_slaves_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, check_slaves, "check_slaves_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, master_abbr, "master_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "n1", "master_template", sdef.SandboxDir, data, true)
	write_script(ReplicationTemplates, "test_replication", "test_replication_template", sdef.SandboxDir, data, true)
	fmt.Println(sdef.SandboxDir + "/" + initialize_slaves)
	common.Run_cmd(sdef.SandboxDir + "/" +initialize_slaves)
	fmt.Printf("Replication directory installed in %s\n", sdef.SandboxDir)
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}

func CreateReplicationSandbox(sdef SandboxDef, origin string, topology string, nodes int, master_ip string) {

	Basedir := sdef.Basedir + "/" + sdef.Version
	if !common.DirExists(Basedir) {
		fmt.Printf("Base directory %s does not exist\n", Basedir)
		os.Exit(1)
	}

	sandbox_dir := sdef.SandboxDir
	switch topology {
	case "master-slave":
		sdef.SandboxDir += "/" + defaults.Defaults().MasterSlavePrefix + common.VersionToName(origin)
	case "group":
		if sdef.SinglePrimary {
			sdef.SandboxDir += "/" + defaults.Defaults().GroupSpPrefix + common.VersionToName(origin)
		} else {
			sdef.SandboxDir += "/" + defaults.Defaults().GroupPrefix + common.VersionToName(origin)
		}
		if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 17}) {
			fmt.Println("Group replication requires MySQL 5.7.17 or greater")
			os.Exit(1)
		}
	case "fan-in":
		if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 9}) {
			fmt.Println("multi-source replication requires MySQL 5.7.9 or greater")
			os.Exit(1)
		}
		sdef.SandboxDir += "/" + defaults.Defaults().FanInPrefix + common.VersionToName(origin)
	case "all-masters":
		if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 9}) {
			fmt.Println("multi-source replication requires MySQL 5.7.9 or greater")
			os.Exit(1)
		}
		sdef.SandboxDir += "/" + defaults.Defaults().AllMastersPrefix + common.VersionToName(origin)
	default:
		fmt.Println("Unrecognized topology. Accepted: 'master-slave', 'group'")
		os.Exit(1)
	}
	if sdef.DirName != "" {
		sdef.SandboxDir = sandbox_dir + "/" + sdef.DirName
	}

	if common.DirExists(sdef.SandboxDir) {
		sdef = CheckDirectory(sdef)
	}

	switch topology {
	case "master-slave":
		CreateMasterSlaveReplication(sdef, origin, nodes, master_ip)
	case "group":
		CreateGroupReplication(sdef, origin, nodes, master_ip)
	case "fan-in":
		// CreateFanInReplication(sdef, origin, nodes)
		fmt.Println("fan-in replication is not implemented yet")
		os.Exit(0)
	case "all-masters":
		// CreateAllMastersReplication(sdef, origin, nodes)
		fmt.Println("all-masters replication is not implemented yet")
		os.Exit(0)
	}
}
