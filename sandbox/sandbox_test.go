package sandbox

import (
	"github.com/datacharmer/dbdeployer/common"
	"testing"
)

func ok_executable_exists(t *testing.T, dir, executable string) {
	full_path := dir + "/" + executable
	if common.ExecExists(full_path) {
		t.Logf("ok - %s exists\n", full_path)
	} else {
		t.Logf("not ok - %s does not exist\n", full_path)
		t.Fail()
	}
}

func ok_dir_exists(t *testing.T, dir string) {
	if common.DirExists(dir) {
		t.Logf("ok - %s exists\n", dir)
	} else {
		t.Logf("not ok - %s does not exist\n", dir)
		t.Fail()
	}
}

type version_rec struct {
	version string
	path    string
	port    int
}

func TestCreateSandbox(t *testing.T) {
	set_mock_environment("mock_dir")
	var versions = []version_rec{
		{"5.0.89", "5_0_89", 5089},
		{"5.1.67", "5_1_67", 5167},
		{"5.5.48", "5_5_48", 5548},
		{"5.6.78", "5_6_78", 5678},
		{"5.7.22", "5_7_22", 5722},
		{"8.0.11", "8_0_11", 8011},
	}
	for _, v := range versions {
		mysql_version := v.version
		path_version := v.path
		port := v.port
		create_mock_version(mysql_version)
		var sdef = SandboxDef{
			Version:        mysql_version,
			Basedir:        mock_sandbox_binary + "/" + mysql_version,
			SandboxDir:     mock_sandbox_home,
			DirName:        "msb_" + path_version,
			LoadGrants:     true,
			InstalledPorts: []int{1186, 3306, 33060},
			Port:           port,
			DbUser:         "msandbox",
			RplUser:        "rsandbox",
			DbPassword:     "msandbox",
			RplPassword:    "rsandbox",
			RemoteAccess:   "127.%",
			BindAddress:    "127.0.0.1",
		}

		CreateSingleSandbox(sdef)
		//exec_list := CreateSingleSandbox(sdef)
		//t.Logf("%#v", exec_list)
		ok_dir_exists(t, sdef.Basedir)
		sandbox_dir := sdef.SandboxDir + "/msb_" + path_version
		ok_dir_exists(t, sandbox_dir)
		t.Logf("%#v", sandbox_dir)
		ok_dir_exists(t, sandbox_dir+"/data")
		ok_dir_exists(t, sandbox_dir+"/tmp")
		ok_executable_exists(t, sandbox_dir, "start")
		ok_executable_exists(t, sandbox_dir, "use")
		ok_executable_exists(t, sandbox_dir, "stop")
	}
	remove_mock_environment("mock_dir")
}
