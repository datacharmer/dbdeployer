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
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/downloads"
	"github.com/datacharmer/dbdeployer/ops"
	"github.com/rogpeppe/go-internal/testscript"
)

var (
	dryRun     bool = os.Getenv("DRY_RUN") != ""
	testDebug  bool = os.Getenv("TEST_DEBUG") != ""
	onlyLatest bool = os.Getenv("ONLY_LATEST") != ""
)

func dbdeployerSetup(t *testing.T, dir string) func(env *testscript.Env) error {
	return func(env *testscript.Env) error {
		readFile := func(fileName string) (string, error) {
			wantedFile := path.Join(dir, fileName)
			if !common.FileExists(wantedFile) {
				return "", fmt.Errorf("no %s file found in %s", fileName, dir)
			}
			text, err := os.ReadFile(wantedFile) // #nosec G304
			if err != nil {
				return "", fmt.Errorf("error reading file %s: %s", wantedFile, err)
			}
			if len(text) == 0 {
				return "", fmt.Errorf("file %s was empty", wantedFile)
			}
			return string(text), nil
		}
		versionText, err := readFile("DB_VERSION")
		if err != nil {
			return fmt.Errorf("error reading version file in %s", dir)
		}
		flavorText, err := readFile("DB_FLAVOR")
		if err != nil {
			return fmt.Errorf("error reading flavor file in %s", dir)
		}
		env.Setenv("db_version", versionText)
		env.Setenv("db_flavor", flavorText)

		env.Values["db_version"] = versionText
		env.Values["db_flavor"] = flavorText
		env.Values["testingT"] = t

		return nil
	}
}

func preliminaryChecks() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error determining current directory\n")
		os.Exit(1)
	}
	if common.FileExists(defaults.SandboxRegistry) {
		sc, err := defaults.ReadCatalog()
		if err != nil {
			fmt.Printf("Error getting information on sandboxes file %s\n", defaults.SandboxRegistry)
			os.Exit(1)
		}
		if len(sc) > 0 {
			fmt.Printf("sandboxes file %s should be empty but it contains sandbox information\n", defaults.SandboxRegistry)
			os.Exit(1)
		}
	}
	if common.FileExists(downloads.TarballFileRegistry) {
		fmt.Printf("downloads file list %s found. Tests may fail for non-standard configuration\n", downloads.TarballFileRegistry)
		fmt.Println("run 'dbdeployer downloads reset' to remove it")
		os.Exit(1)
	}

	if common.FileExists(defaults.ConfigurationFile) {
		fmt.Printf("configuration file %s found. Tests may fail for non-standard configuration.\n", defaults.SandboxRegistry)
		fmt.Println("run 'dbdeployer defaults reset' to remove it")
		os.Exit(1)
	}
	sandboxHome := defaults.Defaults().SandboxHome
	filesInSandboxHome, err := filepath.Glob(path.Join(sandboxHome, "*"))
	if err != nil {
		fmt.Printf("Error getting information on sandboxes: %s\n", err)
		os.Exit(1)
	}
	if len(filesInSandboxHome) > 0 {
		fmt.Printf("found files in $SANDBOX_HOME (%s) - The test needs a clean directory\n", sandboxHome)
		fmt.Printf("%v\n", filesInSandboxHome)
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
	flavor, err := os.ReadFile(filePath) // #nosec G304
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
		"data-load-single":          common.EnhancedGTID,
		"circular-replication":      common.CircularReplication,
		"use-admin":                 common.AdminAddress,
		"dd-expose-tables":          common.DataDict,
		"replication-gtid":          common.EnhancedGTID,
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
		err := os.Mkdir(dataDir, 0750)
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
		contents, err := os.ReadFile(f) // #nosec G304
		if err != nil {
			return fmt.Errorf("[buildTests] error reading file %s: %s", f, err)
		}

		subDataDir := path.Join(dataDir, sectionName)
		if !common.DirExists(subDataDir) {
			err := os.Mkdir(subDataDir, 0750)
			if err != nil {
				return fmt.Errorf("[buildTests] error creating directory %s: %s", subDataDir, err)
			}
		}
		endDataDir := path.Join(subDataDir, label)
		if !common.DirExists(endDataDir) {
			err := os.Mkdir(endDataDir, 0750)
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
			err = os.WriteFile(versionFile, []byte(data["DbVersion"]), 0600)
			if err != nil {
				return fmt.Errorf("[buildTests] error writing version file %s: %s", versionFile, err)
			}
		}
		flavorFile := path.Join(endDataDir, "DB_FLAVOR")
		if !common.FileExists(flavorFile) {
			err = os.WriteFile(flavorFile, []byte(data["DbFlavor"]), 0600)
			if err != nil {
				return fmt.Errorf("[buildTests] error writing flavor file %s: %s", flavorFile, err)
			}
		}
		testName := path.Join(endDataDir, fName+"_"+label+".txtar")
		err = os.WriteFile(testName, buf.Bytes(), 0600)
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
