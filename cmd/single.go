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
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func replace_template(template_name string, file_name string) {
	group, _, contents := FindTemplate(template_name)
	if !common.FileExists(file_name) {
		common.Exit(1, fmt.Sprintf("File %s not found\n", file_name))
	}
	fmt.Printf("Replacing template %s.%s [%d chars] with contents of file %s\n", group, template_name, len(contents), file_name)
	new_contents := common.SlurpAsString(file_name)
	if len(new_contents) == 0 {
		common.Exit(1, fmt.Sprintf("File %s is empty\n", file_name))
	}
	var new_rec sandbox.TemplateDesc = sandbox.TemplateDesc{
		Description: sandbox.AllTemplates[group][template_name].Description,
		Notes:       sandbox.AllTemplates[group][template_name].Notes,
		Contents:    new_contents,
	}
	sandbox.AllTemplates[group][template_name] = new_rec
}

func check_template_change_request(request string) (template_name, file_name string) {
	re := regexp.MustCompile(`(\w+):(\S+)`)
	reqList := re.FindAllStringSubmatch(request, -1)
	if len(reqList) == 0 {
		common.Exit(1, fmt.Sprintf("request '%s' invalid. Required format is 'template_name:file_name'", request))
	}
	template_name = reqList[0][1]
	file_name = reqList[0][2]
	return
}

func process_defaults(new_defaults []string) {
	for _, nd := range new_defaults {
		list := strings.Split(nd, ":")
		if list != nil && len(list) == 2 {
			label := list[0]
			value := list[1]
			defaults.UpdateDefaults(label, value, false)
		}
	}
}

func GetAbsolutePathFromFlag(cmd *cobra.Command, name string) string {
	flags := cmd.Flags()
	value, err := flags.GetString(name)
	if err != nil {
		common.Exit(1, fmt.Sprintf("Error getting flag value for --%s", name))
	}
	value, err = filepath.Abs(value)
	if err != nil {
		common.Exit(1, fmt.Sprintf("Error getting absolute path for %s", value))
	}
	return value
}

func FillSdef(cmd *cobra.Command, args []string) sandbox.SandboxDef {
	var sd sandbox.SandboxDef

	flags := cmd.Flags()
	template_requests, _ := flags.GetStringSlice("use-template")
	for _, request := range template_requests {
		tname, fname := check_template_change_request(request)
		replace_template(tname, fname)
	}
	sd.BasedirName = args[0]
	sd.Version, _ = flags.GetString("binary-version")
	if sd.Version == "" {
		sd.Version = args[0]
	}

	sd.Port = common.VersionToPort(sd.Version)
	if sd.Port < 0 {
		common.Exit(1, fmt.Sprintf("Unsupported version format (%s)", sd.Version))
	}
	sd.UserPort, _ = flags.GetInt("port")
	sd.BasePort, _ = flags.GetInt("base-port")
	sd.DirName, _ = flags.GetString("sandbox-directory")

	if sd.UserPort > 0 {
		sd.Port = sd.UserPort
	}

	basedir := GetAbsolutePathFromFlag(cmd, "sandbox-binary")

	// sd.Basedir = path.Join(basedir, sd.Version)
	sd.Basedir = path.Join(basedir, args[0])
	if !common.DirExists(sd.Basedir) {
		common.Exit(1, fmt.Sprintf("basedir '%s' not found", sd.Basedir))
	}
	sd.SandboxDir = GetAbsolutePathFromFlag(cmd, "sandbox-home")

	common.CheckSandboxDir(sd.SandboxDir)
	sd.InstalledPorts = common.GetInstalledPorts(sd.SandboxDir)
	for _, p := range defaults.Defaults().ReservedPorts {
		sd.InstalledPorts = append(sd.InstalledPorts, p)
	}
	sd.LoadGrants = true
	sd.SkipStart, _ = flags.GetBool("skip-start")
	skip_load_grants, _ := flags.GetBool("skip-load-grants")
	if skip_load_grants || sd.SkipStart {
		sd.LoadGrants = false
	}
	sd.SkipReportHost, _ = flags.GetBool("skip-report-host")
	sd.SkipReportPort, _ = flags.GetBool("skip-report-port")
	sd.DisableMysqlX, _ = flags.GetBool("disable-mysqlx")
	sd.EnableMysqlX, _ = flags.GetBool("enable-mysqlx")
	sd.DbUser, _ = flags.GetString("db-user")
	sd.DbPassword, _ = flags.GetString("db-password")
	sd.RplUser, _ = flags.GetString("rpl-user")
	sd.RplPassword, _ = flags.GetString("rpl-password")
	sd.RemoteAccess, _ = flags.GetString("remote-access")
	sd.BindAddress, _ = flags.GetString("bind-address")
	sd.CustomMysqld, _ = flags.GetString("custom-mysqld")
	sd.InitOptions, _ = flags.GetStringSlice("init-options")
	sd.MyCnfOptions, _ = flags.GetStringSlice("my-cnf-options")
	sd.PreGrantsSqlFile, _ = flags.GetString("pre-grants-sql-file")
	sd.PreGrantsSql, _ = flags.GetStringSlice("pre-grants-sql")
	sd.PostGrantsSql, _ = flags.GetStringSlice("post-grants-sql")
	sd.PostGrantsSqlFile, _ = flags.GetString("post-grants-sql-file")
	sd.MyCnfFile, _ = flags.GetString("my-cnf-file")
	sd.NativeAuthPlugin, _ = flags.GetBool("native-auth-plugin")
	sd.KeepUuid, _ = flags.GetBool("keep-server-uuid")
	sd.Force, _ = flags.GetBool("force")
	sd.ExposeDdTables, _ = flags.GetBool("expose-dd-tables")
	sd.InitGeneralLog, _ = flags.GetBool("init-general-log")
	sd.EnableGeneralLog, _ = flags.GetBool("enable-general-log")

	if sd.DisableMysqlX && sd.EnableMysqlX {
		common.Exit(1, "flags --enable-mysqlx and --disable-mysqlx cannot be used together")
	}
	sd.RunConcurrently, _ = flags.GetBool("concurrent")
	if os.Getenv("RUN_CONCURRENTLY") != "" {
		sd.RunConcurrently = true
	}

	new_defaults, _ := flags.GetStringSlice("defaults")
	process_defaults(new_defaults)

	var gtid bool
	var master bool
	master, _ = flags.GetBool("master")
	gtid, _ = flags.GetBool("gtid")
	if master {
		sd.ReplOptions = sandbox.SingleTemplates["replication_options"].Contents
		sd.ServerId = sd.Port
	}
	if gtid {
		if common.GreaterOrEqualVersion(sd.Version, []int{5, 6, 9}) {
			sd.GtidOptions = sandbox.SingleTemplates["gtid_options"].Contents
			sd.ReplOptions = sandbox.SingleTemplates["replication_options"].Contents
			sd.ServerId = sd.Port
		} else {
			common.Exit(1, "--gtid requires version 5.6.9+")
		}
	}
	return sd
}

func SingleSandbox(cmd *cobra.Command, args []string) {
	var sd sandbox.SandboxDef
	common.CheckOrigin(args)
	sd = FillSdef(cmd, args)
	// When deploying a single sandbox, we disable concurrency
	sd.RunConcurrently = false
	sandbox.CreateSingleSandbox(sd)
}

var singleCmd = &cobra.Command{
	Use: "single MySQL-Version",
	// Args:  cobra.ExactArgs(1),
	Short: "deploys a single sandbox",
	Long: `single installs a sandbox and creates useful scripts for its use.
MySQL-Version is in the format x.x.xx, and it refers to a directory named after the version
containing an unpacked tarball. The place where these directories are found is defined by 
--sandbox-binary (default: $HOME/opt/mysql.)
For example:
	dbdeployer deploy single 5.7.21

For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
Use the "unpack" command to get the tarball into the right directory.
`,
	Run: SingleSandbox,
}

func init() {
	deployCmd.AddCommand(singleCmd)
	singleCmd.PersistentFlags().Bool("master", false, "Make the server replication ready")

}
