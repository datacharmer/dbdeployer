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

package sandbox

import (
	"encoding/json"
	"fmt"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/pkg/errors"
	"os"
	"path"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
)

type SandboxDef struct {
	DirName              string           // Name of the directory containing the sandbox
	SBType               string           // Type of sandbox (single, multiple, replication-node, group-node)
	Multi                bool             // Either single or part of a multiple sandbox
	NodeNum              int              // In multiple sandboxes, which node is this
	Version              string           // MySQL version
	Basedir              string           // Where to get binaries from (e.g. $HOME/opt/mysql/8.0.11)
	BasedirName          string           // The bare name of the directory containing the binaries (e.g. 8.0.11)
	SandboxDir           string           // Target directory for sandboxes
	LoadGrants           bool             // Should we load grants?
	SkipReportHost       bool             // Do not add report-host to my.sandbox.cnf
	SkipReportPort       bool             // Do not add report-port to my.sandbox.cnf
	SkipStart            bool             // Do not start the server after deployment
	InstalledPorts       []int            // Which ports should be skipped in port assignment for this SB
	Port                 int              // Port assigned to this sandbox
	MysqlXPort           int              // XPlugin port for this sandbox
	UserPort             int              // Custom port provided by user
	BasePort             int              // Base port for calculating more ports in multiple SB
	MorePorts            []int            // Additional ports that belong to this sandbox
	Prompt               string           // Prompt to use in "mysql" client
	DbUser               string           // Database user name
	RplUser              string           // Replication user name
	DbPassword           string           // Database password
	RplPassword          string           // Replication password
	RemoteAccess         string           // What access have the users created for this SB (127.%)
	BindAddress          string           // Bind address for this sandbox (127.0.0.1)
	CustomMysqld         string           // Use an alternative mysqld executable
	ServerId             int              // Server ID (for replication)
	ReplOptions          string           // Replication options, as string to append to my.sandbox.cnf
	GtidOptions          string           // Options needed for GTID
	ReplCrashSafeOptions string           // Options needed for Replication crash safe
	SemiSyncOptions      string           // Options for semi-synchronous replication
	InitOptions          []string         // Options to be added to the initialization command
	MyCnfOptions         []string         // Options to be added to my.sandbox.cnf
	PreGrantsSql         []string         // SQL statements to execute before grants assignment
	PreGrantsSqlFile     string           // SQL file to load before grants assignment
	PostGrantsSql        []string         // SQL statements to run after grants assignment
	PostGrantsSqlFile    string           // SQL file to load after grants assignment
	MyCnfFile            string           // options file to merge with the SB my.sandbox.cnf
	HistoryDir           string           // Where to store the MySQL client history
	LogFileName          string           // Where to log operations for this sandbox
	Logger               *defaults.Logger // Carries a logger across sandboxes
	InitGeneralLog       bool             // Enable general log during server initialization
	EnableGeneralLog     bool             // Enable general log for regular usage
	NativeAuthPlugin     bool             // Use the native password plugin for MySQL 8.0.4+
	DisableMysqlX        bool             // Disable Xplugin (MySQL 8.0.11+)
	EnableMysqlX         bool             // Enable Xplugin (MySQL 5.7.12+)
	KeepUuid             bool             // Do not change UUID
	SinglePrimary        bool             // Use single primary for group replication
	Force                bool             // Overwrite an existing sandbox with same target
	ExposeDdTables       bool             // Show hidden data dictionary tables (MySQL 8.0.0+)
	RunConcurrently      bool             // Run multiple sandbox creation concurrently
}

type ScriptDef struct {
	scriptName     string
	templateName   string
	makeExecutable bool
}

type ScriptBatch struct {
	tc         TemplateCollection
	logger     *defaults.Logger
	sandboxDir string
	data       common.StringMap
	scripts    []ScriptDef
}

var emptyExecutionList = []concurrent.ExecutionList{}

func GetOptionsFromFile(filename string) (options []string, err error) {
	skipOptions := map[string]bool{
		"user":         true,
		"port":         true,
		"socket":       true,
		"datadir":      true,
		"basedir":      true,
		"tmpdir":       true,
		"pid-file":     true,
		"server-id":    true,
		"bind-address": true,
		"log-error":    true,
	}
	config, err := common.ParseConfigFile(filename)
	if err != nil {
		return globals.EmptyStrings, err
	}
	for _, kv := range config["mysqld"] {

		if skipOptions[kv.Key] {
			continue
		}
		options = append(options, fmt.Sprintf("%s = %s", kv.Key, kv.Value))
		//fmt.Printf("%d %s : %s \n", N, kv.key, kv.value)
	}
	return options, nil
}

func SandboxDefToJson(sd SandboxDef) string {
	b, err := json.MarshalIndent(sd, " ", "\t")
	if err != nil {
		return "Sandbox definition could not be encoded\n"
	}
	return fmt.Sprintf("%s", b)
}

func StringMapToJson(data common.StringMap) string {
	copyright := data["Copyright"]
	data["Copyright"] = "[skipped] (See 'Copyright' template for full text)"
	b, err := json.MarshalIndent(data, " ", "\t")
	data["Copyright"] = copyright
	if err != nil {
		return "String map could not be encoded"
	}
	return fmt.Sprintf("%s", b)
}

func isLocked(sbDir string) bool {
	return common.FileExists(path.Join(sbDir, globals.ScriptNoClear)) || common.FileExists(path.Join(sbDir, globals.ScriptNoClearAll))
}

func CheckDirectory(sandboxDef SandboxDef) (SandboxDef, error) {
	sandboxDir := sandboxDef.SandboxDir
	if common.DirExists(sandboxDir) {
		if sandboxDef.Force {
			if isLocked(sandboxDir) {
				return sandboxDef, fmt.Errorf("sandbox in %s is locked. Cannot be overwritten\nYou can unlock it with 'dbdeployer admin unlock %s'\n", sandboxDir, common.DirName(sandboxDir))
			}
			fmt.Printf("Overwriting directory %s\n", sandboxDir)
			stopCommand := path.Join(sandboxDir, globals.ScriptStop)
			if !common.ExecExists(stopCommand) {
				stopCommand = path.Join(sandboxDir, globals.ScriptStopAll)
			}
			if !common.ExecExists(stopCommand) {
				fmt.Printf("Neither 'stop' or 'stop_all' found in %s\n", sandboxDir)
			}

			usedPortsList, err := common.GetInstalledPorts(sandboxDir)
			if err != nil {
				return sandboxDef, err
			}
			myUsedPorts := make(map[int]bool)
			for _, p := range usedPortsList {
				myUsedPorts[p] = true
			}

			logDirectory, err := getLogDirFromSbDescription(sandboxDir)
			if err != nil {
				return sandboxDef, err
			}
			_, err = common.RunCmd(stopCommand)
			if err != nil {
				return sandboxDef, fmt.Errorf(globals.ErrWhileStoppingSandbox, sandboxDir)
			}
			_, err = common.RunCmdWithArgs("rm", []string{"-rf", sandboxDir})
			if err != nil {
				return sandboxDef, fmt.Errorf(globals.ErrWhileDeletingSandbox, sandboxDir)
			}
			if logDirectory != "" {
				_, err = common.RunCmdWithArgs("rm", []string{"-rf", logDirectory})
				if err != nil {
					return sandboxDef, fmt.Errorf("error while deleting log directory %s", logDirectory)
				}
			}
			err = defaults.DeleteFromCatalog(sandboxDir)
			if err != nil {
				return SandboxDef{}, err
			}
			var newInstalledPorts []int
			for _, port := range sandboxDef.InstalledPorts {
				if !myUsedPorts[port] {
					newInstalledPorts = append(newInstalledPorts, port)
				}
			}
			sandboxDef.InstalledPorts = newInstalledPorts
		} else {
			return sandboxDef, fmt.Errorf("directory %s already exists. Use --force to override.", sandboxDir)
		}
	}
	return sandboxDef, nil
}

func CheckPort(caller string, sandboxType string, installedPorts []int, port int) error {
	conflict := 0
	for _, p := range installedPorts {
		if p == port {
			conflict = p
		}
	}
	if conflict > 0 {
		return fmt.Errorf("port conflict detected for %s (%s). Port %d is already used", sandboxType, caller, conflict)
	}
	return nil
}

func FixServerUuid(sandboxDef SandboxDef) (uuidDef string, uuidFile string, err error) {
	// 5.6.9
	isMinimumGtid, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumGtidVersion)
	if err != nil {
		return globals.EmptyString, globals.EmptyString, err
	}
	if !isMinimumGtid {
		// Before server UUID was implemented, we don't need to do anything
		return globals.EmptyString, globals.EmptyString, nil
	}
	var newUuid string
	newUuid, err = common.MakeCustomizedUuid(sandboxDef.Port, sandboxDef.NodeNum)
	if err != nil {
		return globals.EmptyString, globals.EmptyString, err
	}
	uuidDef = fmt.Sprintf("server-uuid=%s", newUuid)
	operationDir := path.Join(sandboxDef.SandboxDir, globals.DataDirName)
	uuidFile = path.Join(operationDir, globals.AutoCnfName)
	return uuidDef, uuidFile, nil
}

func sliceToText(stringSlice []string) string {
	var text string = ""
	for _, v := range stringSlice {
		if len(v) > 0 {
			text += fmt.Sprintf("%s\n", v)
		}
	}
	return text
}

func setMysqlxProperties(sandboxDef SandboxDef, globalTmpDir string) (SandboxDef, error) {
	mysqlxPort := sandboxDef.MysqlXPort
	if mysqlxPort == 0 {
		var err error
		mysqlxPort, err = common.FindFreePort(sandboxDef.Port+defaults.Defaults().MysqlXPortDelta, sandboxDef.InstalledPorts, 1)
		if err != nil {
			return SandboxDef{}, errors.Wrapf(err, "error detecting free port for MySQLX")
		}
	}
	sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("mysqlx-port=%d", mysqlxPort))
	sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("mysqlx-socket=%s/mysqlx-%d.sock", globalTmpDir, mysqlxPort))
	sandboxDef.MorePorts = append(sandboxDef.MorePorts, mysqlxPort)
	sandboxDef.MysqlXPort = mysqlxPort
	return sandboxDef, nil
}

func CreateChildSandbox(sandboxDef SandboxDef) (execList []concurrent.ExecutionList, err error) {
	return createSingleSandbox(sandboxDef)
}
func CreateStandaloneSandbox(sandboxDef SandboxDef) (err error) {
	_, err = createSingleSandbox(sandboxDef)
	return err
}

func sbError(reason, format string, args ...interface{}) error {
	return fmt.Errorf(reason+" "+format, args...)
}

func createSingleSandbox(sandboxDef SandboxDef) (execList []concurrent.ExecutionList, err error) {

	var sandboxDir string

	if sandboxDef.SBType == "" {
		sandboxDef.SBType = "single"
	}
	logName := sandboxDef.SBType
	if sandboxDef.NodeNum > 0 {
		logName = fmt.Sprintf("%s-%d", logName, sandboxDef.NodeNum)
	}
	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		var fileName string
		logger, fileName, err = defaults.NewLogger(common.LogDirName(), logName)
		if err != nil {
			return emptyExecutionList, sbError("logger", "%s", err)
		}
		sandboxDef.LogFileName = common.ReplaceLiteralHome(fileName)
	}
	logger.Printf("Single Sandbox Definition: %s\n", SandboxDefToJson(sandboxDef))
	if !common.DirExists(sandboxDef.Basedir) {
		return emptyExecutionList, fmt.Errorf(globals.ErrBaseDirectoryNotFound, sandboxDef.Basedir)
	}

	if sandboxDef.Port <= 1024 {
		return emptyExecutionList, fmt.Errorf("port for sandbox must be > 1024 (given:%d)", sandboxDef.Port)
	}

	versionFname := common.VersionToName(sandboxDef.Version)
	if sandboxDef.Prompt == "" {
		sandboxDef.Prompt = "mysql"
	}
	if sandboxDef.DirName == "" {
		if sandboxDef.Version != sandboxDef.BasedirName {
			sandboxDef.DirName = defaults.Defaults().SandboxPrefix + sandboxDef.BasedirName
		} else {
			sandboxDef.DirName = defaults.Defaults().SandboxPrefix + versionFname
		}
	}
	sandboxDir = path.Join(sandboxDef.SandboxDir, sandboxDef.DirName)
	sandboxDef.SandboxDir = sandboxDir
	logger.Printf("Single Sandbox directory defined as %s\n", sandboxDef.SandboxDir)
	dataDir := path.Join(sandboxDir, globals.DataDirName)
	tmpDir := path.Join(sandboxDir, "tmp")
	globalTmpDir := os.Getenv("TMPDIR")
	if globalTmpDir == "" {
		globalTmpDir = "/tmp"
	}
	if !common.DirExists(globalTmpDir) {
		return emptyExecutionList, fmt.Errorf("TMP directory %s does not exist", globalTmpDir)
	}
	if sandboxDef.NodeNum == 0 && !sandboxDef.Force {
		sandboxDef.Port, err = common.FindFreePort(sandboxDef.Port, sandboxDef.InstalledPorts, 1)
		if err != nil {
			return emptyExecutionList, errors.Wrapf(err, "error detecting free port for single sandbox")
		}
		logger.Printf("Port defined as %d using FindFreePort \n", sandboxDef.Port)
	}
	usingPlugins := false
	rightPluginDir := true // Assuming we can use the right plugin directory
	if sandboxDef.EnableMysqlX {
		// 5.7.12
		isMinimumMySQLX, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxVersion)
		if err != nil {
			return emptyExecutionList, err
		}
		if !isMinimumMySQLX {
			return emptyExecutionList, fmt.Errorf(globals.ErrOptionRequiresVersion, "enable-mysqlx", common.IntSliceToDottedString(globals.MinimumMysqlxVersion))
		}
		// If the version is 8.0.11 or later, MySQL X is enabled already
		// 8.0.11
		isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
		if err != nil {
			return emptyExecutionList, err
		}
		if !isMinimumMySQLXDefault {
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, "plugin_load=mysqlx=mysqlx.so")
			sandboxDef, err = setMysqlxProperties(sandboxDef, globalTmpDir)
			if err != nil {
				return emptyExecutionList, err
			}
			logger.Printf("Added mysqlx plugin to my.cnf\n")
		}
		usingPlugins = true
	}
	// 8.0.11
	isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
	if isMinimumMySQLXDefault && !sandboxDef.DisableMysqlX {
		usingPlugins = true
	}
	if sandboxDef.ExposeDdTables {
		// 8.0.0
		isMinimumDataDictionary, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumDataDictionaryVersion)
		if err != nil {
			return emptyExecutionList, err
		}
		if !isMinimumDataDictionary {
			return emptyExecutionList, fmt.Errorf(globals.ErrOptionRequiresVersion, "expose-dd-tables", common.IntSliceToDottedString(globals.MinimumDataDictionaryVersion))
		}
		sandboxDef.PostGrantsSql = append(sandboxDef.PostGrantsSql, SingleTemplates["expose_dd_tables"].Contents)
		if sandboxDef.CustomMysqld != "" && sandboxDef.CustomMysqld != "mysqld-debug" {
			return emptyExecutionList, fmt.Errorf("--expose-dd-tables requires mysqld-debug. A different file was indicated (--custom-mysqld=%s)\n%s",
				sandboxDef.CustomMysqld, "Either use \"mysqld-debug\" or remove --custom-mysqld")
		}
		sandboxDef.CustomMysqld = "mysqld-debug"
		logger.Printf("Using mysqld-debug for this sandbox\n")
	}
	if sandboxDef.CustomMysqld != "" {
		customMysqld := path.Join(sandboxDef.Basedir, "bin", sandboxDef.CustomMysqld)
		if !common.ExecExists(customMysqld) {
			return emptyExecutionList, fmt.Errorf("File %s not found or not executable\n"+
				"The file \"%s\" (defined with --custom-mysqld) must be in the same directory as the regular mysqld",
				customMysqld, sandboxDef.CustomMysqld)
		}
		pluginDebugDir := fmt.Sprintf("%s/lib/plugin/debug", sandboxDef.Basedir)
		if sandboxDef.CustomMysqld == "mysqld-debug" && common.DirExists(pluginDebugDir) {
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("plugin-dir=%s", pluginDebugDir))
		} else {
			rightPluginDir = false
		}
	}
	// 5.1.0
	isMinimumDynVariables, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumDynVariablesVersion)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumDynVariables {
		if sandboxDef.EnableGeneralLog {
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, "general_log=1")
			logger.Printf("Enabling general log\n")
		}
		if sandboxDef.InitGeneralLog {
			sandboxDef.InitOptions = append(sandboxDef.InitOptions, "--general_log=1")
			logger.Printf("Enabling general log during initialization\n")
		}
	}
	// 8.0.4
	isMinimumNativeAuthPlugin, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumNativeAuthPluginVersion)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumNativeAuthPlugin {
		if sandboxDef.NativeAuthPlugin == true {
			sandboxDef.InitOptions = append(sandboxDef.InitOptions, "--default_authentication_plugin=mysql_native_password")
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, "default_authentication_plugin=mysql_native_password")
			logger.Printf("Using mysql_native_password for authentication\n")
		}
	}
	// 8.0.11
	isMinimumMySQLXDefault, err = common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumMySQLXDefault {
		if sandboxDef.DisableMysqlX {
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, "mysqlx=OFF")
			logger.Printf("Disabling MySQLX\n")
		} else {
			sandboxDef, err = setMysqlxProperties(sandboxDef, globalTmpDir)
			if err != nil {
				return emptyExecutionList, err
			}
		}
	}
	mysqlshExecutable := fmt.Sprintf("%s/bin/mysqlsh", sandboxDef.Basedir)
	if !common.ExecExists(mysqlshExecutable) {
		mysqlshExecutable = "mysqlsh"
	}
	if sandboxDef.MyCnfFile != "" {
		if !common.FileExists(sandboxDef.MyCnfFile) {
			return emptyExecutionList, fmt.Errorf(globals.ErrFileNotFound, sandboxDef.MyCnfFile)
		}
		options, err := GetOptionsFromFile(sandboxDef.MyCnfFile)
		if err != nil {
			return emptyExecutionList, errors.Wrapf(err, "error reading provided configuration file")
		}
		if len(options) > 0 {
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("# options retrieved from %s", sandboxDef.MyCnfFile))
		}
		for _, option := range options {
			// fmt.Printf("[%s]\n", option)
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, option)
		}
	}
	if common.Includes(sliceToText(sandboxDef.MyCnfOptions), "plugin.load") {
		usingPlugins = true
	}
	if common.Includes(sandboxDef.SemiSyncOptions, "plugin.load") {
		usingPlugins = true
	}
	if usingPlugins {
		if !rightPluginDir {
			return emptyExecutionList, fmt.Errorf("the request of using mysqld-debug can't be honored.\n" +
				"This deployment is using a plugin, but the debug\n" +
				"directory for plugins was not found")
		}
	}
	timestamp := time.Now()
	verList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, errors.Wrapf(err, "")
	}
	var data = common.StringMap{"Basedir": sandboxDef.Basedir,
		"Copyright":            SingleTemplates["Copyright"].Contents,
		"AppVersion":           common.VersionDef,
		"DateTime":             timestamp.Format(time.UnixDate),
		"SandboxDir":           sandboxDir,
		"CustomMysqld":         sandboxDef.CustomMysqld,
		"Port":                 sandboxDef.Port,
		"MysqlXPort":           sandboxDef.MysqlXPort,
		"MysqlShell":           mysqlshExecutable,
		"BasePort":             sandboxDef.BasePort,
		"Prompt":               sandboxDef.Prompt,
		"Version":              sandboxDef.Version,
		"VersionMajor":         verList[0],
		"VersionMinor":         verList[1],
		"VersionRev":           verList[2],
		"Datadir":              dataDir,
		"Tmpdir":               tmpDir,
		"GlobalTmpDir":         globalTmpDir,
		"DbUser":               sandboxDef.DbUser,
		"DbPassword":           sandboxDef.DbPassword,
		"RplUser":              sandboxDef.RplUser,
		"RplPassword":          sandboxDef.RplPassword,
		"RemoteAccess":         sandboxDef.RemoteAccess,
		"BindAddress":          sandboxDef.BindAddress,
		"OsUser":               os.Getenv("USER"),
		"ReplOptions":          sandboxDef.ReplOptions,
		"GtidOptions":          sandboxDef.GtidOptions,
		"ReplCrashSafeOptions": sandboxDef.ReplCrashSafeOptions,
		"SemiSyncOptions":      sandboxDef.SemiSyncOptions,
		"ExtraOptions":         sliceToText(sandboxDef.MyCnfOptions),
		"ReportHost":           fmt.Sprintf("report-host=single-%d", sandboxDef.Port),
		"ReportPort":           fmt.Sprintf("report-port=%d", sandboxDef.Port),
		"HistoryDir":           sandboxDef.HistoryDir,
	}
	if sandboxDef.NodeNum != 0 {
		data["ReportHost"] = fmt.Sprintf("report-host = node-%d", sandboxDef.NodeNum)
	}
	if sandboxDef.SkipReportHost || sandboxDef.SBType == "group-node" {
		data["ReportHost"] = ""
	}
	if sandboxDef.SkipReportPort {
		data["ReportPort"] = ""
	}
	if sandboxDef.ServerId > 0 {
		data["ServerId"] = fmt.Sprintf("server-id=%d", sandboxDef.ServerId)
	} else {
		data["ServerId"] = ""
	}
	if common.DirExists(sandboxDir) {
		sandboxDef, err = CheckDirectory(sandboxDef)
		if err != nil {
			return emptyExecutionList, sbError("check directory", "%s", err)
		}
	}
	logger.Printf("Checking port %d using CheckPort\n", sandboxDef.Port)
	err = CheckPort("createSingleSandbox", sandboxDef.SBType, sandboxDef.InstalledPorts, sandboxDef.Port)
	if err != nil {
		return emptyExecutionList, sbError("check port", "%s", err)
	}

	err = os.Mkdir(sandboxDir, 0755)
	if err != nil {
		return emptyExecutionList, sbError("sandbox dir creation", "%s", err)
	}

	logger.Printf("Created directory %s\n", sandboxDef.SandboxDir)
	logger.Printf("Single Sandbox template data: %s\n", StringMapToJson(data))

	err = os.Mkdir(dataDir, 0755)
	if err != nil {
		return emptyExecutionList, sbError("data dir creation", "%s", err)
	}
	logger.Printf("Created directory %s\n", dataDir)
	err = os.Mkdir(tmpDir, 0755)
	if err != nil {
		return emptyExecutionList, sbError("tmp dir creation", "%s", err)
	}
	logger.Printf("Created directory %s\n", tmpDir)
	script := path.Join(sandboxDef.Basedir, "scripts", "mysql_install_db")
	initScriptFlags := ""
	isMinimumDefaultInitialize, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumDefaultInitializeVersion)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumDefaultInitialize {
		script = path.Join(sandboxDef.Basedir, "bin", "mysqld")
		if sandboxDef.CustomMysqld != "" {
			script = path.Join(sandboxDef.Basedir, "bin", sandboxDef.CustomMysqld)
		}
		initScriptFlags = "--initialize-insecure"
	}
	if !common.ExecExists(script) {
		fmt.Printf("SCRIPT\n")
		return emptyExecutionList, fmt.Errorf(globals.ErrScriptNotFound, script)
	}
	if len(sandboxDef.InitOptions) > 0 {
		for _, op := range sandboxDef.InitOptions {
			initScriptFlags += " " + op
		}
	}
	data["InitScript"] = script
	data["InitDefaults"] = "--no-defaults"
	if initScriptFlags != "" {
		initScriptFlags = fmt.Sprintf("\\\n    %s", initScriptFlags)
	}
	data["ExtraInitFlags"] = initScriptFlags
	data["FixUuidFile1"] = ""
	data["FixUuidFile2"] = ""

	if !sandboxDef.KeepUuid {
		newUuid, uuidFname, err := FixServerUuid(sandboxDef)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "# UUID non-blocking failure: %s\n", err)
		}
		if uuidFname != "" {
			data["FixUuidFile1"] = fmt.Sprintf(`echo "[auto]" > %s`, uuidFname)
			data["FixUuidFile2"] = fmt.Sprintf(`echo "%s" >> %s`, newUuid, uuidFname)
			logger.Printf("Created custom UUID %s\n", newUuid)
		}
	}

	err = writeScript(logger, SingleTemplates, globals.ScriptInitDb, "init_db_template", sandboxDir, data, true)
	if err != nil {
		return emptyExecutionList, err
	}
	if sandboxDef.RunConcurrently {
		var eCommand = concurrent.ExecCommand{
			Cmd:  path.Join(sandboxDir, globals.ScriptInitDb),
			Args: []string{},
		}
		logger.Printf("Added init_db script to execution list\n")
		execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 0, Command: eCommand})
	} else {
		logger.Printf("Running init_db script \n")
		initDbScript := path.Join(sandboxDir, globals.ScriptInitDb)
		if !common.FileExists(initDbScript) {
			return emptyExecutionList, fmt.Errorf(globals.ErrFileNotFound, initDbScript)
		}
		_, err := common.RunCmdCtrl(initDbScript, true)
		if err == nil {
			if !sandboxDef.Multi {
				if globals.UsingDbDeployer {
					fmt.Printf("Database installed in %s\n", common.ReplaceLiteralHome(sandboxDir))
					fmt.Printf("run 'dbdeployer usage single' for basic instructions'\n")
				}
			}
		} else {
			return emptyExecutionList, fmt.Errorf("InitDb failure: %s\n", err)
		}
	}

	sbItem := defaults.SandboxItem{
		Origin:      sandboxDef.Basedir,
		SBType:      sandboxDef.SBType,
		Version:     sandboxDef.Version,
		Port:        []int{sandboxDef.Port},
		Nodes:       []string{},
		Destination: sandboxDir,
	}

	if sandboxDef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sandboxDef.LogFileName)
	}
	sbDesc := common.SandboxDescription{
		Basedir: sandboxDef.Basedir,
		SBType:  sandboxDef.SBType,
		Version: sandboxDef.Version,
		Port:    []int{sandboxDef.Port},
		Nodes:   0,
		NodeNum: sandboxDef.NodeNum,
		LogFile: sandboxDef.LogFileName,
	}
	if len(sandboxDef.MorePorts) > 0 {
		for _, port := range sandboxDef.MorePorts {
			sbDesc.Port = append(sbDesc.Port, port)
			sbItem.Port = append(sbItem.Port, port)
		}
	}
	logger.Printf("Writing single sandbox description\n")
	err = common.WriteSandboxDescription(sandboxDir, sbDesc)
	if err != nil {
		return emptyExecutionList, errors.Wrapf(err, "unable to write sandbox description")
	}
	if sandboxDef.SBType == "single" {
		err = defaults.UpdateCatalog(sandboxDir, sbItem)
		if err != nil {
			return emptyExecutionList, errors.Wrapf(err, "error updating catalog")
		}
	}
	logger.Printf("Writing single sandbox scripts\n")
	sb := ScriptBatch{
		sandboxDir: sandboxDir,
		data:       data,
		logger:     logger,
		tc:         SingleTemplates,
		scripts: []ScriptDef{
			{globals.ScriptStart, "start_template", true},
			{globals.ScriptStatus, "status_template", true},
			{globals.ScriptStop, "stop_template", true},
			{globals.ScriptClear, "clear_template", true},
			{globals.ScriptUse, "use_template", true},
			{globals.ScriptShowLog, "show_log_template", true},
			{globals.ScriptShowBinlog, "show_binlog_template", true},
			{globals.ScriptShowRelayLog, "show_relaylog_template", true},
			{globals.ScriptSendKill, "send_kill_template", true},
			{globals.ScriptRestart, "restart_template", true},
			{globals.ScriptLoadGrants, "load_grants_template", true},
			{globals.ScriptAddOption, "add_option_template", true},
			{globals.ScriptMy, "my_template", true},
			{globals.ScriptTestSb, "test_sb_template", true},
			{globals.ScriptMySandboxCnf, "my_cnf_template", true},
		},
	}
	if sandboxDef.MysqlXPort != 0 {
		sb.scripts = append(sb.scripts, ScriptDef{globals.ScriptMysqlsh, "mysqlsh_template", true})
	}
	var grantsTemplateName string = ""
	isMinimumRoles, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumRolesVersion)
	if err != nil {
		return emptyExecutionList, err
	}
	isMinimumCreateUserVersion, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumCreateUserVersion)
	if err != nil {
		return emptyExecutionList, err
	}
	switch {
	// 8.0.0
	case isMinimumRoles:
		grantsTemplateName = "grants_template8x"
		// 5.7.6
	case isMinimumCreateUserVersion:
		grantsTemplateName = "grants_template57"
	default:
		grantsTemplateName = "grants_template5x"
	}
	sb.scripts = append(sb.scripts, ScriptDef{globals.ScriptGrantsMysql, grantsTemplateName, false})
	sb.scripts = append(sb.scripts, ScriptDef{globals.ScriptSbInclude, "sb_include_template", false})

	err = writeScripts(sb)
	if err != nil {
		return emptyExecutionList, err
	}
	preGrantSqlFile := path.Join(sandboxDir, globals.ScriptPreGrantsSql)
	postGrantSqlFile := path.Join(sandboxDir, globals.ScriptPostGrantsSql)
	if sandboxDef.PreGrantsSqlFile != "" {
		err := common.CopyFile(sandboxDef.PreGrantsSqlFile, preGrantSqlFile)
		if err != nil {
			return emptyExecutionList, err
		}
	}
	if sandboxDef.PostGrantsSqlFile != "" {
		err := common.CopyFile(sandboxDef.PostGrantsSqlFile, postGrantSqlFile)
		if err != nil {
			return emptyExecutionList, err
		}
	}

	if len(sandboxDef.PreGrantsSql) > 0 {
		if common.FileExists(preGrantSqlFile) {
			err = common.AppendStrings(sandboxDef.PreGrantsSql, preGrantSqlFile, ";")
		} else {
			err = common.WriteStrings(sandboxDef.PreGrantsSql, preGrantSqlFile, ";")
		}
		if err != nil {
			return emptyExecutionList, err
		}
	}
	if len(sandboxDef.PostGrantsSql) > 0 {
		if common.FileExists(postGrantSqlFile) {
			err = common.AppendStrings(sandboxDef.PostGrantsSql, postGrantSqlFile, ";")
		} else {
			err = common.WriteStrings(sandboxDef.PostGrantsSql, postGrantSqlFile, ";")
		}
		if err != nil {
			return emptyExecutionList, err
		}
	}
	if !sandboxDef.SkipStart && sandboxDef.RunConcurrently {
		var eCommand2 = concurrent.ExecCommand{
			Cmd:  path.Join(sandboxDir, globals.ScriptStart),
			Args: []string{},
		}
		logger.Printf("Adding start command to execution list\n")
		execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 2, Command: eCommand2})
		if sandboxDef.LoadGrants {
			var eCommand3 = concurrent.ExecCommand{
				Cmd:  path.Join(sandboxDir, globals.ScriptLoadGrants),
				Args: []string{globals.ScriptPreGrantsSql},
			}
			var eCommand4 = concurrent.ExecCommand{
				Cmd:  path.Join(sandboxDir, globals.ScriptLoadGrants),
				Args: []string{},
			}
			var eCommand5 = concurrent.ExecCommand{
				Cmd:  path.Join(sandboxDir, globals.ScriptLoadGrants),
				Args: []string{globals.ScriptPostGrantsSql},
			}
			logger.Printf("Adding pre grants command to execution list\n")
			logger.Printf("Adding load grants command to execution list\n")
			logger.Printf("Adding post grants command to execution list\n")
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 3, Command: eCommand3})
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 4, Command: eCommand4})
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 5, Command: eCommand5})
		}
	} else {
		if !sandboxDef.SkipStart {
			logger.Printf("Running start script\n")
			_, err = common.RunCmd(path.Join(sandboxDir, globals.ScriptStart))
			if err != nil {
				return emptyExecutionList, err
			}
			if sandboxDef.LoadGrants {
				logger.Printf("Running pre grants script\n")
				_, err = common.RunCmdWithArgs(path.Join(sandboxDir, globals.ScriptLoadGrants), []string{globals.ScriptPreGrantsSql})
				if err != nil {
					return emptyExecutionList, err
				}
				logger.Printf("Running load grants script\n")
				_, err = common.RunCmd(path.Join(sandboxDir, globals.ScriptLoadGrants))
				if err != nil {
					return emptyExecutionList, err
				}
				logger.Printf("Running post grants script\n")
				_, err = common.RunCmdWithArgs(path.Join(sandboxDir, globals.ScriptLoadGrants), []string{globals.ScriptPostGrantsSql})
				if err != nil {
					return emptyExecutionList, err
				}
			}
		}
	}
	return execList, err
}

func writeScripts(scriptBatch ScriptBatch) error {
	for _, scriptDef := range scriptBatch.scripts {
		err := writeScript(scriptBatch.logger, scriptBatch.tc, scriptDef.scriptName, scriptDef.templateName,
			scriptBatch.sandboxDir, scriptBatch.data, scriptDef.makeExecutable)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeScript(logger *defaults.Logger, tempVar TemplateCollection, name, templateName, directory string, data common.StringMap, makeExecutable bool) error {
	if directory == "" {
		return fmt.Errorf("writeScript (%s): missing directory", name)
	}
	_, ok := tempVar[templateName]
	if !ok {
		return fmt.Errorf("writeScript (%s): template %s not found", name, templateName)
	}
	template := tempVar[templateName].Contents
	template = common.TrimmedLines(template)
	data["TemplateName"] = templateName
	text := common.TemplateFill(template, data)
	executableStatus := ""
	var err error
	if makeExecutable {
		err = writeExec(name, text, directory)
		executableStatus = " executable"
	} else {
		_, err = writeRegularFile(name, text, directory)
	}
	if err != nil {
		return err
	}
	if logger != nil {
		logger.Printf("Creating %s script '%s/%s' using template '%s'\n", executableStatus, common.ReplaceLiteralHome(directory), name, templateName)
	}
	return nil
}

func writeExec(filename, text, directory string) error {
	fname, err := writeRegularFile(filename, text, directory)
	if err != nil {
		return err
	}
	return os.Chmod(fname, 0744)
}

func writeRegularFile(fileName, text, directory string) (string, error) {
	fullName := path.Join(directory, fileName)
	err := common.WriteString(text, fullName)
	if err != nil {
		return "", err
	}
	return fullName, nil
}

func getLogDirFromSbDescription(fullPath string) (string, error) {
	sbDescription := path.Join(fullPath, globals.SandboxDescriptionName)
	logFile := ""
	logDirectory := ""
	if common.FileExists(sbDescription) {
		sbd, err := common.ReadSandboxDescription(fullPath)
		if err != nil {
			return "", err
		}
		logFile = sbd.LogFile
		if logFile != "" {
			logFile = common.ReplaceHomeVar(logFile)
			logDirectory = common.DirName(logFile)
			if !common.DirExists(logDirectory) {
				logDirectory = ""
			}
		}
	}
	return logDirectory, nil
}

func RemoveSandbox(sandboxDir, sandbox string, runConcurrently bool) (execList []concurrent.ExecutionList, err error) {
	fullPath := path.Join(sandboxDir, sandbox)
	if !common.DirExists(fullPath) {
		return emptyExecutionList, fmt.Errorf(globals.ErrDirectoryNotFound, fullPath)
	}
	preserve := path.Join(fullPath, globals.ScriptNoClearAll)
	if !common.ExecExists(preserve) {
		preserve = path.Join(fullPath, globals.ScriptNoClear)
	}
	if common.ExecExists(preserve) {
		fmt.Printf("The sandbox %s is locked\n", sandbox)
		fmt.Printf("You need to unlock it with \"dbdeployer admin unlock\"\n")
		return emptyExecutionList, err
	}
	logDirectory, err := getLogDirFromSbDescription(fullPath)
	if err != nil {
		return emptyExecutionList, err
	}
	stop := path.Join(fullPath, globals.ScriptStopAll)
	if !common.ExecExists(stop) {
		stop = path.Join(fullPath, globals.ScriptStop)
	}
	if !common.ExecExists(stop) {
		return emptyExecutionList, fmt.Errorf(globals.ErrExecutableNotFound, stop)
	}

	if runConcurrently {
		var eCommand1 = concurrent.ExecCommand{
			Cmd:  stop,
			Args: []string{},
		}
		execList = append(execList, concurrent.ExecutionList{Logger: nil, Priority: 0, Command: eCommand1})
	} else {
		if globals.UsingDbDeployer {
			fmt.Printf("Running %s\n", stop)
		}
		_, err := common.RunCmd(stop)
		if err != nil {
			return emptyExecutionList, fmt.Errorf(globals.ErrWhileStoppingSandbox, fullPath)
		}
	}

	rmTargets := []string{fullPath, logDirectory}

	for _, target := range rmTargets {
		if target == "" {
			continue
		}
		cmdStr := "rm"
		rmArgs := []string{"-rf", target}
		if runConcurrently {
			var eCommand2 = concurrent.ExecCommand{
				Cmd:  cmdStr,
				Args: rmArgs,
			}
			execList = append(execList, concurrent.ExecutionList{Logger: nil, Priority: 1, Command: eCommand2})
		} else {
			for _, item := range rmArgs {
				cmdStr += " " + item
			}
			if globals.UsingDbDeployer && target != logDirectory {
				fmt.Printf("Running %s\n", cmdStr)
			}
			_, err := common.RunCmdWithArgs("rm", rmArgs)
			if err != nil {
				return emptyExecutionList, fmt.Errorf(globals.ErrWhileDeletingSandbox, target)
			}
			if globals.UsingDbDeployer && target != logDirectory {
				fmt.Printf("Directory %s deleted\n", target)
			}
		}
	}
	// fmt.Printf("%#v\n",execList)
	return execList, nil
}
