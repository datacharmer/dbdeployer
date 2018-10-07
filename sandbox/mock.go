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

const default_mock_dir string = "mock_dir"

var (
	save_home           string
	save_sandbox_home   string
	save_sandbox_binary string
	save_sleep_time     string
	mock_sandbox_binary string
	mock_sandbox_home   string
)

func set_mock_environment(mock_upper_dir string) {
	if mock_upper_dir == "" {
		mock_upper_dir = default_mock_dir
	}
	if common.DirExists(mock_upper_dir) {
		common.Exitf(1, "Mock directory %s already exists. Aborting", mock_upper_dir)
	}
	PWD := os.Getenv("PWD")
	home := fmt.Sprintf("%s/%s/home", PWD, mock_upper_dir)
	sandbox_binary_upper := fmt.Sprintf("%s/opt", home)
	mock_sandbox_binary = fmt.Sprintf("%s/opt/mysql", home)
	mock_sandbox_home = fmt.Sprintf("%s/sandboxes", home)
	common.Mkdir(mock_upper_dir)
	common.Mkdir(home)
	common.Mkdir(sandbox_binary_upper)
	common.Mkdir(mock_sandbox_binary)
	common.Mkdir(mock_sandbox_home)
	os.Setenv("HOME", home)
	os.Setenv("SANDBOX_HOME", mock_sandbox_home)
	os.Setenv("SANDBOX_BINARY", mock_sandbox_binary)
	os.Setenv("HOME", home)
	os.Setenv("SLEEP_TIME", "0")
	defaults.ConfigurationDir = home + ".dbdeployer"
	defaults.ConfigurationFile = home + ".dbdeployer/config.json"
	defaults.SandboxRegistry = home + ".dbdeployer/sandboxes.json"
	defaults.SandboxRegistryLock = home + ".dbdeployer/sandboxes.lock"
}

func remove_mock_environment(mock_upper_dir string) {
	if !common.DirExists(mock_upper_dir) {
		common.Exitf(1, "Mock directory %s doesn't exist. Aborting", mock_upper_dir)
	}
	os.RemoveAll(mock_upper_dir)
	os.Setenv("HOME", save_home)
	os.Setenv("SANDBOX_HOME", save_sandbox_home)
	os.Setenv("SANDBOX_BINARY", save_sandbox_binary)
	os.Setenv("SLEEP_TIME", save_sleep_time)
}

func create_mock_version(version string) {
	if mock_sandbox_binary == "" {
		common.Exit(1, "Mock directory not set yet. - Call set_mock_environment() first")
	}
	_, logger := defaults.NewLogger(common.LogDirName(), "mock")
	version_dir := fmt.Sprintf("%s/%s", mock_sandbox_binary, version)
	common.Mkdir(version_dir)
	common.Mkdir(version_dir + "/bin")
	common.Mkdir(version_dir + "/scripts")
	common.Mkdir(version_dir + "/lib")
	var empty_data = common.StringMap{}
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
	libmysqlclient_file_name := fmt.Sprintf("libmysqlclient.%s", extension)
	write_script(logger, MockTemplates, "mysqld", "no_op_mock_template",
		version_dir+"/bin", empty_data, true)
	write_script(logger, MockTemplates, "mysql", "no_op_mock_template",
		version_dir+"/bin", empty_data, true)
	write_script(logger, MockTemplates, "mysql_install_db", "no_op_mock_template",
		version_dir+"/scripts", empty_data, true)
	write_script(logger, MockTemplates, libmysqlclient_file_name, "no_op_mock_template",
		version_dir+"/lib", empty_data, true)
	write_script(logger, MockTemplates, "mysqld_safe", "mysqld_safe_mock_template",
		version_dir+"/bin", empty_data, true)
}

func init() {
	mock_sandbox_binary = os.Getenv("SANDBOX_BINARY")
	if mock_sandbox_binary != "" {
		save_sandbox_binary = mock_sandbox_binary
	}
	mock_sandbox_home = os.Getenv("SANDBOX_HOME")
	if mock_sandbox_home != "" {
		save_sandbox_home = mock_sandbox_home
	}
	home := os.Getenv("HOME")
	if home != "" {
		save_home = home
	}
	sleep_time := os.Getenv("SLEEP_TIME")
	if sleep_time != "" {
		save_sleep_time = sleep_time
	}
}
