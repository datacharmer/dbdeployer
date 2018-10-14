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
	"os"
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
	MorePorts            []int            // Additional ports that belong to thos sandbox
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

func GetOptionsFromFile(filename string) (options []string) {
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
	config := common.ParseConfigFile(filename)
	for _, kv := range config["mysqld"] {

		if skipOptions[kv.Key] {
			continue
		}
		options = append(options, fmt.Sprintf("%s = %s", kv.Key, kv.Value))
		//fmt.Printf("%d %s : %s \n", N, kv.key, kv.value)
	}
	return options
}

func SandboxDefToJson(sd SandboxDef) string {
	b, err := json.MarshalIndent(sd, " ", "\t")
	common.ErrCheckExitf(err, 1, "error encoding sandbox definition: %s", err)
	return fmt.Sprintf("%s", b)
}

func SmapToJson(data common.StringMap) string {
	copyright := data["Copyright"]
	data["Copyright"] = "[skipped] (See 'Copyright' template for full text)"
	b, err := json.MarshalIndent(data, " ", "\t")
	data["Copyright"] = copyright
	common.ErrCheckExitf(err, 1, "error encoding data: %s", err)
	return fmt.Sprintf("%s", b)
}

func isLocked(sbDir string) bool {
	return common.FileExists(sbDir+"/no_clear") || common.FileExists(sbDir+"/no_clear_all")
}

func CheckDirectory(sdef SandboxDef) SandboxDef {
	sandboxDir := sdef.SandboxDir
	if common.DirExists(sandboxDir) {
		if sdef.Force {
			if isLocked(sandboxDir) {
				common.Exitf(1, "Sandbox in %s is locked. Cannot be overwritten\nYou can unlock it with 'dbdeployer admin unlock %s'\n", sandboxDir, common.DirName(sandboxDir))
			}
			fmt.Printf("Overwriting directory %s\n", sandboxDir)
			stopCommand := sandboxDir + "/stop"
			if !common.ExecExists(stopCommand) {
				stopCommand = sandboxDir + "/stop_all"
			}
			if !common.ExecExists(stopCommand) {
				fmt.Printf("Neither 'stop' or 'stop_all' found in %s\n", sandboxDir)
			}

			usedPortsList := common.GetInstalledPorts(sandboxDir)
			myUsedPorts := make(map[int]bool)
			for _, p := range usedPortsList {
				myUsedPorts[p] = true
			}

			logDirectory := getLogDirFromSbDescription(sandboxDir)
			common.RunCmd(stopCommand)
			err, _ := common.RunCmdWithArgs("rm", []string{"-rf", sandboxDir})
			common.ErrCheckExitf(err, 1, "Error while deleting sandbox %s", sandboxDir)
			if logDirectory != "" {
				err, _ = common.RunCmdWithArgs("rm", []string{"-rf", logDirectory})
				common.ErrCheckExitf(err, 1, "Error while deleting log directory %s", logDirectory)
			}
			var newInstalledPorts []int
			for _, port := range sdef.InstalledPorts {
				if !myUsedPorts[port] {
					newInstalledPorts = append(newInstalledPorts, port)
				}
			}
			sdef.InstalledPorts = newInstalledPorts
		} else {
			common.Exitf(1, "Directory %s already exists. Use --force to override.", sandboxDir)
		}
	}
	return sdef
}

func CheckPort(caller string, sandboxType string, installedPorts []int, port int) {
	conflict := 0
	for _, p := range installedPorts {
		if p == port {
			conflict = p
		}
	}
	if conflict > 0 {
		common.Exitf(1, "Port conflict detected for %s (%s). Port %d is already used", sandboxType, caller, conflict)
	}
}

func FixServerUuid(sdef SandboxDef) (uuidFile, newUuid string) {
	if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 6, 9}) {
		return
	}
	newUuid = fmt.Sprintf("server-uuid=%s", common.MakeCustomizedUuid(sdef.Port, sdef.NodeNum))
	operationDir := sdef.SandboxDir + "/data"
	uuidFile = operationDir + "/auto.cnf"
	return
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

func setMysqlxProperties(sdef SandboxDef, globalTmpDir string) SandboxDef {
	mysqlxPort := sdef.MysqlXPort
	if mysqlxPort == 0 {
		mysqlxPort = common.FindFreePort(sdef.Port+defaults.Defaults().MysqlXPortDelta, sdef.InstalledPorts, 1)
	}
	sdef.MyCnfOptions = append(sdef.MyCnfOptions, fmt.Sprintf("mysqlx-port=%d", mysqlxPort))
	sdef.MyCnfOptions = append(sdef.MyCnfOptions, fmt.Sprintf("mysqlx-socket=%s/mysqlx-%d.sock", globalTmpDir, mysqlxPort))
	sdef.MorePorts = append(sdef.MorePorts, mysqlxPort)
	sdef.MysqlXPort = mysqlxPort
	return sdef
}

func CreateSingleSandbox(sdef SandboxDef) (execList []concurrent.ExecutionList) {

	var sandboxDir string

	if sdef.SBType == "" {
		sdef.SBType = "single"
	}
	logName := sdef.SBType
	if sdef.NodeNum > 0 {
		logName = fmt.Sprintf("%s-%d", logName, sdef.NodeNum)
	}
	fname, logger := defaults.NewLogger(common.LogDirName(), logName)
	sdef.LogFileName = common.ReplaceLiteralHome(fname)
	logger.Printf("Single Sandbox Definition: %s\n", SandboxDefToJson(sdef))
	if !common.DirExists(sdef.Basedir) {
		common.Exitf(1, "Base directory %s does not exist", sdef.Basedir)
	}

	if sdef.Port <= 1024 {
		common.Exitf(1, "Port for sandbox must be > 1024 (given:%d)", sdef.Port)
	}

	versionFname := common.VersionToName(sdef.Version)
	if sdef.Prompt == "" {
		sdef.Prompt = "mysql"
	}
	if sdef.DirName == "" {
		if sdef.Version != sdef.BasedirName {
			sdef.DirName = defaults.Defaults().SandboxPrefix + sdef.BasedirName
		} else {
			sdef.DirName = defaults.Defaults().SandboxPrefix + versionFname
		}
	}
	sandboxDir = sdef.SandboxDir + "/" + sdef.DirName
	sdef.SandboxDir = sandboxDir
	logger.Printf("Single Sandbox directory defined as %s\n", sdef.SandboxDir)
	datadir := sandboxDir + "/data"
	tmpdir := sandboxDir + "/tmp"
	globalTmpDir := os.Getenv("TMPDIR")
	if globalTmpDir == "" {
		globalTmpDir = "/tmp"
	}
	if !common.DirExists(globalTmpDir) {
		common.Exitf(1, "TMP directory %s does not exist", globalTmpDir)
	}
	if sdef.NodeNum == 0 && !sdef.Force {
		sdef.Port = common.FindFreePort(sdef.Port, sdef.InstalledPorts, 1)
		logger.Printf("Port defined as %d using FindFreePort \n", sdef.Port)
	}
	usingPlugins := false
	rightPluginDir := true // Assuming we can use the right plugin directory
	if sdef.EnableMysqlX {
		if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 12}) {
			common.Exit(1, "option --enable-mysqlx requires version 5.7.12+")
		}
		// If the version is 8.0.11 or later, MySQL X is enabled already
		if !common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, "plugin_load=mysqlx=mysqlx.so")
			sdef = setMysqlxProperties(sdef, globalTmpDir)
			logger.Printf("Added mysqlx plugin to my.cnf\n")
		}
		usingPlugins = true
	}
	if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) && !sdef.DisableMysqlX {
		usingPlugins = true
	}
	if sdef.ExposeDdTables {
		if !common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 0}) {
			common.Exit(1, "--expose-dd-tables requires MySQL 8.0.0+")
		}
		sdef.PostGrantsSql = append(sdef.PostGrantsSql, SingleTemplates["expose_dd_tables"].Contents)
		if sdef.CustomMysqld != "" && sdef.CustomMysqld != "mysqld-debug" {
			common.Exit(1,
				fmt.Sprintf("--expose-dd-tables requires mysqld-debug. A different file was indicated (--custom-mysqld=%s)", sdef.CustomMysqld),
				"Either use \"mysqld-debug\" or remove --custom-mysqld")
		}
		sdef.CustomMysqld = "mysqld-debug"
		logger.Printf("Using mysqld-debug for this sandbox\n")
	}
	if sdef.CustomMysqld != "" {
		customMysqld := sdef.Basedir + "/bin/" + sdef.CustomMysqld
		if !common.ExecExists(customMysqld) {
			common.Exit(1,
				fmt.Sprintf("File %s not found or not executable", customMysqld),
				fmt.Sprintf("The file \"%s\" (defined with --custom-mysqld) must be in the same directory as the regular mysqld", sdef.CustomMysqld))
		}
		pluginDebugDir := fmt.Sprintf("%s/lib/plugin/debug", sdef.Basedir)
		if sdef.CustomMysqld == "mysqld-debug" && common.DirExists(pluginDebugDir) {
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, fmt.Sprintf("plugin-dir=%s", pluginDebugDir))
		} else {
			rightPluginDir = false
		}
	}
	if common.GreaterOrEqualVersion(sdef.Version, []int{5, 1, 0}) {
		if sdef.EnableGeneralLog {
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, "general_log=1")
			logger.Printf("Enabling general log\n")
		}
		if sdef.InitGeneralLog {
			sdef.InitOptions = append(sdef.InitOptions, "--general_log=1")
			logger.Printf("Enabling general log during initialization\n")
		}
	}
	if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 4}) {
		if sdef.NativeAuthPlugin == true {
			sdef.InitOptions = append(sdef.InitOptions, "--default_authentication_plugin=mysql_native_password")
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, "default_authentication_plugin=mysql_native_password")
			logger.Printf("Using mysql_native_password for authentication\n")
		}
	}
	if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 11}) {
		if sdef.DisableMysqlX {
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, "mysqlx=OFF")
			logger.Printf("Disabling MySQLX\n")
		} else {
			sdef = setMysqlxProperties(sdef, globalTmpDir)
		}
	}
	mysqlshExecutable := fmt.Sprintf("%s/bin/mysqlsh", sdef.Basedir)
	if !common.ExecExists(mysqlshExecutable) {
		mysqlshExecutable = "mysqlsh"
	}
	if sdef.MyCnfFile != "" {
		options := GetOptionsFromFile(sdef.MyCnfFile)
		if len(options) > 0 {
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, fmt.Sprintf("# options retrieved from %s", sdef.MyCnfFile))
		}
		for _, option := range options {
			// fmt.Printf("[%s]\n", option)
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, option)
		}
	}
	if common.Includes(sliceToText(sdef.MyCnfOptions), "plugin.load") {
		usingPlugins = true
	}
	if common.Includes(sdef.SemiSyncOptions, "plugin.load") {
		usingPlugins = true
	}
	if usingPlugins {
		if !rightPluginDir {
			common.Exit(1,
				"The request of using mysqld-debug can't be honored.",
				"This deployment is using a plugin, but the debug",
				"directory for plugins was not found")
		}
	}
	timestamp := time.Now()
	var data common.StringMap = common.StringMap{"Basedir": sdef.Basedir,
		"Copyright":            SingleTemplates["Copyright"].Contents,
		"AppVersion":           common.VersionDef,
		"DateTime":             timestamp.Format(time.UnixDate),
		"SandboxDir":           sandboxDir,
		"CustomMysqld":         sdef.CustomMysqld,
		"Port":                 sdef.Port,
		"MysqlXPort":           sdef.MysqlXPort,
		"MysqlShell":           mysqlshExecutable,
		"BasePort":             sdef.BasePort,
		"Prompt":               sdef.Prompt,
		"Version":              sdef.Version,
		"Datadir":              datadir,
		"Tmpdir":               tmpdir,
		"GlobalTmpDir":         globalTmpDir,
		"DbUser":               sdef.DbUser,
		"DbPassword":           sdef.DbPassword,
		"RplUser":              sdef.RplUser,
		"RplPassword":          sdef.RplPassword,
		"RemoteAccess":         sdef.RemoteAccess,
		"BindAddress":          sdef.BindAddress,
		"OsUser":               os.Getenv("USER"),
		"ReplOptions":          sdef.ReplOptions,
		"GtidOptions":          sdef.GtidOptions,
		"ReplCrashSafeOptions": sdef.ReplCrashSafeOptions,
		"SemiSyncOptions":      sdef.SemiSyncOptions,
		"ExtraOptions":         sliceToText(sdef.MyCnfOptions),
		"ReportHost":           fmt.Sprintf("report-host=single-%d", sdef.Port),
		"ReportPort":           fmt.Sprintf("report-port=%d", sdef.Port),
		"HistoryDir":           sdef.HistoryDir,
	}
	if sdef.NodeNum != 0 {
		data["ReportHost"] = fmt.Sprintf("report-host = node-%d", sdef.NodeNum)
	}
	if sdef.SkipReportHost || sdef.SBType == "group-node" {
		data["ReportHost"] = ""
	}
	if sdef.SkipReportPort {
		data["ReportPort"] = ""
	}
	if sdef.ServerId > 0 {
		data["ServerId"] = fmt.Sprintf("server-id=%d", sdef.ServerId)
	} else {
		data["ServerId"] = ""
	}
	if common.DirExists(sandboxDir) {
		sdef = CheckDirectory(sdef)
	}
	logger.Printf("Checking port %d using CheckPort\n", sdef.Port)
	CheckPort("CreateSingleSandbox", sdef.SBType, sdef.InstalledPorts, sdef.Port)

	//fmt.Printf("creating: %s\n", sandbox_dir)
	common.Mkdir(sandboxDir)

	logger.Printf("Created directory %s\n", sdef.SandboxDir)
	logger.Printf("Single Sandbox template data: %s\n", SmapToJson(data))

	// fmt.Printf("creating: %s\n", datadir)
	common.Mkdir(datadir)
	logger.Printf("Created directory %s\n", datadir)
	// fmt.Printf("creating: %s\n", tmpdir)
	common.Mkdir(tmpdir)
	logger.Printf("Created directory %s\n", tmpdir)
	script := sdef.Basedir + "/scripts/mysql_install_db"
	initScriptFlags := ""
	if common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 0}) {
		script = sdef.Basedir + "/bin/mysqld"
		initScriptFlags = "--initialize-insecure"
	}
	// fmt.Printf("Script: %s\n", script)
	if !common.ExecExists(script) {
		common.Exitf(1, "Script '%s' not found", script)
	}
	if len(sdef.InitOptions) > 0 {
		for _, op := range sdef.InitOptions {
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

	if !sdef.KeepUuid {
		uuidFname, newUuid := FixServerUuid(sdef)
		if uuidFname != "" {
			data["FixUuidFile1"] = fmt.Sprintf(`echo "[auto]" > %s`, uuidFname)
			data["FixUuidFile2"] = fmt.Sprintf(`echo "%s" >> %s`, newUuid, uuidFname)
			logger.Printf("Created custom UUID %s\n", newUuid)
		}
	}

	writeScript(logger, SingleTemplates, "init_db", "init_db_template", sandboxDir, data, true)
	if sdef.RunConcurrently {
		var eCommand = concurrent.ExecCommand{
			Cmd:  sandboxDir + "/init_db",
			Args: []string{},
		}
		logger.Printf("Added init_db script to execution list\n")
		execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 0, Command: eCommand})
	} else {
		logger.Printf("Running init_db script \n")
		err, _ := common.RunCmdCtrl(sandboxDir+"/init_db", true)
		if err == nil {
			if !sdef.Multi {
				if defaults.UsingDbDeployer {
					fmt.Printf("Database installed in %s\n", common.ReplaceLiteralHome(sandboxDir))
					fmt.Printf("run 'dbdeployer usage single' for basic instructions'\n")
				}
			}
		} else {
			fmt.Printf("err: %s\n", err)
		}
	}

	sbItem := defaults.SandboxItem{
		Origin:      sdef.Basedir,
		SBType:      sdef.SBType,
		Version:     sdef.Version,
		Port:        []int{sdef.Port},
		Nodes:       []string{},
		Destination: sandboxDir,
	}

	if sdef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sdef.LogFileName)
	}
	sbDesc := common.SandboxDescription{
		Basedir: sdef.Basedir,
		SBType:  sdef.SBType,
		Version: sdef.Version,
		Port:    []int{sdef.Port},
		Nodes:   0,
		NodeNum: sdef.NodeNum,
		LogFile: sdef.LogFileName,
	}
	if len(sdef.MorePorts) > 0 {
		for _, port := range sdef.MorePorts {
			sbDesc.Port = append(sbDesc.Port, port)
			sbItem.Port = append(sbItem.Port, port)
		}
	}
	logger.Printf("Writing single sandbox description\n")
	common.WriteSandboxDescription(sandboxDir, sbDesc)
	if sdef.SBType == "single" {
		defaults.UpdateCatalog(sandboxDir, sbItem)
	}
	logger.Printf("Writing single sandbox scripts\n")
	writeScript(logger, SingleTemplates, "start", "start_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "status", "status_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "stop", "stop_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "clear", "clear_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "use", "use_template", sandboxDir, data, true)
	if sdef.MysqlXPort != 0 {
		writeScript(logger, SingleTemplates, "mysqlsh", "mysqlsh_template", sandboxDir, data, true)
	}
	writeScript(logger, SingleTemplates, "show_log", "show_log_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "send_kill", "send_kill_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "restart", "restart_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "load_grants", "load_grants_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "add_option", "add_option_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "my", "my_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "show_binlog", "show_binlog_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "show_relaylog", "show_relaylog_template", sandboxDir, data, true)
	writeScript(logger, SingleTemplates, "test_sb", "test_sb_template", sandboxDir, data, true)

	writeScript(logger, SingleTemplates, "my.sandbox.cnf", "my_cnf_template", sandboxDir, data, false)
	switch {
	case common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 0}):
		writeScript(logger, SingleTemplates, "grants.mysql", "grants_template8x", sandboxDir, data, false)
	case common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 6}):
		writeScript(logger, SingleTemplates, "grants.mysql", "grants_template57", sandboxDir, data, false)
	default:
		writeScript(logger, SingleTemplates, "grants.mysql", "grants_template5x", sandboxDir, data, false)
	}
	writeScript(logger, SingleTemplates, "sb_include", "sb_include_template", sandboxDir, data, false)

	preGrantSqlFile := sandboxDir + "/pre_grants.sql"
	postGrantSqlFile := sandboxDir + "/post_grants.sql"
	if sdef.PreGrantsSqlFile != "" {
		common.CopyFile(sdef.PreGrantsSqlFile, preGrantSqlFile)
	}
	if sdef.PostGrantsSqlFile != "" {
		common.CopyFile(sdef.PostGrantsSqlFile, postGrantSqlFile)
	}

	if len(sdef.PreGrantsSql) > 0 {
		if common.FileExists(preGrantSqlFile) {
			common.AppendStrings(sdef.PreGrantsSql, preGrantSqlFile, ";")
		} else {
			common.WriteStrings(sdef.PreGrantsSql, preGrantSqlFile, ";")
		}
	}
	if len(sdef.PostGrantsSql) > 0 {
		if common.FileExists(postGrantSqlFile) {
			common.AppendStrings(sdef.PostGrantsSql, postGrantSqlFile, ";")
		} else {
			common.WriteStrings(sdef.PostGrantsSql, postGrantSqlFile, ";")
		}
	}
	//common.Run_cmd(sandbox_dir + "/start", []string{})
	if !sdef.SkipStart && sdef.RunConcurrently {
		var eCommand2 = concurrent.ExecCommand{
			Cmd:  sandboxDir + "/start",
			Args: []string{},
		}
		logger.Printf("Adding start command to execution list\n")
		execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 2, Command: eCommand2})
		if sdef.LoadGrants {
			var eCommand3 = concurrent.ExecCommand{
				Cmd:  sandboxDir + "/load_grants",
				Args: []string{"pre_grants.sql"},
			}
			var eCommand4 = concurrent.ExecCommand{
				Cmd:  sandboxDir + "/load_grants",
				Args: []string{},
			}
			var eCommand5 = concurrent.ExecCommand{
				Cmd:  sandboxDir + "/load_grants",
				Args: []string{"post_grants.sql"},
			}
			logger.Printf("Adding pre grants command to execution list\n")
			logger.Printf("Adding load grants command to execution list\n")
			logger.Printf("Adding post grants command to execution list\n")
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 3, Command: eCommand3})
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 4, Command: eCommand4})
			execList = append(execList, concurrent.ExecutionList{Logger: logger, Priority: 5, Command: eCommand5})
		}
	} else {
		if !sdef.SkipStart {
			logger.Printf("Running start script\n")
			common.RunCmd(sandboxDir + "/start")
			if sdef.LoadGrants {
				logger.Printf("Running pre grants script\n")
				common.RunCmdWithArgs(sandboxDir+"/load_grants", []string{"pre_grants.sql"})
				logger.Printf("Running load grants script\n")
				common.RunCmd(sandboxDir + "/load_grants")
				logger.Printf("Running post grants script\n")
				common.RunCmdWithArgs(sandboxDir+"/load_grants", []string{"post_grants.sql"})
			}
		}
	}
	return
}

func writeScript(logger *defaults.Logger, tempVar TemplateCollection, name, templateName, directory string, data common.StringMap, makeExecutable bool) {
	template := tempVar[templateName].Contents
	template = common.TrimmedLines(template)
	data["TemplateName"] = templateName
	text := common.TemplateFill(template, data)
	executableStatus := ""
	if makeExecutable {
		writeExec(name, text, directory)
		executableStatus = " executable"
	} else {
		writeRegularFile(name, text, directory)
	}
	if logger != nil {
		logger.Printf("Creating%s script '%s/%s' using template '%s'\n", executableStatus, common.ReplaceLiteralHome(directory), name, templateName)
	}
}

func writeExec(filename, text, directory string) {
	fname := writeRegularFile(filename, text, directory)
	os.Chmod(fname, 0744)
}

func writeRegularFile(filename, text, directory string) string {
	fname := directory + "/" + filename
	common.WriteString(text, fname)
	return fname
}

func getLogDirFromSbDescription(fullPath string) string {
	sbdescription := fullPath + "/sbdescription.json"
	logFile := ""
	logDirectory := ""
	if common.FileExists(sbdescription) {
		sbd := common.ReadSandboxDescription(fullPath)
		logFile = sbd.LogFile
		if logFile != "" {
			logFile = common.ReplaceHomeVar(logFile)
			logDirectory = common.DirName(logFile)
			if !common.DirExists(logDirectory) {
				logDirectory = ""
			}
		}
	}
	return logDirectory
}

func RemoveSandbox(sandboxDir, sandbox string, runConcurrently bool) (execList []concurrent.ExecutionList) {
	fullPath := sandboxDir + "/" + sandbox
	if !common.DirExists(fullPath) {
		common.Exitf(1, "Directory '%s' not found", fullPath)
	}
	preserve := fullPath + "/no_clear_all"
	if !common.ExecExists(preserve) {
		preserve = fullPath + "/no_clear"
	}
	if common.ExecExists(preserve) {
		fmt.Printf("The sandbox %s is locked\n", sandbox)
		fmt.Printf("You need to unlock it with \"dbdeployer admin unlock\"\n")
		return
	}
	logDirectory := getLogDirFromSbDescription(fullPath)
	stop := fullPath + "/stop_all"
	if !common.ExecExists(stop) {
		stop = fullPath + "/stop"
	}
	if !common.ExecExists(stop) {
		common.Exitf(1, "Executable '%s' not found", stop)
	}

	if runConcurrently {
		var eCommand1 = concurrent.ExecCommand{
			Cmd:  stop,
			Args: []string{},
		}
		execList = append(execList, concurrent.ExecutionList{Logger: nil, Priority: 0, Command: eCommand1})
	} else {
		if defaults.UsingDbDeployer {
			fmt.Printf("Running %s\n", stop)
		}
		err, _ := common.RunCmd(stop)
		common.ErrCheckExitf(err, 1, "Error while stopping sandbox %s", fullPath)
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
			if defaults.UsingDbDeployer && target != logDirectory {
				fmt.Printf("Running %s\n", cmdStr)
			}
			err, _ := common.RunCmdWithArgs("rm", rmArgs)
			common.ErrCheckExitf(err, 1, "Error while deleting directory %s", target)
			if defaults.UsingDbDeployer && target != logDirectory {
				fmt.Printf("Directory %s deleted\n", target)
			}
		}
	}
	// fmt.Printf("%#v\n",exec_list)
	return
}
