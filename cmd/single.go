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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
	"fmt"
	"os"
	"regexp"
)

func replace_template(template_name string, file_name string) {
	group, contents := FindTemplate(template_name)
	if ! common.FileExists(file_name) {
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
		Description : sandbox.AllTemplates[group][template_name].Description,
		Notes : sandbox.AllTemplates[group][template_name].Notes,
		Contents : new_contents,
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

func FillSdef (cmd *cobra.Command, args []string) sandbox.SandboxDef {
	var sd sandbox.SandboxDef

	flags := cmd.Flags()

	template_requests, _ := flags.GetStringSlice("use-template")
	for _, request := range template_requests {
		tname, fname := check_template_change_request(request)
		replace_template(tname, fname)
	}
	sd.Port = sandbox.VersionToPort(args[0])
	sd.Version = args[0]
	sd.Basedir, _ = flags.GetString("sandbox-binary")
	sd.SandboxDir, _ = flags.GetString("sandbox-home")
	common.CheckSandboxDir(sd.SandboxDir)
	sd.LoadGrants = true
	sd.DbUser, _ = flags.GetString("db-user")
	sd.DbPassword, _ = flags.GetString("db-password")
	sd.RplUser, _ = flags.GetString("rpl-user")
	sd.RplPassword, _ = flags.GetString("rpl-password")
	sd.RemoteAccess, _ = flags.GetString("remote-access")
	sd.BindAddress, _ = flags.GetString("bind-address")
	sd.InitOptions, _ = flags.GetStringSlice("init-options")
	sd.MyCnfOptions, _ = flags.GetStringSlice("my-cnf-options")
	sd.KeepAuthPlugin, _ = flags.GetBool("keep-auth-plugin")

	var gtid bool
	var master bool
	master, _ = flags.GetBool("master")
	gtid, _ = flags.GetBool("gtid")
	if master {
		sd.ReplOptions = sandbox.ReplOptions
		sd.ServerId = sd.Port
	}
	if gtid {
		if sandbox.GreaterOrEqualVersion(sd.Version, []int{5, 6, 9}) {
			sd.GtidOptions = sandbox.GtidOptions
			sd.ReplOptions = sandbox.ReplOptions
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
	// fmt.Println("SINGLE")
	// fmt.Printf("Cmd: %#v\n", cmd)
	// fmt.Printf("\nArgs: %#v\n", args)
	common.CheckOrigin(args)
	sd = FillSdef(cmd, args)
	sandbox.CreateSingleSandbox(sd, args[0])
}

// singleCmd represents the single command
var singleCmd = &cobra.Command{
	Use:   "single MySQL-Version",
	// Args:  cobra.ExactArgs(1),
	Short: "deploys a single sandbox",
	Long:  `single installs a sandbox and creates useful scripts for its use.
MySQL-Version is in the format x.x.xx, and it refers to a directory named after the version
containing an unpacked tarball. The place where these directories are found is defined by 
--sandbox-binary (default: $HOME/opt/mysql.)
For example:
	dbdeployer single 5.7.21

For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
Use the "unpack" command to get the tarball into the right directory.
`,
	Run:   SingleSandbox,
}

func init() {
	rootCmd.AddCommand(singleCmd)
	singleCmd.PersistentFlags().Bool("master",  false, "Make the server replication ready")

}

/*
	Cmd: &cobra.Command{Use:"single",
	Aliases:[]string(nil),
	Short:"deploys a single sandbox",
	Long:"Installs a sandbox and creates useful scripts for its use.",
	Example:"",
	ValidArgs:[]string(nil),
	Args:(cobra.PositionalArgs)(0x1194300),
	ArgAliases:[]string(nil),
	BashCompletionFunction:"",
	Deprecated:"",
	Hidden:false,
	Annotations:map[string]string(nil),
	Version:"",
	PersistentPreRun:(func(*cobra.Command, []string))(nil),
	PersistentPreRunE:(func(*cobra.Command, []string) error)(nil),
	PreRun:(func(*cobra.Command, []string))(nil),
	PreRunE:(func(*cobra.Command, []string) error)(nil),
	Run:(func(*cobra.Command, []string))(0x138fda0),
	RunE:(func(*cobra.Command, []string) error)(nil),
	PostRun:(func(*cobra.Command, []string))(nil),
	PostRunE:(func(*cobra.Command, []string) error)(nil),
	PersistentPostRun:(func(*cobra.Command, []string))(nil),
	PersistentPostRunE:(func(*cobra.Command, []string) error)(nil),
	SilenceErrors:false,
	SilenceUsage:false,
	DisableFlagParsing:false,
	DisableAutoGenTag:false,
	DisableFlagsInUseLine:false,
	DisableSuggestions:false,
	SuggestionsMinimumDistance:0,
	TraverseChildren:false,
	commands:[]*cobra.Command(nil),
	parent:(*cobra.Command)(0x16a3140),
	commandsMaxUseLen:0,
	commandsMaxCommandPathLen:0,
	commandsMaxNameLen:0,
	commandsAreSorted:false,
	args:[]string(nil),
	flagErrorBuf:(*bytes.Buffer)(0xc420125810),
	flags:(*pflag.FlagSet)(0xc42008c5a0),
	pflags:(*pflag.FlagSet)(0xc42008c690),
	lflags:(*pflag.FlagSet)(nil),
	iflags:(*pflag.FlagSet)(nil),
	parentsPflags:(*pflag.FlagSet)(0xc42008c4b0),
	globNormFunc:(func(*pflag.FlagSet, string) pflag.NormalizedName)(nil),
	output:io.Writer(nil),
	usageFunc:(func(*cobra.Command) error)(nil),
	usageTemplate:"",
	flagErrorFunc:(func(*cobra.Command, error) error)(nil),
	helpTemplate:"",
	helpFunc:(func(*cobra.Command, []string))(nil),
	helpCommand:(*cobra.Command)(nil),
	versionTemplate:""}
*/
