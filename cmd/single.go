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
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/pkg/errors"
	"os"
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
		common.Exitf(1, globals.ErrFileNotFound, fileName)
	}
	fmt.Printf("Replacing template %s.%s [%d chars] with contents of file %s\n", group, templateName, len(contents), fileName)
	newContents, err := common.SlurpAsString(fileName)
	if err != nil {
		common.Exitf(1, "%+v", err)
	}
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

func GetAbsolutePathFromFlag(cmd *cobra.Command, name string) (string, error) {
	flags := cmd.Flags()
	value, err := flags.GetString(name)
	common.ErrCheckExitf(err, 1, "error getting flag value for --%s", name)
	filePath, err := common.AbsolutePath(value)
	if err == nil {
		return filePath, nil
	} else {
		return "", errors.Wrap(err, "")
	}
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

func FillSdef(cmd *cobra.Command, args []string) (sandbox.SandboxDef, error) {
	var sd sandbox.SandboxDef

	flags := cmd.Flags()

	logSbOperations, _ := flags.GetBool(globals.LogSBOperationsLabel)
	defaults.LogSBOperations = logSbOperations

	logDir, err := GetAbsolutePathFromFlag(cmd, globals.LogLogDirectoryLabel)
	if err != nil {
		return sd, err
	}
	if logDir != "" {
		defaults.UpdateDefaults(globals.LogLogDirectoryLabel, logDir, false)
	}

	templateRequests, _ := flags.GetStringSlice(globals.UseTemplateLabel)
	for _, request := range templateRequests {
		tname, fname := checkTemplateChangeRequest(request)
		replaceTemplate(tname, fname)
	}

	basedir, err := GetAbsolutePathFromFlag(cmd, globals.SandboxBinaryLabel)
	if err != nil {
		return sd, err
	}
	if os.Getenv("SANDBOX_BINARY") == "" {
		_ = os.Setenv("SANDBOX_BINARY", basedir)
	}
	sd.BasedirName = args[0]
	versionFromOption := false
	sd.Version, _ = flags.GetString(globals.BinaryVersionLabel)
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
		var err error
		sd.BasedirName = common.RemoveTrailingSlash(sd.BasedirName)
		sd.BasedirName, err = common.AbsolutePath(sd.BasedirName)
		if err != nil {
			return sd, errors.Wrapf(err, "couldn't get an absolute path for %s", sd.BasedirName)
		}
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

	sd.Port, err = common.VersionToPort(sd.Version)
	common.ErrCheckExitf(err, 1, "can't convert '%s' into port number", sd.Version)
	if sd.Port < 0 {
		common.Exitf(1, "unsupported version format (%s)", sd.Version)
	}
	sd.UserPort, _ = flags.GetInt(globals.PortLabel)
	sd.BasePort, _ = flags.GetInt(globals.BasePortLabel)
	sd.DirName, _ = flags.GetString(globals.SandboxDirectoryLabel)

	if sd.UserPort > 0 {
		sd.Port = sd.UserPort
	}

	sd.Basedir = path.Join(basedir, sd.BasedirName)
	// sd.Basedir = path.Join(basedir, args[0])
	if !common.DirExists(sd.Basedir) {
		common.Exitf(1, "basedir '%s' not found", sd.Basedir)
	}

	err = common.CheckTarballOperatingSystem(sd.Basedir)
	common.ErrCheckExitf(err, 1, "incorrect tarball detected")

	sd.SandboxDir, err = GetAbsolutePathFromFlag(cmd, globals.SandboxHomeLabel)
	if err != nil {
		return sd, err
	}

	err = common.CheckSandboxDir(sd.SandboxDir)
	if err != nil {
		return sd, err
	}
	sd.InstalledPorts, err = common.GetInstalledPorts(sd.SandboxDir)
	if err != nil {
		return sd, err
	}

	for _, p := range defaults.Defaults().ReservedPorts {
		sd.InstalledPorts = append(sd.InstalledPorts, p)
	}
	sd.LoadGrants = true
	sd.SkipStart, _ = flags.GetBool(globals.SkipStartLabel)
	skipLoadGrants, _ := flags.GetBool(globals.SkipLoadGrantsLabel)
	if skipLoadGrants || sd.SkipStart {
		sd.LoadGrants = false
	}
	sd.SkipReportHost, _ = flags.GetBool(globals.SkipReportHostLabel)
	sd.SkipReportPort, _ = flags.GetBool(globals.SkipReportPortLabel)
	sd.DisableMysqlX, _ = flags.GetBool(globals.DisableMysqlXLabel)
	sd.EnableMysqlX, _ = flags.GetBool(globals.EnableMysqlXLabel)
	sd.HistoryDir, _ = flags.GetString(globals.HistoryDirLabel)
	sd.DbUser, _ = flags.GetString(globals.DbUserLabel)
	sd.DbPassword, _ = flags.GetString(globals.DbPasswordLabel)
	sd.RplUser, _ = flags.GetString(globals.RplUserLabel)

	checkForRootValue(sd.DbUser, globals.DbUserLabel, globals.DbUserValue)
	checkForRootValue(sd.RplUser, globals.RplUserLabel, globals.RplUserValue)

	sd.RplPassword, _ = flags.GetString(globals.RplPasswordLabel)
	sd.RemoteAccess, _ = flags.GetString(globals.RemoteAccessLabel)
	sd.BindAddress, _ = flags.GetString(globals.BindAddressLabel)
	sd.CustomMysqld, _ = flags.GetString(globals.CustomMysqldLabel)
	sd.InitOptions, _ = flags.GetStringSlice(globals.InitOptionsLabel)
	sd.MyCnfOptions, _ = flags.GetStringSlice(globals.MyCnfOptionsLabel)
	sd.PreGrantsSqlFile, _ = flags.GetString(globals.PreGrantsSqlFileLabel)
	sd.PreGrantsSql, _ = flags.GetStringSlice(globals.PreGrantsSqlLabel)
	sd.PostGrantsSql, _ = flags.GetStringSlice(globals.PostGrantsSqlLabel)
	sd.PostGrantsSqlFile, _ = flags.GetString(globals.PostGrantsSqlFileLabel)
	sd.MyCnfFile, _ = flags.GetString(globals.MyCnfFileLabel)
	sd.NativeAuthPlugin, _ = flags.GetBool(globals.NativeAuthPluginLabel)
	sd.KeepUuid, _ = flags.GetBool(globals.KeepServerUuidLabel)
	sd.Force, _ = flags.GetBool(globals.ForceLabel)
	sd.ExposeDdTables, _ = flags.GetBool(globals.ExposeDdTablesLabel)
	sd.InitGeneralLog, _ = flags.GetBool(globals.InitGeneralLogLabel)
	sd.EnableGeneralLog, _ = flags.GetBool(globals.EnableGeneralLogLabel)

	if sd.DisableMysqlX && sd.EnableMysqlX {
		common.Exit(1, "flags --enable-mysqlx and --disable-mysqlx cannot be used together")
	}
	sd.RunConcurrently, _ = flags.GetBool(globals.ConcurrentLabel)
	if common.IsEnvSet("RUN_CONCURRENTLY") {
		sd.RunConcurrently = true
	}

	newDefaults, _ := flags.GetStringSlice(globals.DefaultsLabel)
	processDefaults(newDefaults)

	var gtid bool
	var master bool
	var replCrashSafe bool
	master, _ = flags.GetBool(globals.MasterLabel)
	gtid, _ = flags.GetBool(globals.GtidLabel)
	replCrashSafe, _ = flags.GetBool(globals.ReplCrashSafeLabel)
	if master {
		sd.ReplOptions = sandbox.SingleTemplates["replication_options"].Contents
		sd.ServerId = sd.Port
	}
	if gtid {
		templateName := "gtid_options_56"
		// 5.7.0
		isEnhancedGtid, err := common.GreaterOrEqualVersion(sd.Version, globals.MinimumEnhancedGtidVersion)
		common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
		if isEnhancedGtid {
			templateName = "gtid_options_57"
		}
		// 5.6.9
		isMinimumGtid, err := common.GreaterOrEqualVersion(sd.Version, globals.MinimumGtidVersion)
		common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
		if isMinimumGtid {
			sd.GtidOptions = sandbox.SingleTemplates[templateName].Contents
			sd.ReplCrashSafeOptions = sandbox.SingleTemplates["repl_crash_safe_options"].Contents
			sd.ReplOptions = sandbox.SingleTemplates["replication_options"].Contents
			sd.ServerId = sd.Port
		} else {
			common.Exitf(1, globals.ErrOptionRequiresVersion, globals.GtidLabel, common.IntSliceToDottedString(globals.MinimumGtidVersion))
		}
	}
	if replCrashSafe && sd.ReplCrashSafeOptions == "" {
		// 5.6.2

		isMinimumCrashSafe, err := common.GreaterOrEqualVersion(sd.Version, globals.MinimumCrashSafeVersion)
		common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
		if isMinimumCrashSafe {
			sd.ReplCrashSafeOptions = sandbox.SingleTemplates["repl_crash_safe_options"].Contents
		} else {
			common.Exitf(1, globals.ErrOptionRequiresVersion, globals.ReplCrashSafeLabel, common.IntSliceToDottedString(globals.MinimumCrashSafeVersion))
		}
	}
	return sd, nil
}

func SingleSandbox(cmd *cobra.Command, args []string) {
	var sd sandbox.SandboxDef
	var err error
	common.CheckOrigin(args)
	sd, err = FillSdef(cmd, args)
	if err != nil {
		common.Exitf(1, "error while filling the sandbox definition: %+v", err)
	}
	// When deploying a single sandbox, we disable concurrency
	sd.RunConcurrently = false
	err = sandbox.CreateStandaloneSandbox(sd)
	if err != nil {
		common.Exitf(1, globals.ErrCreatingSandbox, err)
	}
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
	singleCmd.PersistentFlags().Bool(globals.MasterLabel, false, "Make the server replication ready")

}
