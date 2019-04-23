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
	"os"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/spf13/cobra"
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
	// If the command line was not set in the abbreviations module,
	// we save it here, before it is processed by Cobra
	if len(common.CommandLineArgs) == 0 {

		common.CommandLineArgs = append(common.CommandLineArgs, os.Args...)
	}
	if err := rootCmd.Execute(); err != nil {
		common.Exitf(1, "%s", err)
	}
}

func setPflag(cmd *cobra.Command, key string, abbr string, envVar string, defaultVar string, helpStr string, isSlice bool) {
	var defaultValue string
	if envVar != "" {
		defaultValue = os.Getenv(envVar)
	}
	if defaultValue == "" {
		defaultValue = defaultVar
	}
	if isSlice {
		cmd.PersistentFlags().StringSliceP(key, abbr, []string{defaultValue}, helpStr)
	} else {
		cmd.PersistentFlags().StringP(key, abbr, defaultValue, helpStr)
	}
}

func checkDefaultsFile() {
	flags := rootCmd.Flags()
	defaults.CustomConfigurationFile, _ = flags.GetString(globals.ConfigLabel)
	if defaults.CustomConfigurationFile != defaults.ConfigurationFile {
		if common.FileExists(defaults.CustomConfigurationFile) {
			defaults.ConfigurationFile = defaults.CustomConfigurationFile
		} else {
			common.Exitf(1, globals.ErrFileNotFound, defaults.CustomConfigurationFile)
		}
	}
	defaults.LoadConfiguration()
	loadTemplates()
}

func init() {
	cobra.OnInitialize(checkDefaultsFile)
	// spew.Dump(rootCmd)
	rootCmd.PersistentFlags().StringVar(&defaults.CustomConfigurationFile, globals.ConfigLabel, defaults.ConfigurationFile, "configuration file")
	setPflag(rootCmd, globals.SandboxHomeLabel, "", "SANDBOX_HOME", defaults.Defaults().SandboxHome, "Sandbox deployment directory", false)
	setPflag(rootCmd, globals.SandboxBinaryLabel, "", "SANDBOX_BINARY", defaults.Defaults().SandboxBinary, "Binary repository", false)

	rootCmd.InitDefaultVersionFlag()

	// Indicates that we're using dbdeployer command line interface
	// rather than calling its sandbox creation functions from other apps.
	globals.UsingDbDeployer = true
}
