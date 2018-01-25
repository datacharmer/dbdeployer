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

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dbdeployer",
	Short: "Installs multiple MySQL servers on the same host",
	Long: `Makes MySQL server installation an easy task.
	Runs single, multiple, and replicated sandboxes.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	Version: versionDef,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func set_pflag(key string, env_var string, default_var string, help_str string) {
	var default_value string
	if env_var != "" {
		default_value = os.Getenv(env_var)
	}
	if default_value == "" {
		default_value = default_var
	}
	rootCmd.PersistentFlags().String(key, default_value, help_str)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./sandbox.json", "config file")
	set_pflag("sandbox-home", "SANDBOX_HOME", os.Getenv("HOME")+"/sandboxes", "Sandbox deployment direcory")
	set_pflag("sandbox-binary", "SANDBOX_BINARY", os.Getenv("HOME")+"/opt/mysql", "Binary repository")

	set_pflag("remote-access", "", "127.%", "defines the database access ")
	set_pflag("bind-address", "", "127.0.0.1", "defines the database bind-address ")
	set_pflag("init-options", "INIT_OPTIONS", "", "mysqld options to run during initialization")
	set_pflag("my-cnf-options", "MY_CNF_OPTIONS", "", "mysqld options to add to my.sandbox.cnf")
	// This option will allow to merge the template with an external my.cnf
	// The options that are essential for the sandbox will be preserved
	//set_pflag("my-cnf-file", "MY_CNF_file", "", "Alternate source file for my.sandbox.cnf")
	set_pflag("db-user", "", "msandbox", "database user")
	set_pflag("rpl-user", "", "rsandbox", "replication user")
	set_pflag("db-password", "", "msandbox", "database password")
	set_pflag("rpl-password", "", "rsandbox", "replication password")
	rootCmd.PersistentFlags().Bool("gtid", false, "enables GTID")

	rootCmd.InitDefaultVersionFlag()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".dbdeployer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".dbdeployer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
