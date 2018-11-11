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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"path"
	"testing"
)

func okPortExists(t *testing.T, dirName string, port int) {
	sandboxList := defaults.ReadCatalog()
	// In the sandbox catalog (a map of sandbox structures),
	// each entry is indexed with the full path of the sandbox
	// directory.
	for name, sb := range sandboxList {
		if name == dirName {
			// A sandbox can have more than one port
			// We loop through it to find the requested one
			for _, p := range sb.Port {
				if p == port {
					t.Logf("ok - port %d found in %s\n", port, dirName)
					return
				}
			}
		}
	}
	// If we reach this point, the port was not found
	t.Logf("not ok - port %d not found in %s\n", port, dirName)
	t.Fail()
}

func okExecutableExists(t *testing.T, dir, executable string) {
	fullPath := path.Join(dir, executable)
	if common.ExecExists(fullPath) {
		t.Logf("ok - %s exists\n", fullPath)
	} else {
		t.Logf("not ok - %s does not exist\n", fullPath)
		t.Fail()
	}
}

func okDirExists(t *testing.T, dir string) {
	if common.DirExists(dir) {
		t.Logf("ok - %s exists\n", dir)
	} else {
		t.Logf("not ok - %s does not exist\n", dir)
		t.Fail()
	}
}

type versionRec struct {
	version string
	path    string
	port    int
}

func TestCreateSandbox(t *testing.T) {
	setMockEnvironment("mock_dir")
	var versions = []versionRec{
		{"5.0.89", "5_0_89", 5089},
		{"5.1.67", "5_1_67", 5167},
		{"5.5.48", "5_5_48", 5548},
		{"5.6.78", "5_6_78", 5678},
		{"5.7.22", "5_7_22", 5722},
		{"8.0.11", "8_0_11", 8011},
	}
	for _, v := range versions {
		mysqlVersion := v.version
		pathVersion := v.path
		port := v.port
		createMockVersion(mysqlVersion)
		var sandboxDef = SandboxDef{
			Version:    mysqlVersion,
			Basedir:    path.Join(mockSandboxBinary, mysqlVersion),
			SandboxDir: mockSandboxHome,
			DirName:    defaults.Defaults().SandboxPrefix + pathVersion,
			LoadGrants: true,
			// InstalledPorts: []int{1186, 3306, 33060},
			InstalledPorts: defaults.Defaults().ReservedPorts,
			Port:           port,
			DbUser:         defaults.DbUserValue,
			RplUser:        defaults.RplUserValue,
			DbPassword:     defaults.DbPasswordValue,
			RplPassword:    defaults.RplPasswordValue,
			RemoteAccess:   defaults.RemoteAccessValue,
			BindAddress:    defaults.BindAddressValue,
		}

		CreateSingleSandbox(sandboxDef)
		okDirExists(t, sandboxDef.Basedir)
		sandboxDir := path.Join(sandboxDef.SandboxDir, "msb_"+pathVersion)
		okDirExists(t, sandboxDir)
		t.Logf("%#v", sandboxDir)
		okDirExists(t, path.Join(sandboxDir, "data"))
		okDirExists(t, path.Join(sandboxDir, "tmp"))
		okExecutableExists(t, sandboxDir, "start")
		okExecutableExists(t, sandboxDir, "use")
		okExecutableExists(t, sandboxDir, "stop")
		okPortExists(t, sandboxDir, sandboxDef.Port)
	}
	removeMockEnvironment("mock_dir")
}
