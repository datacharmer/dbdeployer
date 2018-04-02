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
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
)

func replace_template(template_name string, file_name string) {
	group, _, contents := FindTemplate(template_name)
	if !common.FileExists(file_name) {
		fmt.Printf("File %s not found\n", file_name)
		os.Exit(1)
	}
	fmt.Printf("Replacing template %s.%s [%d chars] with contents of file %s\n", group, template_name, len(contents), file_name)
	new_contents := common.SlurpAsString(file_name)
	if len(new_contents) == 0 {
		fmt.Printf("File %s is empty\n", file_name)
		os.Exit(1)
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
		//fmt.Printf("%v\n", reqList)
		fmt.Printf("request '%s' invalid. Required format is 'template_name:file_name'\n", request)
		os.Exit(1)
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

func FillSdef(cmd *cobra.Command, args []string) sandbox.SandboxDef {
	var sd sandbox.SandboxDef

	flags := cmd.Flags()
	template_requests, _ := flags.GetStringSlice("use-template")
	for _, request := range template_requests {
		tname, fname := check_template_change_request(request)
		replace_template(tname, fname)
	}
	sd.Port = common.VersionToPort(args[0])

	sd.UserPort, _ = flags.GetInt("port")
	sd.BasePort, _ = flags.GetInt("base-port")
	sd.DirName, _ = flags.GetString("sandbox-directory")
	if sd.UserPort > 0 {
		sd.Port = sd.UserPort
	}

	sd.Version = args[0]
	sd.Basedir, _ = flags.GetString("sandbox-binary")
	sd.SandboxDir, _ = flags.GetString("sandbox-home")
	common.CheckSandboxDir(sd.SandboxDir)
	sd.InstalledPorts = common.GetInstalledPorts(sd.SandboxDir)
	sd.LoadGrants = true
	skip_load_grants, _ := flags.GetBool("skip-load-grants")
	if skip_load_grants {
		sd.LoadGrants = false
	}
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
			fmt.Println("--gtid requires version 5.6.9+")
			os.Exit(1)
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
	sandbox.CreateSingleSandbox(sd, args[0])
}

/*
func ReplacedCmd(cmd *cobra.Command, args []string) {
	invoked := cmd.Use
	fmt.Printf("The command \"%s\" has been replaced.\n",invoked)
	fmt.Printf("Use \"dbdeployer deploy %s\" instead.\n",invoked)
	os.Exit(0)
}
*/

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

/*
var (
	hiddenSingleCmd = &cobra.Command{
		Use: "single",
		Short: "REMOVED: use 'deploy single' instead",
		Hidden: true,
		Run: ReplacedCmd,
	}
	hiddenReplicationCmd = &cobra.Command{
		Use: "replication",
		Short: "REMOVED: use 'deploy replication' instead",
		Hidden: true,
		Run: ReplacedCmd,
	}

	hiddenMultipleCmd = &cobra.Command{
		Use: "multiple",
		Short: "REMOVED: use 'deploy multiple' instead",
		Hidden: true,
		Run: ReplacedCmd,
	}
)
*/

func init() {
	//rootCmd.AddCommand(hiddenSingleCmd)
	//rootCmd.AddCommand(hiddenReplicationCmd)
	//rootCmd.AddCommand(hiddenMultipleCmd)
	deployCmd.AddCommand(singleCmd)
	singleCmd.PersistentFlags().Bool("master", false, "Make the server replication ready")

}
