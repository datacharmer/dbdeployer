package sandbox

import (
	"github.com/datacharmer/dbdeployer/common"
	"fmt"
	"os"
)

type Node struct {
	Node       int
	Port       int
	ServerId   int
	Name       string
}

func CreateMultipleSandbox(sdef SandboxDef, origin string, nodes int) {

	Basedir := sdef.Basedir + "/" + sdef.Version
	if !common.DirExists(Basedir) {
		fmt.Printf("Base directory %s does not exist\n", Basedir)
		os.Exit(1)
	}

	sdef.SandboxDir += "/" + MultiplePrefix + VersionToName(origin)

	err := os.Mkdir(sdef.SandboxDir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sdef.ReplOptions = ReplOptions
	vList := VersionToList(sdef.Version)
	rev := vList[2]
	base_port := sdef.Port + MultipleBasePort + (rev * 100)
	base_server_id := 0
	if nodes < 2 {
		fmt.Println("For single sandbox deployment, use the 'single' command")
		os.Exit(1)
	}
	var data common.Smap = common.Smap{
		"Copyright":  Copyright,
		"SandboxDir": sdef.SandboxDir,
		"Nodes":     []common.Smap{},
	}

	for i := 1; i <= nodes; i++ {
		data["Nodes"] = append(data["Nodes"].([]common.Smap), common.Smap{
			"Node": i,
			"SandboxDir":  sdef.SandboxDir,
		 })
		sdef.LoadGrants = true
		sdef.DirName = fmt.Sprintf("node%d", i)
		sdef.Port = base_port + i + 1
		sdef.ServerId = (base_server_id + i) * 100
		fmt.Printf("Installing and starting node %d\n", i)
		sdef.Multi = true
		sdef.Prompt = fmt.Sprintf("node%d", i)
		CreateSingleSandbox(sdef, origin)
		var data_node  common.Smap = common.Smap{
			"Node" : i,
			"SandboxDir" : sdef.SandboxDir,
			"Copyright" : Copyright,
		}
		write_script(fmt.Sprintf("n%d",i), node_template, sdef.SandboxDir, data_node, true)
	}
	sb_desc := common.SandboxDescription{
		Basedir : Basedir,
		SBType	: "multiple",
		Version : sdef.Version,
		Port	: 0,
		Nodes 	: nodes,
	}
	common.WriteSandboxDescription(sdef.SandboxDir, sb_desc)

	write_script("start_all", start_multi_template, sdef.SandboxDir, data, true)
	write_script("restart_all", restart_multi_template, sdef.SandboxDir, data, true)
	write_script("status_all", status_multi_template, sdef.SandboxDir, data, true)
	write_script("stop_all", stop_multi_template, sdef.SandboxDir, data, true)
	write_script("send_kill_all", send_kill_multi_template, sdef.SandboxDir, data, true)
	write_script("use_all", use_multi_template, sdef.SandboxDir, data, true)
	fmt.Printf("Multiple directory installed in %s\n", sdef.SandboxDir)
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}

