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
	"path"
	"regexp"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func replaceTemplate(templateName string, fileName string) {
	group, _, contents := FindTemplate(templateName)
	if !common.FileExists(fileName) {
		common.Exitf(1, defaults.ErrFileNotFound, fileName)
	}
	fmt.Printf("Replacing template %s.%s [%d chars] with contents of file %s\n", group, templateName, len(contents), fileName)
	newContents := common.SlurpAsString(fileName)
	if len(newContents) == 0 {
		common.Exitf(1, "file %s is empty\n", fileName)
	}
	var newRec = sandbox.TemplateDesc{
		Description: sandbox.AllTemplates[group][templateName].Description,
		Notes:       sandbox.AllTemplates[group][templateName].Notes,
		Contents:    newContents,
	}
	sandbox.AllTemplates[group][templateName] = newRec
}

func checkTemplateChangeRequest(request string) (templateName, fileName string) {
	re := regexp.MustCompile(`(\w+):(\S+)`)
	reqList := re.FindAllStringSubmatch(request, -1)
	if len(reqList) == 0 {
		common.Exitf(1, "request '%s' invalid. Required format is 'template_name:file_name'", request)
	}
	templateName = reqList[0][1]
	fileName = reqList[0][2]
	return
}

func processDefaults(newDefaults []string) {
	for _, nd := range newDefaults {
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
	common.ErrCheckExitf(err, 1, "error getting flag value for --%s", name)
	return common.AbsolutePath(value)
}

func checkIfAbridgedVersion(version, basedir string) string {
	fullPattern := regexp.MustCompile(`\d\.\d+\.\d+$`)
	if fullPattern.MatchString(version) {
		return version
	}
	validPattern := regexp.MustCompile(`\d+\.\d+$`)
	if !validPattern.MatchString(version) {
		return version
	}
	fullVersion := common.LatestVersion(basedir, version)
	if fullVersion == "" {
		common.Exitf(1, "FATAL: no full version found for %s in %s\n", version, basedir)
	} else {
		fmt.Printf("# %s => %s\n", version, fullVersion)
		version = fullVersion
	}
	return version
}

func checkForRootValue(value, label, defaultVal string) {
	if value == "root" {
		common.Exit(1, fmt.Sprintf("option --%s cannot be 'root'", label),
			"The 'root' user will be initialized regardless,",
			"using the same password defined for the default db-user.",
			fmt.Sprintf("The default user for this option is '%s'.", defaultVal))
	}
}

func FillSdef(cmd *cobra.Command, args []string) sandbox.SandboxDef {
	var sd sandbox.SandboxDef

	flags := cmd.Flags()

	logSbOperations, _ := flags.GetBool(defaults.LogSBOperationsLabel)
	defaults.LogSBOperations = logSbOperations

	logDir := GetAbsolutePathFromFlag(cmd, defaults.LogLogDirectoryLabel)
	if logDir != "" {
		defaults.UpdateDefaults(defaults.LogLogDirectoryLabel, logDir, false)
	}

	templateRequests, _ := flags.GetStringSlice(defaults.UseTemplateLabel)
	for _, request := range templateRequests {
		tname, fname := checkTemplateChangeRequest(request)
		replaceTemplate(tname, fname)
	}

	basedir := GetAbsolutePathFromFlag(cmd, defaults.SandboxBinaryLabel)

	sd.BasedirName = args[0]
	versionFromOption := false
	sd.Version, _ = flags.GetString(defaults.BinaryVersionLabel)
	if sd.Version == "" {
		sd.Version = args[0]
		oldVersion := sd.Version
		sd.Version = checkIfAbridgedVersion(sd.Version, basedir)
		if oldVersion != sd.Version {
			sd.BasedirName = sd.Version
		}
	} else {
		versionFromOption = true
	}

	if common.DirExists(sd.BasedirName) {
		sd.BasedirName = common.RemoveTrailingSlash(sd.BasedirName)
		sd.BasedirName = common.AbsolutePath(sd.BasedirName)
		// fmt.Printf("OLD bd <%s> - v: <%s>\n",basedir, sd.Version )
		target := sd.BasedirName
		oldBasedir := basedir
		basedir = common.DirName(sd.BasedirName)
		if oldBasedir != defaults.Defaults().SandboxBinary {
			// basedir was set using either an environment variable
			// or a command line option
			if oldBasedir != basedir {
				// The new basedir is different from the one given by command line or env
				common.Exit(1, "the Sandbox Binary directory was set twice,",
					fmt.Sprintf(" using conflicting values: '%s' and '%s' ", oldBasedir, basedir))
			}
		}
		sd.BasedirName = common.BaseName(sd.BasedirName)
		if !versionFromOption {
			sd.Version = sd.BasedirName
		}
		if !common.IsVersion(sd.Version) {
			common.Exitf(1, "no version detected for directory %s", target)
		}
		// fmt.Printf("NEW bd <%s> - v: <%s>\n",basedir, sd.Version )
	}

	sd.Port = common.VersionToPort(sd.Version)
	if sd.Port < 0 {
		common.Exitf(1, "unsupported version format (%s)", sd.Version)
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
		common.Exitf(1, "basedir '%s' not found", sd.Basedir)
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
	skipLoadGrants, _ := flags.GetBool(defaults.SkipLoadGrantsLabel)
	if skipLoadGrants || sd.SkipStart {
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

	checkForRootValue(sd.DbUser, defaults.DbUserLabel, defaults.DbUserValue)
	checkForRootValue(sd.RplUser, defaults.RplUserLabel, defaults.RplUserValue)

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
	if common.IsEnvSet("RUN_CONCURRENTLY") {
		sd.RunConcurrently = true
	}

	newDefaults, _ := flags.GetStringSlice(defaults.DefaultsLabel)
	processDefaults(newDefaults)

	var gtid bool
	var master bool
	var replCrashSafe bool
	master, _ = flags.GetBool(defaults.MasterLabel)
	gtid, _ = flags.GetBool(defaults.GtidLabel)
	replCrashSafe, _ = flags.GetBool(defaults.ReplCrashSafeLabel)
	if master {
		sd.ReplOptions = sandbox.SingleTemplates["replication_options"].Contents
		sd.ServerId = sd.Port
	}
	if gtid {
		templateName := "gtid_options_56"
		// 5.7.0
		if common.GreaterOrEqualVersion(sd.Version, defaults.MinimumEnhancedGtidVersion) {
			templateName = "gtid_options_57"
		}
		// 5.6.9
		if common.GreaterOrEqualVersion(sd.Version, defaults.MinimumGtidVersion) {
			sd.GtidOptions = sandbox.SingleTemplates[templateName].Contents
			sd.ReplCrashSafeOptions = sandbox.SingleTemplates["repl_crash_safe_options"].Contents
			sd.ReplOptions = sandbox.SingleTemplates["replication_options"].Contents
			sd.ServerId = sd.Port
		} else {
			common.Exitf(1, defaults.ErrOptionRequiresVersion, defaults.GtidLabel, common.IntSliceToDottedString(defaults.MinimumGtidVersion))
		}
	}
	if replCrashSafe && sd.ReplCrashSafeOptions == "" {
		// 5.6.2
		if common.GreaterOrEqualVersion(sd.Version, defaults.MinimumCrashSafeVersion) {
			sd.ReplCrashSafeOptions = sandbox.SingleTemplates["repl_crash_safe_options"].Contents
		} else {
			common.Exitf(1, defaults.ErrOptionRequiresVersion, defaults.ReplCrashSafeLabel, common.IntSliceToDottedString(defaults.MinimumCrashSafeVersion))
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
