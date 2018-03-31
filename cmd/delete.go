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
	"github.com/spf13/cobra"
	"os"
)

func RemoveSandbox(sandbox_dir, sandbox string, run_concurrently bool) (exec_list []concurrent.ExecutionList) {
	full_path := sandbox_dir + "/" + sandbox
	if !common.DirExists(full_path) {
		fmt.Printf("Directory '%s' not found\n", full_path)
		os.Exit(1)
	}
	preserve := full_path + "/no_clear_all"
	if !common.ExecExists(preserve) {
		preserve = full_path + "/no_clear"
	}
	if common.ExecExists(preserve) {
		fmt.Printf("The sandbox %s is locked\n",sandbox)
		fmt.Printf("You need to unlock it with \"dbdeployer admin unlock\"\n",)
		return
	}
	stop := full_path + "/stop_all"
	if !common.ExecExists(stop) {
		stop = full_path + "/stop"
	}
	if !common.ExecExists(stop) {
		fmt.Printf("Executable '%s' not found\n", stop)
		os.Exit(1)
	}

	if run_concurrently {
		var eCommand1 = concurrent.ExecCommand{
			Cmd : stop,
			Args : []string{},
		}
		exec_list = append(exec_list, concurrent.ExecutionList{0, eCommand1})
	} else {
		fmt.Printf("Running %s\n", stop)
		err, _ := common.Run_cmd(stop)
		if err != nil {
			fmt.Printf("Error while stopping sandbox %s\n", full_path)
			os.Exit(1)
		}
	}

	cmd_str := "rm"
	rm_args := []string{"-rf", full_path}
	if run_concurrently {
		var eCommand2 = concurrent.ExecCommand{
			Cmd : cmd_str,
			Args : rm_args,
		}
		exec_list = append(exec_list, concurrent.ExecutionList{1, eCommand2})
	} else {
		for _, item := range rm_args {
			cmd_str += " " + item
		}
		fmt.Printf("Running %s\n", cmd_str)
		err, _ := common.Run_cmd_with_args("rm", rm_args)
		if err != nil {
			fmt.Printf("Error while deleting sandbox %s\n", full_path)
			os.Exit(1)
		}
		fmt.Printf("Sandbox %s deleted\n", full_path)
	}
	// fmt.Printf("%#v\n",exec_list)
	return
}

func DeleteSandbox(cmd *cobra.Command, args []string) {
	var exec_lists []concurrent.ExecutionList
	if len(args) < 1 {
		fmt.Println("Sandbox name (or \"ALL\") required.")
		fmt.Println("You can run 'dbdeployer sandboxes for a list of available deployments'")
		os.Exit(1)
	}
	flags := cmd.Flags()
	sandbox := args[0]
	confirm, _ := flags.GetBool("confirm")
	run_concurrently, _ := flags.GetBool("concurrent")
	if os.Getenv("RUN_CONCURRENTLY") != "" {
		run_concurrently = true
	}
	skip_confirm, _ := flags.GetBool("skip-confirm")
	sandbox_dir, _ := flags.GetString("sandbox-home")
	deletion_list := []common.SandboxInfo{common.SandboxInfo{sandbox, false}}
	if sandbox == "ALL" || sandbox == "all" {
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
				fmt.Println("Execution interrupted by user")
				os.Exit(0)
			}
		}
	}
	for _, sb := range deletion_list {
		if sb.Locked {
			fmt.Printf("Sandbox %s is locked\n",sb.SandboxName)
		} else {
			exec_list := RemoveSandbox(sandbox_dir, sb.SandboxName, run_concurrently)
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
