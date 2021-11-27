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

package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/concurrent"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

type SandboxDef struct {
	DirName              string           // Name of the directory containing the sandbox
	SBType               string           // Type of sandbox (single, multiple, replication-node, group-node)
	Multi                bool             // CoalesceString single or part of a multiple sandbox
	NodeNum              int              // In multiple sandboxes, which node is this
	Version              string           // MySQL version
	Basedir              string           // Where to get binaries from (e.g. $HOME/opt/mysql/8.0.11)
	SbHost               string           // The host for this sandbox (default 127.0.0.1)
	Imported             bool             // The server is being imported
	ClientBasedir        string           // Where to get client binaries from (e.g. $HOME/opt/mysql/8.0.15)
	BasedirName          string           // The bare name of the directory containing the binaries (e.g. 8.0.11)
	SandboxDir           string           // Target directory for sandboxes
	ShellPath            string           // The Bash interpreter to use for generated scripts
	StartArgs            []string         // Arguments passed to the 'start' command
	LoadGrants           bool             // Should we load grants?
	SkipReportHost       bool             // Do not add report-host to my.sandbox.cnf
	SkipReportPort       bool             // Do not add report-port to my.sandbox.cnf
	SkipStart            bool             // Do not start the server after deployment
	InstalledPorts       []int            // Which ports should be skipped in port assignment for this SB
	Port                 int              // Port assigned to this sandbox
	MysqlXPort           int              // XPlugin port for this sandbox
	AdminPort            int              // Admin port for this sandbox (8.0.14+)
	UserPort             int              // Custom port provided by user
	BasePort             int              // Base port for calculating more ports in multiple SB
	MorePorts            []int            // Additional ports that belong to this sandbox
	Prompt               string           // Prompt to use in "mysql" client
	DbUser               string           // Database user name
	RplUser              string           // Replication user name
	DbPassword           string           // Database password
	RplPassword          string           // Replication password
	DefaultRole          string           // Role assigned to default user
	CustomRoleName       string           // Custom role name
	CustomRolePrivileges string           // Custom role privileges (such as 'SELECT, INSERT')
	CustomRoleTarget     string           // Custom role target (such as 'dbName.*', or '*.*')
	CustomRoleExtra      string           // Custom role extra (such as 'with grant option')
	TaskUser             string           // Additional user to be created on demand
	TaskUserRole         string           // Role to be assigned to task user
	RemoteAccess         string           // What access have the users created for this SB (127.%)
	BindAddress          string           // Bind address for this sandbox (127.0.0.1)
	CustomMysqld         string           // Use an alternative mysqld executable
	ServerId             int              // Server ID (for single sandbox)
	BaseServerId         int              // Base Server ID (for multiple sandboxes)
	ReplOptions          string           // Replication options, as string to append to my.sandbox.cnf
	GtidOptions          string           // Options needed for GTID
	ReplCrashSafeOptions string           // Options needed for Replication crash safe
	SemiSyncOptions      string           // Options for semi-synchronous replication
	ReadOnlyOptions      string           // Options for read-only passed to child sandboxes
	InitOptions          []string         // Options to be added to the initialization command
	MyCnfOptions         []string         // Options to be added to my.sandbox.cnf
	ChangeMasterOptions  []string         // Options to be added to CHANGE MASTER TO
	PreGrantsSql         []string         // SQL statements to execute before grants assignment
	PreGrantsSqlFile     string           // SQL file to load before grants assignment
	PostGrantsSql        []string         // SQL statements to run after grants assignment
	PostGrantsSqlFile    string           // SQL file to load after grants assignment
	MyCnfFile            string           // options file to merge with the SB my.sandbox.cnf
	HistoryDir           string           // Where to store the MySQL client history
	LogFileName          string           // Where to log operations for this sandbox
	Flavor               string           // The flavor of the binaries (MySQL, Percona, NDB, etc)
	PortAsServerId       bool             // Whether we use the port number as server ID
	FlavorInPrompt       bool             // Add flavor to prompt
	SocketInDatadir      bool             // Whether we want the socket in the data directory
	SlavesReadOnly       bool             // Whether slaves will set the read_only flag
	SlavesSuperReadOnly  bool             // Whether slaves will set the super_read_only flag
	Logger               *defaults.Logger // Carries a logger across sandboxes
	InitGeneralLog       bool             // Enable general log during server initialization
	EnableGeneralLog     bool             // Enable general log for regular usage
	NativeAuthPlugin     bool             // Use the native password plugin for MySQL 8.0.4+
	DisableMysqlX        bool             // Disable Xplugin (MySQL 8.0.11+)
	EnableMysqlX         bool             // Enable Xplugin (MySQL 5.7.12+)
	EnableAdminAddress   bool             // Enable Admin address (MySQL 8.0.14+)
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

func getOptionsFromFile(filename string) (options []string, err error) {
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
		//common.CondPrintf("%d %s : %s \n", N, kv.key, kv.value)
	}
	return options, nil
}

func sandboxDefToJson(sd SandboxDef) string {
	b, err := json.MarshalIndent(sd, " ", "\t")
	if err != nil {
		return "Sandbox definition could not be encoded\n"
	}
	return string(b)
}

func stringMapToJson(data common.StringMap) string {
	copyright := data[globals.TmplCopyright]
	data[globals.TmplCopyright] = "[skipped] (See 'copyright' template for full text)"
	b, err := json.MarshalIndent(data, " ", "\t")
	data[globals.TmplCopyright] = copyright
	if err != nil {
		return "String map could not be encoded"
	}
	return string(b)
}

func isLocked(sbDir string) bool {
	return common.FileExists(path.Join(sbDir, globals.ScriptNoClear)) || common.FileExists(path.Join(sbDir, globals.ScriptNoClearAll))
}

func checkDirectory(sandboxDef SandboxDef) (SandboxDef, error) {
	sandboxDir := sandboxDef.SandboxDir
	if common.DirExists(sandboxDir) {
		if sandboxDef.Force {
			if isLocked(sandboxDir) {
				return sandboxDef, fmt.Errorf("sandbox in %s is locked. Cannot be overwritten\nYou can unlock it with 'dbdeployer admin unlock %s'", sandboxDir, common.DirName(sandboxDir))
			}
			common.CondPrintf("Overwriting directory %s\n", sandboxDir)
			stopCommand := path.Join(sandboxDir, globals.ScriptStop)
			if !common.ExecExists(stopCommand) {
				stopCommand = path.Join(sandboxDir, globals.ScriptStopAll)
			}
			if !common.ExecExists(stopCommand) {
				common.CondPrintf("Neither 'stop' or 'stop_all' found in %s\n", sandboxDir)
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
			if !sandboxDef.Imported {
				var newInstalledPorts []int
				for _, port := range sandboxDef.InstalledPorts {
					if !myUsedPorts[port] {
						newInstalledPorts = append(newInstalledPorts, port)
					}
				}
				sandboxDef.InstalledPorts = newInstalledPorts
			}
		} else {
			return sandboxDef, fmt.Errorf("directory %s already exists. Use --force to override", sandboxDir)
		}
	}
	return sandboxDef, nil
}

func checkPortAvailability(caller string, sandboxType string, installedPorts []int, port int) error {
	if sandboxType == globals.SbTypeSingleImported {
		return nil
	}
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

func fixServerUuid(sandboxDef SandboxDef) (uuidDef string, uuidFile string, err error) {
	// 5.6.9
	// isMinimumGtid, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumGtidVersion)
	isMinimumGtid, err := common.HasCapability(sandboxDef.Flavor, common.GTID, sandboxDef.Version)
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

func setMysqlxProperties(sandboxDef SandboxDef, socketDir string) (SandboxDef, error) {
	if sandboxDef.Imported {
		return sandboxDef, nil
	}
	mysqlxPort := sandboxDef.MysqlXPort
	if mysqlxPort == 0 {
		var err error
		mysqlxPort, err = common.FindFreePort(sandboxDef.Port+defaults.Defaults().MysqlXPortDelta, sandboxDef.InstalledPorts, 1)
		if err != nil {
			return SandboxDef{}, errors.Wrapf(err, "error detecting free port for MySQLX")
		}
	}
	sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("mysqlx-port=%d", mysqlxPort))
	sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("mysqlx-socket=%s/mysqlx-%d.sock", socketDir, mysqlxPort))
	sandboxDef.MorePorts = append(sandboxDef.MorePorts, mysqlxPort)
	sandboxDef.MysqlXPort = mysqlxPort
	return sandboxDef, nil
}

func setAdminPortProperties(sandboxDef SandboxDef) (SandboxDef, error) {
	if sandboxDef.Imported {
		return sandboxDef, nil
	}
	if !sandboxDef.EnableAdminAddress {
		return sandboxDef, nil
	}
	isMinimumAdminAddress, err := common.HasCapability(sandboxDef.Flavor, common.AdminAddress, sandboxDef.Version)
	if err != nil {
		return sandboxDef, err
	}
	if !isMinimumAdminAddress {
		return sandboxDef, fmt.Errorf(globals.ErrOptionRequiresVersion, globals.EnableAdminAddressLabel,
			common.IntSliceToDottedString(globals.MinimumAdminAddressVersion))
	}
	adminPort := sandboxDef.AdminPort
	if adminPort == 0 {
		var err error
		adminPort, err = common.FindFreePort(sandboxDef.Port+defaults.Defaults().AdminPortDelta, sandboxDef.InstalledPorts, 1)
		if err != nil {
			return SandboxDef{}, errors.Wrapf(err, "error detecting free admin port")
		}
	}
	sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("admin-port=%d", adminPort))
	sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("admin-address=%s", sandboxDef.SbHost))
	sandboxDef.MorePorts = append(sandboxDef.MorePorts, adminPort)
	sandboxDef.AdminPort = adminPort
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
		sandboxDef.SBType = globals.SbTypeSingle
	}
	// Assuming a default flavor for backward compatibility
	if sandboxDef.Flavor == "" {
		sandboxDef.Flavor = common.MySQLFlavor
	}

	if sandboxDef.Flavor == common.TiDbFlavor {
		// Ensures that we can run a client.
		// Since TiDB tarballs don't include a client, we need to use one from MySQL
		// In theory, it could be possible to circumvent this necessity by using
		// the environment variable $MYSQL_EDITOR, but this would potentially
		// interfere with other sandboxes isolation. So I better leave this feature
		// undocumented and only to be used in emergencies.

		if sandboxDef.Prompt == "" || sandboxDef.Prompt == globals.PromptValue {
			if !sandboxDef.FlavorInPrompt {
				sandboxDef.Prompt = "TiDB"
			}
		}
		if sandboxDef.ClientBasedir == "" {
			return emptyExecutionList,
				fmt.Errorf("flavor '%s' requires option --'%s'", common.TiDbFlavor, globals.ClientFromLabel)
		}

		// Replaces main templates with the ones needed for TiDB
		for name, templateDesc := range TidbTemplates {
			re := regexp.MustCompile(`^` + tidbPrefix)
			singleName := re.ReplaceAllString(name, "")
			SingleTemplates[singleName] = templateDesc
		}
	}
	if sandboxDef.Imported {
		if sandboxDef.ClientBasedir == "" {
			return emptyExecutionList,
				fmt.Errorf("imported sandbox requires option --%s", globals.ClientFromLabel)
		}
		// Replaces main templates with the ones needed for imported
		for name, templateDesc := range ImportTemplates {
			re := regexp.MustCompile(`^` + importPrefix)
			singleName := re.ReplaceAllString(name, "")
			SingleTemplates[singleName] = templateDesc
		}
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
		sandboxDef.Logger = logger
	}
	logger.Printf("Single Sandbox Definition: %s\n", sandboxDefToJson(sandboxDef))
	if !common.DirExists(sandboxDef.Basedir) && !sandboxDef.Imported {
		return emptyExecutionList, fmt.Errorf(globals.ErrBaseDirectoryNotFound, sandboxDef.Basedir)
	}

	if sandboxDef.Port <= 1024 {
		return emptyExecutionList, fmt.Errorf("port for sandbox must be > 1024 (given:%d)", sandboxDef.Port)
	}

	versionFname := common.VersionToName(sandboxDef.Version)
	if sandboxDef.Prompt == "" && !sandboxDef.FlavorInPrompt {
		sandboxDef.Prompt = "mysql"
	}
	if sandboxDef.FlavorInPrompt {
		sandboxDef.Prompt = sandboxDef.Flavor + "-" + sandboxDef.Prompt
	}
	if sandboxDef.DirName == "" {
		prefix := defaults.Defaults().SandboxPrefix
		if sandboxDef.Imported {
			prefix = defaults.Defaults().ImportedSandboxPrefix
		}
		if sandboxDef.Version != sandboxDef.BasedirName {
			sandboxDef.DirName = prefix + sandboxDef.BasedirName
		} else {
			sandboxDef.DirName = prefix + versionFname
		}
	}
	if sandboxDef.DirName == globals.ForbiddenDirName {
		return emptyExecutionList, fmt.Errorf("the name %s cannot be used for a sandbox", sandboxDef.DirName)
	}
	sandboxDir = path.Join(sandboxDef.SandboxDir, sandboxDef.DirName)
	sandboxDef.SandboxDir = sandboxDir
	logger.Printf("Single Sandbox directory defined as %s\n", sandboxDef.SandboxDir)
	dataDir := path.Join(sandboxDir, globals.DataDirName)
	tmpDir := path.Join(sandboxDir, "tmp")

	globalTmpDir := common.GlobalTempDir()

	if !common.DirExists(globalTmpDir) {
		return emptyExecutionList, fmt.Errorf("TMP directory %s does not exist", globalTmpDir)
	}
	socketDir := globalTmpDir
	if sandboxDef.SocketInDatadir {
		socketDir = dataDir
	}
	if sandboxDef.NodeNum == 0 && !sandboxDef.Force && !sandboxDef.Imported {
		sandboxDef.Port, err = common.FindFreePort(sandboxDef.Port, sandboxDef.InstalledPorts, 1)
		if err != nil {
			return emptyExecutionList, errors.Wrapf(err, "error detecting free port for single sandbox")
		}
		logger.Printf("Port defined as %d using FindFreePort \n", sandboxDef.Port)
	}
	usingPlugins := false
	rightPluginDir := true // Assuming we can use the right plugin directory

	// 8.0.14
	sandboxDef, err = setAdminPortProperties(sandboxDef)
	if err != nil {
		return emptyExecutionList, err
	}

	if sandboxDef.EnableMysqlX && !sandboxDef.Imported {
		// 5.7.12
		// isMinimumMySQLX, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxVersion)
		isMinimumMySQLX, err := common.HasCapability(sandboxDef.Flavor, common.MySQLX, sandboxDef.Version)
		if err != nil {
			return emptyExecutionList, err
		}
		if !isMinimumMySQLX {
			return emptyExecutionList, fmt.Errorf(globals.ErrOptionRequiresVersion, globals.EnableMysqlXLabel,
				common.IntSliceToDottedString(globals.MinimumMysqlxVersion))
		}
		// If the version is 8.0.11 or later, MySQL X is enabled already
		// 8.0.11
		// isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
		isMinimumMySQLXDefault, err := common.HasCapability(sandboxDef.Flavor, common.MySQLXDefault, sandboxDef.Version)
		if err != nil {
			return emptyExecutionList, err
		}
		if !isMinimumMySQLXDefault {
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, "plugin_load_add=mysqlx=mysqlx.so")
			sandboxDef, err = setMysqlxProperties(sandboxDef, socketDir)
			if err != nil {
				return emptyExecutionList, err
			}
			logger.Printf("Added mysqlx plugin to my.cnf\n")
		}
		usingPlugins = true
	}
	// 8.0.11
	// isMinimumMySQLXDefault, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
	isMinimumMySQLXDefault, err := common.HasCapability(sandboxDef.Flavor, common.MySQLXDefault, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumMySQLXDefault && !sandboxDef.DisableMysqlX {
		usingPlugins = true
	}
	if sandboxDef.ExposeDdTables && !sandboxDef.Imported {
		// 8.0.0
		// isMinimumDataDictionary, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumDataDictionaryVersion)
		isMinimumDataDictionary, err := common.HasCapability(sandboxDef.Flavor, common.DataDict, sandboxDef.Version)
		if err != nil {
			return emptyExecutionList, err
		}
		if !isMinimumDataDictionary {
			return emptyExecutionList, fmt.Errorf(globals.ErrOptionRequiresVersion, "expose-dd-tables", common.IntSliceToDottedString(globals.MinimumDataDictionaryVersion))
		}
		sandboxDef.PostGrantsSql = append(sandboxDef.PostGrantsSql, SingleTemplates[globals.TmplExposeDdTables].Contents)
		if sandboxDef.CustomMysqld != "" && sandboxDef.CustomMysqld != "mysqld-debug" {
			return emptyExecutionList, fmt.Errorf("--expose-dd-tables requires mysqld-debug. A different file was indicated (--custom-mysqld=%s)\n%s",
				sandboxDef.CustomMysqld, "CoalesceString use \"mysqld-debug\" or remove --custom-mysqld")
		}
		sandboxDef.CustomMysqld = "mysqld-debug"
		logger.Printf("Using mysqld-debug for this sandbox\n")
	}
	if sandboxDef.CustomMysqld != "" {
		customMysqld := path.Join(sandboxDef.Basedir, "bin", sandboxDef.CustomMysqld)
		if !common.ExecExists(customMysqld) {
			return emptyExecutionList, fmt.Errorf("file %s not found or not executable\n"+
				"The file \"%s\" (defined with --custom-mysqld) must be in the same directory as the regular mysqld",
				customMysqld, sandboxDef.CustomMysqld)
		}
		pluginDebugDir := fmt.Sprintf("%s/lib/plugin/debug", sandboxDef.Basedir)
		if sandboxDef.CustomMysqld == "mysqld-debug" {
			if common.DirExists(pluginDebugDir) {
				sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("plugin-dir=%s", pluginDebugDir))
			} else {
				rightPluginDir = false
			}
		}
	}
	// 5.1.0
	// isMinimumDynVariables, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumDynVariablesVersion)
	isMinimumDynVariables, err := common.HasCapability(sandboxDef.Flavor, common.DynVariables, sandboxDef.Version)
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
	// isMinimumNativeAuthPlugin, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumNativeAuthPluginVersion)
	isMinimumNativeAuthPlugin, err := common.HasCapability(sandboxDef.Flavor, common.NativeAuth, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumNativeAuthPlugin {
		if sandboxDef.NativeAuthPlugin {
			sandboxDef.InitOptions = append(sandboxDef.InitOptions, "--default_authentication_plugin=mysql_native_password")
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, "default_authentication_plugin=mysql_native_password")
			logger.Printf("Using mysql_native_password for authentication\n")
		}
	}
	// 8.0.17
	isMinimumClonePlugin, err := common.HasCapability(sandboxDef.Flavor, common.CloneServer, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	// MariaDB 10.4.3 defaults to socket auth
	isMinimumRootAuth, err := common.HasCapability(sandboxDef.Flavor, common.RootAuth, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumRootAuth {
		sandboxDef.InitOptions = append(sandboxDef.InitOptions, "--auth-root-authentication-method=normal")
	}
	// 8.0.11
	// isMinimumMySQLXDefault, err = common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumMysqlxDefaultVersion)
	isMinimumMySQLXDefault, err = common.HasCapability(sandboxDef.Flavor, common.MySQLXDefault, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumMySQLXDefault {
		if sandboxDef.DisableMysqlX {
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, "mysqlx=OFF")
			logger.Printf("Disabling MySQLX\n")
		} else {
			sandboxDef, err = setMysqlxProperties(sandboxDef, socketDir)
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
		options, err := getOptionsFromFile(sandboxDef.MyCnfFile)
		if err != nil {
			return emptyExecutionList, errors.Wrapf(err, "error reading provided configuration file")
		}
		if len(options) > 0 {
			sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, fmt.Sprintf("# options retrieved from %s", sandboxDef.MyCnfFile))
		}
		sandboxDef.MyCnfOptions = append(sandboxDef.MyCnfOptions, options...)
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
	mysqlxSocket := ""
	if sandboxDef.MysqlXPort != 0 {
		mysqlxSocket = fmt.Sprintf("%s/mysqlx-%d.sock", socketDir, sandboxDef.MysqlXPort)
	}
	timestamp := time.Now()
	verList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, errors.Wrapf(err, "")
	}
	shortVersion := fmt.Sprintf("%d.%d", verList[0], verList[1])
	if sandboxDef.ClientBasedir == "" {
		sandboxDef.ClientBasedir = sandboxDef.Basedir
	}

	userDef, err := user.Current()

	if err != nil {
		return emptyExecutionList, fmt.Errorf("could not get information about current OS user")
	}

	if sandboxDef.SbHost == "" {
		sandboxDef.SbHost = "127.0.0.1"
	}

	var data = common.StringMap{
		"ShellPath":            sandboxDef.ShellPath,
		"SbHost":               sandboxDef.SbHost,
		"Basedir":              sandboxDef.Basedir,
		"ClientBasedir":        sandboxDef.ClientBasedir,
		"Copyright":            SingleTemplates[globals.TmplCopyright].Contents,
		"AppVersion":           common.VersionDef,
		"DateTime":             timestamp.Format(time.UnixDate),
		"SandboxDir":           sandboxDir,
		"CustomMysqld":         sandboxDef.CustomMysqld,
		"Port":                 sandboxDef.Port,
		"MysqlXPort":           sandboxDef.MysqlXPort,
		"MysqlXSocket":         mysqlxSocket,
		"AdminPort":            sandboxDef.AdminPort,
		"MysqlShell":           mysqlshExecutable,
		"BasePort":             sandboxDef.BasePort,
		"Prompt":               sandboxDef.Prompt,
		"Version":              sandboxDef.Version,
		"Flavor":               sandboxDef.Flavor,
		"SandboxType":          sandboxDef.SBType,
		"VersionMajor":         verList[0],
		"VersionMinor":         verList[1],
		"VersionRev":           verList[2],
		"SortableVersion":      fmt.Sprintf("%03d%03d%03d", verList[0], verList[1], verList[2]),
		"Datadir":              dataDir,
		"Tmpdir":               tmpDir,
		"GlobalTmpDir":         socketDir,
		"SocketFile":           path.Join(socketDir, fmt.Sprintf("mysql_sandbox%d.sock", sandboxDef.Port)),
		"DbUser":               sandboxDef.DbUser,
		"DbPassword":           sandboxDef.DbPassword,
		"RplUser":              sandboxDef.RplUser,
		"RplPassword":          sandboxDef.RplPassword,
		"DefaultRole":          sandboxDef.DefaultRole,
		"CustomRoleName":       sandboxDef.CustomRoleName,
		"CustomRolePrivileges": sandboxDef.CustomRolePrivileges,
		"CustomRoleTarget":     sandboxDef.CustomRoleTarget,
		"CustomRoleExtra":      sandboxDef.CustomRoleExtra,
		"TaskUser":             sandboxDef.TaskUser,
		"TaskUserRole":         sandboxDef.TaskUserRole,
		"RemoteAccess":         sandboxDef.RemoteAccess,
		"BindAddress":          sandboxDef.BindAddress,
		"OsUser":               userDef.Username,
		"ReplOptions":          sandboxDef.ReplOptions,
		"GtidOptions":          sandboxDef.GtidOptions,
		"ReplCrashSafeOptions": sandboxDef.ReplCrashSafeOptions,
		"SemiSyncOptions":      sandboxDef.SemiSyncOptions,
		"ReadOnlyOptions":      sandboxDef.ReadOnlyOptions,
		"ExtraOptions":         sliceToText(sandboxDef.MyCnfOptions),
		"ReportHost":           fmt.Sprintf("report-host=single-%d", sandboxDef.Port),
		"ReportPort":           fmt.Sprintf("report-port=%d", sandboxDef.Port),
		"HistoryDir":           sandboxDef.HistoryDir,
	}

	if sandboxDef.TaskUser != "" {
		if sandboxDef.TaskUserRole == "" {
			return emptyExecutionList, fmt.Errorf("task user defined but task role is empty")
		}

		data["TemplateName"] = globals.TmplTaskUserGrants
		taskUserText, err := common.SafeTemplateFill(
			globals.TmplTaskUserGrants,
			SingleTemplates[globals.TmplTaskUserGrants].Contents,
			data)
		if err != nil {
			return emptyExecutionList, fmt.Errorf("error filling task user template: %s", err)
		}
		data["TaskUserGrants"] = taskUserText
		delete(data, "TemplateName")
	} else {
		data["TaskUserGrants"] = ""
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
	if sandboxDef.PortAsServerId {
		sandboxDef.ServerId = sandboxDef.Port
	}
	if sandboxDef.ServerId > 0 {
		data["ServerId"] = fmt.Sprintf("server-id=%d", sandboxDef.ServerId)
	} else {
		data["ServerId"] = ""
	}
	if common.DirExists(sandboxDir) {
		sandboxDef, err = checkDirectory(sandboxDef)
		if err != nil {
			return emptyExecutionList, sbError("check directory", "%s", err)
		}
	}
	logger.Printf("Checking port %d using checkPortAvailability\n", sandboxDef.Port)
	err = checkPortAvailability("createSingleSandbox", sandboxDef.SBType, sandboxDef.InstalledPorts, sandboxDef.Port)
	if err != nil {
		return emptyExecutionList, sbError("check port", "%s", err)
	}

	err = os.Mkdir(sandboxDir, globals.PublicDirectoryAttr)
	if err != nil {
		return emptyExecutionList, sbError("sandbox dir creation", "%s", err)
	}

	logger.Printf("Created directory %s\n", sandboxDef.SandboxDir)
	logger.Printf("Single Sandbox template data: %s\n", stringMapToJson(data))

	err = os.Mkdir(dataDir, globals.PublicDirectoryAttr)
	if err != nil {
		return emptyExecutionList, sbError("data dir creation", "%s", err)
	}
	logger.Printf("Created directory %s\n", dataDir)
	err = os.Mkdir(tmpDir, globals.PublicDirectoryAttr)
	if err != nil {
		return emptyExecutionList, sbError("tmp dir creation", "%s", err)
	}
	logger.Printf("Created directory %s\n", tmpDir)
	script := ""
	initScriptFlags := ""
	// isMinimumDefaultInitialize, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumDefaultInitializeVersion)
	isMinimumDefaultInitialize, err := common.HasCapability(sandboxDef.Flavor, common.Initialize, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	if isMinimumDefaultInitialize {
		script = path.Join(sandboxDef.Basedir, "bin", globals.FnMysqld)
		if sandboxDef.CustomMysqld != "" {
			script = path.Join(sandboxDef.Basedir, "bin", sandboxDef.CustomMysqld)
		}
		initScriptFlags = "--initialize-insecure"
	}
	usesMysqlInstallDb, err := common.HasCapability(sandboxDef.Flavor, common.InstallDb, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	if usesMysqlInstallDb && !sandboxDef.Imported {
		script = path.Join(sandboxDef.Basedir, "scripts", globals.FnMysqlInstallDb)
	}
	if script != "" && !common.ExecExists(script) && !sandboxDef.Imported {
		common.CondPrintf("SCRIPT\n")
		return emptyExecutionList, fmt.Errorf(globals.ErrScriptNotFound, script)
	}
	if len(sandboxDef.InitOptions) > 0 {
		for _, op := range sandboxDef.InitOptions {
			initScriptFlags += " " + op
		}
	}
	data["InitScript"] = script
	data["InitDefaults"] = ""
	if script != "" {
		data["InitDefaults"] = "--no-defaults"
	}
	if initScriptFlags != "" {
		initScriptFlags = fmt.Sprintf("\\\n    %s", initScriptFlags)
	}
	data["ExtraInitFlags"] = initScriptFlags
	data["FixUuidFile1"] = ""
	data["FixUuidFile2"] = ""

	if !sandboxDef.KeepUuid {
		newUuid, uuidFname, err := fixServerUuid(sandboxDef)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "# UUID non-blocking failure: %s\n", err)
		}
		if uuidFname != "" {
			data["FixUuidFile1"] = fmt.Sprintf(`echo "[auto]" > %s`, uuidFname)
			data["FixUuidFile2"] = fmt.Sprintf(`echo "%s" >> %s`, newUuid, uuidFname)
			logger.Printf("Created custom UUID %s\n", newUuid)
		}
	}

	err = writeScript(logger, SingleTemplates, globals.ScriptInitDb, globals.TmplInitDb, sandboxDir, data, true)
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
		initOutput, err := common.RunCmdCtrl(initDbScript, true)
		if err == nil {
			if !sandboxDef.Multi {
				if globals.UsingDbDeployer {
					common.CondPrintf("Database installed in %s\n", common.ReplaceLiteralHome(sandboxDir))
					common.CondPrintf("run 'dbdeployer usage single' for basic instructions'\n")
				}
			}
		} else {
			common.CondPrintf("%s output: %s\n", globals.ScriptInitDb, initOutput)
			return emptyExecutionList, fmt.Errorf("%s failure: %s", globals.ScriptInitDb, err)
		}
	}

	sbItem := defaults.SandboxItem{
		Origin:      sandboxDef.Basedir,
		SBType:      sandboxDef.SBType,
		Version:     sandboxDef.Version,
		Flavor:      sandboxDef.Flavor,
		Host:        sandboxDef.SbHost,
		Port:        []int{sandboxDef.Port},
		Nodes:       []string{},
		Destination: sandboxDir,
	}

	if sandboxDef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sandboxDef.LogFileName)
	}
	sbDesc := common.SandboxDescription{
		Basedir:       sandboxDef.Basedir,
		ClientBasedir: sandboxDef.ClientBasedir,
		SBType:        sandboxDef.SBType,
		Version:       sandboxDef.Version,
		Flavor:        sandboxDef.Flavor,
		Host:          sandboxDef.SbHost,
		Port:          []int{sandboxDef.Port},
		Nodes:         0,
		NodeNum:       sandboxDef.NodeNum,
		LogFile:       sandboxDef.LogFileName,
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
	if sandboxDef.SBType == globals.SbTypeSingle || sandboxDef.SBType == globals.SbTypeSingleImported {
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
			{globals.ScriptStart, globals.TmplStart, true},
			{globals.ScriptStatus, globals.TmplStatus, true},
			{globals.ScriptStop, globals.TmplStop, true},
			{globals.ScriptClear, globals.TmplClear, true},
			{globals.ScriptUse, globals.TmplUse, true},
			{globals.ScriptShowLog, globals.TmplShowLog, true},
			{globals.ScriptShowBinlog, globals.TmplShowBinlog, true},
			{globals.ScriptShowRelayLog, globals.TmplShowRelaylog, true},
			{globals.ScriptSendKill, globals.TmplSendKill, true},
			{globals.ScriptRestart, globals.TmplRestart, true},
			{globals.ScriptLoadGrants, globals.TmplLoadGrants, true},
			{globals.ScriptAddOption, globals.TmplAddOption, true},
			{globals.ScriptMy, globals.TmplMy, true},
			{globals.ScriptTestSb, globals.TmplTestSb, true},
			{globals.ScriptMySandboxCnf, globals.TmplMyCnf, false},
			{globals.ScriptAfterStart, globals.TmplAfterStart, true},
			{globals.ScriptConnectionSql, globals.TmplConnectionInfoSql, false},
			{globals.ScriptConnectionConf, globals.TmplConnectionInfoConf, false},
			{globals.ScriptConnectionJson, globals.TmplConnectionInfoJson, false},
			{globals.ScriptReplicateFrom, globals.TmplReplicateFrom, true},
			{globals.ScriptMetadata, globals.TmplMetadata, true},
			{globals.ScriptSysbench, globals.TmplSysbench, true},
			{globals.ScriptSysbenchReady, globals.TmplSysbenchReady, true},
			{globals.ScriptWipeAndRestart, globals.TmplWipeAndRestart, true},
		},
	}
	if sandboxDef.EnableAdminAddress {
		sb.scripts = append(sb.scripts, ScriptDef{globals.ScriptUseAdmin, globals.TmplUseAdmin, true})
	}
	if sandboxDef.MysqlXPort != 0 {
		sb.scripts = append(sb.scripts, ScriptDef{globals.ScriptMysqlsh, globals.TmplMysqlsh, true})
	}
	if isMinimumClonePlugin {
		sb.scripts = append(sb.scripts, ScriptDef{
			globals.ScriptCloneFrom, globals.TmplCloneFrom, true})
		sb.scripts = append(sb.scripts, ScriptDef{
			globals.ScriptCloneConnectionSql, globals.TmplCloneConnectionSql, false})
		logger.Printf("enabling clone scripts")
	}
	var grantsTemplateName string = ""
	// isMinimumRoles, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumRolesVersion)
	isMinimumRoles, err := common.HasCapability(sandboxDef.Flavor, common.Roles, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	// isMinimumCreateUserVersion, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumCreateUserVersion)
	isMinimumCreateUserVersion, err := common.HasCapability(sandboxDef.Flavor, common.CreateUser, sandboxDef.Version)
	if err != nil {
		return emptyExecutionList, err
	}
	switch {
	// 8.0.0
	case shortVersion == "7.4":
		grantsTemplateName = globals.TmplGrants5x
	case isMinimumRoles:
		grantsTemplateName = globals.TmplGrants8x
		// 5.7.6
	case isMinimumCreateUserVersion:
		grantsTemplateName = globals.TmplGrants57
	default:
		grantsTemplateName = globals.TmplGrants5x
	}
	sb.scripts = append(sb.scripts, ScriptDef{globals.ScriptGrantsMysql, grantsTemplateName, false})
	sb.scripts = append(sb.scripts, ScriptDef{globals.ScriptSbInclude, globals.TmplSbInclude, false})

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
			err = common.AppendStrings(sandboxDef.PreGrantsSql, preGrantSqlFile, ";\n")
		} else {
			err = common.WriteStrings(sandboxDef.PreGrantsSql, preGrantSqlFile, ";\n")
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
			Args: sandboxDef.StartArgs,
		}
		logger.Printf("Adding start command to execution list\n")
		execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 2, Command: eCommand2})
		if sandboxDef.LoadGrants {
			var (
				eCmdAfterStart = concurrent.ExecCommand{
					Cmd:  path.Join(sandboxDir, globals.ScriptAfterStart),
					Args: []string{},
				}
				eCmdPreGrants = concurrent.ExecCommand{
					Cmd:  path.Join(sandboxDir, globals.ScriptLoadGrants),
					Args: []string{globals.ScriptPreGrantsSql},
				}
				eCmdLoadGrants = concurrent.ExecCommand{
					Cmd:  path.Join(sandboxDir, globals.ScriptLoadGrants),
					Args: []string{},
				}
				eCmdPostGrants = concurrent.ExecCommand{
					Cmd:  path.Join(sandboxDir, globals.ScriptLoadGrants),
					Args: []string{globals.ScriptPostGrantsSql},
				}
			)
			logger.Printf("Adding after start command to execution list\n")
			logger.Printf("Adding pre grants command to execution list\n")
			logger.Printf("Adding load grants command to execution list\n")
			logger.Printf("Adding post grants command to execution list\n")
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 3, Command: eCmdAfterStart})
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 4, Command: eCmdPreGrants})
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 5, Command: eCmdLoadGrants})
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 6, Command: eCmdPostGrants})
		}
	} else {
		if !sandboxDef.SkipStart {
			logger.Printf("Running start script\n")
			if len(sandboxDef.StartArgs) > 0 {
				_, err = common.RunCmdWithArgs(path.Join(sandboxDir, globals.ScriptStart), sandboxDef.StartArgs)
			} else {
				_, err = common.RunCmd(path.Join(sandboxDir, globals.ScriptStart))
			}
			if err != nil {
				return emptyExecutionList, err
			}
			logger.Printf("Running after start script\n")
			_, err = common.RunCmd(path.Join(sandboxDir, globals.ScriptAfterStart))
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

func writeScript(logger *defaults.Logger, tempVar TemplateCollection, scriptName, templateName, directory string,
	data common.StringMap, makeExecutable bool) error {
	if directory == "" {
		return fmt.Errorf("writeScript (%s): missing directory", scriptName)
	}
	_, ok := tempVar[templateName]
	if !ok {
		return fmt.Errorf("writeScript (%s): template %s not found", scriptName, templateName)
	}
	template := tempVar[templateName].Contents
	template = common.TrimmedLines(template)
	data["TemplateName"] = templateName
	var err error
	text, err := common.SafeTemplateFill(templateName, template, data)
	if err != nil {
		return err
	}
	executableStatus := ""
	if makeExecutable {
		err = writeExec(scriptName, text, directory)
		executableStatus = " executable"
	} else {
		_, err = writeRegularFile(scriptName, text, directory)
	}
	if err != nil {
		return err
	}
	if logger != nil {
		logger.Printf("Creating %s script '%s/%s' using template '%s'\n", executableStatus, common.ReplaceLiteralHome(directory), scriptName, templateName)
	}
	return nil
}

func writeExec(filename, text, directory string) error {
	fname, err := writeRegularFile(filename, text, directory)
	if err != nil {
		return err
	}
	return os.Chmod(fname, globals.ExecutableFileAttr)
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

// Deprecated: use RemoveCustomSandbox instead
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
		common.CondPrintf("The sandbox %s is locked\n", sandbox)
		common.CondPrintf("You need to unlock it with \"dbdeployer admin unlock\"\n")
		return emptyExecutionList, err
	}
	logDirectory, err := getLogDirFromSbDescription(fullPath)
	if err != nil {
		return emptyExecutionList, err
	}

	sandboxDefaultMarker := path.Join(fullPath, "is_default")
	stop := path.Join(fullPath, globals.ScriptStopAll)
	useKill := common.IsEnvSet("KILL_AS_STOP")
	if useKill {
		stop = path.Join(fullPath, globals.ScriptSendKillAll)
	}
	if !common.ExecExists(stop) {
		stop = path.Join(fullPath, globals.ScriptStop)
		if useKill {
			stop = path.Join(fullPath, globals.ScriptSendKill)
		}
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
		common.CondPrintf("Running %s\n", stop)
		_, err = common.RunCmd(stop)
		if err != nil {
			return emptyExecutionList, fmt.Errorf(globals.ErrWhileStoppingSandbox, fullPath)
		}
	}

	if common.FileExists(sandboxDefaultMarker) {
		defaultExecutable, err := common.SlurpAsString(sandboxDefaultMarker)
		if err != nil {
			return emptyExecutionList, fmt.Errorf("error reading default address from %s: %s",
				sandboxDefaultMarker, err)
		}
		defaultExecutable = strings.TrimSpace(defaultExecutable)
		if common.FileExists(defaultExecutable) {
			_ = os.Remove(defaultExecutable)
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
				common.CondPrintf("Running %s\n", cmdStr)
			}
			_, err = common.RunCmdWithArgs("rm", rmArgs)
			if err != nil {
				return emptyExecutionList, fmt.Errorf(globals.ErrWhileDeletingSandbox, target)
			}
			if globals.UsingDbDeployer && target != logDirectory {
				common.CondPrintf("Directory %s deleted\n", target)
			}
		}
	}
	return execList, nil
}

func RemoveCustomSandbox(sandboxDir, sandbox string, runConcurrently, useStop bool) (execList []concurrent.ExecutionList, err error) {
	fullPath := path.Join(sandboxDir, sandbox)
	if !common.DirExists(fullPath) {
		return emptyExecutionList, fmt.Errorf(globals.ErrDirectoryNotFound, fullPath)
	}
	preserve := path.Join(fullPath, globals.ScriptNoClearAll)
	if !common.ExecExists(preserve) {
		preserve = path.Join(fullPath, globals.ScriptNoClear)
	}
	if common.ExecExists(preserve) {
		common.CondPrintf("The sandbox %s is locked\n", sandbox)
		common.CondPrintf("You need to unlock it with \"dbdeployer admin unlock\"\n")
		return emptyExecutionList, err
	}
	logDirectory, err := getLogDirFromSbDescription(fullPath)
	if err != nil {
		return emptyExecutionList, err
	}

	stopCmd := ""
	sandboxDefaultMarker := path.Join(fullPath, "is_default")
	var stopArgs []string
	stop := path.Join(fullPath, globals.ScriptStop)
	stopAll := path.Join(fullPath, globals.ScriptStopAll)
	sendKillAll := path.Join(fullPath, globals.ScriptSendKillAll)
	sendKill := path.Join(fullPath, globals.ScriptSendKill)

	if useStop {
		if common.ExecExists(stopAll) {
			stopCmd = stopAll
		} else {
			if common.ExecExists(stop) {
				stopCmd = stop
			} else {
				return emptyExecutionList, fmt.Errorf(globals.ErrExecutableNotFound, stop)
			}
		}
	} else {
		if common.ExecExists(sendKillAll) {
			stopCmd = sendKillAll
		} else {
			if common.ExecExists(sendKill) {
				stopCmd = sendKill
			} else {
				return emptyExecutionList, fmt.Errorf(globals.ErrExecutableNotFound, sendKill)
			}
		}
		stopArgs = append(stopArgs, "destroy")
	}

	if runConcurrently {
		var eCommand1 = concurrent.ExecCommand{
			Cmd:  stopCmd,
			Args: stopArgs,
		}
		execList = append(execList, concurrent.ExecutionList{Logger: nil, Priority: 0, Command: eCommand1})
	} else {
		if useStop {
			common.CondPrintf("Running %s\n", stopCmd)
			_, err = common.RunCmd(stopCmd)
		} else {
			common.CondPrintf("Running %s destroy\n", stopCmd)
			_, err = common.RunCmdCtrlWithArgs(stopCmd, stopArgs, false)
		}
		if err != nil {
			return emptyExecutionList, fmt.Errorf(globals.ErrWhileStoppingSandbox, fullPath)
		}
	}

	if common.FileExists(sandboxDefaultMarker) {
		defaultExecutable, err := common.SlurpAsString(sandboxDefaultMarker)
		if err != nil {
			return emptyExecutionList, fmt.Errorf("error reading default address from %s: %s",
				sandboxDefaultMarker, err)
		}
		defaultExecutable = strings.TrimSpace(defaultExecutable)
		if common.FileExists(defaultExecutable) {
			_ = os.Remove(defaultExecutable)
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
				common.CondPrintf("Running %s\n", cmdStr)
			}
			_, err = common.RunCmdWithArgs("rm", rmArgs)
			if err != nil {
				return emptyExecutionList, fmt.Errorf(globals.ErrWhileDeletingSandbox, target)
			}
			if globals.UsingDbDeployer && target != logDirectory {
				common.CondPrintf("Directory %s deleted\n", target)
			}
		}
	}
	return execList, nil
}
