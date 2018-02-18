// Copyright © 2017-2018 Giuseppe Maxia
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

func set_pflag(key string, abbr string, env_var string, default_var string, help_str string, is_slice bool) {
	var default_value string
	if env_var != "" {
		default_value = os.Getenv(env_var)
	}
	if default_value == "" {
		default_value = default_var
	}
	if is_slice {
		rootCmd.PersistentFlags().StringSliceP(key, abbr, []string{default_value}, help_str)
	} else {
		rootCmd.PersistentFlags().StringP(key, abbr, default_value, help_str)
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
	rootCmd.PersistentFlags().StringVar(&defaults.CustomConfigurationFile, "config", defaults.ConfigurationFile, "configuration file")
	set_pflag("sandbox-home", "", "SANDBOX_HOME", defaults.Defaults().SandboxHome, "Sandbox deployment direcory", false)
	set_pflag("sandbox-binary", "", "SANDBOX_BINARY", defaults.Defaults().SandboxBinary, "Binary repository", false)

	set_pflag("remote-access", "", "", "127.%", "defines the database access ", false)
	set_pflag("bind-address", "", "", "127.0.0.1", "defines the database bind-address ", false)
	set_pflag("custom-mysqld", "", "", "", "Uses an alternative mysqld (must be in the same directory as regular mysqld)", false)
	set_pflag("init-options", "i", "INIT_OPTIONS", "", "mysqld options to run during initialization", true)
	set_pflag("my-cnf-options", "c", "MY_CNF_OPTIONS", "", "mysqld options to add to my.sandbox.cnf", true)
	set_pflag("pre-grants-sql-file", "", "", "", "SQL file to run before loading grants", false)
	set_pflag("pre-grants-sql", "", "", "", "SQL queries to run before loading grants", true)
	set_pflag("post-grants-sql", "", "", "", "SQL queries to run after loading grants", true)
	set_pflag("post-grants-sql-file", "", "", "", "SQL file to run after loading grants", false)
	// This option will allow to merge the template with an external my.cnf
	// The options that are essential for the sandbox will be preserved
	set_pflag("my-cnf-file", "", "MY_CNF_FILE", "", "Alternative source file for my.sandbox.cnf", false)
	set_pflag("db-user", "u", "", "msandbox", "database user", false)
	set_pflag("rpl-user", "", "", "rsandbox", "replication user", false)
	set_pflag("db-password", "p", "", "msandbox", "database password", false)
	set_pflag("rpl-password", "", "", "rsandbox", "replication password", false)
	set_pflag("use-template", "", "", "", "[template_name:file_name] Replace existing template with one from file", true)
	set_pflag("sandbox-directory", "", "", "", "Changes the default sandbox directory", false)
	rootCmd.PersistentFlags().Int("port", 0, "Overrides default port")
	rootCmd.PersistentFlags().Int("base-port", 0, "Overrides default base-port (for multiple sandboxes)")
	rootCmd.PersistentFlags().Bool("gtid", false, "enables GTID")
	rootCmd.PersistentFlags().Bool("keep-auth-plugin", false, "in 8.0.4+, does not change the auth plugin")
	rootCmd.PersistentFlags().Bool("keep-server-uuid", false, "Does not change the server UUID")
	rootCmd.PersistentFlags().Bool("force", false, "If a destination sandbox already exists, it will be overwritten")
	rootCmd.PersistentFlags().Bool("skip-load-grants", false, "Does not load the grants")
	rootCmd.PersistentFlags().Bool("expose-dd-tables", false, "In MySQL 8.0+ shows data dictionary tables")
	// TODO rootCmd.PersistentFlags().Bool("check-port", false, "Check if the port is already in use, and find a free one")

	rootCmd.InitDefaultVersionFlag()
}
