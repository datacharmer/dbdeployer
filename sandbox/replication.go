package sandbox

import (
	"github.com/datacharmer/dbdeployer/common"
	"fmt"
	"os"
)


type Slave struct {
	Node       int
	Port       int
	ServerId   int
	Name       string
	MasterPort int
}



func CreateMasterSlaveReplication(sdef SandboxDef, origin string, nodes int) {

	sdef.ReplOptions = ReplOptions
	vList := VersionToList(sdef.Version)
	rev := vList[2]
	base_port := sdef.Port + MasterSlaveBasePort + (rev * 100)
	base_server_id := 0
	sdef.DirName = "master"
	sdef.Port = base_port + 1
	sdef.ServerId = (base_server_id + 1) * 100
	sdef.LoadGrants = false
	master_port := sdef.Port
	if nodes < 2 {
		fmt.Println("Can't run replication with less than 2 nodes")
		os.Exit(1)
	}
	slaves := nodes - 1
	var data common.Smap = common.Smap{
		"Copyright":  Copyright,
		"SandboxDir": sdef.SandboxDir,
		"Slaves":     []common.Smap{},
	}

	fmt.Println("Installing and starting master")
	sdef.LoadGrants = true
	sdef.Multi = true
	sdef.Prompt = "master"
	CreateSingleSandbox(sdef, origin)
	for i := 1; i <= slaves; i++ {
		data["Slaves"] = append(data["Slaves"].([]common.Smap), common.Smap{
			"Node": i,
			"SandboxDir":  sdef.SandboxDir,
			"MasterPort":  master_port,
			"RplUser":     sdef.RplUser,
			"RplPassword": sdef.RplPassword})
		sdef.LoadGrants = false
		sdef.Prompt = fmt.Sprintf("slave%d", i)
		sdef.DirName = fmt.Sprintf("node%d", i)
		sdef.Port = base_port + i + 1
		sdef.ServerId = (base_server_id + i + 1) * 100
		fmt.Printf("Installing and starting slave %d\n", i)
		CreateSingleSandbox(sdef, origin)
		var data_slave  common.Smap = common.Smap{
			"Node" : i,
			"SandboxDir" : sdef.SandboxDir,
			"Copyright" : Copyright,
		}
		write_script(fmt.Sprintf("s%d",i), slave_template, sdef.SandboxDir, data_slave, true)
		write_script(fmt.Sprintf("n%d",i+1), slave_template, sdef.SandboxDir, data_slave, true)
	}
	sb_desc := common.SandboxDescription{
		Basedir : sdef.Basedir + "/" + sdef.Version,
		SBType	: "master-slave",
		Version : sdef.Version,
		Port	: 0,
		Nodes 	: slaves,
	}
	common.WriteSandboxDescription(sdef.SandboxDir, sb_desc)

	write_script("start_all", start_all_template, sdef.SandboxDir, data, true)
	write_script("restart_all", restart_all_template, sdef.SandboxDir, data, true)
	write_script("status_all", status_all_template, sdef.SandboxDir, data, true)
	write_script("stop_all", stop_all_template, sdef.SandboxDir, data, true)
	write_script("send_kill_all", send_kill_all_template, sdef.SandboxDir, data, true)
	write_script("use_all", use_all_template, sdef.SandboxDir, data, true)
	write_script("initialize_slaves", init_slaves_template, sdef.SandboxDir, data, true)
	write_script("check_slaves", check_slaves_template, sdef.SandboxDir, data, true)
	write_script("m", master_template, sdef.SandboxDir, data, true)
	write_script("n1", master_template, sdef.SandboxDir, data, true)
	fmt.Println(sdef.SandboxDir + "/initialize_slaves")
	common.Run_cmd(sdef.SandboxDir + "/initialize_slaves")
	fmt.Printf("Replication directory installed in %s\n", sdef.SandboxDir)
	fmt.Printf("run 'dbdeployer usage multiple' for basic instructions'\n")
}

func CreateReplicationSandbox(sdef SandboxDef, origin string, topology string, nodes int) {

	Basedir := sdef.Basedir + "/" + sdef.Version
	if !common.DirExists(Basedir) {
		fmt.Printf("Base directory %s does not exist\n", Basedir)
		os.Exit(1)
	}

	switch topology {
	case "master-slave":
		sdef.SandboxDir += "/" + MasterSlavePrefix + VersionToName(origin)
	case "group":
		sdef.SandboxDir += "/" + GroupPrefix + VersionToName(origin)
		if ! GreaterOrEqualVersion(sdef.Version, []int{5, 7, 17}) {
			fmt.Println("Group replication requires MySQL 5.7.17 or greater")
			os.Exit(1)
		}
	default:
		fmt.Println("Unrecognized topology. Accepted: 'master-slave', 'group'")
		os.Exit(1)
	}

	if common.DirExists(sdef.SandboxDir) {
		fmt.Printf("Directory %s already exists\n", sdef.SandboxDir)
		os.Exit(1)
	}

	err := os.Mkdir(sdef.SandboxDir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch topology {
	case "master-slave":
		CreateMasterSlaveReplication(sdef, origin, nodes)
	case "group":
		CreateGroupReplication(sdef, origin, nodes)
	}
}
