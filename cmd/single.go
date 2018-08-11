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
	return common.AbsolutePath(value)
}

func check_if_abridged_version(version, basedir string) string {
	full_pattern := regexp.MustCompile(`\d\.\d+\.\d+$`)
	if full_pattern.MatchString(version) {
		return version
	}
	valid_pattern := regexp.MustCompile(`\d+\.\d+$`)
	if !valid_pattern.MatchString(version) {
		return version
	}
	full_version := common.LatestVersion(basedir, version)
	if full_version == "" {
		common.Exit(1, fmt.Sprintf("FATAL: no full version found for %s in %s\n", version, basedir))
	} else {
		fmt.Printf("# %s => %s\n", version, full_version)
		version = full_version
	}
	return version
}

func FillSdef(cmd *cobra.Command, args []string) sandbox.SandboxDef {
	var sd sandbox.SandboxDef

	flags := cmd.Flags()
	template_requests, _ := flags.GetStringSlice(defaults.UseTemplateLabel)
	for _, request := range template_requests {
		tname, fname := check_template_change_request(request)
		replace_template(tname, fname)
	}

	basedir := GetAbsolutePathFromFlag(cmd, defaults.SandboxBinaryLabel)

	sd.BasedirName = args[0]
	version_from_option := false
	sd.Version, _ = flags.GetString(defaults.BinaryVersionLabel)
	if sd.Version == "" {
		sd.Version = args[0]
		old_version := sd.Version
		sd.Version = check_if_abridged_version(sd.Version, basedir)
		if old_version != sd.Version {
			sd.BasedirName = sd.Version
		}
	} else {
		version_from_option = true
	}

	if common.DirExists(sd.BasedirName) {
		sd.BasedirName = common.RemoveTrailingSlash(sd.BasedirName)
		sd.BasedirName = common.AbsolutePath(sd.BasedirName)
		// fmt.Printf("OLD bd <%s> - v: <%s>\n",basedir, sd.Version )
		target := sd.BasedirName
		old_basedir := basedir
		basedir = common.DirName(sd.BasedirName)
		if old_basedir != defaults.Defaults().SandboxBinary {
			// basedir was set using either an environment variable
			// or a command line option
			if old_basedir != basedir {
				// The new basedir is different from the one given by command line or env
				common.Exit(1, "The Sandbox Binary directory was set twice,",
					fmt.Sprintf(" using conflicting values: '%s' and '%s' ", old_basedir, basedir))
			}
		}
		sd.BasedirName = common.BaseName(sd.BasedirName)
		if !version_from_option {
			sd.Version = sd.BasedirName
		}
		if !common.IsVersion(sd.Version) {
			common.Exit(1, fmt.Sprintf("No version detected for directory %s", target))
		}
		// fmt.Printf("NEW bd <%s> - v: <%s>\n",basedir, sd.Version )
	}

	sd.Port = common.VersionToPort(sd.Version)
	if sd.Port < 0 {
		common.Exit(1, fmt.Sprintf("Unsupported version format (%s)", sd.Version))
	}
	sd.UserPort, _ = flags.GetInt(defaults.PortLabel)
	sd.BasePort, _ = flags.GetInt(defaults.BasePortLabel)
	sd.DirName, _ = flags.GetString(defaults.SandboxDirectoryLabel)

	if sd.UserPort > 0 {
		sd.Port = sd.UserPort
	}

	sd.Basedir = path.Join(basedir, sd.BasedirName)
	// sd.Basedir = path.Join(basedir, args[0])
	if !common.DirExists(sd.Basedir) {
		common.Exit(1, fmt.Sprintf("basedir '%s' not found", sd.Basedir))
	}

	common.CheckTarballOperatingSystem(sd.Basedir)

	sd.SandboxDir = GetAbsolutePathFromFlag(cmd, defaults.SandboxHomeLabel)

	common.CheckSandboxDir(sd.SandboxDir)
	sd.InstalledPorts = common.GetInstalledPorts(sd.SandboxDir)
	for _, p := range defaults.Defaults().ReservedPorts {
		sd.InstalledPorts = append(sd.InstalledPorts, p)
	}
	sd.LoadGrants = true
	sd.SkipStart, _ = flags.GetBool(defaults.SkipStartLabel)
	skip_load_grants, _ := flags.GetBool(defaults.SkipLoadGrantsLabel)
	if skip_load_grants || sd.SkipStart {
		sd.LoadGrants = false
	}
	sd.SkipReportHost, _ = flags.GetBool(defaults.SkipReportHostLabel)
	sd.SkipReportPort, _ = flags.GetBool(defaults.SkipReportPortLabel)
	sd.DisableMysqlX, _ = flags.GetBool(defaults.DisableMysqlXLabel)
	sd.EnableMysqlX, _ = flags.GetBool(defaults.EnableMysqlXLabel)
	sd.HistoryDir, _ = flags.GetString(defaults.HistoryDirLabel)
	sd.DbUser, _ = flags.GetString(defaults.DbUserLabel)
	sd.DbPassword, _ = flags.GetString(defaults.DbPasswordLabel)
	sd.RplUser, _ = flags.GetString(defaults.RplUserLabel)

	sd.RplPassword, _ = flags.GetString(defaults.RplPasswordLabel)
	sd.RemoteAccess, _ = flags.GetString(defaults.RemoteAccessLabel)
	sd.BindAddress, _ = flags.GetString(defaults.BindAddressLabel)
	sd.CustomMysqld, _ = flags.GetString(defaults.CustomMysqldLabel)
	sd.InitOptions, _ = flags.GetStringSlice(defaults.InitOptionsLabel)
	sd.MyCnfOptions, _ = flags.GetStringSlice(defaults.MyCnfOptionsLabel)
	sd.PreGrantsSqlFile, _ = flags.GetString(defaults.PreGrantsSqlFileLabel)
	sd.PreGrantsSql, _ = flags.GetStringSlice(defaults.PreGrantsSqlLabel)
	sd.PostGrantsSql, _ = flags.GetStringSlice(defaults.PostGrantsSqlLabel)
	sd.PostGrantsSqlFile, _ = flags.GetString(defaults.PostGrantsSqlFileLabel)
	sd.MyCnfFile, _ = flags.GetString(defaults.MyCnfFileLabel)
	sd.NativeAuthPlugin, _ = flags.GetBool(defaults.NativeAuthPluginLabel)
	sd.KeepUuid, _ = flags.GetBool(defaults.KeepServerUuidLabel)
	sd.Force, _ = flags.GetBool(defaults.ForceLabel)
	sd.ExposeDdTables, _ = flags.GetBool(defaults.ExposeDdTablesLabel)
	sd.InitGeneralLog, _ = flags.GetBool(defaults.InitGeneralLogLabel)
	sd.EnableGeneralLog, _ = flags.GetBool(defaults.EnableGeneralLogLabel)

	if sd.DisableMysqlX && sd.EnableMysqlX {
		common.Exit(1, "flags --enable-mysqlx and --disable-mysqlx cannot be used together")
	}
	sd.RunConcurrently, _ = flags.GetBool(defaults.ConcurrentLabel)
	if os.Getenv("RUN_CONCURRENTLY") != "" {
		sd.RunConcurrently = true
	}

	new_defaults, _ := flags.GetStringSlice(defaults.DefaultsLabel)
	process_defaults(new_defaults)

	var gtid bool
	var master bool
	master, _ = flags.GetBool(defaults.MasterLabel)
	gtid, _ = flags.GetBool(defaults.GtidLabel)
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
	dbdeployer deploy single 5.7     # deploys the latest release of 5.7.x
	dbdeployer deploy single 5.7.21  # deploys a specific release
	dbdeployer deploy single /path/to/5.7.21  # deploys a specific release in a given path

For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
Use the "unpack" command to get the tarball into the right directory.
`,
	Run: SingleSandbox,
}

func init() {
	deployCmd.AddCommand(singleCmd)
	singleCmd.PersistentFlags().Bool(defaults.MasterLabel, false, "Make the server replication ready")

}
