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
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/spf13/cobra"
)

func sandboxesUsingBinariesDir(basedir, binariesDir string) []string {
	var sandboxes []string
	var sandboxList defaults.SandboxCatalog
	var err error
	sandboxList, err = defaults.ReadCatalog()
	common.ErrCheckExitf(err, 1, "error getting sandboxes from catalog: %s", err)
	fullPath := path.Join(basedir, binariesDir)
	for _, sb := range sandboxList {
		if sb.Origin == fullPath {
			sandboxes = append(sandboxes, sb.Destination)
		}
	}
	return sandboxes
}

func deleteBinaries(basedir, binariesDir string, confirm bool) (deleted bool, err error) {
	fullPath := path.Join(basedir, binariesDir)
	if !common.DirExists(fullPath) {
		common.Exitf(1, globals.ErrDirectoryNotFound, fullPath)
	}

	sandboxesUsingBinaries := sandboxesUsingBinariesDir(basedir, binariesDir)
	if len(sandboxesUsingBinaries) > 0 {
		return false, fmt.Errorf("\nbinaries directory %s is used by the following deployments:\n%s",
			fullPath, strings.Join(sandboxesUsingBinaries, "\n"))
	}
	if confirm {
		common.CondPrintf("Do you want to delete %s? y/[N] ", binariesDir)
		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()
		if err != nil {
			fmt.Println(err)
		} else {
			answer := string(line)
			if answer == "y" || answer == "Y" {
				fmt.Println("Proceeding with deletion")
			} else {
				fmt.Println("Deletion skipped at user request")
				return false, nil
			}
		}
	}
	_, err = common.RunCmdWithArgs("rm", []string{"-rf", fullPath})
	if err != nil {
		return false, fmt.Errorf("error removing %s", fullPath)
	}
	if common.DirExists(fullPath) {
		return false, fmt.Errorf("directory %s was not removed - Reason unknown", fullPath)
	}
	return true, nil
}

func runDeleteBinaries(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1,
			"binaries directory name required.",
			"You can run 'dbdeployer versions for a list of available binaries'")
	}
	flags := cmd.Flags()
	binariesDir := args[0]
	skipConfirm, _ := flags.GetBool(globals.SkipConfirmLabel)

	basedir, err := getAbsolutePathFromFlag(cmd, "sandbox-binary")
	common.ErrCheckExitf(err, 1, "error finding absolute path for 'sandbox-binary'")

	isDeleted, err := deleteBinaries(basedir, binariesDir, !skipConfirm)
	if !isDeleted {
		common.ErrCheckExitf(err, 1, "%s", err)
		return
	}
	fmt.Printf("Directory %s removed\n", binariesDir)
}

var deleteBinariesCmd = &cobra.Command{
	Use:   "delete-binaries binaries_dir_name",
	Short: "delete an expanded tarball",
	Example: `
	$ dbdeployer delete-binaries 8.0.4
	$ dbdeployer delete ps5.7.25`,
	Long: `Removes the given directory and all its subdirectories.
It will fail if the directory is still used by any sandbox.
Warning: this command is irreversible!`,
	Run:         runDeleteBinaries,
	Annotations: map[string]string{"export": makeExportArgs(globals.ExportVersionDir, 1)},
}

func init() {
	rootCmd.AddCommand(deleteBinariesCmd)
	deleteBinariesCmd.Flags().BoolP(globals.SkipConfirmLabel, "", false, "Skips confirmation.")
}
