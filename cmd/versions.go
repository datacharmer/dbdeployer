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

package cmd

import (
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/ops"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/spf13/cobra"
)

// Shows the MySQL versions available in $SANDBOX_BINARY
// (default $HOME/opt/mysql)
func showVersions(cmd *cobra.Command, args []string) {
	var err error
	basedir, err := getAbsolutePathFromFlag(cmd, "sandbox-binary")
	common.ErrCheckExitf(err, 1, "error getting absolute path for 'sandbox-binary'")
	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	byFlavor, _ := cmd.Flags().GetBool(globals.ByFlavorLabel)

	err = ops.ShowVersions(ops.VersionOptions{
		SandboxBinary: basedir,
		Flavor:        flavor,
		ByFlavor:      byFlavor,
	})
	if err != nil {
		common.Exitf(1, "error showing versions: %s", err)
	}
}

// versionsCmd represents the versions command
var versionsCmd = &cobra.Command{
	Use:     "versions",
	Aliases: []string{"available"},
	Short:   "List available versions",
	Long:    ``,
	Run:     showVersions,
}

func init() {
	rootCmd.AddCommand(versionsCmd)
	setPflag(versionsCmd, globals.FlavorLabel, "", "", "", "Get only versions of the given flavor", false)
	versionsCmd.Flags().BoolP(globals.ByFlavorLabel, "", false, "Shows versions list by flavor")
}
