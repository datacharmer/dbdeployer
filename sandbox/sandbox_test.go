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
	"github.com/datacharmer/dbdeployer/compare"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"os"
	"path"
	"strings"
	"testing"
)

func okPortExists(t *testing.T, dirName string, port int) {
	sandboxList, err := defaults.ReadCatalog()
	if err != nil {
		t.Logf("not ok - error reading catalog \n")
		t.Fail()
		return
	}
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

var singleScriptNames = []string{
	globals.ScriptStart,
	globals.ScriptStop,
	globals.ScriptStatus,
	globals.ScriptRestart,
	globals.ScriptClear,
	globals.ScriptSendKill,
	globals.ScriptUse,
}

func testCreateMockSandbox(t *testing.T) {
	err := setMockEnvironment("mock_dir")
	if err != nil {
		t.Fatal("mock dir creation failed")
	}
	compare.OkIsNil("mock creation", err, t)
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
		err = createMockVersion(mysqlVersion)
		compare.OkIsNil("version creation", err, t)
		var sandboxDef = SandboxDef{
			Version:    mysqlVersion,
			Basedir:    path.Join(mockSandboxBinary, mysqlVersion),
			SandboxDir: mockSandboxHome,
			DirName:    defaults.Defaults().SandboxPrefix + pathVersion,
			LoadGrants: true,
			// InstalledPorts: []int{1186, 3306, 33060},
			InstalledPorts: defaults.Defaults().ReservedPorts,
			Port:           port,
			DbUser:         globals.DbUserValue,
			RplUser:        globals.RplUserValue,
			DbPassword:     globals.DbPasswordValue,
			RplPassword:    globals.RplPasswordValue,
			RemoteAccess:   globals.RemoteAccessValue,
			BindAddress:    globals.BindAddressValue,
		}

		err := CreateStandaloneSandbox(sandboxDef)
		if err != nil {
			t.Logf("Sandbox %s %s\n", mysqlVersion, pathVersion)
			t.Logf(globals.ErrCreatingSandbox, err)
			t.Fail()
		}
		okDirExists(t, sandboxDef.Basedir)
		sandboxDir := path.Join(sandboxDef.SandboxDir, defaults.Defaults().SandboxPrefix+pathVersion)
		okDirExists(t, sandboxDir)
		t.Logf("%#v", sandboxDir)
		okDirExists(t, path.Join(sandboxDir, "data"))
		okDirExists(t, path.Join(sandboxDir, "tmp"))
		for _, script := range singleScriptNames {
			okExecutableExists(t, sandboxDir, script)
		}
		okPortExists(t, sandboxDir, sandboxDef.Port)
	}
	err = removeMockEnvironment("mock_dir")
	compare.OkIsNil("removal", err, t)
}

func testCreateStandaloneSandbox(t *testing.T) {

	latestVersion := preCreationChecks(t)
	t.Logf("latest: %s\n", latestVersion)
	pathVersion := strings.Replace(latestVersion, ".", "_", -1)
	port, err := common.VersionToPort(latestVersion)
	if err != nil {
		t.Fatalf("can't convert version %s to port", latestVersion)
	}
	var sandboxDef = SandboxDef{
		Version:        latestVersion,
		Basedir:        path.Join(defaults.Defaults().SandboxBinary, latestVersion),
		SandboxDir:     defaults.Defaults().SandboxHome,
		DirName:        defaults.Defaults().SandboxPrefix + pathVersion,
		LoadGrants:     true,
		InstalledPorts: defaults.Defaults().ReservedPorts,
		Port:           port,
		DbUser:         globals.DbUserValue,
		RplUser:        globals.RplUserValue,
		DbPassword:     globals.DbPasswordValue,
		RplPassword:    globals.RplPasswordValue,
		RemoteAccess:   globals.RemoteAccessValue,
		BindAddress:    globals.BindAddressValue,
	}

	if common.IsEnvSet("SHOW_SANDBOX_DEF") {
		t.Logf("%s", sandboxDefToJson(sandboxDef))
	}

	err = CreateStandaloneSandbox(sandboxDef)
	if err != nil {
		t.Fatal(fmt.Sprintf(globals.ErrCreatingSandbox, err))
	}

	sandboxDir := path.Join(sandboxDef.SandboxDir, defaults.Defaults().SandboxPrefix+pathVersion)
	okDirExists(t, sandboxDir)
	okDirExists(t, path.Join(sandboxDir, "data"))
	okDirExists(t, path.Join(sandboxDir, "tmp"))
	for _, script := range singleScriptNames {
		okExecutableExists(t, sandboxDir, script)
	}
	okPortExists(t, sandboxDir, sandboxDef.Port)
	_, err = RemoveSandbox(defaults.Defaults().SandboxHome, sandboxDef.DirName, false)
	if err != nil {
		t.Fatal(fmt.Sprint(globals.ErrWhileRemoving, sandboxDef.SandboxDir, err))
	}
	err = defaults.DeleteFromCatalog(sandboxDir)
	if err != nil {
		t.Fatal(fmt.Sprintf(globals.ErrRemovingFromCatalog, sandboxDef.SandboxDir))
	}
}

func preCreationChecks(t *testing.T) string {
	if common.IsEnvSet("SKIP_REAL_SANDBOX_TEST") || common.IsEnvSet("TRAVIS") {
		t.Skip("User choice")
	}
	sandboxBinary := os.Getenv("SANDBOX_BINARY")
	if sandboxBinary == "" {
		sandboxBinary = defaults.Defaults().SandboxBinary
	}
	if !common.DirExists(sandboxBinary) {
		t.Skip("SANDBOX_BINARY directory not found")
	}
	versions, err := common.GetVersionsFromDir(sandboxBinary)
	if err != nil || len(versions) == 0 {
		t.Skip("error while retrieving versions")
	}
	wantedVersion := os.Getenv("WANTED_VERSION")
	if wantedVersion == "" {
		wantedVersion = "5.7"
	}
	sortedVersions := common.SortVersionsSubset(versions, wantedVersion)
	if len(sortedVersions) < 1 {
		t.Skip("no items found for version ", wantedVersion, "\n")
	}
	latestVersion := sortedVersions[len(sortedVersions)-1]
	return latestVersion
}

func testCreateReplicationSandbox(t *testing.T) {
	latestVersion := preCreationChecks(t)

	t.Logf("latest: %s\n", latestVersion)
	pathVersion := strings.Replace(latestVersion, ".", "_", -1)
	t.Logf("path: %s\n", pathVersion)
	var sandboxDef = SandboxDef{
		Version:        latestVersion,
		Basedir:        path.Join(defaults.Defaults().SandboxBinary, latestVersion),
		SandboxDir:     defaults.Defaults().SandboxHome,
		DirName:        defaults.Defaults().MasterSlavePrefix + pathVersion,
		LoadGrants:     false,
		InstalledPorts: defaults.Defaults().ReservedPorts,
		DbUser:         globals.DbUserValue,
		RplUser:        globals.RplUserValue,
		DbPassword:     globals.DbPasswordValue,
		RplPassword:    globals.RplPasswordValue,
		RemoteAccess:   globals.RemoteAccessValue,
		BindAddress:    globals.BindAddressValue,
	}

	if common.IsEnvSet("SHOW_SANDBOX_DEF") {
		t.Logf("%s", sandboxDefToJson(sandboxDef))
	}

	err := CreateReplicationSandbox(sandboxDef, latestVersion, globals.MasterSlaveLabel, 3, "127.0.0.1", "1", "2,3")
	if err != nil {
		t.Fatal(fmt.Sprintf(globals.ErrCreatingSandbox, err))
	}

	sandboxDir := path.Join(sandboxDef.SandboxDir, defaults.Defaults().MasterSlavePrefix+pathVersion)
	okDirExists(t, sandboxDir)
	dirs := []string{
		defaults.Defaults().MasterName,
		defaults.Defaults().NodePrefix + "1",
		defaults.Defaults().NodePrefix + "2",
	}
	for _, dir := range dirs {
		okDirExists(t, path.Join(sandboxDir, dir))
		okDirExists(t, path.Join(sandboxDir, dir, "data"))
		okDirExists(t, path.Join(sandboxDir, dir, "tmp"))
		for _, script := range singleScriptNames {
			okExecutableExists(t, path.Join(sandboxDir, dir), script)
		}
	}
	for _, script := range singleScriptNames {
		okExecutableExists(t, sandboxDir, script+"_all")
	}
	_, err = RemoveSandbox(defaults.Defaults().SandboxHome, sandboxDef.DirName, false)
	if err != nil {
		t.Fatal(fmt.Sprint(globals.ErrWhileRemoving, sandboxDef.SandboxDir, err))
	}
	t.Logf("sandbox to delete: %s\n", sandboxDef.SandboxDir)
	err = defaults.DeleteFromCatalog(sandboxDir)
	if err != nil {
		t.Fatal(fmt.Sprintf(globals.ErrRemovingFromCatalog, sandboxDef.SandboxDir))
	}
}

func TestCreateSandbox(t *testing.T) {
	t.Run("single", testCreateStandaloneSandbox)
	t.Run("replication", testCreateReplicationSandbox)
	t.Run("mock", testCreateMockSandbox)
}
