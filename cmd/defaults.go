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
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/spf13/cobra"
)

func ShowDefaults(cmd *cobra.Command, args []string) {
	defaults.ShowDefaults(defaults.Defaults())
}

func WriteDefaults(cmd *cobra.Command, args []string) {
	defaults.WriteDefaultsFile(defaults.ConfigurationFile, defaults.Defaults())
	fmt.Printf("# Default values exported to %s\n", defaults.ConfigurationFile)
}

func RemoveDefaults(cmd *cobra.Command, args []string) {
	defaults.RemoveDefaultsFile()
}

func LoadDefaults(cmd *cobra.Command, args []string) {
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
	fmt.Printf("Defaults imported from %s into %s\n", filename, defaults.ConfigurationFile)
}

func ExportDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, "'export' requires a file name")
	}
	filename := args[0]
	if common.FileExists(filename) {
		common.Exitf(1, "file '%s' already exists. Will not overwrite", filename)
	}
	defaults.WriteDefaultsFile(filename, defaults.Defaults())
	fmt.Printf("# Defaults exported to file %s\n", filename)
}

func UpdateDefaults(cmd *cobra.Command, args []string) {
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
		Run:     ShowDefaults,
	}

	defaultsLoadCmd = &cobra.Command{
		Use:     "load file_name",
		Short:   "Load defaults from file",
		Aliases: []string{"import"},
		Long:    fmt.Sprintf(`Reads defaults from file and saves them to dbdeployer configuration file (%s)`, defaults.ConfigurationFile),
		Run:     LoadDefaults,
	}

	defaultsUpdateCmd = &cobra.Command{
		Use:   "update label value",
		Short: "Load defaults from file",
		Example: `
	$ dbdeployer defaults update master-slave-base-port 17500		
`,
		Long: `Updates one field of the defaults. Stores the result in the dbdeployer configuration file.
Use "dbdeployer defaults show" to see which values are available`,
		Run: UpdateDefaults,
	}

	defaultsExportCmd = &cobra.Command{
		Use:   "export filename",
		Short: "Export current defaults to a given file",
		Long:  `Saves current defaults to a user-defined file`,
		Run:   ExportDefaults,
	}

	defaultsStoreCmd = &cobra.Command{
		Use:   "store",
		Short: "Store current defaults",
		Long:  fmt.Sprintf(`Saves current defaults to dbdeployer configuration file (%s)`, defaults.ConfigurationFile),
		Run:   WriteDefaults,
	}

	defaultsRemoveCmd = &cobra.Command{
		Use:     "reset",
		Aliases: []string{"remove"},
		Short:   "Remove current defaults file",
		Long: fmt.Sprintf(`Removes current dbdeployer configuration file (%s)`, defaults.ConfigurationFile) + `
Afterwards, dbdeployer will use the internally stored defaults.
`,
		Run: RemoveDefaults,
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
}
