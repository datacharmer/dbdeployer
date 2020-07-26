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
	"regexp"

	"github.com/alexeyco/simpletable"

	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"

	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
)

func showDefaults(cmd *cobra.Command, args []string) {
	camelCase, _ := cmd.Flags().GetBool(globals.CamelCase)
	if camelCase {
		results := defaults.DefaultsToMap()
		reDash := regexp.MustCompile(`-`)
		reCapital := regexp.MustCompile(`^[A-Z]`)
		for k, v := range results {
			if reDash.MatchString(k) {
				continue
			}
			if !reCapital.MatchString(k) {
				continue
			}
			fmt.Printf("%s %v\n", k, v)
		}
	} else {
		defaults.ShowDefaults(defaults.Defaults())
	}
}

func writeDefaults(cmd *cobra.Command, args []string) {
	defaults.WriteDefaultsFile(defaults.ConfigurationFile, defaults.Defaults())
	common.CondPrintf("# Default values exported to %s\n", defaults.ConfigurationFile)
}

func removeDefaults(cmd *cobra.Command, args []string) {
	defaults.RemoveDefaultsFile()
}

func loadDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, "'load' requires a file name")
	}
	filename := args[0]
	newDefaults := defaults.ReadDefaultsFile(filename)
	if defaults.ValidateDefaults(newDefaults) {
		defaults.WriteDefaultsFile(defaults.ConfigurationFile, newDefaults)
	} else {
		return
	}
	common.CondPrintf("Defaults imported from %s into %s\n", filename, defaults.ConfigurationFile)
}

func exportDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, "'export' requires a file name")
	}
	filename := args[0]
	if common.FileExists(filename) {
		common.Exitf(1, "file '%s' already exists. Will not overwrite", filename)
	}
	defaults.WriteDefaultsFile(filename, defaults.Defaults())
	common.CondPrintf("# Defaults exported to file %s\n", filename)
}

func updateDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		common.Exit(1,
			"'update' requires a label and a value",
			"Example: dbdeployer defaults update master-slave-base-port 17500")
	}
	label := args[0]
	value := args[1]
	defaults.UpdateDefaults(label, value, true)
	defaults.ShowDefaults(defaults.Defaults())
}

func processBashCompletionEnabling(useRemote, runIt bool, remoteUrl, completionFile string) error {
	useLocal := completionFile != ""

	var bashCompletionScript string
	var bashCompletionScripts = []string{
		path.Join("/etc", "bash_completion"),
		path.Join("/usr", "local", "etc", "bash_completion"),
		path.Join("/etc", "profile.d", "bash_completion.sh"),
	}
	destinationDir := path.Join("/etc", "bash_completion.d")
	alternateDestinationDir := path.Join("/usr", "local", "etc", "bash_completion.d")
	if !common.DirExists(destinationDir) {
		if common.DirExists(alternateDestinationDir) {
			destinationDir = alternateDestinationDir
		} else {
			return fmt.Errorf("neither %s or %s found", destinationDir, alternateDestinationDir)
		}
	}

	for _, script := range bashCompletionScripts {
		if common.FileExists(script) {
			bashCompletionScript = script
			break
		}
	}
	if bashCompletionScript == "" {
		return fmt.Errorf("none of bash completion scripts found (%v)", bashCompletionScripts)
	}
	if completionFile == "" {
		completionFile = globals.CompletionFileValue
	}
	if useLocal && useRemote {
		return fmt.Errorf("only one of '--%s' or '--%s' should be used", globals.CompletionFileValue, globals.RemoteLabel)
	}
	if !useRemote {
		useLocal = true
	}
	if useLocal {
		defaultCompletionFile := path.Join(os.Getenv("PWD"), globals.CompletionFileValue)
		defaultSecondCompletionFile := path.Join(os.Getenv("PWD"), "docs", globals.CompletionFileValue)
		completionFile, _ = common.AbsolutePath(completionFile)
		if completionFile == defaultCompletionFile {
			if !common.FileExists(completionFile) {
				if common.FileExists(defaultSecondCompletionFile) {
					completionFile = defaultSecondCompletionFile
				}
			}
		}
	}
	if useRemote {
		if remoteUrl == "" {
			return fmt.Errorf("remote URL at '--%s' cannot be empty", globals.RemoteUrlLabel)
		}
		if common.FileExists(completionFile) {
			return fmt.Errorf(globals.ErrFileAlreadyExists, completionFile)
		}
		err := rest.DownloadFile(completionFile, remoteUrl, true, globals.MB)
		if err != nil {
			return fmt.Errorf("error downloading %s: %s", completionFile, err)
		}
		fmt.Printf("Download of file %s was successful\n", completionFile)
	}

	if !common.FileExists(completionFile) {
		return fmt.Errorf(globals.ErrFileNotFound, completionFile)
	}

	fmt.Printf("# completion file: %s\n", completionFile)
	bareCompletionFileName := common.BaseName(completionFile)
	destinationFile := path.Join(destinationDir, bareCompletionFileName)
	if common.FileExists(destinationFile) {
		// Get the checksum of both files, so we can skip the copy if they are already the same
		sourceChecksum, err := common.GetFileSha256(completionFile)
		if err != nil {
			return fmt.Errorf("error getting checksum from file %s", completionFile)
		}
		destChecksum, err := common.GetFileSha256(destinationFile)
		if err != nil {
			return fmt.Errorf("error getting checksum from file %s", destinationFile)
		}
		if sourceChecksum == destChecksum {
			fmt.Printf("Files '%s' and '%s' have the same checksum - Copy is not needed\n", completionFile, destinationFile)
			return nil
		}
	}

	if runIt {
		command := "cp"
		argsList := []string{completionFile, destinationDir}
		sudo := common.Which("sudo")
		if sudo != "" {
			command = sudo
			argsList = []string{"cp", completionFile, destinationDir}
		}
		fmt.Printf("# Running: sudo cp %s %s\n", completionFile, destinationDir)

		output, err := common.RunCmdWithArgs(command, argsList)
		if err != nil {
			fmt.Printf("%s\n", output)
			return fmt.Errorf("error copying bash completion file into %s: %s", destinationDir, err)
		}
		if !common.FileExists(destinationFile) {
			return fmt.Errorf("error after copying bash completion file: "+globals.ErrFileNotFound, destinationFile)
		}
		fmt.Printf("# File copied to %s\n", destinationFile)
	} else {

		fmt.Printf("# Run the command: sudo cp %s %s\n", completionFile, destinationDir)
	}
	fmt.Printf("# Run the command 'source %s'\n", bashCompletionScript)
	return nil
}

func enableBashCompletion(cmd *cobra.Command, args []string) {

	flags := cmd.Flags()
	useRemote, _ := flags.GetBool(globals.RemoteLabel)
	runIt, _ := flags.GetBool(globals.RunItLabel)
	remoteUrl, _ := flags.GetString(globals.RemoteUrlLabel)
	completionFile, _ := flags.GetString(globals.CompletionFileLabel)
	err := processBashCompletionEnabling(useRemote, runIt, remoteUrl, completionFile)
	if err != nil {
		common.Exitf(1, "%s", err)
	}
}

func showFlagAliases(cmd *cobra.Command, args []string) {
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "command"},
			{Align: simpletable.AlignCenter, Text: "flag"},
			{Align: simpletable.AlignCenter, Text: "alias"},
		},
	}

	for _, alias := range globals.FlagAliases {
		var cells []*simpletable.Cell
		cells = append(cells, &simpletable.Cell{Text: alias.Command})
		cells = append(cells, &simpletable.Cell{Text: "--" + alias.FlagName})
		cells = append(cells, &simpletable.Cell{Text: "--" + alias.Alias})
		table.Body.Cells = append(table.Body.Cells, cells)
	}

	table.SetStyle(simpletable.StyleDefault)
	table.Println()
}

var (
	defaultsCmd = &cobra.Command{
		Use:     "defaults",
		Short:   "tasks related to dbdeployer defaults",
		Aliases: []string{"config"},
		Long: `Runs commands related to the administration of dbdeployer,
such as showing the defaults and saving new ones.`,
	}

	defaultsShowCmd = &cobra.Command{
		Use:     "show",
		Short:   "shows defaults",
		Aliases: []string{"list"},
		Long:    `Shows currently defined defaults`,
		Run:     showDefaults,
	}

	defaultsLoadCmd = &cobra.Command{
		Use:         "load file_name",
		Short:       "Load defaults from file",
		Aliases:     []string{"import"},
		Long:        fmt.Sprintf(`Reads defaults from file and saves them to dbdeployer configuration file (%s)`, defaults.ConfigurationFile),
		Run:         loadDefaults,
		Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
	}

	defaultsUpdateCmd = &cobra.Command{
		Use:   "update label value",
		Short: "Change defaults value",
		Example: `
	$ dbdeployer defaults update master-slave-base-port 17500		
`,
		Long: `Updates one field of the defaults. Stores the result in the dbdeployer configuration file.
Use "dbdeployer defaults show" to see which values are available`,
		Run:         updateDefaults,
		Annotations: map[string]string{"export": ExportAnnotationToJson(DoubleStringExport)},
	}

	defaultsExportCmd = &cobra.Command{
		Use:         "export filename",
		Short:       "Export current defaults to a given file",
		Long:        `Saves current defaults to a user-defined file`,
		Run:         exportDefaults,
		Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
	}

	defaultsStoreCmd = &cobra.Command{
		Use:   "store",
		Short: "Store current defaults",
		Long:  fmt.Sprintf(`Saves current defaults to dbdeployer configuration file (%s)`, defaults.ConfigurationFile),
		Run:   writeDefaults,
	}

	defaultsRemoveCmd = &cobra.Command{
		Use:     "reset",
		Aliases: []string{"remove"},
		Short:   "Remove current defaults file",
		Long: fmt.Sprintf(`Removes current dbdeployer configuration file (%s)`, defaults.ConfigurationFile) + `
Afterwards, dbdeployer will use the internally stored defaults.
`,
		Run: removeDefaults,
	}

	defaultsEnableCompletionCmd = &cobra.Command{
		Use:   "enable-bash-completion",
		Short: "Enables bash-completion for dbdeployer",
		Long:  `Enables bash completion using either a local copy of dbdeployer_completion.sh or a remote one`,
		Run:   enableBashCompletion,
	}
	defaultsFlagAliasesCmd = &cobra.Command{
		Use:     "flag-aliases",
		Aliases: []string{"option-aliases", "aliases"},
		Short:   "Shows flag aliases",
		Long:    `Shows the aliases available for some flags`,
		Run:     showFlagAliases,
	}
)

func init() {
	rootCmd.AddCommand(defaultsCmd)
	defaultsCmd.AddCommand(defaultsStoreCmd)
	defaultsCmd.AddCommand(defaultsShowCmd)
	defaultsCmd.AddCommand(defaultsRemoveCmd)
	defaultsCmd.AddCommand(defaultsLoadCmd)
	defaultsCmd.AddCommand(defaultsUpdateCmd)
	defaultsCmd.AddCommand(defaultsExportCmd)
	defaultsCmd.AddCommand(defaultsFlagAliasesCmd)
	defaultsCmd.AddCommand(defaultsEnableCompletionCmd)

	setPflag(defaultsEnableCompletionCmd,
		globals.RemoteUrlLabel, "", "", defaults.Defaults().RemoteCompletionUrl,
		fmt.Sprintf("Where to downloads %s from", globals.CompletionFileValue), false)
	defaultsEnableCompletionCmd.PersistentFlags().Bool(globals.RemoteLabel, false,
		fmt.Sprintf("Download %s from GitHub", globals.CompletionFileValue))
	setPflag(defaultsEnableCompletionCmd, globals.CompletionFileLabel, "", "", "",
		"Use this file as completion", false)
	defaultsEnableCompletionCmd.PersistentFlags().Bool(globals.RunItLabel, false, "Run the command instead of just showing it")
	defaultsShowCmd.PersistentFlags().Bool(globals.CamelCase, false, "Show defaults in CamelCase format")
}
