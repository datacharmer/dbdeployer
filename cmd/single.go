// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
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
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func replaceTemplate(templateName string, fileName string) {
	// Use the canonical name for the template returned by findTemplate
	group, templateName, contents := findTemplate(templateName)
	if !common.FileExists(fileName) {
		common.Exitf(1, globals.ErrFileNotFound, fileName)
	}
	common.CondPrintf("Replacing template %s.%s [%d chars] with contents of file %s\n", group, templateName, len(contents), fileName)
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
		if list == nil {
			return
		}
		if len(list) == 2 {
			label := list[0]
			value := list[1]
			defaults.UpdateDefaults(label, value, false)
		}
	}
}

func getAbsolutePathFromFlag(cmd *cobra.Command, name string) (string, error) {
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
		common.CondPrintf("# %s => %s\n", version, fullVersion)
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

// Gets the Database flavor
// If none is found, defaults to MySQL
func getFlavor(userDefinedFlavor, basedir string) string {
	flavorOrigin := ""
	flavor := userDefinedFlavor
	if userDefinedFlavor != "" {
		flavorOrigin = "flag"
	}
	flavorFile := path.Join(basedir, globals.FlavorFileName)
	if common.FileExists(flavorFile) {
		flavorText, err := common.SlurpAsString(flavorFile)
		if err != nil {
			common.Exitf(1, "error reading flavor file %s: %s", flavorFile, err)
		}
		flavorText = strings.TrimSpace(flavorText)
		if userDefinedFlavor != "" && userDefinedFlavor != flavorText {
			common.Exitf(1, "user defined flavor %s doesn't match found flavor %s", userDefinedFlavor, flavorText)
		}
		flavor = flavorText
		flavorOrigin = "FLAVOR file"
	}
	// Flavor detection based on tarball contents
	if flavor == "" {
		flavor = common.DetectBinaryFlavor(basedir)
		flavorOrigin = "Binary examination"
	}
	err := common.CheckFlavorSupport(flavor)
	if err != nil {
		common.Exitf(1, "flavor detected from %s unsupported: %s", flavorOrigin, err)
	}
	return flavor
}

func fillSandboxDefinition(cmd *cobra.Command, args []string, usingImport bool) (sandbox.SandboxDef, error) {
	var sd sandbox.SandboxDef
	var err error
	sd.Imported = usingImport
	err = common.CheckPrerequisites("dbdeployer needed tools", globals.NeededExecutables)
	if err != nil {
		return sd, err
	}
	flags := cmd.Flags()

	sd.SbHost = "127.0.0.1"
	logSbOperations, _ := flags.GetBool(globals.LogSBOperationsLabel)
	defaults.LogSBOperations = logSbOperations

	if !sd.Imported {
		logDir, err := getAbsolutePathFromFlag(cmd, globals.LogLogDirectoryLabel)
		if err != nil {
			return sd, err
		}
		if logDir != "" {
			defaults.UpdateDefaults(globals.LogLogDirectoryLabel, logDir, false)
		}
	}
	templateRequests, _ := flags.GetStringArray(globals.UseTemplateLabel)
	for _, request := range templateRequests {
		tname, fname := checkTemplateChangeRequest(request)
		replaceTemplate(tname, fname)
	}

	basedir, err := getAbsolutePathFromFlag(cmd, globals.SandboxBinaryLabel)
	if err != nil {
		return sd, err
	}
	if !common.DirExists(basedir) {
		if common.FileExists(basedir) {
			return sd, fmt.Errorf("the path indicated as SANDBOX_BINARY (%s) is a file, not a directory", basedir)
		}
		return sd, fmt.Errorf("sandbox binary directory %s not found\n"+
			"Use 'dbdeployer init' to initialize it", basedir)
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
		// common.CondPrintf("OLD bd <%s> - v: <%s>\n",basedir, sd.Version )
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
		// common.CondPrintf("NEW bd <%s> - v: <%s>\n",basedir, sd.Version )
	}

	sd.Port, err = common.VersionToPort(sd.Version)
	common.ErrCheckExitf(err, 1, "can't convert '%s' into port number", sd.Version)
	if sd.Port < 0 {
		common.Exitf(1, "unsupported version format (%s)", sd.Version)
	}
	sd.UserPort, _ = flags.GetInt(globals.PortLabel)
	sd.BasePort, _ = flags.GetInt(globals.BasePortLabel)
	sd.BaseServerId, _ = flags.GetInt(globals.BaseServerIdLabel)
	sd.DirName, _ = flags.GetString(globals.SandboxDirectoryLabel)

	if sd.UserPort > 0 {
		sd.Port = sd.UserPort
	}

	sd.Basedir = path.Join(basedir, sd.BasedirName)
	if !common.DirExists(sd.Basedir) && !sd.Imported {
		common.Exitf(1, "basedir '%s' not found", sd.Basedir)
	}

	skipLibraryCheck, _ := flags.GetBool(globals.SkipLibraryCheck)
	if os.Getenv("SB_MOCKING") != "" {
		skipLibraryCheck = true
	}
	if !skipLibraryCheck {
		err = common.CheckLibraries(sd.Basedir)
		if err != nil {
			return sd, err
		}
	}

	sd.ClientBasedir, _ = flags.GetString(globals.ClientFromLabel)
	if sd.ClientBasedir != "" {
		clientBasedir := path.Join(basedir, sd.ClientBasedir)
		if !common.DirExists(clientBasedir) {
			common.Exitf(1, globals.ErrDirectoryNotFound, clientBasedir)
		}
		sd.ClientBasedir = clientBasedir
	}
	if !sd.Imported {
		err = common.CheckTarballOperatingSystem(sd.Basedir)
		common.ErrCheckExitf(err, 1, "incorrect tarball detected")
	}
	sd.SandboxDir, err = getAbsolutePathFromFlag(cmd, globals.SandboxHomeLabel)
	if err != nil {
		return sd, err
	}

	if sd.SandboxDir == sd.Basedir {
		return sd, fmt.Errorf("sandbox-binary and sandbox-home cannot be the same directory (%s)", sd.SandboxDir)
	}
	err = common.CheckSandboxDir(sd.SandboxDir)
	if err != nil {
		return sd, err
	}
	sd.InstalledPorts, err = common.GetInstalledPorts(sd.SandboxDir)
	if err != nil {
		return sd, err
	}

	sd.InstalledPorts = append(sd.InstalledPorts, defaults.Defaults().ReservedPorts...)
	sd.LoadGrants = true
	sd.SkipStart, _ = flags.GetBool(globals.SkipStartLabel)
	skipLoadGrants, _ := flags.GetBool(globals.SkipLoadGrantsLabel)
	if skipLoadGrants || sd.SkipStart {
		sd.LoadGrants = false
	}
	sd.SlavesReadOnly, _ = flags.GetBool(globals.ReadOnlyLabel)
	sd.SlavesSuperReadOnly, _ = flags.GetBool(globals.SuperReadOnlyLabel)
	sd.SkipReportHost, _ = flags.GetBool(globals.SkipReportHostLabel)
	sd.SkipReportPort, _ = flags.GetBool(globals.SkipReportPortLabel)
	sd.DisableMysqlX, _ = flags.GetBool(globals.DisableMysqlXLabel)
	sd.EnableMysqlX, _ = flags.GetBool(globals.EnableMysqlXLabel)
	sd.EnableAdminAddress, _ = flags.GetBool(globals.EnableAdminAddressLabel)
	sd.SocketInDatadir, _ = flags.GetBool(globals.SocketInDatadirLabel)
	sd.PortAsServerId, _ = flags.GetBool(globals.PortAsServerIdLabel)
	sd.ServerId, _ = flags.GetInt(globals.ServerIdLabel)
	if common.IsEnvSet("SOCKET_IN_DATADIR") {
		sd.SocketInDatadir = true
	}
	if sd.PortAsServerId && sd.ServerId != 0 {
		return sd, fmt.Errorf("options --%s and --%s should not be provided together",
			globals.PortAsServerIdLabel, globals.ServerIdLabel)
	}
	sd.FlavorInPrompt, _ = flags.GetBool(globals.FlavorInPromptLabel)
	sd.HistoryDir, _ = flags.GetString(globals.HistoryDirLabel)
	sd.DbUser, _ = flags.GetString(globals.DbUserLabel)
	sd.DbPassword, _ = flags.GetString(globals.DbPasswordLabel)
	sd.RplUser, _ = flags.GetString(globals.RplUserLabel)
	sd.DefaultRole, _ = flags.GetString(globals.DefaultRoleLabel)
	sd.CustomRoleName, _ = flags.GetString(globals.CustomRoleNameLabel)
	sd.CustomRolePrivileges, _ = flags.GetString(globals.CustomRolePrivilegesLabel)
	sd.CustomRoleTarget, _ = flags.GetString(globals.CustomRoleTargetLabel)
	sd.CustomRoleExtra, _ = flags.GetString(globals.CustomRoleExtraLabel)
	sd.TaskUser, _ = flags.GetString(globals.TaskUserLabel)
	sd.TaskUserRole, _ = flags.GetString(globals.TaskUserRoleLabel)
	sd.Flavor, _ = flags.GetString(globals.FlavorLabel)

	sd.Flavor = getFlavor(sd.Flavor, sd.Basedir)
	checkForRootValue(sd.DbUser, globals.DbUserLabel, globals.DbUserValue)
	checkForRootValue(sd.RplUser, globals.RplUserLabel, globals.RplUserValue)

	sd.RplPassword, _ = flags.GetString(globals.RplPasswordLabel)
	sd.RemoteAccess, _ = flags.GetString(globals.RemoteAccessLabel)
	sd.BindAddress, _ = flags.GetString(globals.BindAddressLabel)
	sd.CustomMysqld, _ = flags.GetString(globals.CustomMysqldLabel)
	sd.InitOptions, _ = flags.GetStringArray(globals.InitOptionsLabel)
	sd.MyCnfOptions, _ = flags.GetStringArray(globals.MyCnfOptionsLabel)
	sd.ChangeMasterOptions, _ = flags.GetStringArray(globals.ChangeMasterOptions)
	sd.PreGrantsSqlFile, _ = flags.GetString(globals.PreGrantsSqlFileLabel)
	sd.PreGrantsSql, _ = flags.GetStringArray(globals.PreGrantsSqlLabel)
	sd.PostGrantsSql, _ = flags.GetStringArray(globals.PostGrantsSqlLabel)
	sd.PostGrantsSqlFile, _ = flags.GetString(globals.PostGrantsSqlFileLabel)
	sd.MyCnfFile, _ = flags.GetString(globals.MyCnfFileLabel)
	sd.NativeAuthPlugin, _ = flags.GetBool(globals.NativeAuthPluginLabel)
	sd.KeepUuid, _ = flags.GetBool(globals.KeepServerUuidLabel)
	sd.Force, _ = flags.GetBool(globals.ForceLabel)
	sd.ExposeDdTables, _ = flags.GetBool(globals.ExposeDdTablesLabel)
	sd.InitGeneralLog, _ = flags.GetBool(globals.InitGeneralLogLabel)
	sd.EnableGeneralLog, _ = flags.GetBool(globals.EnableGeneralLogLabel)
	sd.ShellPath = defaults.Defaults().ShellPath

	if sd.DisableMysqlX && sd.EnableMysqlX {
		common.Exit(1, "flags --enable-mysqlx and --disable-mysqlx cannot be used together")
	}
	sd.RunConcurrently, _ = flags.GetBool(globals.ConcurrentLabel)
	if common.IsEnvSet("RUN_CONCURRENTLY") {
		sd.RunConcurrently = true
	}

	newDefaults, _ := flags.GetStringArray(globals.DefaultsLabel)
	processDefaults(newDefaults)

	var gtid bool
	var master bool
	var replCrashSafe bool
	master, _ = flags.GetBool(globals.MasterLabel)
	gtid, _ = flags.GetBool(globals.GtidLabel)
	replCrashSafe, _ = flags.GetBool(globals.ReplCrashSafeLabel)
	if master {
		sd.ReplOptions = sandbox.SingleTemplates[globals.TmplReplicationOptions].Contents
		if sd.ServerId == 0 {
			sd.PortAsServerId = true
		} else {
			sd.PortAsServerId = false
		}
	}
	if gtid {
		templateName := globals.TmplGtidOptions56
		// 5.7.0
		// isEnhancedGtid, err := common.GreaterOrEqualVersion(sd.Version, globals.MinimumEnhancedGtidVersion)
		isEnhancedGtid, err := common.HasCapability(sd.Flavor, common.EnhancedGTID, sd.Version)
		common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
		if isEnhancedGtid {
			templateName = globals.TmplGtidOptions57
		}
		// 5.6.9
		//isMinimumGtid, err := common.GreaterOrEqualVersion(sd.Version, globals.MinimumGtidVersion)
		isMinimumGtid, err := common.HasCapability(sd.Flavor, common.GTID, sd.Version)
		common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
		if isMinimumGtid {
			sd.GtidOptions = sandbox.SingleTemplates[templateName].Contents
			sd.ReplCrashSafeOptions = sandbox.SingleTemplates[globals.TmplReplCrashSafeOptions].Contents
			sd.ReplOptions = sandbox.SingleTemplates[globals.TmplReplicationOptions].Contents
			if sd.ServerId == 0 {
				sd.PortAsServerId = true
			} else {
				sd.PortAsServerId = false
			}
		} else {
			common.Exitf(1, globals.ErrOptionRequiresVersion, globals.GtidLabel, common.IntSliceToDottedString(globals.MinimumGtidVersion))
		}
	}
	if replCrashSafe && sd.ReplCrashSafeOptions == "" {
		// 5.6.2

		// isMinimumCrashSafe, err := common.GreaterOrEqualVersion(sd.Version, globals.MinimumCrashSafeVersion)
		isMinimumCrashSafe, err := common.HasCapability(sd.Flavor, common.CrashSafe, sd.Version)
		common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
		if isMinimumCrashSafe {
			sd.ReplCrashSafeOptions = sandbox.SingleTemplates[globals.TmplReplCrashSafeOptions].Contents
		} else {
			common.Exitf(1, globals.ErrOptionRequiresVersion, globals.ReplCrashSafeLabel, common.IntSliceToDottedString(globals.MinimumCrashSafeVersion))
		}
	}
	if flags.Changed(globals.DefaultRoleLabel) ||
		flags.Changed(globals.CustomRoleNameLabel) ||
		flags.Changed(globals.CustomRolePrivilegesLabel) ||
		flags.Changed(globals.TaskUserLabel) ||
		flags.Changed(globals.TaskUserRoleLabel) ||
		flags.Changed(globals.CustomRoleTargetLabel) ||
		flags.Changed(globals.CustomRoleExtraLabel) {
		isRoleEnabled, err := common.HasCapability(sd.Flavor, common.Roles, sd.Version)
		common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
		if !isRoleEnabled {
			common.Exitf(1, "options about roles requires version 8.0+")
		}
	}
	return sd, nil
}

func singleSandbox(cmd *cobra.Command, args []string) {
	var sd sandbox.SandboxDef
	var err error
	common.CheckOrigin(args)
	sd, err = fillSandboxDefinition(cmd, args, false)
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
	Use:   "single MySQL-Version",
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
	Run:         singleSandbox,
	Annotations: map[string]string{"export": makeExportArgs(globals.ExportVersionDir, 1)},
}

func init() {
	deployCmd.AddCommand(singleCmd)
	singleCmd.PersistentFlags().Bool(globals.MasterLabel, false, "Make the server replication ready")
	singleCmd.PersistentFlags().Int(globals.ServerIdLabel, 0, "Overwrite default server-id")
	setPflag(singleCmd, globals.PromptLabel, "", "", globals.PromptValue, "Default prompt for the single client", false)
}
