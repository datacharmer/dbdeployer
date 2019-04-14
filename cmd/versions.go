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

package cmd

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/globals"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/spf13/cobra"
)

func listVersions(dirs []string, basedir, flavor string, iteration int) {
	maxWidth := 80
	maxLen := 0
	for _, dir := range dirs {
		if len(dir) > maxLen {
			maxLen = len(dir)
		}
	}
	var header string

	if basedir != "" && iteration == 0 {
		header = fmt.Sprintf("Basedir: %s\n", basedir)
	}
	if flavor != "" {
		header += fmt.Sprintf("(Flavor: %s)\n", flavor)
	}
	if header != "" {
		fmt.Printf("%s", header)
	}
	columns := int(maxWidth / (maxLen + 2))
	mask := fmt.Sprintf("%%-%ds", maxLen+2)
	count := 0
	for _, dir := range dirs {
		fmt.Printf(mask, dir)
		count += 1
		if count > columns {
			count = 0
			fmt.Println("")
		}
	}
	fmt.Println("")
}

// Shows the MySQL versions available in $SANDBOX_BINARY
// (default $HOME/opt/mysql)
func showVersions(cmd *cobra.Command, args []string) {
	var err error
	basedir, err := getAbsolutePathFromFlag(cmd, "sandbox-binary")
	common.ErrCheckExitf(err, 1, "error getting absolute path for 'sandbox-binary'")
	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	byFlavor, _ := cmd.Flags().GetBool(globals.ByFlavorLabel)

	var versionInfoList []common.VersionInfo
	var dirs []string
	var flavoredLists = make(map[string][]string)

	versionInfoList = common.GetVersionInfoFromDir(basedir)
	if byFlavor {
		for _, verInfo := range versionInfoList {
			flavoredLists[verInfo.Flavor] = append(flavoredLists[verInfo.Flavor], verInfo.Version)
		}
		count := 0
		for f, versions := range flavoredLists {
			listVersions(versions, basedir, f, count)
			count++
			fmt.Println("")
		}
	} else {
		for _, verInfo := range versionInfoList {
			if flavor == verInfo.Flavor || flavor == "" {
				dirs = append(dirs, verInfo.Version)
			}
		}
		listVersions(dirs, basedir, flavor, 0)
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
