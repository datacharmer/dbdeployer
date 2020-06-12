// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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
	"math/rand"
	"os"
	"path"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/spf13/cobra"
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
		// #nosec G404 We don't care if the number isn't really random
		_ = os.Setenv("MYSQL_TEST_LOGIN_FILE", fmt.Sprintf("/tmp/dont_break_my_sandboxes%d", rand.Int()))
	}
	rootCmd.AddCommand(deployCmd)
	deployCmd.PersistentFlags().Int(globals.PortLabel, 0, "Overrides default port")
	deployCmd.PersistentFlags().Int(globals.BasePortLabel, 0, "Overrides default base-port (for multiple sandboxes)")
	deployCmd.PersistentFlags().Int(globals.BaseServerIdLabel, 0, "Overrides default server_id (for multiple sandboxes)")
	deployCmd.PersistentFlags().Bool(globals.GtidLabel, false, "enables GTID")
	deployCmd.PersistentFlags().Bool(globals.ReplCrashSafeLabel, false, "enables Replication crash safe")
	deployCmd.PersistentFlags().Bool(globals.NativeAuthPluginLabel, false, "in 8.0.4+, uses the native password auth plugin")
	deployCmd.PersistentFlags().Bool(globals.KeepServerUuidLabel, false, "Does not change the server UUID")
	deployCmd.PersistentFlags().Bool(globals.ForceLabel, false, "If a destination sandbox already exists, it will be overwritten")
	deployCmd.PersistentFlags().Bool(globals.SkipStartLabel, false, "Does not start the database server")
	deployCmd.PersistentFlags().Bool(globals.DisableMysqlXLabel, false, "Disable MySQLX plugin (8.0.11+)")
	deployCmd.PersistentFlags().Bool(globals.EnableMysqlXLabel, false, "Enables MySQLX plugin (5.7.12+)")
	deployCmd.PersistentFlags().Bool(globals.EnableAdminAddressLabel, false, "Enables admin address (8.0.14+)")
	deployCmd.PersistentFlags().Bool(globals.SkipLoadGrantsLabel, false, "Does not load the grants")
	deployCmd.PersistentFlags().Bool(globals.SkipReportHostLabel, false, "Does not include report host in my.sandbox.cnf")
	deployCmd.PersistentFlags().Bool(globals.SkipReportPortLabel, false, "Does not include report port in my.sandbox.cnf")
	deployCmd.PersistentFlags().Bool(globals.ExposeDdTablesLabel, false, "In MySQL 8.0+ shows data dictionary tables")
	deployCmd.PersistentFlags().Bool(globals.ConcurrentLabel, false, "Runs multiple sandbox deployments concurrently")
	deployCmd.PersistentFlags().Bool(globals.EnableGeneralLogLabel, false, "Enables general log for the sandbox (MySQL 5.1+)")
	deployCmd.PersistentFlags().Bool(globals.InitGeneralLogLabel, false, "uses general log during initialization (MySQL 5.1+)")
	deployCmd.PersistentFlags().Bool(globals.LogSBOperationsLabel, defaults.LogSBOperations, "Logs sandbox operations to a file")
	deployCmd.PersistentFlags().Bool(globals.SocketInDatadirLabel, false, "Create socket in datadir instead of $TMPDIR")
	deployCmd.PersistentFlags().Bool(globals.FlavorInPromptLabel, false, "Add flavor values to prompt")
	deployCmd.PersistentFlags().Bool(globals.PortAsServerIdLabel, false, "Use the port number as server ID")

	setPflag(deployCmd, globals.LogLogDirectoryLabel, "", "", defaults.Defaults().LogDirectory, "Where to store dbdeployer logs", false)
	setPflag(deployCmd, globals.RemoteAccessLabel, "", "", globals.RemoteAccessValue, "defines the database access ", false)
	setPflag(deployCmd, globals.BindAddressLabel, "", "", globals.BindAddressValue, "defines the database bind-address ", false)
	setPflag(deployCmd, globals.CustomMysqldLabel, "", "", "", "Uses an alternative mysqld (must be in the same directory as regular mysqld)", false)
	setPflag(deployCmd, globals.BinaryVersionLabel, "", "", "", "Specifies the version when the basedir directory name does not contain it (i.e. it is not x.x.xx)", false)
	setPflag(deployCmd, globals.DefaultsLabel, "", "", "", "Change defaults on-the-fly (--defaults=label:value)", true)
	setPflag(deployCmd, globals.InitOptionsLabel, "i", "INIT_OPTIONS", "", "mysqld options to run during initialization", true)
	setPflag(deployCmd, globals.MyCnfOptionsLabel, "c", "MY_CNF_OPTIONS", "", "mysqld options to add to my.sandbox.cnf", true)
	setPflag(deployCmd, globals.PreGrantsSqlFileLabel, "", "", "", "SQL file to run before loading grants", false)
	setPflag(deployCmd, globals.PreGrantsSqlLabel, "", "", "", "SQL queries to run before loading grants", true)
	setPflag(deployCmd, globals.PostGrantsSqlLabel, "", "", "", "SQL queries to run after loading grants", true)
	setPflag(deployCmd, globals.PostGrantsSqlFileLabel, "", "", "", "SQL file to run after loading grants", false)
	// This option will allow to merge the template with an external my.cnf
	// The options that are essential for the sandbox will be preserved
	setPflag(deployCmd, globals.MyCnfFileLabel, "", "MY_CNF_FILE", "", "Alternative source file for my.sandbox.cnf", false)
	setPflag(deployCmd, globals.DbUserLabel, "u", "", globals.DbUserValue, "database user", false)
	setPflag(deployCmd, globals.RplUserLabel, "", "", globals.RplUserValue, "replication user", false)
	setPflag(deployCmd, globals.DbPasswordLabel, "p", "", globals.DbPasswordValue, "database password", false)
	setPflag(deployCmd, globals.RplPasswordLabel, "", "", globals.RplPasswordValue, "replication password", false)
	setPflag(deployCmd, globals.UseTemplateLabel, "", "", "", "[template_name:file_name] Replace existing template with one from file", true)
	setPflag(deployCmd, globals.SandboxDirectoryLabel, "", "", "", "Changes the default name of the sandbox directory", false)
	setPflag(deployCmd, globals.HistoryDirLabel, "", "", "", "Where to store mysql client history (default: in sandbox directory)", false)
	setPflag(deployCmd, globals.FlavorLabel, "", "", "", "Defines the tarball flavor (MySQL, NDB, Percona Server, etc)", false)
	setPflag(deployCmd, globals.ClientFromLabel, "", "", "", "Where to get the client binaries from", false)
	setPflag(deployCmd, globals.DefaultRoleLabel, "", "", "R_DO_IT_ALL", "Which role to assign to default user (8.0+)", false)
	setPflag(deployCmd, globals.TaskUserLabel, "", "", "", "Task user to be created (8.0+)", false)
	setPflag(deployCmd, globals.TaskUserRoleLabel, "", "", "", "Role to be assigned to task user (8.0+)", false)
	setPflag(deployCmd, globals.CustomRoleNameLabel, "", "", "R_CUSTOM", "Name for custom role (8.0+)", false)
	setPflag(deployCmd, globals.CustomRolePrivilegesLabel, "", "", "ALL PRIVILEGES", "Privileges for custom role (8.0+)", false)
	setPflag(deployCmd, globals.CustomRoleTargetLabel, "", "", "*.*", "Target for custom role (8.0+)", false)
	setPflag(deployCmd, globals.CustomRoleExtraLabel, "", "", "WITH GRANT OPTION", "Extra instructions for custom role (8.0+)", false)
}
