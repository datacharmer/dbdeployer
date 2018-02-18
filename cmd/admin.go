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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/spf13/cobra"
	"os"
)

func ShowDefaults(cmd *cobra.Command, args []string) {
	defaults.ShowDefaults(defaults.Defaults())
}

func WriteDefaults(cmd *cobra.Command, args []string) {
	defaults.WriteDefaultsFile(defaults.ConfigurationFile, defaults.Defaults())
}

func RemoveDefaults(cmd *cobra.Command, args []string) {
	defaults.RemoveDefaultsFile()
}

func LoadDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("'load' requires a file name\n")
		os.Exit(1)
	}
	filename := args[0]
	new_defaults := defaults.ReadDefaultsFile(filename)
	if defaults.ValidateDefaults(new_defaults) {
		defaults.WriteDefaultsFile(defaults.ConfigurationFile, new_defaults)
	}
}

func ExportDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("'export' requires a file name\n")
		os.Exit(1)
	}
	filename := args[0]
	if common.FileExists(filename) {
		fmt.Printf("File %s already exists. Will not overwrite\n", filename)
		os.Exit(1)
	}
	defaults.WriteDefaultsFile(filename, defaults.Defaults())
	fmt.Printf("# Defaults exported to file %s\n", filename)
}

func UpdateDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("'update' requires a label and a value\n")
		fmt.Printf("Example: dbdeployer admin update master-slave-base-port 17500")
		os.Exit(1)
	}
	label := args[0]
	value := args[1]
	defaults.UpdateDefaults(label, value)
	defaults.ShowDefaults(defaults.Defaults())
}

var (
	adminCmd = &cobra.Command{
		Use:   "admin",
		Short: "administrative tasks",
		Long: `Runs commands related to the administration of dbdeployer,
such as showing the defaults and saving new ones.`,
	}

	adminShowCmd = &cobra.Command{
		Use:     "show",
		Short:   "shows defaults",
		Aliases: []string{"list"},
		Long:    `Shows currently defined defaults`,
		Run:     ShowDefaults,
	}

	adminLoadCmd = &cobra.Command{
		Use:   "load file_name",
		Short: "Load defaults from file",
		Long:  fmt.Sprintf(`Reads defaults from file and saves them to dbdeployer configuration file (%s)`, defaults.ConfigurationFile),
		Run:   LoadDefaults,
	}

	adminUpdateCmd = &cobra.Command{
		Use:   "update label value",
		Short: "Load defaults from file",
		Example: `
	$ dbdeployer admin update master-slave-base-port 17500		
`,
		Long: `Updates one field of the defaults. Stores the result in the dbdeployer configuration file.
Use "dbdeployer admin show" to see which values are available`,
		Run: UpdateDefaults,
	}

	adminExportCmd = &cobra.Command{
		Use:   "export filename",
		Short: "Export current defaults to a given file",
		Long:  `Saves current defaults to a fileer-defined file`,
		Run:   ExportDefaults,
	}

	adminStoreCmd = &cobra.Command{
		Use:   "store",
		Short: "Store current defaults",
		Long:  fmt.Sprintf(`Saves current defaults to dbdeployer configuration file (%s)`, defaults.ConfigurationFile),
		Run:   WriteDefaults,
	}

	adminRemoveCmd = &cobra.Command{
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
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminStoreCmd)
	adminCmd.AddCommand(adminShowCmd)
	adminCmd.AddCommand(adminRemoveCmd)
	adminCmd.AddCommand(adminLoadCmd)
	adminCmd.AddCommand(adminUpdateCmd)
	adminCmd.AddCommand(adminExportCmd)
}
