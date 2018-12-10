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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/spf13/cobra"
	"path"
)

func globalRunCommand(cmd *cobra.Command, executable string, args []string, requireArgs bool, skipMissing bool) {
	sandboxDir, err := getAbsolutePathFromFlag(cmd, "sandbox-home")
	common.ErrCheckExitf(err, 1, "error defining absolute path for 'sandbox-home'")
	sandboxList, err := common.GetInstalledSandboxes(sandboxDir)
	common.ErrCheckExitf(err, 1, globals.ErrRetrievingSandboxList, err)
	runList := common.SandboxInfoToFileNames(sandboxList)
	if len(runList) == 0 {
		common.Exitf(1, "no sandboxes found in %s", sandboxDir)
	}
	if requireArgs && len(args) < 1 {
		common.Exitf(1, "arguments required for command %s", executable)
	}
	for _, sb := range runList {
		singleUse := true
		fullDirPath := path.Join(sandboxDir, sb)
		cmdFile := path.Join(fullDirPath, executable)
		realExecutable := executable
		if !common.ExecExists(cmdFile) {
			cmdFile = path.Join(fullDirPath, executable+"_all")
			realExecutable = executable + "_all"
			singleUse = false
		}
		if !common.ExecExists(cmdFile) {
			if skipMissing {
				common.CondPrintf("# Sandbox %s: executable %s not found\n", fullDirPath, executable)
				continue
			}
			common.Exitf(1, "no %s or %s found in %s", executable, executable+"_all", fullDirPath)
		}
		var cmdArgs []string

		if singleUse && executable == "use" {
			cmdArgs = append(cmdArgs, "-e")
		}
		for _, arg := range args {
			cmdArgs = append(cmdArgs, arg)
		}
		var err error
		common.CondPrintf("# Running \"%s\" on %s\n", realExecutable, sb)
		if len(cmdArgs) > 0 {
			_, err = common.RunCmdWithArgs(cmdFile, cmdArgs)
		} else {
			_, err = common.RunCmd(cmdFile)
		}
		common.ErrCheckExitf(err, 1, "error while running %s\n", cmdFile)
		fmt.Println("")
	}
}

func startAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, globals.ScriptStart, args, false, false)
}

func restartAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, globals.ScriptRestart, args, false, false)
}

func stopAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, globals.ScriptStop, args, false, false)
}

func statusAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, globals.ScriptStatus, args, false, false)
}

func testAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, globals.ScriptTestSb, args, false, false)
}

func testReplicationAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, globals.ScriptTestReplication, args, false, true)
}

func useAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, globals.ScriptUse, args, true, false)
}

var (
	globalCmd = &cobra.Command{
		Use:   "global",
		Short: "Runs a given command in every sandbox",
		Long:  `This command can propagate the given action through all sandboxes.`,
		Example: `
	$ dbdeployer global use "select version()"
	$ dbdeployer global status
	$ dbdeployer global stop
	`,
	}

	globalStartCmd = &cobra.Command{
		Use:   "start [options]",
		Short: "Starts all sandboxes",
		Long:  ``,
		Run:   startAllSandboxes,
	}

	globalRestartCmd = &cobra.Command{
		Use:   "restart [options]",
		Short: "Restarts all sandboxes",
		Long:  ``,
		Run:   restartAllSandboxes,
	}

	globalStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops all sandboxes",
		Long:  ``,
		Run:   stopAllSandboxes,
	}
	globalStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Shows the status in all sandboxes",
		Long:  ``,
		Run:   statusAllSandboxes,
	}

	globalTestCmd = &cobra.Command{
		Use:     "test",
		Aliases: []string{"test_sb", "test-sb"},
		Short:   "Tests all sandboxes",
		Long:    ``,
		Run:     testAllSandboxes,
	}

	globalTestReplicationCmd = &cobra.Command{
		Use:     "test-replication",
		Aliases: []string{"test_replication"},
		Short:   "Tests replication in all sandboxes",
		Long:    ``,
		Run:     testReplicationAllSandboxes,
	}

	globalUseCmd = &cobra.Command{
		Use:   "use {query}",
		Short: "Runs a query in all sandboxes",
		Long: `Runs a query in all sandboxes.
It does not check if the query is compatible with every version deployed.
For example, a query using @@port won't run in MySQL 5.0.x`,
		Example: `
	$ dbdeployer global use "select @@server_id, @@port"`,
		Run: useAllSandboxes,
	}
)

func init() {
	rootCmd.AddCommand(globalCmd)
	globalCmd.AddCommand(globalStartCmd)
	globalCmd.AddCommand(globalRestartCmd)
	globalCmd.AddCommand(globalStopCmd)
	globalCmd.AddCommand(globalStatusCmd)
	globalCmd.AddCommand(globalTestCmd)
	globalCmd.AddCommand(globalTestReplicationCmd)
	globalCmd.AddCommand(globalUseCmd)

}
