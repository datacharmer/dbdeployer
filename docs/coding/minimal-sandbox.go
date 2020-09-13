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
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/sandbox"
	"os"
	"path"
)

func main() {

	// Searches for expanded sandboxes in $HOME/opt/mysql
	sandboxBinary := path.Join(os.Getenv("HOME"), "opt", "mysql")

	// For this to work, we need to have
	// a MySQL tarball expanded in $HOME/opt/mysql/5.7.22
	version := "5.7.22"

	// Creates sandboxes in $HOME/sandboxes
	sandboxHome := path.Join(os.Getenv("HOME"), "sandboxes")

	// MySQL will look for binaries in $HOME/opt/mysql/5.7.22
	basedir := path.Join(sandboxBinary, version)

	// The unique port for this sandbox
	port := 5722

	sandboxName := "msb_5_7_22"

	if !common.DirExists(sandboxHome) {
		err := os.Mkdir(sandboxHome, globals.PublicDirectoryAttr)
		if err != nil {
			common.Exitf(1, globals.ErrCreatingDirectory, sandboxHome, err)
		}
	}

	// Minimum data to be filled for a simple sandbox.
	// See sandbox/sandbox.go for the full description
	// of this data structure
	var sandboxDef = sandbox.SandboxDef{
		Version:        version,
		Basedir:        basedir,
		SandboxDir:     sandboxHome,
		DirName:        sandboxName,
		LoadGrants:     true,
		InstalledPorts: []int{1186, 3306, 33060},
		Port:           port,
		DbUser:         globals.DbUserValue,      // "msandbox"
		DbPassword:     globals.DbPasswordValue,  // "msandbox"
		RplUser:        globals.RplUserValue,     // "rsandbox"
		RplPassword:    globals.RplPasswordValue, // "rsandbox"
		RemoteAccess:   "127.%",
		BindAddress:    "127.0.0.1",
	}

	// Calls the sandbox creation
	err := sandbox.CreateStandaloneSandbox(sandboxDef)
	if err != nil {
		common.Exitf(1, globals.ErrCreatingSandbox, err)
	}

	// Invokes the sandbox self-testing script
	_, err = common.RunCmdCtrl(path.Join(sandboxHome, "msb_5_7_22", "test_sb"), false)
	if err != nil {
		common.Exitf(1, "error executing sandbox test: %s", err)
	}

	// Removes the sandbox from disk
	_, err = sandbox.RemoveCustomSandbox(sandboxHome, "msb_5_7_22", false, false)
	if err != nil {
		common.Exitf(1, globals.ErrWhileDeletingSandbox, err)
	}

	// Removes the sandbox from dbdeployer catalog
	err = defaults.DeleteFromCatalog(path.Join(sandboxHome, "msb_5_7_22"))
	if err != nil {
		common.Exitf(1, globals.ErrRemovingFromCatalog, sandboxName)
	}
}
