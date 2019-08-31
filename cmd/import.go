// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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

package cmd

import (
	"fmt"
	"path"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/importing"
	"github.com/datacharmer/dbdeployer/sandbox"
)

func importSingleSandbox(cmd *cobra.Command, args []string) {
	// args will be at least 4, as ensured by MinimumNArgs
	host := args[0]
	strPort := args[1]
	user := args[2]
	password := args[3]

	port := common.Atoi(strPort)
	var config = importing.ParamsToConfig(host, user, password, port)

	db, err := importing.Connect(config)
	if err != nil {
		common.Exitf(1, "error connecting to server %s:%d - %s", host, port, err)
	}
	var versionString string

	err = db.GetSingleResult(config, "SELECT version()", &versionString)
	if err != nil {
		common.Exitf(1, "error getting version from server %s:%d - %s", host, port, err)
	}
	if versionString == "" {
		common.Exitf(1, "empty version returned from server %s:%d", host, port)
	}

	reVersion := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	versionList := reVersion.FindAllString(versionString, -1)
	if len(versionList) > 0 {
		versionString = versionList[0]
	}

	fmt.Printf("detected: %s\n", versionString)
	var sd sandbox.SandboxDef
	sd, err = fillSandboxDefinition(cmd, []string{versionString}, true)
	if err != nil {
		common.Exitf(1, "error while filling the sandbox definition: %+v", err)
	}
	// When deploying a single sandbox, we disable concurrency
	sd.RunConcurrently = false
	sd.SbHost = host
	sd.Port = port
	sd.DbUser = user
	sd.DbPassword = password
	sd.SBType = globals.SbTypeSingleImported
	sd.Imported = true

	if sd.ClientBasedir == "" {
		clientVersion, err := common.GetCompatibleClientVersion(defaults.Defaults().SandboxBinary, versionString)
		if err != nil {
			common.Exitf(1,
				"no suitable client version found - use --'%s' to designate one : %s ", globals.ClientFromLabel, err)
		}
		sd.ClientBasedir = path.Join(defaults.Defaults().SandboxBinary, clientVersion)
	}
	// Removes imported server port from list of installed ones
	var newPortList []int
	for _, uPort := range defaults.Defaults().ReservedPorts {
		if uPort != port {
			newPortList = append(newPortList, uPort)
		}
	}
	sd.InstalledPorts = newPortList
	sd.LoadGrants = false
	sd.SkipStart = true
	sd.RplUser = user
	sd.RplPassword = password

	err = sandbox.CreateStandaloneSandbox(sd)
	if err != nil {
		common.Exitf(1, globals.ErrCreatingSandbox, err)
	}
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "imports one or more MySQL servers into a sandbox",
	Long:  ``,
}

// TODO:
// Detect which client to use from remote version
// add host to `sandboxes` output
// add host and import type to catalog and sandbox description
// check that we are not importing from a sandbox already in this host
// Make the whole import procedure a library function
// Add more sandbox creation options

var importSingleCmd = &cobra.Command{
	Use:   "single host port user password",
	Short: "imports a MySQL server into a sandbox",
	Args:  cobra.MinimumNArgs(4),
	Long: `Imports an existing (local or remote) server into a sandbox,
so that it can be used with the usual sandbox scripts.
Requires host, port, user, password.
`,
	Run: importSingleSandbox,
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.AddCommand(importSingleCmd)
	setPflag(importCmd, globals.ClientFromLabel, "", "", "", "Where to get the client binaries from", false)
}
