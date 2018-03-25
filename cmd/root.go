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

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/spf13/cobra"
	//"github.com/davecgh/go-spew/spew"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dbdeployer",
	Short: "Installs multiple MySQL servers on the same host",
	Long: `dbdeployer makes MySQL server installation an easy task.
Runs single, multiple, and replicated sandboxes.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	Version: common.VersionDef,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func set_pflag(cmd *cobra.Command, key string, abbr string, env_var string, default_var string, help_str string, is_slice bool) {
	var default_value string
	if env_var != "" {
		default_value = os.Getenv(env_var)
	}
	if default_value == "" {
		default_value = default_var
	}
	if is_slice {
		cmd.PersistentFlags().StringSliceP(key, abbr, []string{default_value}, help_str)
	} else {
		cmd.PersistentFlags().StringP(key, abbr, default_value, help_str)
	}
}

func checkDefaultsFile() {
	flags := rootCmd.Flags()
	defaults.CustomConfigurationFile, _ = flags.GetString("config")
	if defaults.CustomConfigurationFile != defaults.ConfigurationFile {
		if common.FileExists(defaults.CustomConfigurationFile) {
			defaults.ConfigurationFile = defaults.CustomConfigurationFile
		} else {
			fmt.Printf("*** File %s not found\n", defaults.CustomConfigurationFile)
			os.Exit(1)
		}
	}
	defaults.LoadConfiguration()
	LoadTemplates()
}

func init() {
	cobra.OnInitialize(checkDefaultsFile)
	// spew.Dump(rootCmd)
	rootCmd.PersistentFlags().StringVar(&defaults.CustomConfigurationFile, "config", defaults.ConfigurationFile, "configuration file")
	set_pflag(rootCmd,"sandbox-home", "", "SANDBOX_HOME", defaults.Defaults().SandboxHome, "Sandbox deployment direcory", false)
	set_pflag(rootCmd,"sandbox-binary", "", "SANDBOX_BINARY", defaults.Defaults().SandboxBinary, "Binary repository", false)

	rootCmd.InitDefaultVersionFlag()
}
