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
// to create two MySQL sandboxes using dbdeployer code
// from another Go program.
package main

import (
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
	"os"
	"path"
)

func main() {
	// Searches for expanded sandboxes in $HOME/opt/mysql
	sandbox_binary := path.Join(os.Getenv("HOME"), "opt", "mysql")

	// Creates sandboxes in $HOME/sandboxes
	sandbox_home := path.Join(os.Getenv("HOME"), "sandboxes")

	// For this to work, we need to have
	// a MySQL tarball expanded in $HOME/opt/mysql/5.7.22
	version1 := "5.7.22"
	version2 := "5.6.33"

	sandboxName1 := "msb_5_7_22"
	sandboxName2 := "msb_5_6_33"

	// The unique ports for these sandboxes
	port1 := 5722
	port2 := 5633

	// MySQL will look for binaries in $HOME/opt/mysql/5.7.22
	basedir1 := path.Join(sandbox_binary, version1) // This is what dbdeployer expects
	// i.e. a name containing the full version

	// MySQL will look for binaries in $HOME/opt/mysql/my-5.6
	basedir2 := path.Join(sandbox_binary, "my-5.6")
	basedir2a := path.Join(sandbox_binary, "5.6.33")
	if !common.DirExists(basedir2) {
		if common.DirExists(basedir2a) {
			err := os.Symlink(basedir2a, basedir2)
			if err != nil {
				common.Exitf(1, "error creating a symlink between %s and %s", basedir2a, basedir2)
			}
		} else {
			common.Exitf(1, defaults.ErrDirectoryNotFound, basedir2)
		}
	}
	// This is a deviation from dbdeployer
	// paradigm, using a non-standard name
	// for the base directory

	// Creates the base target directory if it doesn't exist
	if !common.DirExists(sandbox_home) {
		common.Mkdir(sandbox_home)
	}

	// Minimum data to be filled for a simple sandbox.
	// See sandbox/sandbox.go for the full description
	// of this data structure
	var sdef = sandbox.SandboxDef{
		Version:    version1,
		Basedir:    basedir1,
		SandboxDir: sandbox_home,
		DirName:    sandboxName1,
		LoadGrants: true,
		// This is the list of ports to ignore
		// when checking if the designated port is
		// used or not.
		// Try changing the Port item to 3306.
		// You will see that the sandbox will install using 3307
		InstalledPorts: []int{1186, 3306, 33060},
		Port:           port1,
		DbUser:         defaults.DbUserValue,      // "msandbox"
		DbPassword:     defaults.DbPasswordValue,  // "msandbox"
		RplUser:        defaults.RplUserValue,     // "rsandbox"
		RplPassword:    defaults.RplPasswordValue, // "rsandbox"
		RemoteAccess:   "127.%",
		BindAddress:    "127.0.0.1",
	}

	// Calls the sandbox creation
	err := sandbox.CreateStandaloneSandbox(sdef)
	if err != nil {
		common.Exitf(1, defaults.ErrCreatingSandbox, err)
	}

	sdef.Version = version2
	sdef.Basedir = basedir2
	sdef.DirName = sandboxName2
	sdef.Port = port2

	// Calls the sandbox creation for the second sandbox
	err = sandbox.CreateStandaloneSandbox(sdef)
	if err != nil {
		common.Exitf(1, defaults.ErrCreatingSandbox, err)
	}

	// Invokes the sandbox self-testing script
	err, _ = common.RunCmd(path.Join(sandbox_home, sandboxName1, "test_sb"))
	if err != nil {
		common.Exitf(1, "error running test for sandbox 2: %s", err)
	}
	err, _ = common.RunCmd(path.Join(sandbox_home, sandboxName2, "test_sb"))
	if err != nil {
		common.Exitf(1, "error running test for sandbox 2: %s", err)
	}

	// Removes the sandbox from disk
	err, _ = sandbox.RemoveSandbox(sandbox_home, sandboxName1, false)
	if err != nil {
		common.Exitf(1, defaults.ErrWhileDeletingSandbox, err)
	}
	err, _ = sandbox.RemoveSandbox(sandbox_home, sandboxName2, false)
	if err != nil {
		common.Exitf(1, defaults.ErrWhileDeletingSandbox, err)
	}

	// Removes the sandbox from dbdeployer catalog
	err = defaults.DeleteFromCatalog(path.Join(sandbox_home, sandboxName1))
	if err != nil {
		common.Exitf(1, defaults.ErrRemovingFromCatalog, sandboxName1)
	}
	err = defaults.DeleteFromCatalog(path.Join(sandbox_home, sandboxName2))
	if err != nil {
		common.Exitf(1, defaults.ErrRemovingFromCatalog, sandboxName2)
	}
}
