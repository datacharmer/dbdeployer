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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"

	"github.com/datacharmer/dbdeployer/cmd"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/ops"

	"github.com/rogpeppe/go-internal/testscript"
)

var dryRun bool

func TestDbDeployer(t *testing.T) {
	t.Logf("entering TestDbDeployer")
	if dryRun {
		t.Skip("Dry Run")
	}
	if !common.DirExists("testdata") {
		t.Skip("no testdata found")
	}
	// Directories in testdata are created by the setup code in TestMain
	dirs, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Skip("no directories found in testdata")
	}
	t.Logf("Directories: %v\n", dirs)
	for _, dir := range dirs {
		t.Logf("entering TestDbDeployer/%s", dir)
		t.Run(path.Base(dir), func(t *testing.T) {
			testscript.Run(t, testscript.Params{
				Dir:       dir,
				Cmds:      customCommands(),
				Condition: customConditions,
				Setup:     dbdeployerSetup(t, dir),
			})
		})
	}
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
	flavor, err := ioutil.ReadFile(filePath)
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
		if latest != "" {
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
		if latest != "" {
			fmt.Printf("found latest %s: %s\n", v, latest)
			continue
		}
		err := ops.GetRemoteTarball(ops.DownloadsOptions{
			SandboxBinary:     sandboxBinary,
			TarballOS:         runtime.GOOS,
			Flavor:            common.MySQLFlavor,
			Version:           latest,
			Newest:            true,
			Minimal:           strings.EqualFold(runtime.GOOS, "linux"),
			Unpack:            true,
			DeleteAfterUnpack: true,
		})
		if err != nil {
			fmt.Printf("error getting tarball for version %s\n", latest)
			return err
		}
		fmt.Printf("retrieved tarball for version %s\n", latest)
	}
	return nil
}

func TestMain(m *testing.M) {
	flag.BoolVar(&dryRun, "dry", false, "creates testdata without running tests")

	//shortVersions := []string{"5.0", "5.1", "5.5", "5.6", "5.7", "8.0"}
	shortVersions := []string{"5.7", "8.0"}
	fmt.Printf("short versions: %v\n", shortVersions)
	err := initializeEnv(shortVersions)
	if err != nil {
		fmt.Printf("error initializing the environment - Skipping tests: %s\n", err)
		os.Exit(0)
	}
	versions := getVersionList(shortVersions)

	fmt.Printf("versions: %v\n", versions)
	for _, v := range versions {
		label := strings.Replace(v, ".", "_", -1)
		fmt.Printf("building test: %s\n", label)
		err := buildTests("templates", "testdata", label, map[string]string{
			"DbVersion": v,
			"DbFlavor":  getFlavor(v),
			"DbPathVer": label,
			"Home":      os.Getenv("HOME"),
			"TmpDir":    "/tmp",
		})
		if err != nil {
			fmt.Printf("error creating the tests for %s :%s\n", label, err)
			os.Exit(1)
		}
	}
	fmt.Printf("TestMain: starting tests\n")
	exitCode := testscript.RunMain(m, map[string]func() int{
		"dbdeployer": cmd.Execute,
	})

	if common.DirExists("testdata") && !dryRun {
		_ = os.RemoveAll("testdata")
	}
	os.Exit(exitCode)
}

// buildTests takes all the files from templateDir and populates several data directories
// Each directory is named with the combination of the bare name of the template file + the label
// for example, from the data directory "testdata", file "single.tmpl", and label "8_0_29" we get the file
// "single_8_0_29.txt" under "testdata/8_0_29"
func buildTests(templateDir, dataDir, label string, data map[string]string) error {

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
	files, err := filepath.Glob(templateDir + "/*.tmpl")

	if err != nil {
		return fmt.Errorf("[buildTests] error retrieving template files: %s", err)
	}
	for _, f := range files {
		fName := strings.Replace(path.Base(f), ".tmpl", "", 1)

		contents, err := ioutil.ReadFile(f)
		if err != nil {
			return fmt.Errorf("[buildTests] error reading file %s: %s", f, err)
		}

		subDataDir := path.Join(dataDir, label)
		if !common.DirExists(subDataDir) {
			err := os.Mkdir(subDataDir, 0755)
			if err != nil {
				return fmt.Errorf("[buildTests] error creating directory %s: %s", subDataDir, err)
			}
		}
		processTemplate := template.Must(template.New(label).Parse(string(contents)))
		buf := &bytes.Buffer{}

		if err := processTemplate.Execute(buf, data); err != nil {
			return fmt.Errorf("[buildTests] error processing template from %s: %s", f, err)
		}
		versionFile := path.Join(subDataDir, "DB_VERSION")
		if !common.FileExists(versionFile) {
			err = ioutil.WriteFile(versionFile, []byte(data["DbVersion"]), 0644)
			if err != nil {
				return fmt.Errorf("[buildTests] error writing version file %s: %s", versionFile, err)
			}
		}
		flavorFile := path.Join(subDataDir, "DB_FLAVOR")
		if !common.FileExists(flavorFile) {
			err = ioutil.WriteFile(flavorFile, []byte(data["DbFlavor"]), 0644)
			if err != nil {
				return fmt.Errorf("[buildTests] error writing flavor file %s: %s", flavorFile, err)
			}
		}
		testName := path.Join(subDataDir, fName+"_"+label+".txt")
		err = ioutil.WriteFile(testName, buf.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("[buildTests] error writing text file %s: %s", testName, err)
		}
	}
	return nil
}
