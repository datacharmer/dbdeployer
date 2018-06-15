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
	"bufio"
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
	"os"
)



func DeleteSandbox(cmd *cobra.Command, args []string) {
	var exec_lists []concurrent.ExecutionList
	if len(args) < 1 {
		common.Exit(1, 
			"Sandbox name (or \"ALL\") required.",
			"You can run 'dbdeployer sandboxes for a list of available deployments'")
	}
	flags := cmd.Flags()
	sandbox_name := args[0]
	confirm, _ := flags.GetBool("confirm")
	run_concurrently, _ := flags.GetBool("concurrent")
	if os.Getenv("RUN_CONCURRENTLY") != "" {
		run_concurrently = true
	}
	skip_confirm, _ := flags.GetBool("skip-confirm")
	sandbox_dir := GetAbsolutePathFromFlag(cmd, "sandbox-home")

	deletion_list := []common.SandboxInfo{common.SandboxInfo{sandbox_name, false}}
	if sandbox_name == "ALL" || sandbox_name == "all" {
		confirm = true
		if skip_confirm {
			confirm = false
		}
		deletion_list = common.GetInstalledSandboxes(sandbox_dir)
	}
	if len(deletion_list) == 0 {
		fmt.Printf("Nothing to delete in %s\n", sandbox_dir)
		return
	}
	if len(deletion_list) > 60 && run_concurrently {
		fmt.Println("# Concurrency disabled. Can't run more than 60 concurrent operations")
		run_concurrently = false
	}
	fmt.Printf("List of deployed sandboxes:\n")
	unlocked_found := false
	for _, sb := range deletion_list {
		locked := ""
		if sb.Locked {
			locked = "(*LOCKED*)"
		} else {
			unlocked_found = true
		}
		fmt.Printf("%s/%s %s\n", sandbox_dir, sb.SandboxName, locked)
	}
	 if !unlocked_found {
	 	fmt.Printf("No unlocked sandboxes found.\n")
		return
	 }
	if confirm {
		fmt.Printf("Do you confirm? y/[N] ")

		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()
		if err != nil {
			fmt.Println(err)
		} else {
			answer := string(line)
			if answer == "y" || answer == "Y" {
				fmt.Println("Proceeding with deletion")
			} else {
				common.Exit(0, "Execution interrupted by user")
			}
		}
	}
	for _, sb := range deletion_list {
		if sb.Locked {
			fmt.Printf("Sandbox %s is locked\n",sb.SandboxName)
		} else {
			exec_list := sandbox.RemoveSandbox(sandbox_dir, sb.SandboxName, run_concurrently)
			for _, list := range exec_list {
				exec_lists = append(exec_lists, list)
			}
		}
	}
	concurrent.RunParallelTasksByPriority(exec_lists)
	for _, sb := range deletion_list {
		full_path := sandbox_dir + "/" + sb.SandboxName
		if !sb.Locked {
			defaults.DeleteFromCatalog(full_path)
		}
	}
}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete sandbox_name (or \"ALL\")",
	Short:   "delete an installed sandbox",
	Aliases: []string{"remove", "destroy"},
	Example: `
	$ dbdeployer delete msb_8_0_4
	$ dbdeployer delete rsandbox_5_7_21`,
	Long: `Stops the sandbox (and its depending sandboxes, if any), and removes it.
Warning: this command is irreversible!`,
	Run: DeleteSandbox,
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolP("skip-confirm", "", false, "Skips confirmation with multiple deletions.")
	deleteCmd.Flags().BoolP("confirm", "", false, "Requires confirmation.")
	deleteCmd.Flags().BoolP("concurrent", "", false, "Runs multiple deletion tasks concurrently.")
}
