package main

import (
    "os"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
)

func main() {
	version := "5.7.22"
	sandbox_home := os.Getenv("HOME") + "/sandboxes"
	sandbox_binary := os.Getenv("HOME") + "/opt/mysql"
	basedir := sandbox_binary + "/" + version
	port := 5722
	user := "msandbox"
	password := "psandbox"

	if !common.DirExists(sandbox_home) {
		common.Mkdir(sandbox_home)
	}

	var sdef =	sandbox.SandboxDef{
		 Version: version, 
		 Basedir: basedir,
		 SandboxDir: sandbox_home,
		 DirName: "msb_5_7_22",
		 LoadGrants:true,
		 InstalledPorts:[]int{1186, 3306, 33060},
		 Port: port,
		 DbUser: user,
		 DbPassword: password,
		 RplUser: "r" + user,
		 RplPassword: "r" + password,
		 RemoteAccess:"127.%",
		 BindAddress:"127.0.0.1",
	}

	sandbox.CreateSingleSandbox(sdef)
	common.Run_cmd(sandbox_home + "/msb_5_7_22/test_sb")
	sandbox.RemoveSandbox(sandbox_home, "msb_5_7_22", false)
	defaults.DeleteFromCatalog(sandbox_home+"/msb_5_7_22")
}
