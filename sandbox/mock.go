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
	"os"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"runtime"
)

// Code in this module creates a fake directory structure that allows
// the testing of sandboxes without having MySQL packages.

const defaultMockDir string = "mock_dir"

var (
	saveHome          string
	saveSandboxHome   string
	saveSandboxBinary string
	saveSleepTime     string
	mockSandboxBinary string
	mockSandboxHome   string
)

func setMockEnvironment(mockUpperDir string) {
	if mockUpperDir == "" {
		mockUpperDir = defaultMockDir
	}
	if common.DirExists(mockUpperDir) {
		common.Exitf(1, "Mock directory %s already exists. Aborting", mockUpperDir)
	}
	PWD := os.Getenv("PWD")
	home := fmt.Sprintf("%s/%s/home", PWD, mockUpperDir)
	sandboxBinaryUpper := fmt.Sprintf("%s/opt", home)
	mockSandboxBinary = fmt.Sprintf("%s/opt/mysql", home)
	mockSandboxHome = fmt.Sprintf("%s/sandboxes", home)
	common.Mkdir(mockUpperDir)
	common.Mkdir(home)
	common.Mkdir(sandboxBinaryUpper)
	common.Mkdir(mockSandboxBinary)
	common.Mkdir(mockSandboxHome)
	os.Setenv("HOME", home)
	os.Setenv("SANDBOX_HOME", mockSandboxHome)
	os.Setenv("SANDBOX_BINARY", mockSandboxBinary)
	os.Setenv("HOME", home)
	os.Setenv("SLEEP_TIME", "0")
	defaults.ConfigurationDir = home + ".dbdeployer"
	defaults.ConfigurationFile = home + ".dbdeployer/config.json"
	defaults.SandboxRegistry = home + ".dbdeployer/sandboxes.json"
	defaults.SandboxRegistryLock = home + ".dbdeployer/sandboxes.lock"
}

func removeMockEnvironment(mockUpperDir string) {
	if !common.DirExists(mockUpperDir) {
		common.Exitf(1, "Mock directory %s doesn't exist. Aborting", mockUpperDir)
	}
	os.RemoveAll(mockUpperDir)
	os.Setenv("HOME", saveHome)
	os.Setenv("SANDBOX_HOME", saveSandboxHome)
	os.Setenv("SANDBOX_BINARY", saveSandboxBinary)
	os.Setenv("SLEEP_TIME", saveSleepTime)
}

func createMockVersion(version string) {
	if mockSandboxBinary == "" {
		common.Exit(1, "Mock directory not set yet. - Call setMockEnvironment() first")
	}
	_, logger := defaults.NewLogger(common.LogDirName(), "mock")
	versionDir := fmt.Sprintf("%s/%s", mockSandboxBinary, version)
	common.Mkdir(versionDir)
	common.Mkdir(versionDir + "/bin")
	common.Mkdir(versionDir + "/scripts")
	common.Mkdir(versionDir + "/lib")
	var emptyData = common.StringMap{}
	currentOs := runtime.GOOS
	extension := ""
	switch currentOs {
	case "linux":
		extension = "so"
	case "darwin":
		extension = "dylib"
	default:
		common.Exitf(1, "unhandled operating system %s", currentOs)
	}
	libmysqlclientFileName := fmt.Sprintf("libmysqlclient.%s", extension)
	writeScript(logger, MockTemplates, "mysqld", "no_op_mock_template",
		versionDir+"/bin", emptyData, true)
	writeScript(logger, MockTemplates, "mysql", "no_op_mock_template",
		versionDir+"/bin", emptyData, true)
	writeScript(logger, MockTemplates, "mysql_install_db", "no_op_mock_template",
		versionDir+"/scripts", emptyData, true)
	writeScript(logger, MockTemplates, libmysqlclientFileName, "no_op_mock_template",
		versionDir+"/lib", emptyData, true)
	writeScript(logger, MockTemplates, "mysqld_safe", "mysqld_safe_mock_template",
		versionDir+"/bin", emptyData, true)
}

func init() {
	mockSandboxBinary = os.Getenv("SANDBOX_BINARY")
	if mockSandboxBinary != "" {
		saveSandboxBinary = mockSandboxBinary
	}
	mockSandboxHome = os.Getenv("SANDBOX_HOME")
	if mockSandboxHome != "" {
		saveSandboxHome = mockSandboxHome
	}
	home := os.Getenv("HOME")
	if home != "" {
		saveHome = home
	}
	sleepTime := os.Getenv("SLEEP_TIME")
	if sleepTime != "" {
		saveSleepTime = sleepTime
	}
}
