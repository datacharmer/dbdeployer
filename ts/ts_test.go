// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2022 Giuseppe Maxia
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

package ts

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"text/template"

	"github.com/datacharmer/dbdeployer/cmd"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/ops"

	"github.com/rogpeppe/go-internal/testscript"
)

var (
	dryRun     bool = os.Getenv("DRY_RUN") != ""
	testDebug  bool = os.Getenv("TEST_DEBUG") != ""
	onlyLatest bool = os.Getenv("ONLY_LATEST") != ""
)

func preTest(t *testing.T, dirName string) []string {
	conditionalPrint("entering %s\n", t.Name())
	if dryRun {
		t.Skip("Dry Run")
	}
	if !common.DirExists("testdata") {
		t.Skip("no testdata found")
	}
	// Directories in testdata are created by the setup code in TestMain
	dirs, err := filepath.Glob("testdata/" + dirName + "/*")
	if err != nil {
		t.Skipf("no directories found in testdata/%s", dirName)
	}
	conditionalPrint("Directories: %v\n", dirs)
	return dirs
}

func testDbDeployer(t *testing.T, name string, parallel bool) {
	if parallel {
		t.Parallel()
	}
	dirs := preTest(t, name)
	for _, dir := range dirs {
		subTestName := path.Base(dir)
		conditionalPrint("entering %s/%s", t.Name(), subTestName)
		t.Run(subTestName, func(t *testing.T) {
			testscript.Run(t, testscript.Params{
				Dir:                 dir,
				Cmds:                customCommands(),
				Condition:           customConditions,
				Setup:               dbdeployerSetup(t, dir),
				RequireExplicitExec: true,
			})
		})
	}
}

func TestFeature(t *testing.T) {
	testDbDeployer(t, "feature", true)
}

func TestReplication(t *testing.T) {
	testDbDeployer(t, "replication", true)
}

func TestMultiSource(t *testing.T) {
	testDbDeployer(t, "multi-source", false)
}

func TestGroup(t *testing.T) {
	testDbDeployer(t, "group", false)
}

func getFlavor(version string) string {
	sandboxBinary := os.Getenv("SANDBOX_BINARY")
	if sandboxBinary == "" {
		sandboxBinary = defaults.Defaults().SandboxBinary
	}
	filePath := path.Join(sandboxBinary, version, "FLAVOR")
	if !common.FileExists(filePath) {
		return common.MySQLFlavor
	}
	flavor, err := os.ReadFile(filePath)
	if err != nil || string(flavor) == "" {
		return common.MySQLFlavor
	}

	return strings.TrimSpace(string(flavor))
}

func getVersionList(shortVersions []string) []string {
	sandboxBinary := os.Getenv("SANDBOX_BINARY")
	if sandboxBinary == "" {
		sandboxBinary = defaults.Defaults().SandboxBinary
	}
	var versionList []string
	for _, sv := range shortVersions {
		latest := common.GetLatestVersion(sandboxBinary, sv, common.MySQLFlavor)
		if latest != "" && strings.HasPrefix(latest, sv) {
			versionList = append(versionList, latest)
		}
	}
	return versionList
}

func initializeEnv(versionList []string) error {
	sandboxBinary := os.Getenv("SANDBOX_BINARY")
	if sandboxBinary == "" {
		sandboxBinary = defaults.Defaults().SandboxBinary
	}
	sandboxHome := os.Getenv("SANDBOX_HOME")
	if sandboxHome == "" {
		sandboxHome = defaults.Defaults().SandboxHome
	}
	needInitializing := false
	if !common.DirExists(sandboxBinary) || !common.DirExists(sandboxHome) {
		needInitializing = true
	}

	if needInitializing {
		err := ops.InitEnvironment(ops.InitOptions{
			SandboxBinary:  sandboxBinary,
			SandboxHome:    sandboxHome,
			SkipCompletion: true,
		})
		if err != nil {
			return err
		}
	}
	for _, v := range versionList {
		latest := common.GetLatestVersion(sandboxBinary, v, common.MySQLFlavor)

		if latest != "" && strings.HasPrefix(latest, v) {
			conditionalPrint("found latest %s: %s\n", v, latest)
			continue
		}
		err := ops.GetRemoteTarball(ops.DownloadsOptions{
			SandboxBinary:     sandboxBinary,
			TarballOS:         runtime.GOOS,
			Flavor:            common.MySQLFlavor,
			Version:           v,
			Newest:            true,
			Minimal:           strings.EqualFold(runtime.GOOS, "linux"),
			Unpack:            true,
			DeleteAfterUnpack: true,
			Retries:           2,
		})
		if err != nil {
			conditionalPrint("no tarball retrieved for version %s\n", v)
			return nil
		}
		conditionalPrint("retrieved tarball for version %s\n", v)
	}
	return nil
}

func TestMain(m *testing.M) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error determining current directory\n")
		os.Exit(1)
	}
	shortVersions := []string{"4.1", "5.0", "5.1", "5.5", "5.6", "5.7", "8.0"}
	if os.Getenv("GITHUB_ACTIONS") != "" {
		shortVersions = []string{"5.6", "5.7", "8.0"}
	}
	customShortVersions := os.Getenv("TEST_SHORT_VERSIONS")
	if customShortVersions != "" {
		shortVersions = strings.Split(customShortVersions, ",")
	}
	if onlyLatest {
		shortVersions = []string{"8.0"}
	}
	conditionalPrint("short versions: %v\n", shortVersions)
	err = initializeEnv(shortVersions)
	if err != nil {
		conditionalPrint("error initializing the environment - Skipping tests: %s\n", err)
		os.Exit(0)
	}
	versions := getVersionList(shortVersions)

	conditionalPrint("versions: %v\n", versions)
	for _, v := range versions {
		_ = os.Chdir(currentDir)
		label := strings.Replace(v, ".", "_", -1)
		intendedPort := strings.Replace(v, ".", "", -1)
		increasedPort, err := strconv.Atoi(intendedPort)
		if err != nil {
			fmt.Printf("error converting version %s to number", intendedPort)
			os.Exit(1)
		}
		conditionalPrint("building test: %s\n", label)
		err = buildTests("templates", "testdata", label, map[string]string{
			"DbVersion":       v,
			"DbFlavor":        getFlavor(v),
			"DbPathVer":       label,
			"Home":            os.Getenv("HOME"),
			"TmpDir":          "/tmp",
			"DbIncreasedPort": fmt.Sprintf("%d", increasedPort+101),
		})
		if err != nil {
			fmt.Printf("error creating the tests for %s :%s\n", label, err)
			os.Exit(1)
		}
	}
	conditionalPrint("TestMain: starting tests\n")
	exitCode := testscript.RunMain(m, map[string]func() int{
		"dbdeployer": cmd.Execute,
	})

	if common.DirExists("testdata") && !dryRun {
		_ = os.RemoveAll("testdata")
	}
	os.Exit(exitCode)
}

var deltaPort = 0

// buildTests takes all the files from templateDir and populates several data directories
// Each directory is named with the combination of the bare name of the template file + the label
// for example, from the data directory "testdata", file "single.tmpl", and label "8_0_29" we get the file
// "single_8_0_29.txtar" under "testdata/8_0_29"
func buildTests(templateDir, dataDir, label string, data map[string]string) error {

	var templateNameToFeature = map[string]string{
		"single":                    "",
		"single-skip-start":         "",
		"single-custom-credentials": "",
		"replication":               "",
		"multiple":                  "",
		"circular-replication":      common.CircularReplication,
		"use-admin":                 common.AdminAddress,
		"dd-expose-tables":          common.DataDict,
		"replication-gtid":          common.GTID,
		"group":                     common.GroupReplication,
		"group_sp":                  common.GroupReplication,
		"semisync":                  common.SemiSynch,
		"fan-in":                    common.MultiSource,
		"all-masters":               common.MultiSource,
	}
	for _, needed := range []string{"DbVersion", "DbFlavor", "DbPathVer", "Home", "TmpDir"} {
		neededTxt, ok := data[needed]
		if !ok {
			return fmt.Errorf("[buildTests] the data must contain a '%s' element", needed)
		}
		if neededTxt == "" {
			return fmt.Errorf("[buildTests] the element '%s' in data is empty", needed)
		}
	}

	homeDir := data["Home"]
	if !common.DirExists(homeDir) {
		return fmt.Errorf("[buildTests] home directory '%s' not found", homeDir)
	}

	tmpDir := data["TmpDir"]
	if !common.DirExists(tmpDir) {
		return fmt.Errorf("[buildTests] temp directory '%s' not found", tmpDir)
	}

	if !common.DirExists(dataDir) {
		err := os.Mkdir(dataDir, 0755)
		if err != nil {
			return fmt.Errorf("[buildTests] error creating directory %s: %s", dataDir, err)
		}
	}
	if !common.DirExists(dataDir) {
		return fmt.Errorf("datadir %s not found after creation", dataDir)
	}
	files, err := filepath.Glob(templateDir + "/*/*.tmpl")

	if err != nil {
		return fmt.Errorf("[buildTests] error retrieving template files: %s", err)
	}

	for _, f := range files {
		dirName := common.DirName(f)
		sectionName := common.BaseName(dirName)

		//fmt.Printf("File %s (%s)\n", f, sectionName)
		fName := strings.Replace(path.Base(f), ".tmpl", "", 1)

		feature, known := templateNameToFeature[fName]
		if !known {
			return fmt.Errorf("file %s.tmpl has no recognized feature", fName)
		}
		var valid = false
		if feature == "" {
			valid = true
		} else {
			valid, err = common.HasCapability(data["DbFlavor"], feature, data["DbVersion"])
			if err != nil {
				return fmt.Errorf("error determining the validity of feature %s for version %s", feature, data["DbVersion"])
			}
		}
		if !valid {
			conditionalPrint("skipping file %s: feature %s is not available for version %s", fName, feature, data["DbVersion"])
			continue
		}
		conditionalPrint("processing file %s\n", fName)
		contents, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("[buildTests] error reading file %s: %s", f, err)
		}

		subDataDir := path.Join(dataDir, sectionName)
		if !common.DirExists(subDataDir) {
			err := os.Mkdir(subDataDir, 0755)
			if err != nil {
				return fmt.Errorf("[buildTests] error creating directory %s: %s", subDataDir, err)
			}
		}
		endDataDir := path.Join(subDataDir, label)
		if !common.DirExists(endDataDir) {
			err := os.Mkdir(endDataDir, 0755)
			if err != nil {
				return fmt.Errorf("[buildTests] error creating directory %s: %s", endDataDir, err)
			}
		}
		processTemplate := template.Must(template.New(label).Parse(string(contents)))
		buf := &bytes.Buffer{}

		dbPort, ok := data["DbIncreasedPort"]
		if ok {
			// make sure that the DbIncreasedPort is unique among the scripts
			port, _ := strconv.Atoi(dbPort)
			deltaPort++
			data["DbIncreasedPort"] = fmt.Sprintf("%d", port+deltaPort)
		}
		if err := processTemplate.Execute(buf, data); err != nil {
			return fmt.Errorf("[buildTests] error processing template from %s: %s", f, err)
		}
		versionFile := path.Join(endDataDir, "DB_VERSION")
		if !common.FileExists(versionFile) {
			err = os.WriteFile(versionFile, []byte(data["DbVersion"]), 0644)
			if err != nil {
				return fmt.Errorf("[buildTests] error writing version file %s: %s", versionFile, err)
			}
		}
		flavorFile := path.Join(endDataDir, "DB_FLAVOR")
		if !common.FileExists(flavorFile) {
			err = os.WriteFile(flavorFile, []byte(data["DbFlavor"]), 0644)
			if err != nil {
				return fmt.Errorf("[buildTests] error writing flavor file %s: %s", flavorFile, err)
			}
		}
		testName := path.Join(endDataDir, fName+"_"+label+".txtar")
		err = os.WriteFile(testName, buf.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("[buildTests] error writing text file %s: %s", testName, err)
		}
		if !common.FileExists(testName) {
			return fmt.Errorf("file %s not found after creation", testName)
		}
	}
	return nil
}

func conditionalPrint(format string, args ...interface{}) {
	if testDebug || os.Getenv("TEST_DEBUG") != "" {
		log.Printf(format, args...)
	}
}
