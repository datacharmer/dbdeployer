// Copyright © 2018 Giuseppe Maxia
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
	"github.com/spf13/cobra"
	"os"
)

func DeleteSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Sandbox name required.")
		fmt.Println("You can run 'dbdeployer sandboxes for a list of available deployments'")
		os.Exit(1)
	}
	flags := cmd.Flags()
	sandbox := args[0]
	confirm, _ := flags.GetBool("confirm")
	sandbox_dir, _ := flags.GetString("sandbox-home")
	full_path := sandbox_dir + "/" + sandbox
	if !common.DirExists(full_path) {
		fmt.Println("Directory '%s' not found\n", full_path)
		os.Exit(1)
	}
	stop := full_path + "/stop_all"
	if !common.ExecExists(stop) {
		stop = full_path + "/stop"
	}
	if !common.ExecExists(stop) {
		fmt.Println("Executable '%s' not found\n", stop)
		os.Exit(1)
	}
	if confirm {
		fmt.Printf("We're about to delete %s\n", full_path)
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
	fmt.Printf("Running %s\n", stop)
	err := common.Run_cmd(stop)
	if err != nil {
		fmt.Printf("Error while stopping sandbox %s\n", full_path)
		os.Exit(1)
	}

	rm_cmd := []string{"-rf", full_path}
	cmd_str := "rm"
	for _, item := range rm_cmd {
		cmd_str += " " + item
	}
	fmt.Printf("Running %s\n", cmd_str)
	err = common.Run_cmd_with_args("rm", rm_cmd)
	if err != nil {
		fmt.Printf("Error while deleting sandbox %s\n", full_path)
		os.Exit(1)
	}
	fmt.Printf("Sandbox %s deleted\n", full_path)
}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete sandbox_name",
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

	deleteCmd.Flags().BoolP("confirm", "", false, "Requires confirmation.")
}
