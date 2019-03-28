// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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
	"fmt"
	"os"
	"path"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
)

type DbdeployerDefaults struct {
	Version           string `json:"version"`
	SandboxHome       string `json:"sandbox-home"`
	SandboxBinary     string `json:"sandbox-binary"`
	UseSandboxCatalog bool   `json:"use-sandbox-catalog"`
	LogSBOperations   bool   `json:"log-sb-operations"`
	LogDirectory      string `json:"log-directory"`
	CookbookDirectory string `json:"cookbook-directory"`

	//UseConcurrency    			   bool   `json:"use-concurrency"`
	MasterSlaveBasePort           int `json:"master-slave-base-port"`
	GroupReplicationBasePort      int `json:"group-replication-base-port"`
	GroupReplicationSpBasePort    int `json:"group-replication-sp-base-port"`
	FanInReplicationBasePort      int `json:"fan-in-replication-base-port"`
	AllMastersReplicationBasePort int `json:"all-masters-replication-base-port"`
	MultipleBasePort              int `json:"multiple-base-port"`
	// GaleraBasePort                 int    `json:"galera-base-port"`
	PxcBasePort       int    `json:"pxc-base-port"`
	NdbBasePort       int    `json:"ndb-base-port"`
	NdbClusterPort    int    `json:"ndb-cluster-port"`
	GroupPortDelta    int    `json:"group-port-delta"`
	MysqlXPortDelta   int    `json:"mysqlx-port-delta"`
	AdminPortDelta    int    `json:"admin-port-delta"`
	MasterName        string `json:"master-name"`
	MasterAbbr        string `json:"master-abbr"`
	NodePrefix        string `json:"node-prefix"`
	SlavePrefix       string `json:"slave-prefix"`
	SlaveAbbr         string `json:"slave-abbr"`
	SandboxPrefix     string `json:"sandbox-prefix"`
	MasterSlavePrefix string `json:"master-slave-prefix"`
	GroupPrefix       string `json:"group-prefix"`
	GroupSpPrefix     string `json:"group-sp-prefix"`
	MultiplePrefix    string `json:"multiple-prefix"`
	FanInPrefix       string `json:"fan-in-prefix"`
	AllMastersPrefix  string `json:"all-masters-prefix"`
	ReservedPorts     []int  `json:"reserved-ports"`
	RemoteRepository  string `json:"remote-repository"`
	RemoteIndexFile   string `json:"remote-index-file"`
	// GaleraPrefix                   string `json:"galera-prefix"`
	PxcPrefix string `json:"pxc-prefix"`
	NdbPrefix string `json:"ndb-prefix"`
	Timestamp string `json:"timestamp"`
}

const (
	minPortValue            int    = 11000
	maxPortValue            int    = 30000
	ConfigurationDirName    string = ".dbdeployer"
	ConfigurationFileName   string = "config.json"
	SandboxRegistryName     string = "sandboxes.json"
	SandboxRegistryLockName string = "sandboxes.lock"
)

var (
	homeDir                 string = os.Getenv("HOME")
	ConfigurationDir        string = path.Join(homeDir, ConfigurationDirName)
	ConfigurationFile       string = path.Join(ConfigurationDir, ConfigurationFileName)
	CustomConfigurationFile string = ""
	SandboxRegistry         string = path.Join(ConfigurationDir, SandboxRegistryName)
	SandboxRegistryLock     string = path.Join(ConfigurationDir, SandboxRegistryLockName)
	LogSBOperations         bool   = common.IsEnvSet("DBDEPLOYER_LOGGING")

	factoryDefaults = DbdeployerDefaults{
		Version:       common.CompatibleVersion,
		SandboxHome:   path.Join(homeDir, "sandboxes"),
		SandboxBinary: path.Join(homeDir, "opt", "mysql"),

		UseSandboxCatalog: true,
		LogSBOperations:   false,
		LogDirectory:      path.Join(homeDir, "sandboxes", "logs"),
		CookbookDirectory: "recipes",
		//UseConcurrency :			   true,
		MasterSlaveBasePort:           11000,
		GroupReplicationBasePort:      12000,
		GroupReplicationSpBasePort:    13000,
		FanInReplicationBasePort:      14000,
		AllMastersReplicationBasePort: 15000,
		MultipleBasePort:              16000,
		PxcBasePort:                   18000,
		// GaleraBasePort:                17000,
		NdbBasePort:       19000,
		NdbClusterPort:    20000,
		GroupPortDelta:    125,
		MysqlXPortDelta:   10000,
		AdminPortDelta:    11000,
		MasterName:        "master",
		MasterAbbr:        "m",
		NodePrefix:        "node",
		SlavePrefix:       "slave",
		SlaveAbbr:         "s",
		SandboxPrefix:     "msb_",
		MasterSlavePrefix: "rsandbox_",
		GroupPrefix:       "group_msb_",
		GroupSpPrefix:     "group_sp_msb_",
		MultiplePrefix:    "multi_msb_",
		FanInPrefix:       "fan_in_msb_",
		AllMastersPrefix:  "all_masters_msb_",
		ReservedPorts: []int{
			1186,  // MySQL Cluster
			3306,  // MySQL Server regular port
			33060, // MySQLX
			33062, // MySQL Server admin port
		},
		RemoteRepository: "https://raw.githubusercontent.com/datacharmer/mysql-docker-minimal/master/dbdata",
		RemoteIndexFile:  "available.json",
		// GaleraPrefix:                  "galera_msb_",
		NdbPrefix: "ndb_msb_",
		PxcPrefix: "pxc_msb_",
		Timestamp: time.Now().Format(time.UnixDate),
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
	jsonString := fmt.Sprintf("%s", b)
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
	var allInts bool
	allInts = checkInt("master-slave-base-port", nd.MasterSlaveBasePort, minPortValue, maxPortValue) &&
		checkInt("group-replication-base-port", nd.GroupReplicationBasePort, minPortValue, maxPortValue) &&
		checkInt("group-replication-sp-base-port", nd.GroupReplicationSpBasePort, minPortValue, maxPortValue) &&
		checkInt("multiple-base-port", nd.MultipleBasePort, minPortValue, maxPortValue) &&
		checkInt("fan-in-base-port", nd.FanInReplicationBasePort, minPortValue, maxPortValue) &&
		checkInt("all-masters-base-port", nd.AllMastersReplicationBasePort, minPortValue, maxPortValue) &&
		// checkInt("galera-base-port", nd.GaleraBasePort, minPortValue, maxPortValue) &&
		checkInt("pxc-base-port", nd.PxcBasePort, minPortValue, maxPortValue) &&
		checkInt("ndb-base-port", nd.NdbBasePort, minPortValue, maxPortValue) &&
		checkInt("ndb-cluster-port", nd.NdbClusterPort, minPortValue, maxPortValue) &&
		checkInt("group-port-delta", nd.GroupPortDelta, 101, 299) &&
		checkInt("mysqlx-port-delta", nd.MysqlXPortDelta, 2000, 15000) &&
		checkInt("admin-port-delta", nd.AdminPortDelta, 2000, 15000)
	if !allInts {
		return false
	}
	var noConflicts bool
	noConflicts = nd.MultipleBasePort != nd.GroupReplicationSpBasePort &&
		nd.MultipleBasePort != nd.GroupReplicationBasePort &&
		nd.MultipleBasePort != nd.MasterSlaveBasePort &&
		nd.MultipleBasePort != nd.FanInReplicationBasePort &&
		nd.MultipleBasePort != nd.AllMastersReplicationBasePort &&
		nd.MultipleBasePort != nd.NdbBasePort &&
		nd.MultipleBasePort != nd.NdbClusterPort &&
		// nd.MultipleBasePort != nd.GaleraBasePort &&
		nd.MultipleBasePort != nd.PxcBasePort &&
		nd.MultiplePrefix != nd.GroupSpPrefix &&
		nd.MultiplePrefix != nd.GroupPrefix &&
		nd.MultiplePrefix != nd.MasterSlavePrefix &&
		nd.MultiplePrefix != nd.SandboxPrefix &&
		nd.MultiplePrefix != nd.FanInPrefix &&
		nd.MultiplePrefix != nd.AllMastersPrefix &&
		nd.MasterAbbr != nd.SlaveAbbr &&
		nd.MultiplePrefix != nd.NdbPrefix &&
		// nd.MultiplePrefix != nd.GaleraPrefix &&
		nd.MultiplePrefix != nd.PxcPrefix &&
		nd.SandboxHome != nd.SandboxBinary
	if !noConflicts {
		common.CondPrintf("Conflicts found in defaults values:\n")
		ShowDefaults(nd)
		return false
	}
	allStrings := nd.SandboxPrefix != "" &&
		nd.MasterSlavePrefix != "" &&
		nd.MasterName != "" &&
		nd.MasterAbbr != "" &&
		nd.NodePrefix != "" &&
		nd.SlavePrefix != "" &&
		nd.SlaveAbbr != "" &&
		nd.GroupPrefix != "" &&
		nd.GroupSpPrefix != "" &&
		nd.MultiplePrefix != "" &&
		nd.PxcPrefix != "" &&
		// nd.GaleraPrefix != "" &&
		nd.NdbPrefix != "" &&
		nd.SandboxHome != "" &&
		nd.SandboxBinary != "" &&
		nd.RemoteIndexFile != "" &&
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
	//case "use-concurrency":
	//	new_defaults.UseConcurrency = common.TextToBool(value)
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
	// case "galera-base-port":
	//	 new_defaults.GaleraBasePort = common.Atoi(value)
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
	case "reserved-ports":
		newDefaults.ReservedPorts = strToSlice("reserved-ports", value)
	// case "galera-prefix":
	// 	new_defaults.GaleraPrefix = value
	case "pxc-prefix":
		newDefaults.PxcPrefix = value
	case "ndb-prefix":
		newDefaults.NdbPrefix = value
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
