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
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy sandboxes",
	Long: `Deploys single, multiple, or replicated sandboxes`,
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.PersistentFlags().Int("port", 0, "Overrides default port")
	deployCmd.PersistentFlags().Int("base-port", 0, "Overrides default base-port (for multiple sandboxes)")
	deployCmd.PersistentFlags().Bool("gtid", false, "enables GTID")
	deployCmd.PersistentFlags().Bool("keep-auth-plugin", false, "in 8.0.4+, does not change the auth plugin")
	deployCmd.PersistentFlags().Bool("keep-server-uuid", false, "Does not change the server UUID")
	deployCmd.PersistentFlags().Bool("force", false, "If a destination sandbox already exists, it will be overwritten")
	deployCmd.PersistentFlags().Bool("skip-load-grants", false, "Does not load the grants")
	deployCmd.PersistentFlags().Bool("expose-dd-tables", false, "In MySQL 8.0+ shows data dictionary tables")

	set_pflag(deployCmd,"remote-access", "", "", "127.%", "defines the database access ", false)
	set_pflag(deployCmd,"bind-address", "", "", "127.0.0.1", "defines the database bind-address ", false)
	set_pflag(deployCmd,"custom-mysqld", "", "", "", "Uses an alternative mysqld (must be in the same directory as regular mysqld)", false)
	set_pflag(deployCmd,"defaults", "", "", "", "Change defaults on-the-fly (--defaults=label:value)", true)
	set_pflag(deployCmd,"init-options", "i", "INIT_OPTIONS", "", "mysqld options to run during initialization", true)
	set_pflag(deployCmd,"my-cnf-options", "c", "MY_CNF_OPTIONS", "", "mysqld options to add to my.sandbox.cnf", true)
	set_pflag(deployCmd,"pre-grants-sql-file", "", "", "", "SQL file to run before loading grants", false)
	set_pflag(deployCmd,"pre-grants-sql", "", "", "", "SQL queries to run before loading grants", true)
	set_pflag(deployCmd,"post-grants-sql", "", "", "", "SQL queries to run after loading grants", true)
	set_pflag(deployCmd,"post-grants-sql-file", "", "", "", "SQL file to run after loading grants", false)
	// This option will allow to merge the template with an external my.cnf
	// The options that are essential for the sandbox will be preserved
	set_pflag(deployCmd,"my-cnf-file", "", "MY_CNF_FILE", "", "Alternative source file for my.sandbox.cnf", false)
	set_pflag(deployCmd,"db-user", "u", "", "msandbox", "database user", false)
	set_pflag(deployCmd,"rpl-user", "", "", "rsandbox", "replication user", false)
	set_pflag(deployCmd,"db-password", "p", "", "msandbox", "database password", false)
	set_pflag(deployCmd,"rpl-password", "", "", "rsandbox", "replication password", false)
	set_pflag(deployCmd,"use-template", "", "", "", "[template_name:file_name] Replace existing template with one from file", true)
	set_pflag(deployCmd,"sandbox-directory", "", "", "", "Changes the default sandbox directory", false)

}
