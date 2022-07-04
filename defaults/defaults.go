// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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

package defaults

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
)

type DbdeployerDefaults struct {
	Version                       string `json:"version"`
	SandboxHome                   string `json:"sandbox-home"`
	SandboxBinary                 string `json:"sandbox-binary"`
	UseSandboxCatalog             bool   `json:"use-sandbox-catalog"`
	LogSBOperations               bool   `json:"log-sb-operations"`
	LogDirectory                  string `json:"log-directory"`
	CookbookDirectory             string `json:"cookbook-directory"`
	ShellPath                     string `json:"shell-path"`
	MasterSlaveBasePort           int    `json:"master-slave-base-port"`
	GroupReplicationBasePort      int    `json:"group-replication-base-port"`
	GroupReplicationSpBasePort    int    `json:"group-replication-sp-base-port"`
	FanInReplicationBasePort      int    `json:"fan-in-replication-base-port"`
	AllMastersReplicationBasePort int    `json:"all-masters-replication-base-port"`
	MultipleBasePort              int    `json:"multiple-base-port"`
	PxcBasePort                   int    `json:"pxc-base-port"`
	NdbBasePort                   int    `json:"ndb-base-port"`
	NdbClusterPort                int    `json:"ndb-cluster-port"`
	GroupPortDelta                int    `json:"group-port-delta"`
	MysqlXPortDelta               int    `json:"mysqlx-port-delta"`
	AdminPortDelta                int    `json:"admin-port-delta"`
	MasterName                    string `json:"master-name"`
	MasterAbbr                    string `json:"master-abbr"`
	NodePrefix                    string `json:"node-prefix"`
	SlavePrefix                   string `json:"slave-prefix"`
	SlaveAbbr                     string `json:"slave-abbr"`
	SandboxPrefix                 string `json:"sandbox-prefix"`
	ImportedSandboxPrefix         string `json:"imported-sandbox-prefix"`
	MasterSlavePrefix             string `json:"master-slave-prefix"`
	GroupPrefix                   string `json:"group-prefix"`
	GroupSpPrefix                 string `json:"group-sp-prefix"`
	MultiplePrefix                string `json:"multiple-prefix"`
	FanInPrefix                   string `json:"fan-in-prefix"`
	AllMastersPrefix              string `json:"all-masters-prefix"`
	ReservedPorts                 []int  `json:"reserved-ports"`
	RemoteRepository              string `json:"remote-repository"`
	RemoteIndexFile               string `json:"remote-index-file"`
	RemoteCompletionUrl           string `json:"remote-completion-url"`
	RemoteTarballUrl              string `json:"remote-tarball-url"`
	PxcPrefix                     string `json:"pxc-prefix"`
	NdbPrefix                     string `json:"ndb-prefix"`
	DefaultSandboxExecutable      string `json:"default-sandbox-executable"`
	DownloadNameLinux             string `json:"download-name-linux"`
	DownloadNameMacOs             string `json:"download-name-macos"`
	DownloadUrl                   string `json:"download-url"`
	Timestamp                     string `json:"timestamp"`
}

const (
	minPortValue            int    = globals.MinAllowedPort // = 1100
	maxPortValue            int    = 30000
	ConfigurationDirName    string = ".dbdeployer"
	ConfigurationFileName   string = "config.json"
	ArchivesFileName        string = "archives.json"
	SandboxRegistryName     string = "sandboxes.json"
	SandboxRegistryLockName string = "sandboxes.lock"
)

var (
	homeDir                 string = os.Getenv("HOME")
	ConfigurationDir        string = path.Join(homeDir, ConfigurationDirName)
	ConfigurationFile       string = path.Join(ConfigurationDir, ConfigurationFileName)
	ArchivesFile            string = path.Join(ConfigurationDir, ArchivesFileName)
	CustomConfigurationFile string = ""
	SandboxRegistry         string = path.Join(ConfigurationDir, SandboxRegistryName)
	SandboxRegistryLock     string = path.Join(common.GlobalTempDir(), SandboxRegistryLockName)
	LogSBOperations         bool   = common.IsEnvSet("DBDEPLOYER_LOGGING")

	factoryDefaults = DbdeployerDefaults{
		Version:                       common.CompatibleVersion,
		SandboxHome:                   path.Join(homeDir, "sandboxes"),
		SandboxBinary:                 path.Join(homeDir, "opt", "mysql"),
		UseSandboxCatalog:             true,
		LogSBOperations:               false,
		LogDirectory:                  path.Join(homeDir, "sandboxes", "logs"),
		CookbookDirectory:             "recipes",
		ShellPath:                     globals.ShellPathValue,
		MasterSlaveBasePort:           11000,
		GroupReplicationBasePort:      12000,
		GroupReplicationSpBasePort:    13000,
		FanInReplicationBasePort:      14000,
		AllMastersReplicationBasePort: 15000,
		MultipleBasePort:              16000,
		PxcBasePort:                   18000,
		NdbBasePort:                   19000,
		NdbClusterPort:                20000,
		GroupPortDelta:                125,
		MysqlXPortDelta:               10000,
		AdminPortDelta:                11000,
		MasterName:                    "master",
		MasterAbbr:                    "m",
		NodePrefix:                    "node",
		SlavePrefix:                   "slave",
		SlaveAbbr:                     "s",
		SandboxPrefix:                 "msb_",
		ImportedSandboxPrefix:         "imp_msb_",
		MasterSlavePrefix:             "rsandbox_",
		GroupPrefix:                   "group_msb_",
		GroupSpPrefix:                 "group_sp_msb_",
		MultiplePrefix:                "multi_msb_",
		FanInPrefix:                   "fan_in_msb_",
		AllMastersPrefix:              "all_masters_msb_",
		ReservedPorts:                 globals.ReservedPorts,
		RemoteRepository:              "https://raw.githubusercontent.com/datacharmer/mysql-docker-minimal/master/dbdata",
		RemoteIndexFile:               "available.json",
		RemoteCompletionUrl:           "https://raw.githubusercontent.com/datacharmer/dbdeployer/master/docs/dbdeployer_completion.sh",
		RemoteTarballUrl:              "https://raw.githubusercontent.com/datacharmer/dbdeployer/master/downloads/tarball_list.json",
		NdbPrefix:                     "ndb_msb_",
		PxcPrefix:                     "pxc_msb_",
		DefaultSandboxExecutable:      "default",
		DownloadNameLinux:             "mysql-{{.Version}}-linux-glibc2.17-x86_64{{.Minimal}}.{{.Ext}}",
		DownloadNameMacOs:             "mysql-{{.Version}}-macos11-x86_64.{{.Ext}}",
		DownloadUrl:                   "https://dev.mysql.com/get/Downloads/MySQL",
		Timestamp:                     time.Now().Format(time.UnixDate),
	}
	currentDefaults DbdeployerDefaults
)

func Defaults() DbdeployerDefaults {
	if currentDefaults.Version == "" {
		if common.FileExists(ConfigurationFile) {
			currentDefaults = ReadDefaultsFile(ConfigurationFile)
		} else {
			currentDefaults = factoryDefaults
		}
	}
	if currentDefaults.LogSBOperations {
		LogSBOperations = true
	}
	return currentDefaults
}

func ShowDefaults(defaults DbdeployerDefaults) {
	defaults = replaceLiteralEnvValues(defaults)
	if common.FileExists(ConfigurationFile) {
		common.CondPrintf("# Configuration file: %s\n", ConfigurationFile)
	} else {
		common.CondPrintln("# Internal values:")
	}
	b, err := json.MarshalIndent(defaults, " ", "\t")
	common.ErrCheckExitf(err, 1, globals.ErrEncodingDefaults, err)
	common.CondPrintf("%s\n", b)
}

func WriteDefaultsFile(filename string, defaults DbdeployerDefaults) {
	defaults = replaceLiteralEnvValues(defaults)
	defaultsDir := common.DirName(filename)
	if !common.DirExists(defaultsDir) {
		common.Mkdir(defaultsDir)
	}
	b, err := json.MarshalIndent(defaults, " ", "\t")
	common.ErrCheckExitf(err, 1, globals.ErrEncodingDefaults, err)
	jsonString := string(b)
	err = common.WriteString(jsonString, filename)
	common.ErrCheckExitf(err, 1, "error writing defaults file")
}

func expandEnvironmentVariables(defaults DbdeployerDefaults) DbdeployerDefaults {
	defaults.SandboxHome = common.ReplaceEnvVar(defaults.SandboxHome, "HOME")
	defaults.SandboxHome = common.ReplaceEnvVar(defaults.SandboxHome, "PWD")
	defaults.SandboxBinary = common.ReplaceEnvVar(defaults.SandboxBinary, "HOME")
	defaults.SandboxBinary = common.ReplaceEnvVar(defaults.SandboxBinary, "PWD")
	return defaults
}

func replaceLiteralEnvValues(defaults DbdeployerDefaults) DbdeployerDefaults {
	defaults.SandboxHome = common.ReplaceLiteralEnvVar(defaults.SandboxHome, "HOME")
	defaults.SandboxHome = common.ReplaceLiteralEnvVar(defaults.SandboxHome, "PWD")
	defaults.SandboxBinary = common.ReplaceLiteralEnvVar(defaults.SandboxBinary, "HOME")
	defaults.SandboxBinary = common.ReplaceLiteralEnvVar(defaults.SandboxBinary, "PWD")
	return defaults
}

func ReadDefaultsFile(filename string) (defaults DbdeployerDefaults) {
	defaultsBlob, err := common.SlurpAsBytes(filename)
	common.ErrCheckExitf(err, 1, "error reading defaults file %s: %s", filename, err)

	err = json.Unmarshal(defaultsBlob, &defaults)
	common.ErrCheckExitf(err, 1, globals.ErrEncodingDefaults, err)
	defaults = expandEnvironmentVariables(defaults)
	return
}

func checkInt(name string, val, min, max int) bool {
	if val >= min && val <= max {
		return true
	}
	common.CondPrintf("Value %s (%d) must be between %d and %d\n", name, val, min, max)
	return false
}

func ValidateDefaults(nd DbdeployerDefaults) bool {
	allIntegers := checkInt("master-slave-base-port", nd.MasterSlaveBasePort, minPortValue, maxPortValue) &&
		checkInt("group-replication-base-port", nd.GroupReplicationBasePort, minPortValue, maxPortValue) &&
		checkInt("group-replication-sp-base-port", nd.GroupReplicationSpBasePort, minPortValue, maxPortValue) &&
		checkInt("multiple-base-port", nd.MultipleBasePort, minPortValue, maxPortValue) &&
		checkInt("fan-in-base-port", nd.FanInReplicationBasePort, minPortValue, maxPortValue) &&
		checkInt("all-masters-base-port", nd.AllMastersReplicationBasePort, minPortValue, maxPortValue) &&
		checkInt("pxc-base-port", nd.PxcBasePort, minPortValue, maxPortValue) &&
		checkInt("ndb-base-port", nd.NdbBasePort, minPortValue, maxPortValue) &&
		checkInt("ndb-cluster-port", nd.NdbClusterPort, minPortValue, maxPortValue) &&
		checkInt("group-port-delta", nd.GroupPortDelta, 101, 299) &&
		checkInt("mysqlx-port-delta", nd.MysqlXPortDelta, 2000, 15000) &&
		checkInt("admin-port-delta", nd.AdminPortDelta, 2000, 15000)
	if !allIntegers {
		return false
	}
	noConflicts := nd.MultipleBasePort != nd.GroupReplicationSpBasePort &&
		nd.MultipleBasePort != nd.GroupReplicationBasePort &&
		nd.MultipleBasePort != nd.MasterSlaveBasePort &&
		nd.MultipleBasePort != nd.FanInReplicationBasePort &&
		nd.MultipleBasePort != nd.AllMastersReplicationBasePort &&
		nd.MultipleBasePort != nd.NdbBasePort &&
		nd.MultipleBasePort != nd.NdbClusterPort &&
		nd.MultipleBasePort != nd.PxcBasePort &&
		nd.MultiplePrefix != nd.GroupSpPrefix &&
		nd.MultiplePrefix != nd.GroupPrefix &&
		nd.MultiplePrefix != nd.MasterSlavePrefix &&
		nd.MultiplePrefix != nd.SandboxPrefix &&
		nd.MultiplePrefix != nd.ImportedSandboxPrefix &&
		nd.MultiplePrefix != nd.FanInPrefix &&
		nd.MultiplePrefix != nd.AllMastersPrefix &&
		nd.MasterAbbr != nd.SlaveAbbr &&
		nd.MultiplePrefix != nd.NdbPrefix &&
		nd.MultiplePrefix != nd.PxcPrefix &&
		nd.SandboxHome != nd.SandboxBinary
	if !noConflicts {
		common.CondPrintf("Conflicts found in defaults values:\n")
		ShowDefaults(nd)
		return false
	}
	allStrings := nd.SandboxPrefix != "" &&
		nd.ImportedSandboxPrefix != "" &&
		nd.MasterSlavePrefix != "" &&
		nd.ShellPath != "" &&
		nd.MasterName != "" &&
		nd.MasterAbbr != "" &&
		nd.NodePrefix != "" &&
		nd.SlavePrefix != "" &&
		nd.SlaveAbbr != "" &&
		nd.GroupPrefix != "" &&
		nd.GroupSpPrefix != "" &&
		nd.MultiplePrefix != "" &&
		nd.PxcPrefix != "" &&
		nd.NdbPrefix != "" &&
		nd.DefaultSandboxExecutable != "" &&
		nd.DownloadUrl != "" &&
		nd.DownloadNameLinux != "" &&
		nd.DownloadNameMacOs != "" &&
		nd.SandboxHome != "" &&
		nd.SandboxBinary != "" &&
		nd.RemoteIndexFile != "" &&
		nd.RemoteCompletionUrl != "" &&
		nd.RemoteTarballUrl != "" &&
		nd.RemoteRepository != ""
	if !allStrings {
		common.CondPrintf("One or more empty values found in defaults\n")
		ShowDefaults(nd)
		return false
	}
	compatibleVersionList, err := common.VersionToList(common.CompatibleVersion)
	if err != nil {
		return false
	}
	versionList, err := common.VersionToList(nd.Version)
	if err != nil {
		return false
	}
	compatibleVersion, err := common.GreaterOrEqualVersionList(versionList, compatibleVersionList)
	common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
	if !compatibleVersion {
		common.CondPrintf("Provided defaults are for version %s. Current version is %s\n", nd.Version, common.CompatibleVersion)
		return false
	}
	return true
}

func RemoveDefaultsFile() {
	if common.FileExists(ConfigurationFile) {
		err := os.Remove(ConfigurationFile)
		common.ErrCheckExitf(err, 1, "%s", err)
		common.CondPrintf("#File %s removed\n", ConfigurationFile)
	} else {
		common.Exitf(1, "configuration file %s not found", ConfigurationFile)
	}
}

func strToSlice(label, s string) []int {
	intList, err := common.StringToIntSlice(s)
	if err != nil {
		common.Exitf(1, "bad input for %s: %s (%s) ", label, s, err)
	}
	return intList
}

func UpdateDefaults(label, value string, storeDefaults bool) {
	newDefaults := Defaults()
	switch label {
	case "version":
		newDefaults.Version = value
	case "sandbox-home":
		newDefaults.SandboxHome = value
	case "sandbox-binary":
		newDefaults.SandboxBinary = value
	case "use-sandbox-catalog":
		newDefaults.UseSandboxCatalog = common.TextToBool(value)
	case "log-sb-operations":
		newDefaults.LogSBOperations = common.TextToBool(value)
	case "log-directory":
		newDefaults.LogDirectory = value
	case "cookbook-directory":
		newDefaults.CookbookDirectory = value
	case "shell-path":
		newDefaults.ShellPath = value
	case "master-slave-base-port":
		newDefaults.MasterSlaveBasePort = common.Atoi(value)
	case "group-replication-base-port":
		newDefaults.GroupReplicationBasePort = common.Atoi(value)
	case "group-replication-sp-base-port":
		newDefaults.GroupReplicationSpBasePort = common.Atoi(value)
	case "multiple-base-port":
		newDefaults.MultipleBasePort = common.Atoi(value)
	case "fan-in-base-port":
		newDefaults.FanInReplicationBasePort = common.Atoi(value)
	case "all-masters-base-port":
		newDefaults.AllMastersReplicationBasePort = common.Atoi(value)
	case "ndb-base-port":
		newDefaults.NdbBasePort = common.Atoi(value)
	case "ndb-cluster-port":
		newDefaults.NdbClusterPort = common.Atoi(value)
	case "pxc-base-port":
		newDefaults.PxcBasePort = common.Atoi(value)
	case "group-port-delta":
		newDefaults.GroupPortDelta = common.Atoi(value)
	case "mysqlx-port-delta":
		newDefaults.MysqlXPortDelta = common.Atoi(value)
	case "admin-port-delta":
		newDefaults.AdminPortDelta = common.Atoi(value)
	case "master-name":
		newDefaults.MasterName = value
	case "master-abbr":
		newDefaults.MasterAbbr = value
	case "node-prefix":
		newDefaults.NodePrefix = value
	case "slave-prefix":
		newDefaults.SlavePrefix = value
	case "slave-abbr":
		newDefaults.SlaveAbbr = value
	case "sandbox-prefix":
		newDefaults.SandboxPrefix = value
	case "imported-sandbox-prefix":
		newDefaults.ImportedSandboxPrefix = value
	case "master-slave-prefix":
		newDefaults.MasterSlavePrefix = value
	case "group-prefix":
		newDefaults.GroupPrefix = value
	case "group-sp-prefix":
		newDefaults.GroupSpPrefix = value
	case "multiple-prefix":
		newDefaults.MultiplePrefix = value
	case "fan-in-prefix":
		newDefaults.FanInPrefix = value
	case "all-masters-prefix":
		newDefaults.AllMastersPrefix = value
	case "remote-repository":
		newDefaults.RemoteRepository = value
	case "remote-index-file":
		newDefaults.RemoteIndexFile = value
	case "remote-completion-url":
		newDefaults.RemoteCompletionUrl = value
	case "remote-tarball-url":
		newDefaults.RemoteTarballUrl = value
	case "reserved-ports":
		newDefaults.ReservedPorts = strToSlice("reserved-ports", value)
	case "pxc-prefix":
		newDefaults.PxcPrefix = value
	case "ndb-prefix":
		newDefaults.NdbPrefix = value
	case "default-sandbox-executable":
		newDefaults.DefaultSandboxExecutable = value
	case "download-url":
		newDefaults.DownloadUrl = value
	case "download-name-linux":
		newDefaults.DownloadNameLinux = value
	case "download-name-macos":
		newDefaults.DownloadNameMacOs = value
	default:
		common.Exitf(1, "unrecognized label %s", label)
	}
	if ValidateDefaults(newDefaults) {
		currentDefaults = newDefaults
		if storeDefaults {
			WriteDefaultsFile(ConfigurationFile, Defaults())
			common.CondPrintf("# Updated %s -> \"%s\"\n", label, value)
		}
	} else {
		common.Exitf(1, "invalid defaults data %s : %s", label, value)
	}
}

func LoadConfiguration() {
	if !common.FileExists(ConfigurationFile) {
		// WriteDefaultsFile(ConfigurationFile, Defaults())
		return
	}
	newDefaults := ReadDefaultsFile(ConfigurationFile)
	if ValidateDefaults(newDefaults) {
		currentDefaults = newDefaults
	} else {
		common.CondPrintln(globals.StarLine)
		common.CondPrintf("Defaults file %s not validated.\n", ConfigurationFile)
		common.CondPrintln("Loading internal defaults")
		common.CondPrintln(globals.StarLine)
		common.CondPrintln("")
		time.Sleep(1000 * time.Millisecond)
	}
}

// Converts the defaults to a string map,
// useful to access single values from other operations
func DefaultsToMap() common.StringMap {
	currentDefaults = Defaults()
	if currentDefaults.ShellPath == "" {
		tempPath, err := common.GetBashPath("")
		if err == nil {
			currentDefaults.ShellPath = tempPath
		} else {
			currentDefaults.ShellPath = globals.ShellPathValue
		}
	}
	return common.StringMap{
		"Version":                           currentDefaults.Version,
		"version":                           currentDefaults.Version,
		"SandboxHome":                       currentDefaults.SandboxHome,
		"sandbox-home":                      currentDefaults.SandboxHome,
		"SandboxBinary":                     currentDefaults.SandboxBinary,
		"sandbox-binary":                    currentDefaults.SandboxBinary,
		"UseSandboxCatalog":                 currentDefaults.UseSandboxCatalog,
		"use-sandbox-catalog":               currentDefaults.UseSandboxCatalog,
		"LogSBOperations":                   currentDefaults.LogSBOperations,
		"log-sb-operations":                 currentDefaults.LogSBOperations,
		"LogDirectory":                      currentDefaults.LogDirectory,
		"log-directory":                     currentDefaults.LogDirectory,
		"ShellPath":                         currentDefaults.ShellPath,
		"shell-path":                        currentDefaults.ShellPath,
		"CookbookDirectory":                 currentDefaults.CookbookDirectory,
		"cookbook-directory":                currentDefaults.CookbookDirectory,
		"MasterSlaveBasePort":               currentDefaults.MasterSlaveBasePort,
		"master-slave-base-port":            currentDefaults.MasterSlaveBasePort,
		"GroupReplicationBasePort":          currentDefaults.GroupReplicationBasePort,
		"group-replication-base-port":       currentDefaults.GroupReplicationBasePort,
		"GroupReplicationSpBasePort":        currentDefaults.GroupReplicationSpBasePort,
		"group-replication-sp-base-port":    currentDefaults.GroupReplicationSpBasePort,
		"FanInReplicationBasePort":          currentDefaults.FanInReplicationBasePort,
		"fan-in-replication-base-port":      currentDefaults.FanInReplicationBasePort,
		"AllMastersReplicationBasePort":     currentDefaults.AllMastersReplicationBasePort,
		"all-masters-replication-base-port": currentDefaults.AllMastersReplicationBasePort,
		"MultipleBasePort":                  currentDefaults.MultipleBasePort,
		"multiple-base-port":                currentDefaults.MultipleBasePort,
		"PxcBasePort":                       currentDefaults.PxcBasePort,
		"pxc-base-port":                     currentDefaults.PxcBasePort,
		"NdbBasePort":                       currentDefaults.NdbBasePort,
		"ndb-base-port":                     currentDefaults.NdbBasePort,
		"NdbClusterPort":                    currentDefaults.NdbClusterPort,
		"ndb-cluster-port":                  currentDefaults.NdbClusterPort,
		"GroupPortDelta":                    currentDefaults.GroupPortDelta,
		"group-port-delta":                  currentDefaults.GroupPortDelta,
		"MysqlXPortDelta":                   currentDefaults.MysqlXPortDelta,
		"mysqlx-port-delta":                 currentDefaults.MysqlXPortDelta,
		"AdminPortDelta":                    currentDefaults.AdminPortDelta,
		"admin-port-delta":                  currentDefaults.AdminPortDelta,
		"MasterName":                        currentDefaults.MasterName,
		"master-name":                       currentDefaults.MasterName,
		"MasterAbbr":                        currentDefaults.MasterAbbr,
		"master-abbr":                       currentDefaults.MasterAbbr,
		"NodePrefix":                        currentDefaults.NodePrefix,
		"node-prefix":                       currentDefaults.NodePrefix,
		"SlavePrefix":                       currentDefaults.SlavePrefix,
		"slave-prefix":                      currentDefaults.SlavePrefix,
		"SlaveAbbr":                         currentDefaults.SlaveAbbr,
		"slave-abbr":                        currentDefaults.SlaveAbbr,
		"SandboxPrefix":                     currentDefaults.SandboxPrefix,
		"sandbox-prefix":                    currentDefaults.SandboxPrefix,
		"ImportedSandboxPrefix":             currentDefaults.ImportedSandboxPrefix,
		"imported-sandbox-prefix":           currentDefaults.ImportedSandboxPrefix,
		"MasterSlavePrefix":                 currentDefaults.MasterSlavePrefix,
		"master-slave-prefix":               currentDefaults.MasterSlavePrefix,
		"GroupPrefix":                       currentDefaults.GroupPrefix,
		"group-prefix":                      currentDefaults.GroupPrefix,
		"GroupSpPrefix":                     currentDefaults.GroupSpPrefix,
		"group-sp-prefix":                   currentDefaults.GroupSpPrefix,
		"MultiplePrefix":                    currentDefaults.MultiplePrefix,
		"multiple-prefix":                   currentDefaults.MultiplePrefix,
		"FanInPrefix":                       currentDefaults.FanInPrefix,
		"fan-in-prefix":                     currentDefaults.FanInPrefix,
		"AllMastersPrefix":                  currentDefaults.AllMastersPrefix,
		"all-masters-prefix":                currentDefaults.AllMastersPrefix,
		"ReservedPorts":                     currentDefaults.ReservedPorts,
		"reserved-ports":                    currentDefaults.ReservedPorts,
		"RemoteRepository":                  currentDefaults.RemoteRepository,
		"remote-repository":                 currentDefaults.RemoteRepository,
		"RemoteIndexFile":                   currentDefaults.RemoteIndexFile,
		"remote-index-file":                 currentDefaults.RemoteIndexFile,
		"RemoteCompletionUrl":               currentDefaults.RemoteCompletionUrl,
		"remote-completion-url":             currentDefaults.RemoteCompletionUrl,
		"RemoteTarballUrl":                  currentDefaults.RemoteTarballUrl,
		"remote-tarball-url":                currentDefaults.RemoteTarballUrl,
		"remote-tarballs":                   currentDefaults.RemoteTarballUrl,
		"remote-github":                     currentDefaults.RemoteTarballUrl,
		"PxcPrefix":                         currentDefaults.PxcPrefix,
		"pxc-prefix":                        currentDefaults.PxcPrefix,
		"NdbPrefix":                         currentDefaults.NdbPrefix,
		"ndb-prefix":                        currentDefaults.NdbPrefix,
		"DefaultSandboxExecutable":          currentDefaults.DefaultSandboxExecutable,
		"default-sandbox-executable":        currentDefaults.DefaultSandboxExecutable,
		"download-url":                      currentDefaults.DownloadUrl,
		"DownloadUrl":                       currentDefaults.DownloadUrl,
		"download-name-macos":               currentDefaults.DownloadNameMacOs,
		"DownloadNameMacOs":                 currentDefaults.DownloadNameMacOs,
		"download-name-linux":               currentDefaults.DownloadNameLinux,
		"DownloadNameLinux":                 currentDefaults.DownloadNameLinux,
		"Timestamp":                         currentDefaults.Timestamp,
		"timestamp":                         currentDefaults.Timestamp,
	}
}

func ResetDefaults() {
	homeDir = os.Getenv("HOME")
	ConfigurationDir = path.Join(homeDir, ConfigurationDirName)
	ConfigurationFile = path.Join(ConfigurationDir, ConfigurationFileName)
	CustomConfigurationFile = ""
	SandboxRegistry = path.Join(ConfigurationDir, SandboxRegistryName)
	SandboxRegistryLock = path.Join(common.GlobalTempDir(), SandboxRegistryLockName)
	LogSBOperations = common.IsEnvSet("DBDEPLOYER_LOGGING")
	factoryDefaults.SandboxBinary = path.Join(homeDir, "opt", "mysql")
	factoryDefaults.SandboxHome = path.Join(homeDir, "sandboxes")
	factoryDefaults.LogDirectory = path.Join(homeDir, "sandboxes", "logs")
	currentDefaults = DbdeployerDefaults{}
}
