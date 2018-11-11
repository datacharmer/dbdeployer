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
	"math/rand"
	"os"
	"path"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy sandboxes",
	Long:  `Deploys single, multiple, or replicated sandboxes`,
}

func init() {
	myloginCnf := path.Join(os.Getenv("HOME"), ".mylogin.cnf")
	if common.FileExists(myloginCnf) {
		// dbdeployer is not compatible with .mylogin.cnf,
		// as it bypasses --defaults-file and --no-defaults.
		// See: https://dev.mysql.com/doc/refman/8.0/en/mysql-config-editor.html
		// The following statement disables .mylogin.cnf
		os.Setenv("MYSQL_TEST_LOGIN_FILE", fmt.Sprintf("/tmp/dont_break_my_sandboxes%d", rand.Int()))
	}
	rootCmd.AddCommand(deployCmd)
	deployCmd.PersistentFlags().Int(defaults.PortLabel, 0, "Overrides default port")
	deployCmd.PersistentFlags().Int(defaults.BasePortLabel, 0, "Overrides default base-port (for multiple sandboxes)")
	deployCmd.PersistentFlags().Bool(defaults.GtidLabel, false, "enables GTID")
	deployCmd.PersistentFlags().Bool(defaults.ReplCrashSafeLabel, false, "enables Replication crash safe")
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
	deployCmd.PersistentFlags().Bool(defaults.LogSBOperationsLabel, defaults.LogSBOperations, "Logs sandbox operations to a file")

	setPflag(deployCmd, defaults.LogLogDirectoryLabel, "", "", defaults.Defaults().LogDirectory, "Where to store dbdeployer logs", false)
	setPflag(deployCmd, defaults.RemoteAccessLabel, "", "", defaults.RemoteAccessValue, "defines the database access ", false)
	setPflag(deployCmd, defaults.BindAddressLabel, "", "", defaults.BindAddressValue, "defines the database bind-address ", false)
	setPflag(deployCmd, defaults.CustomMysqldLabel, "", "", "", "Uses an alternative mysqld (must be in the same directory as regular mysqld)", false)
	setPflag(deployCmd, defaults.BinaryVersionLabel, "", "", "", "Specifies the version when the basedir directory name does not contain it (i.e. it is not x.x.xx)", false)
	setPflag(deployCmd, defaults.DefaultsLabel, "", "", "", "Change defaults on-the-fly (--defaults=label:value)", true)
	setPflag(deployCmd, defaults.InitOptionsLabel, "i", "INIT_OPTIONS", "", "mysqld options to run during initialization", true)
	setPflag(deployCmd, defaults.MyCnfOptionsLabel, "c", "MY_CNF_OPTIONS", "", "mysqld options to add to my.sandbox.cnf", true)
	setPflag(deployCmd, defaults.PreGrantsSqlFileLabel, "", "", "", "SQL file to run before loading grants", false)
	setPflag(deployCmd, defaults.PreGrantsSqlLabel, "", "", "", "SQL queries to run before loading grants", true)
	setPflag(deployCmd, defaults.PostGrantsSqlLabel, "", "", "", "SQL queries to run after loading grants", true)
	setPflag(deployCmd, defaults.PostGrantsSqlFileLabel, "", "", "", "SQL file to run after loading grants", false)
	// This option will allow to merge the template with an external my.cnf
	// The options that are essential for the sandbox will be preserved
	setPflag(deployCmd, defaults.MyCnfFileLabel, "", "MY_CNF_FILE", "", "Alternative source file for my.sandbox.cnf", false)
	setPflag(deployCmd, defaults.DbUserLabel, "u", "", defaults.DbUserValue, "database user", false)
	setPflag(deployCmd, defaults.RplUserLabel, "", "", defaults.RplUserValue, "replication user", false)
	setPflag(deployCmd, defaults.DbPasswordLabel, "p", "", defaults.DbPasswordValue, "database password", false)
	setPflag(deployCmd, defaults.RplPasswordLabel, "", "", defaults.RplPasswordValue, "replication password", false)
	setPflag(deployCmd, defaults.UseTemplateLabel, "", "", "", "[template_name:file_name] Replace existing template with one from file", true)
	setPflag(deployCmd, defaults.SandboxDirectoryLabel, "", "", "", "Changes the default sandbox directory", false)
	setPflag(deployCmd, defaults.HistoryDirLabel, "", "", "", "Where to store mysql client history (default: in sandbox directory)", false)
}
