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
	"github.com/datacharmer/dbdeployer/defaults"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy sandboxes",
	Long: `Deploys single, multiple, or replicated sandboxes`,
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.PersistentFlags().Int(defaults.PortLabel, 0, "Overrides default port")
	deployCmd.PersistentFlags().Int(defaults.BasePortLabel, 0, "Overrides default base-port (for multiple sandboxes)")
	deployCmd.PersistentFlags().Bool(defaults.GtidLabel, false, "enables GTID")
	deployCmd.PersistentFlags().Bool(defaults.NativeAuthPluginLabel, false, "in 8.0.4+, uses the native password auth plugin")
	deployCmd.PersistentFlags().Bool(defaults.KeepServerUuidLabel, false, "Does not change the server UUID")
	deployCmd.PersistentFlags().Bool(defaults.ForceLabel, false, "If a destination sandbox already exists, it will be overwritten")
	deployCmd.PersistentFlags().Bool(defaults.SkipStartLabel, false, "Does not start the database server")
	deployCmd.PersistentFlags().Bool(defaults.DisableMysqlXLabel, false, "Disable MySQLX plugin (8.0.11+)")
	deployCmd.PersistentFlags().Bool(defaults.EnableMysqlXLabel, false, "Enables MySQLX plugin (5.7.12+)")
	deployCmd.PersistentFlags().Bool(defaults.SkipLoadGrantsLabel, false, "Does not load the grants")
	deployCmd.PersistentFlags().Bool(defaults.SkipReportHostLabel, false, "Does not include report host in my.sandbox.cnf")
	deployCmd.PersistentFlags().Bool(defaults.SkipReportPortLabel, false, "Does not include report port in my.sandbox.cnf")
	deployCmd.PersistentFlags().Bool(defaults.ExposeDdTablesLabel, false, "In MySQL 8.0+ shows data dictionary tables")
	deployCmd.PersistentFlags().Bool(defaults.ConcurrentLabel, false, "Runs multiple sandbox deployments concurrently")
	deployCmd.PersistentFlags().Bool(defaults.EnableGeneralLogLabel, false, "Enables general log for the sandbox (MySQL 5.1+)")
	deployCmd.PersistentFlags().Bool(defaults.InitGeneralLogLabel, false, "uses general log during initialization (MySQL 5.1+)")

	set_pflag(deployCmd,defaults.RemoteAccessLabel, "", "", defaults.RemoteAccessValue, "defines the database access ", false)
	set_pflag(deployCmd,defaults.BindAddressLabel, "", "", defaults.BindAddressValue, "defines the database bind-address ", false)
	set_pflag(deployCmd,defaults.CustomMysqldLabel, "", "", "", "Uses an alternative mysqld (must be in the same directory as regular mysqld)", false)
	set_pflag(deployCmd,defaults.BinaryVersionLabel, "", "", "", "Specifies the version when the basedir directory name does not contain it (i.e. it is not x.x.xx)", false)
	set_pflag(deployCmd,defaults.DefaultsLabel, "", "", "", "Change defaults on-the-fly (--defaults=label:value)", true)
	set_pflag(deployCmd,defaults.InitOptionsLabel, "i", "INIT_OPTIONS", "", "mysqld options to run during initialization", true)
	set_pflag(deployCmd,defaults.MyCnfOptionsLabel, "c", "MY_CNF_OPTIONS", "", "mysqld options to add to my.sandbox.cnf", true)
	set_pflag(deployCmd,defaults.PreGrantsSqlFileLabel, "", "", "", "SQL file to run before loading grants", false)
	set_pflag(deployCmd,defaults.PreGrantsSqlLabel, "", "", "", "SQL queries to run before loading grants", true)
	set_pflag(deployCmd,defaults.PostGrantsSqlLabel, "", "", "", "SQL queries to run after loading grants", true)
	set_pflag(deployCmd,defaults.PostGrantsSqlFileLabel, "", "", "", "SQL file to run after loading grants", false)
	// This option will allow to merge the template with an external my.cnf
	// The options that are essential for the sandbox will be preserved
	set_pflag(deployCmd,defaults.MyCnfFileLabel, "", "MY_CNF_FILE", "", "Alternative source file for my.sandbox.cnf", false)
	set_pflag(deployCmd,defaults.DbUserLabel, "u", "", defaults.DbUserValue, "database user", false)
	set_pflag(deployCmd,defaults.RplUserLabel, "", "", defaults.RplUserValue, "replication user", false)
	set_pflag(deployCmd,defaults.DbPasswordLabel, "p", "", defaults.DbPasswordValue, "database password", false)
	set_pflag(deployCmd,defaults.RplPasswordLabel, "", "", defaults.RplPasswordValue, "replication password", false)
	set_pflag(deployCmd,defaults.UseTemplateLabel, "", "", "", "[template_name:file_name] Replace existing template with one from file", true)
	set_pflag(deployCmd,defaults.SandboxDirectoryLabel, "", "", "", "Changes the default sandbox directory", false)
	set_pflag(deployCmd,defaults.HistoryDirLabel, "", "", "", "Where to store mysql client history (default: in sandbox directory)", false)
}
