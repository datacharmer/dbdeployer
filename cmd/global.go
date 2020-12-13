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
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/spf13/cobra"
)

func getRange(s string) (min, max int, negation bool, err error) {
	if s == "" {
		return 0, 0, false, nil
	}
	value, negate := common.OptionComponents(s)
	reRange := regexp.MustCompile(`(\d+)[.,:-](\d+)`)
	rangeList := reRange.FindAllStringSubmatch(value, -1)
	if len(rangeList) == 0 || len(rangeList[0]) == 0 {
		return 0, 0, false, fmt.Errorf("error detecting range. Expected format: xxxx-xxxx")
	}
	minText := rangeList[0][1]
	maxText := rangeList[0][2]
	min = common.Atoi(minText)
	max = common.Atoi(maxText)
	if min > max {
		return 0, 0, false, fmt.Errorf("minimum value (%d) is greater than maximum value (%d)", min, max)
	}
	negation = negate
	return
}

func globalRunCommand(cmd *cobra.Command, executable string, args []string, requireArgs bool, skipMissing bool) {
	sandboxDir, err := getAbsolutePathFromFlag(cmd, "sandbox-home")
	common.ErrCheckExitf(err, 1, "error defining absolute path for 'sandbox-home'")
	sandboxList, err := common.GetInstalledSandboxes(sandboxDir)
	common.ErrCheckExitf(err, 1, globals.ErrRetrievingSandboxList, err)
	flags := cmd.Flags()
	sbFlavor, _ := flags.GetString(globals.FlavorLabel)
	sbPortOpt, _ := flags.GetString(globals.PortLabel)
	sbType, _ := flags.GetString(globals.TypeLabel)
	sbName, _ := flags.GetString(globals.NameLabel)
	sbVersion, _ := flags.GetString(globals.VersionLabel)
	sbShortVersion, _ := flags.GetString(globals.ShortVersionLabel)
	sbPortRange, _ := flags.GetString(globals.PortRangeLabel)
	verbose, _ := flags.GetBool(globals.VerboseLabel)
	dryRun, _ := flags.GetBool(globals.DryRunLabel)
	sbPortValue, sbPortNegation := common.OptionComponents(sbPortOpt)
	sbPort := 0
	if sbPortValue != "" {
		sbPort = common.Atoi(sbPortValue)
	}
	minPort, maxPort, sbPortRangeNegation, err := getRange(sbPortRange)
	if err != nil {
		common.Exitf(1, "error getting ports range: %s", err)

	}
	runList := common.SandboxInfoToFileNames(sandboxList)
	if len(runList) == 0 {
		common.Exitf(1, "no sandboxes found in %s", sandboxDir)
	}
	if requireArgs && len(args) < 1 {
		common.Exitf(1, "arguments required for command %s", executable)
	}
	var sbDescription common.SandboxDescription
	for _, sb := range runList {
		singleUse := true
		fullDirPath := path.Join(sandboxDir, sb)
		sbDescription, err = common.ReadSandboxDescription(fullDirPath)
		if err != nil {
			common.Exitf(1, "error reading sandbox description from %s", fullDirPath)
			return
		}
		if sbFlavor != "" {
			if !common.OptionCompare(sbFlavor, sbDescription.Flavor) { // sbDescription.Flavor != sbFlavor {
				if verbose {
					common.CondPrintf("Skipping %s of flavor %s \n", sb, sbDescription.Flavor)
				}
				continue
			}
		}
		if sbType != "" {
			if !common.OptionCompare(sbType, sbDescription.SBType) {
				if verbose {
					common.CondPrintf("Skipping %s of type %s \n", sb, sbDescription.SBType)
				}
				continue
			}
		}
		if sbVersion != "" {
			if !common.OptionCompare(sbVersion, sbDescription.Version) {
				if verbose {
					common.CondPrintf("Skipping %s of version %s \n", sb, sbDescription.Version)
				}
				continue
			}
		}
		shortVersionList := strings.Split(sbDescription.Version, ".")
		shortVersion := fmt.Sprintf("%s.%s", shortVersionList[0], shortVersionList[1])
		if sbShortVersion != "" {
			if !common.OptionCompare(sbShortVersion, shortVersion) {
				if verbose {
					common.CondPrintf("Skipping %s of short version %s \n", sb, shortVersion)
				}
				continue
			}
		}
		if sbName != "" {
			if !common.OptionCompare(sbName, sb) {
				if verbose {
					common.CondPrintf("Skipping %s (name) \n", sb)
				}
				continue
			}
		}
		if sbPort != 0 {
			found := false
			for _, port := range sbDescription.Port {
				if port == sbPort {
					found = true
				}
			}
			if sbPortNegation {
				found = !found
			}
			if !found {
				if verbose {
					common.CondPrintf("Skipping %s - port not found \n", sb)
				}
				continue
			}
		}
		if minPort > 0 && maxPort > 0 {
			found := false
			for _, port := range sbDescription.Port {
				if port >= minPort && port <= maxPort {
					found = true
				}
			}
			if sbPortRangeNegation {
				found = !found
			}
			if !found {
				if verbose {
					common.CondPrintf("Skipping %s - port range not matched \n", sb)
				}
				continue
			}
		}

		if executable == "exec" {
			if dryRun {
				fmt.Printf("%v\n", args)
			} else {
				currentDir := os.Getenv("PWD")
				err = os.Chdir(fullDirPath)
				if err != nil {
					fmt.Printf("error changing directory to %s:%s\n", fullDirPath, err)
				}
				cmd := args[0]
				var cmdArgs []string
				//if (cmd == "bash" || cmd == "sh") && len(args) >1 {
				//	cmdArgs = append(cmdArgs, "-c")
				//}
				for N := 1; N < len(args); N++ {
					cmdArgs = append(cmdArgs, args[N])
				}
				fmt.Printf("# %s\n", fullDirPath)
				if len(cmdArgs) > 0 {
					_, err = common.RunCmdWithArgs(cmd, cmdArgs)
				} else {
					_, err = common.RunCmd(cmd)
				}
				common.ErrCheckExitf(err, 1, "error while running %s\n", cmd)
				_ = os.Chdir(currentDir)
			}
			continue
		}
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
		cmdArgs = append(cmdArgs, args...)
		var err error
		common.CondPrintf("# Running \"%s\" on %s\n", realExecutable, sb)
		if dryRun {
			argsStr := ""
			for _, elem := range cmdArgs {
				argsStr += " " + elem
			}

			common.CondPrintf("would run '%s %s'\n", cmdFile, argsStr)
		} else {

			if len(cmdArgs) > 0 {
				_, err = common.RunCmdWithArgs(cmdFile, cmdArgs)
			} else {
				_, err = common.RunCmd(cmdFile)
			}
			common.ErrCheckExitf(err, 1, "error while running %s\n", cmdFile)
		}
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

func execAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, "exec", args, true, false)
}

func metadataAllSandboxes(cmd *cobra.Command, args []string) {
	globalRunCommand(cmd, globals.ScriptMetadata, args, true, false)
}

var (
	globalCmd = &cobra.Command{
		Use:   "global",
		Short: "Runs a given command in every sandbox",
		Long:  `This command can propagate the given action through all sandboxes.`,
		Example: `
	$ dbdeployer global use "select version()"
	$ dbdeployer global status
	$ dbdeployer global stop --version=5.7.27
	$ dbdeployer global stop --short-version=8.0
	$ dbdeployer global stop --short-version='!8.0' # or --short-version=no-8.0
	$ dbdeployer global status --port-range=5000-8099
	$ dbdeployer global start --flavor=percona
	$ dbdeployer global start --flavor='!percona' --type=single
	$ dbdeployer global metadata version --flavor='!percona' --type=single
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
		Run:         useAllSandboxes,
		Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
	}

	globalExecCmd = &cobra.Command{
		Use:   "exec {command}",
		Short: "Runs a command in all sandboxes",
		Long: `Runs a command in all sandboxes.
The command will be executed inside the sandbox. Thus, you can reference a file that you know should be there.
There is no check to ensure that a give command is doable.
For example: the command "cat filename" will result in an error if filename is not present in the sandbox directory.
You can run complex shell commands by prepending them with either "bash -- -c" or "sh -- -c". Such command must be quoted. 
Only one command can be passed, although you can use a shell command as described above to overcome this limitation.
IMPORTANT: if your command argument contains flags, you must use a double dash (--) before any of the flags.

You may combine the command passed to "exec" with other drill-down commands in the sandbox directory, 
such as "./exec_all". In this case, you need to make sure that all your sandbox contain the command, or use "--type"
to only run "exec" in the specific topologies.
`,
		Example: `$ dbdeployer global exec grep 'version\|type' sbdescription.json
	$ dbdeployer global exec grep -- -w basedir sbdescription.json
	$ dbdeployer global exec pwd
	$ dbdeployer global exec shell -- -c "if [ -f filename ] ; then echo 'found filename'; fi"
    $ dbdeployer global exec --type=group-multi-primary ./exec_all ls -- -l data/ibdata1
`,
		Run:         execAllSandboxes,
		Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
	}

	globalMetadataCmd = &cobra.Command{
		Use:   "metadata {keyword}",
		Short: "Runs a metadata query in all sandboxes",
		Long:  `Runs a metadata query in all sandboxes`,
		Example: `
	$ dbdeployer global metadata `,
		Run:         metadataAllSandboxes,
		Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
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
	globalCmd.AddCommand(globalExecCmd)
	globalCmd.AddCommand(globalMetadataCmd)

	setPflag(globalCmd, globals.VersionLabel, "", "", "", "Runs command only in sandboxes of the given version", false)
	setPflag(globalCmd, globals.ShortVersionLabel, "", "", "", "Runs command only in sandboxes of the given short version", false)
	setPflag(globalCmd, globals.FlavorLabel, "", "", "", "Runs command only in sandboxes of the given flavor", false)
	setPflag(globalCmd, globals.TypeLabel, "", "", "", "Runs command only in sandboxes of the given type", false)
	setPflag(globalCmd, globals.NameLabel, "", "", "", "Runs command only in sandboxes of the given name", false)
	setPflag(globalCmd, globals.PortRangeLabel, "", "", "", "Runs command only in sandboxes containing a port in the given range", false)
	globalCmd.PersistentFlags().String(globals.PortLabel, "", "Runs commands only in sandboxes containing the given port")
	globalCmd.PersistentFlags().Bool(globals.VerboseLabel, false, "Show what is matched when filters are used")
	globalCmd.PersistentFlags().Bool(globals.DryRunLabel, false, "Show what would be executed, without doing it")
}
