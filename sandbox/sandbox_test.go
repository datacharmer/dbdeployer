package sandbox

import (
	"testing"
	"github.com/datacharmer/dbdeployer/common"
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

func TestCreateSandbox(t *testing.T) {
	set_mock_environment("mock_dir")
	create_mock_version("5.7.22")
	var sdef =	SandboxDef{
		 Version:"5.7.22",
		 Basedir: mock_sandbox_binary + "/5.7.22",
		 SandboxDir: mock_sandbox_home,
		 LoadGrants:true,
		 InstalledPorts:[]int{1186, 3306, 33060},
		 Port:5722,
		 DbUser:"msandbox",
		 RplUser:"rsandbox",
		 DbPassword:"msandbox",
		 RplPassword:"rsandbox",
		 RemoteAccess:"127.%",
		 BindAddress:"127.0.0.1",
	}
	
	exec_list := CreateSingleSandbox(sdef)
	t.Logf("%#v", exec_list)
	ok_dir_exists(t, sdef.Basedir)
	sandbox_dir:= sdef.SandboxDir + "/msb_5_7_22"
	ok_dir_exists(t, sandbox_dir)
	ok_dir_exists(t, sandbox_dir + "/data")
	ok_dir_exists(t, sandbox_dir + "/tmp")
	ok_executable_exists(t, sandbox_dir, "start")
	ok_executable_exists(t, sandbox_dir, "use")
	ok_executable_exists(t, sandbox_dir, "stop")
	
	remove_mock_environment("mock_dir")
}
