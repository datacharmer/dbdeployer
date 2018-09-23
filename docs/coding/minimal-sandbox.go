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

// This is a sample source file that shows how
// to create a MySQL sandbox using dbdeployer code
// from another Go program.
package main

import (
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
	"os"
)

func main() {
	// Searches for expanded sandboxes in $HOME/opt/mysql
	sandbox_binary := os.Getenv("HOME") + "/opt/mysql"

	// For this to work, we need to have
	// a MySQL tarball expanded in $HOME/opt/mysql/5.7.22
	version := "5.7.22"

	// Creates sandboxes in $HOME/sandboxes
	sandbox_home := os.Getenv("HOME") + "/sandboxes"

	// MySQL will look for binaries in $HOME/opt/mysql/5.7.22
	basedir := sandbox_binary + "/" + version

	// The unique port for this sandbox
	port := 5722

	// Username and password for this sandbox
	user := "msandbox"
	password := "msandbox"

	if !common.DirExists(sandbox_home) {
		common.Mkdir(sandbox_home)
	}

	// Minimum data to be filled for a simple sandbox.
	// See sandbox/sandbox.go for the full description
	// of this data structure
	var sdef = sandbox.SandboxDef{
		Version:        version,
		Basedir:        basedir,
		SandboxDir:     sandbox_home,
		DirName:        "msb_5_7_22",
		LoadGrants:     true,
		InstalledPorts: []int{1186, 3306, 33060},
		Port:           port,
		DbUser:         user,
		DbPassword:     password,
		RplUser:        "r" + user,
		RplPassword:    "r" + password,
		RemoteAccess:   "127.%",
		BindAddress:    "127.0.0.1",
	}

	// Calls the sandbox creation
	sandbox.CreateSingleSandbox(sdef)

	// Invokes the sandbox self-testing script
	common.Run_cmd(sandbox_home + "/msb_5_7_22/test_sb")

	// Removes the sandbox from disk
	sandbox.RemoveSandbox(sandbox_home, "msb_5_7_22", false)

	// Removes the sandbox from dbdeployer catalog
	defaults.DeleteFromCatalog(sandbox_home + "/msb_5_7_22")
}
