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
	err := SetMockEnvironment(DefaultMockDir)
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
		err = CreateMockVersion(mysqlVersion)
		compare.OkIsNil("version creation", err, t)
		var sandboxDef = SandboxDef{
			Version:    mysqlVersion,
			Flavor:     common.MySQLFlavor,
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
		okDirExists(t, path.Join(sandboxDir, "data"))
		okDirExists(t, path.Join(sandboxDir, "tmp"))
		for _, script := range singleScriptNames {
			okExecutableExists(t, sandboxDir, script)
		}
		okPortExists(t, sandboxDir, sandboxDef.Port)
	}
	err = RemoveMockEnvironment(DefaultMockDir)
	compare.OkIsNil("removal", err, t)
}

func testDetectFlavor(t *testing.T) {

	err := SetMockEnvironment(DefaultMockDir)
	if err != nil {
		t.Fatal("mock dir creation failed")
	}

	type FlavorDetection struct {
		version  string
		setup    []MockFileSet
		expected string
	}

	var flavorDetectionSet = []FlavorDetection{
		FlavorDetection{"5.0.0",
			MySQLMockSet(false),
			common.MySQLFlavor,
		},
		FlavorDetection{"5.5.0",
			MySQLMockSet(true),
			common.MySQLFlavor,
		},
		FlavorDetection{"3.0.0", []MockFileSet{
			MockFileSet{"bin",
				[]ScriptDef{
					{globals.FnTiDbServer, noOpMockTemplateName, true},
				}},
		},
			common.TiDbFlavor,
		},
		FlavorDetection{"10.0.0", []MockFileSet{
			MockFileSet{"bin",
				[]ScriptDef{
					{"aria_chk", noOpMockTemplateName, true},
				}},
		},
			common.MariaDbFlavor,
		},
		FlavorDetection{"10.3.0", []MockFileSet{
			MockFileSet{"lib",
				[]ScriptDef{
					{globals.FnLibMariadbClientA, noOpMockTemplateName, false},
				}},
		},
			common.MariaDbFlavor,
		},
		FlavorDetection{"8.0.14", []MockFileSet{
			MockFileSet{"lib",
				[]ScriptDef{
					{globals.FnLibPerconaServerClientA, noOpMockTemplateName, false},
				}},
		},
			common.PerconaServerFlavor,
		},
		FlavorDetection{"8.0.12", []MockFileSet{
			MockFileSet{"bin",
				[]ScriptDef{
					{globals.FnGarbd, noOpMockTemplateName, true},
				}},
			MockFileSet{"lib",
				[]ScriptDef{
					{globals.FnLibPerconaServerClientSo, noOpMockTemplateName, false},
					{globals.FnLibGaleraSmmSo, noOpMockTemplateName, false},
				}},
		},
			common.PxcFlavor,
		},
		FlavorDetection{"5.7.77", []MockFileSet{
			MockFileSet{"bin",
				[]ScriptDef{
					{globals.FnGarbd, noOpMockTemplateName, true},
				}},
			MockFileSet{"lib",
				[]ScriptDef{
					{globals.FnLibPerconaServerClientA, noOpMockTemplateName, false},
					{globals.FnLibGaleraSmmA, noOpMockTemplateName, false},
				}},
		},
			common.PxcFlavor,
		},
		FlavorDetection{"6.7.8", []MockFileSet{
			MockFileSet{"bin",
				[]ScriptDef{
					{globals.FnNdbdMgm, noOpMockTemplateName, true},
					{globals.FnNdbdMgmd, noOpMockTemplateName, true},
					{globals.FnNdbd, noOpMockTemplateName, true},
					{globals.FnNdbdMtd, noOpMockTemplateName, true},
				}},
			MockFileSet{"lib",
				[]ScriptDef{
					{globals.FnLibMySQLClientA, noOpMockTemplateName, false},
					{globals.FnNdbdEngineSo, noOpMockTemplateName, false},
				}},
		},
			common.NdbFlavor,
		},
	}

	for _, fd := range flavorDetectionSet {
		err = CreateCustomMockVersion(fd.version, fd.setup)
		compare.OkIsNil("mock creation", err, t)
		basedir := path.Join(mockSandboxBinary, fd.version)
		detectedFlavor := common.DetectBinaryFlavor(basedir)
		compare.OkEqualString(fmt.Sprintf("%s/%s: flavor detected %s",
			fd.setup[0].dir, fd.setup[0].fileSet[0].scriptName, detectedFlavor),
			detectedFlavor, fd.expected, t)
	}

	err = RemoveMockEnvironment(DefaultMockDir)
	compare.OkIsNil("removal", err, t)

}

func testCreateTidbMockSandbox(t *testing.T) {
	err := SetMockEnvironment(DefaultMockDir)
	if err != nil {
		t.Fatal("mock dir creation failed")
	}
	compare.OkIsNil("mock creation", err, t)
	var versions = []versionRec{
		{"3.0.30", "3_0_30", 3030},
		{"5.0.50", "5_0_50", 5050},
		{"8.0.80", "8_0_80", 8080},
	}
	err = CreateMockVersion("5.7.25")
	compare.OkIsNil("MySQL support version creation", err, t)
	for _, v := range versions {
		tidbVersion := v.version
		pathVersion := v.path
		port := v.port
		fileSet := MockFileSet{
			"bin",
			[]ScriptDef{
				{globals.FnTiDbServer, globals.TmplTidbMock, true},
			},
		}
		fileSets := []MockFileSet{fileSet}
		err = CreateCustomMockVersion(tidbVersion, fileSets)
		compare.OkIsNil("TiDB version creation", err, t)
		mockSandboxDir := defaults.Defaults().SandboxPrefix + pathVersion
		var sandboxDef = SandboxDef{
			Version:         tidbVersion,
			Flavor:          common.TiDbFlavor,
			Basedir:         path.Join(mockSandboxBinary, tidbVersion),
			SandboxDir:      mockSandboxHome,
			DirName:         mockSandboxDir,
			SocketInDatadir: true,
			LoadGrants:      true,
			InstalledPorts:  defaults.Defaults().ReservedPorts,
			Port:            port,
			DbUser:          globals.DbUserValue,
			RplUser:         globals.RplUserValue,
			DbPassword:      globals.DbPasswordValue,
			RplPassword:     globals.RplPasswordValue,
			RemoteAccess:    globals.RemoteAccessValue,
			BindAddress:     globals.BindAddressValue,
			ClientBasedir:   path.Join(mockSandboxBinary, "5.7.25"),
		}

		err := CreateStandaloneSandbox(sandboxDef)
		if err != nil {
			t.Logf("Sandbox %s %s\n", tidbVersion, pathVersion)
			t.Logf(globals.ErrCreatingSandbox, err)
			t.Fail()
		}
		okDirExists(t, sandboxDef.Basedir)
		sandboxDir := path.Join(sandboxDef.SandboxDir, defaults.Defaults().SandboxPrefix+pathVersion)
		okDirExists(t, sandboxDir)
		okDirExists(t, path.Join(sandboxDir, "data"))
		okDirExists(t, path.Join(sandboxDir, "tmp"))
		for _, script := range singleScriptNames {
			okExecutableExists(t, sandboxDir, script)
		}
		okPortExists(t, sandboxDir, sandboxDef.Port)

		_, err = RemoveCustomSandbox(mockSandboxHome, sandboxDef.DirName, false, true)
		if err != nil {
			t.Fatal(fmt.Sprintf(globals.ErrWhileRemoving, sandboxDir, err))
		}
		err = defaults.DeleteFromCatalog(sandboxDir)
		if err != nil {
			t.Fatal(fmt.Sprintf(globals.ErrRemovingFromCatalog, sandboxDir))
		}

	}
	err = RemoveMockEnvironment(DefaultMockDir)
	compare.OkIsNil("removal", err, t)
}

func expectFailure(sandboxDef SandboxDef, label, deployment, regex string, args map[string]string, t *testing.T) {
	var topology string
	var masterIp string
	var nodesStr string
	var origin string
	var ok bool
	switch deployment {
	case "single":
		err := CreateStandaloneSandbox(sandboxDef)
		compare.OkIsNotNil(label, err, t)
		if err != nil {
			compare.OkMatchesString(label, err.Error(), regex, t)
		}
	case "replication":
		topology, ok = args["topology"]
		if !ok {
			topology = globals.MasterSlaveLabel
		}
		t.Logf("<%s>\n", topology)
		masterIp, ok = args["masterIp"]
		if !ok {
			masterIp = "127.0.0.1"
		}
		nodesStr, ok = args["nodes"]
		nodes := 3
		if ok {
			nodes = common.Atoi(nodesStr)
		}
		origin, ok = args["origin"]
		if !ok {
			origin = sandboxDef.Version
		}
		err := CreateReplicationSandbox(sandboxDef, origin, ReplicationData{
			Topology:   topology,
			Nodes:      nodes,
			MasterIp:   masterIp,
			MasterList: "",
			SlaveList:  ""})
		compare.OkIsNotNil(label, err, t)
		if err != nil {
			compare.OkMatchesString(label, err.Error(), regex, t)
		}
	}

}

func testFailSandboxConditions(t *testing.T) {
	err := SetMockEnvironment(DefaultMockDir)
	if err != nil {
		t.Fatal("mock dir creation failed")
	}
	compare.OkIsNil("mock creation", err, t)
	mysqlVersion := "5.6.99"
	pathVersion := "5_6_99"
	err = CreateMockVersion(mysqlVersion)
	compare.OkIsNil("version creation", err, t)
	var sandboxDef = SandboxDef{
		Version:        mysqlVersion,
		Flavor:         common.MySQLFlavor,
		Basedir:        path.Join(mockSandboxBinary, mysqlVersion),
		SandboxDir:     mockSandboxHome,
		DirName:        defaults.Defaults().SandboxPrefix + pathVersion,
		LoadGrants:     true,
		InstalledPorts: defaults.Defaults().ReservedPorts,
		DbUser:         globals.DbUserValue,
		RplUser:        globals.RplUserValue,
		DbPassword:     globals.DbPasswordValue,
		RplPassword:    globals.RplPasswordValue,
		RemoteAccess:   globals.RemoteAccessValue,
		BindAddress:    globals.BindAddressValue,
	}

	var emptyMap map[string]string

	sandboxDef.Port = 1000
	expectFailure(sandboxDef, "lower port error", "single", "must be > 1024", emptyMap, t)

	sandboxDef.Port = 5699
	sandboxDef.EnableMysqlX = true

	expectFailure(sandboxDef, "invalid MySQLX", "single", "requires MySQL version '5.7.12'", emptyMap, t)

	sandboxDef.EnableMysqlX = false
	sandboxDef.SlavesSuperReadOnly = true

	var replMap = make(map[string]string)
	replMap["origin"] = mysqlVersion
	expectFailure(sandboxDef, "invalid super read only",
		"replication",
		"requires MySQL version '5.7.8'",
		replMap,
		t)

	sandboxDef.SlavesSuperReadOnly = false
	sandboxDef.ExposeDdTables = true

	expectFailure(sandboxDef, "invalid expose DD tables", "single", "requires MySQL version '8.0.0'", emptyMap, t)

	sandboxDef.ExposeDdTables = false

	replMap["nodes"] = "1"
	expectFailure(sandboxDef, "invalid nodes",
		"replication",
		"less than 2 nodes",
		replMap,
		t)

	replMap["nodes"] = "2"
	replMap["masterIp"] = "127.x.9.1"
	expectFailure(sandboxDef, "invalid master ip",
		"replication",
		"valid IPV4",
		replMap,
		t)

	sandboxDef.CustomMysqld = "DummyFile"
	expectFailure(sandboxDef, "invalid mysqld file", "single", "same directory", emptyMap, t)

	sandboxDef.CustomMysqld = ""
	sandboxDef.MyCnfFile = "dummyFile"

	expectFailure(sandboxDef, "invalid cnf file", "single", `file.*not found`, emptyMap, t)

	sandboxDef.MyCnfFile = ""
	mysqlVersion = "5.0.99"
	pathVersion = "5_0_99"
	err = CreateMockVersion(mysqlVersion)
	compare.OkIsNil("version creation", err, t)
	sandboxDef.SlavesReadOnly = true
	sandboxDef.Port = 5099
	sandboxDef.Version = mysqlVersion

	replMap["nodes"] = "2"
	replMap["origin"] = mysqlVersion
	replMap["masterIp"] = "127.0.0.1"
	expectFailure(sandboxDef, "invalid read only",
		"replication",
		"requires MySQL version '5.1.0'",
		replMap,
		t)

	sandboxDef.MyCnfFile = ""
	mysqlVersion = "5.7.0"
	pathVersion = "5_7_0"
	err = CreateMockVersion(mysqlVersion)
	compare.OkIsNil("version creation", err, t)
	sandboxDef.SlavesReadOnly = true
	sandboxDef.Port = 5700
	sandboxDef.Version = mysqlVersion
	sandboxDef.SlavesReadOnly = false

	replMap["origin"] = mysqlVersion
	replMap["topology"] = globals.GroupLabel
	expectFailure(sandboxDef, "invalid group",
		"replication",
		"requires MySQL version '5.7.17'",
		replMap,
		t)

	replMap["topology"] = globals.FanInLabel
	expectFailure(sandboxDef, "invalid fan-in",
		"replication",
		`multi-source.*requires MySQL version '5.7.9'`,
		replMap,
		t)

	replMap["topology"] = globals.AllMastersLabel
	expectFailure(sandboxDef, "invalid all-masters",
		"replication",
		`multi-source.*requires MySQL version '5.7.9'`,
		replMap,
		t)

	// t.Logf("%+v", err)
	err = RemoveMockEnvironment(DefaultMockDir)
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
		Flavor:         common.MySQLFlavor,
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
	_, err = RemoveCustomSandbox(defaults.Defaults().SandboxHome, sandboxDef.DirName, false, false)
	if err != nil {
		t.Fatal(fmt.Sprint(globals.ErrWhileRemoving, sandboxDef.SandboxDir, err))
	}
	err = defaults.DeleteFromCatalog(sandboxDir)
	if err != nil {
		t.Fatal(fmt.Sprintf(globals.ErrRemovingFromCatalog, sandboxDef.SandboxDir))
	}
}

func preCreationChecks(t *testing.T) string {
	compare.SkipOnDemand("SKIP_REAL_SANDBOX_TEST", t)
	compare.SkipOnDemand("GITHUB_ACTIONS", t)
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
		Flavor:         common.MySQLFlavor,
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

	err := CreateReplicationSandbox(sandboxDef, latestVersion, ReplicationData{
		Topology: globals.MasterSlaveLabel, Nodes: 3, MasterIp: "127.0.0.1", MasterList: "1", SlaveList: "2,3"})
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
	_, err = RemoveCustomSandbox(defaults.Defaults().SandboxHome, sandboxDef.DirName, false, false)
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
	if common.FileExists(defaults.SandboxRegistry) {
		catalog, err := defaults.ReadCatalog()
		if err != nil {
			t.Fatalf("error getting catalog list: %s", err)
		}
		if len(catalog) > 0 {
			t.Fatalf("catalog (%s) not empty", defaults.SandboxRegistry)
		}
	}
	if common.DirExists(defaults.Defaults().SandboxHome) {
		installed, err := common.GetInstalledSandboxes(defaults.Defaults().SandboxHome)
		if err != nil {
			t.Fatalf("error getting sandboxes list: %s", err)
		}
		if len(installed) > 0 {
			t.Fatalf("sandbox home (%s) not empty", defaults.Defaults().SandboxHome)
		}
	}
	t.Run("single", testCreateStandaloneSandbox)
	t.Run("replication", testCreateReplicationSandbox)
	t.Run("mock", testCreateMockSandbox)
	t.Run("mocktidb", testCreateTidbMockSandbox)
	t.Run("expectedFailures", testFailSandboxConditions)
	t.Run("flavors", testDetectFlavor)
}
