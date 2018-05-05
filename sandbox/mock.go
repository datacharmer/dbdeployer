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
)

// Code in this module creates a fake directory structure that allows
// the testing of sandboxes without having MySQL packages.

const default_mock_dir string = "mock_dir"
var (
	save_home string 
	save_sandbox_home string
	save_sandbox_binary string
	save_sleep_time string
	sandbox_binary string
	sandbox_home string
)

func set_mock_environment(mock_upper_dir string) {
	if mock_upper_dir == "" {
		mock_upper_dir = default_mock_dir
	}
	if common.DirExists(mock_upper_dir) {
		common.Exit(1, fmt.Sprintf("Mock directory %s already exists. Aborting",mock_upper_dir))
	}
	PWD := os.Getenv("PWD")
	home := fmt.Sprintf("%s/%s/home", PWD, mock_upper_dir)
	sandbox_binary_upper := fmt.Sprintf("%s/opt", home)
	sandbox_binary = fmt.Sprintf("%s/opt/mysql", home)
	sandbox_home = fmt.Sprintf("%s/sandboxes", home)
	common.Mkdir(mock_upper_dir)
	common.Mkdir(home)
	common.Mkdir(sandbox_binary_upper)
	common.Mkdir(sandbox_binary)
	common.Mkdir(sandbox_home)
	os.Setenv("HOME", home)
	os.Setenv("SANDBOX_HOME", sandbox_home)
	os.Setenv("SANDBOX_BINARY", sandbox_binary)
	os.Setenv("HOME", home)
	os.Setenv("SLEEP_TIME", "0")
	defaults.ConfigurationDir = home + ".dbdeployer"
	defaults.ConfigurationFile = home + ".dbdeployer/config.json"
	defaults.SandboxRegistry = home + ".dbdeployer/sandboxes.json"
	defaults.SandboxRegistryLock = home + ".dbdeployer/sandboxes.lock"
}

func remove_mock_environment(mock_upper_dir string) {
	if !common.DirExists(mock_upper_dir) {
		common.Exit(1, fmt.Sprintf("Mock directory %s doesn't exist. Aborting",mock_upper_dir))
	}
	os.RemoveAll(mock_upper_dir)
	os.Setenv("HOME", save_home)
	os.Setenv("SANDBOX_HOME", save_sandbox_home)
	os.Setenv("SANDBOX_BINARY", save_sandbox_binary)
}

func create_mock_version(version string) {
	if sandbox_binary == "" {
		common.Exit(1, "Mock directory not set yet. - Call set_mock_environment() first")
	}
	version_dir := fmt.Sprintf("%s/%s", sandbox_binary, version)
	common.Mkdir(version_dir)
	common.Mkdir(version_dir+ "/bin")	
	common.Mkdir(version_dir+ "/scripts")
	write_script(MockTemplates, "mysqld", "no_op_mock_template", version_dir + "/bin", common.Smap{}, true) 
	write_script(MockTemplates, "mysql", "no_op_mock_template", version_dir + "/bin", common.Smap{}, true) 
	write_script(MockTemplates, "mysql_install_db", "no_op_mock_template", version_dir + "/scripts", common.Smap{}, true) 
	write_script(MockTemplates, "mysqld_safe", "mysqld_safe_mock_template", version_dir + "/bin", common.Smap{}, true) 
}

func init() {
	sandbox_binary = os.Getenv("SANDBOX_BINARY")
	if sandbox_binary != "" {
		save_sandbox_binary = sandbox_binary
	}
	sandbox_home = os.Getenv("SANDBOX_HOME")
	if sandbox_home != "" {
		save_sandbox_home = sandbox_home
	}
	home := os.Getenv("HOME")
	if home != "" {
		save_home = home
	}
}
