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
	"dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func FillSdef (cmd *cobra.Command, args []string) sandbox.SandboxDef {
	var sd sandbox.SandboxDef
	flags := cmd.Flags()
	sd.Port = sandbox.VersionToPort(args[0])
	sd.Version = args[0]
	sd.Basedir, _ = flags.GetString("sandbox-binary")
	sd.SandboxDir, _ = flags.GetString("sandbox-home")
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
	gtid, _ = flags.GetBool("gtid")
	if gtid {
		sd.GtidOptions = sandbox.GtidOptions
		sd.ReplOptions = sandbox.ReplOptions
		sd.ServerId = sd.Port
	}
	return sd
}

func SingleSandbox(cmd *cobra.Command, args []string) {
	var sd sandbox.SandboxDef
	// fmt.Println("SINGLE")
	// fmt.Printf("Cmd: %#v\n", cmd)
	// fmt.Printf("\nArgs: %#v\n", args)
	sd = FillSdef(cmd, args)
	sandbox.CreateSingleSandbox(sd, args[0])
}

// singleCmd represents the single command
var singleCmd = &cobra.Command{
	Use:   "single MySQL-Version",
	Args:  cobra.ExactArgs(1),
	Short: "deploys a single sandbox",
	Long:  `single installs a sandbox and creates useful scripts for its use.
MySQL-Version is in the format x.x.xx, and it refers to a directory named after the version
containing an unpacked tarball. The place where these directories are found is defined by 
--sandbox-binary (default: $HOME/opt/mysql.)
For example:
	dbdeployer single 5.7.21

For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing 
the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
`,
	Run:   SingleSandbox,
}

func init() {
	rootCmd.AddCommand(singleCmd)

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
