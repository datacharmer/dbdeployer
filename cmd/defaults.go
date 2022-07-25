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
	"fmt"
	"regexp"

	"github.com/alexeyco/simpletable"
	"github.com/datacharmer/dbdeployer/ops"

	"github.com/datacharmer/dbdeployer/globals"
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

func enableBashCompletion(cmd *cobra.Command, args []string) {

	flags := cmd.Flags()
	useRemote, _ := flags.GetBool(globals.RemoteLabel)
	runIt, _ := flags.GetBool(globals.RunItLabel)
	remoteUrl, _ := flags.GetString(globals.RemoteUrlLabel)
	completionFile, _ := flags.GetString(globals.CompletionFileLabel)
	err := ops.ProcessBashCompletionEnabling(useRemote, runIt, remoteUrl, completionFile)
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
