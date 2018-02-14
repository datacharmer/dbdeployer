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
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/datacharmer/dbdeployer/common"
)

func GlobalRunCommand(cmd *cobra.Command, executable string, args []string, require_args bool) {
	flags := cmd.Flags()
	sandbox_dir, _ := flags.GetString("sandbox-home")
	run_list := common.GetInstalledSandboxes(sandbox_dir)
	if len(run_list) == 0 {
		fmt.Printf("No sandboxes found in %s\n", sandbox_dir)
		os.Exit(1)
	}
	if require_args && len(args) < 1 {
		fmt.Printf("Arguments required for command %s\n", executable)
		os.Exit(1)
	}
	for _, sb := range run_list {
		single_use := true
		full_dir_path := sandbox_dir + "/" + sb
		cmd_file := full_dir_path + "/" + executable
		real_executable := executable
		if !common.ExecExists(cmd_file) {
			cmd_file = full_dir_path + "/" + executable + "_all"
			real_executable = executable + "_all"
			single_use = false
		}
		if !common.ExecExists(cmd_file) {
			fmt.Printf("No %s or %s found in %s\n", executable, executable + "_all", full_dir_path)
			os.Exit(1)
		}
		var cmd_args []string
		
		if single_use && executable == "use" {
			cmd_args = append(cmd_args, "-e")
		}
		for _,arg := range args {
			cmd_args = append(cmd_args, arg)
		}
		var err error
		fmt.Printf("# Running \"%s\" on %s\n", real_executable, sb)
		if len(cmd_args) > 0 {
			err,_ = common.Run_cmd_with_args(cmd_file, cmd_args)
		} else {
			err,_ = common.Run_cmd(cmd_file)
		}
		if err != nil {
			fmt.Printf("Error while running %s\n", cmd_file)
			os.Exit(1)
		}
		fmt.Println("")
	}
}

func StartAllSandboxes(cmd *cobra.Command, args []string) {
	GlobalRunCommand(cmd, "start", args, false)
}

func RestartAllSandboxes(cmd *cobra.Command, args []string) {
	GlobalRunCommand(cmd, "restart", args, false)
}

func StopAllSandboxes(cmd *cobra.Command, args []string) {
	GlobalRunCommand(cmd, "stop", args, false)
}

func StatusAllSandboxes(cmd *cobra.Command, args []string) {
	GlobalRunCommand(cmd, "status", args, false)
}

func TestAllSandboxes(cmd *cobra.Command, args []string) {
	GlobalRunCommand(cmd, "test_sb", args, false)
}

func UseAllSandboxes(cmd *cobra.Command, args []string) {
	GlobalRunCommand(cmd, "use", args, true)
}

var (
	globalCmd = &cobra.Command{
		Use:   "global",
		Short: "Runs a given command in every sandbox",
		Long: `This command can propagate the given action through all sandboxes.`,
		Example: `
	$ dbdeployer global use "select version()"
	$ dbdeployer global status
	$ dbdeployer global stop
	`,
	}

	globalStartCmd = &cobra.Command{
		Use:   "start [options]",
		Short: "Starts all sandboxes",
		Long: ``,
		Run: StartAllSandboxes,
	}

	globalRestartCmd = &cobra.Command{
		Use:   "restart [options]",
		Short: "Restarts all sandboxes",
		Long: ``,
		Run: RestartAllSandboxes,
	}

	globalStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops all sandboxes",
		Long: ``,
		Run: StopAllSandboxes,
	}
	globalStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Shows the status in all sandboxes",
		Long: ``,
		Run: StatusAllSandboxes,
	}

	globalTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Tests all sandboxes",
		Long: ``,
		Run: TestAllSandboxes,
	}

	globalUseCmd = &cobra.Command{
		Use:   "use {query}",
		Short: "Runs a query in all sandboxes",
		Long: `Runs a query in all sandboxes.
It does not check if the query is compatible with every version deployed.
For example, a query using @@port won't run in MySQL 5.0.x`,
		Example: `
	$ dbdeployer global use "select @@server_id, @@port"`,
		Run: UseAllSandboxes,
	}
)

func init() {
	rootCmd.AddCommand(globalCmd)
	globalCmd.AddCommand(globalStartCmd)
	globalCmd.AddCommand(globalRestartCmd)
	globalCmd.AddCommand(globalStopCmd)
	globalCmd.AddCommand(globalStatusCmd)
	globalCmd.AddCommand(globalTestCmd)
	globalCmd.AddCommand(globalUseCmd)

}
