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

package defaults

import (
	"encoding/json"
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"os"
	"strconv"
	"strings"
	"time"
)

type DbdeployerDefaults struct {
	Version           string `json:"version"`
	SandboxHome       string `json:"sandbox-home"`
	SandboxBinary     string `json:"sandbox-binary"`
	UseSandboxCatalog bool   `json:"use-sandbox-catalog"`
	//UseConcurrency    			   bool   `json:"use-concurrency"`
	MasterSlaveBasePort           int `json:"master-slave-base-port"`
	GroupReplicationBasePort      int `json:"group-replication-base-port"`
	GroupReplicationSpBasePort    int `json:"group-replication-sp-base-port"`
	FanInReplicationBasePort      int `json:"fan-in-replication-base-port"`
	AllMastersReplicationBasePort int `json:"all-masters-replication-base-port"`
	MultipleBasePort              int `json:"multiple-base-port"`
	// GaleraBasePort                 int    `json:"galera-base-port"`
	// PXCBasePort                    int    `json:"pxc-base-port"`
	// NdbBasePort                    int    `json:"ndb-base-port"`
	GroupPortDelta    int    `json:"group-port-delta"`
	MysqlXPortDelta   int    `json:"mysqlx-port-delta"`
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
	// GaleraPrefix                   string `json:"galera-prefix"`
	// PxcPrefix                      string `json:"pxc-prefix"`
	// NdbPrefix                      string `json:"ndb-prefix"`
	Timestamp string `json:"timestamp"`
}

const (
	min_port_value int = 11000
	max_port_value int = 30000
	LineLength     int = 80
)

var (
	home_dir                string = os.Getenv("HOME")
	ConfigurationDir        string = home_dir + "/.dbdeployer"
	ConfigurationFile       string = ConfigurationDir + "/config.json"
	CustomConfigurationFile string = ""
	SandboxRegistry         string = ConfigurationDir + "/sandboxes.json"
	SandboxRegistryLock     string = ConfigurationDir + "/sandboxes.lock"
	StarLine                string = strings.Repeat("*", LineLength)
	DashLine                string = strings.Repeat("-", LineLength)
	HashLine                string = strings.Repeat("#", LineLength)

	// This variable is changed to true when the "cmd" package is activated,
	// meaning that we're using the command line interface of dbdeployer.
	// It is used to make decisions whether to write messages to the screen
	// when calling sandbox creation functions from other apps.
	UsingDbDeployer bool = false

	factoryDefaults = DbdeployerDefaults{
		Version:       common.CompatibleVersion,
		SandboxHome:   home_dir + "/sandboxes",
		SandboxBinary: home_dir + "/opt/mysql",

		UseSandboxCatalog: true,
		//UseConcurrency :			   true,
		MasterSlaveBasePort:           11000,
		GroupReplicationBasePort:      12000,
		GroupReplicationSpBasePort:    13000,
		FanInReplicationBasePort:      14000,
		AllMastersReplicationBasePort: 15000,
		MultipleBasePort:              16000,
		// GaleraBasePort:                17000,
		// PxcBasePort:                   18000,
		// NdbBasePort:                   19000,
		GroupPortDelta:    125,
		MysqlXPortDelta:   10000,
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
		ReservedPorts:     []int{1186, 3306, 33060},
		// GaleraPrefix:                  "galera_msb_",
		// NdbPrefix:                     "ndb_msb_",
		// PxcPrefix:                     "pxc_msb_",
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
	return currentDefaults
}

func ShowDefaults(defaults DbdeployerDefaults) {
	defaults = replace_literal_env_values(defaults)
	if common.FileExists(ConfigurationFile) {
		fmt.Printf("# Configuration file: %s\n", ConfigurationFile)
	} else {
		fmt.Println("# Internal values:")
	}
	b, err := json.MarshalIndent(defaults, " ", "\t")
	if err != nil {
		common.Exitf(1, "error encoding defaults: %s", err)
	}
	fmt.Printf("%s\n", b)
}

func WriteDefaultsFile(filename string, defaults DbdeployerDefaults) {
	defaults = replace_literal_env_values(defaults)
	defaults_dir := common.DirName(filename)
	if !common.DirExists(defaults_dir) {
		common.Mkdir(defaults_dir)
	}
	b, err := json.MarshalIndent(defaults, " ", "\t")
	if err != nil {
		common.Exitf(1, "error encoding defaults: %s", err)
	}
	json_string := fmt.Sprintf("%s", b)
	common.WriteString(json_string, filename)
}

func expand_environment_variables(defaults DbdeployerDefaults) DbdeployerDefaults {
	defaults.SandboxHome = common.ReplaceEnvVar(defaults.SandboxHome, "HOME")
	defaults.SandboxHome = common.ReplaceEnvVar(defaults.SandboxHome, "PWD")
	defaults.SandboxBinary = common.ReplaceEnvVar(defaults.SandboxBinary, "HOME")
	defaults.SandboxBinary = common.ReplaceEnvVar(defaults.SandboxBinary, "PWD")
	return defaults
}

func replace_literal_env_values(defaults DbdeployerDefaults) DbdeployerDefaults {
	defaults.SandboxHome = common.ReplaceLiteralEnvVar(defaults.SandboxHome, "HOME")
	defaults.SandboxHome = common.ReplaceLiteralEnvVar(defaults.SandboxHome, "PWD")
	defaults.SandboxBinary = common.ReplaceLiteralEnvVar(defaults.SandboxBinary, "HOME")
	defaults.SandboxBinary = common.ReplaceLiteralEnvVar(defaults.SandboxBinary, "PWD")
	return defaults
}

func ReadDefaultsFile(filename string) (defaults DbdeployerDefaults) {
	defaults_blob := common.SlurpAsBytes(filename)

	err := json.Unmarshal(defaults_blob, &defaults)
	if err != nil {
		common.Exitf(1, "error decoding defaults: %s", err)
	}
	defaults = expand_environment_variables(defaults)
	return
}

func check_int(name string, val, min, max int) bool {
	if val >= min && val <= max {
		return true
	}
	fmt.Printf("Value %s (%d) must be between %d and %d\n", name, val, min, max)
	return false
}

func ValidateDefaults(nd DbdeployerDefaults) bool {
	var all_ints bool
	all_ints = check_int("master-slave-base-port", nd.MasterSlaveBasePort, min_port_value, max_port_value) &&
		check_int("group-replication-base-port", nd.GroupReplicationBasePort, min_port_value, max_port_value) &&
		check_int("group-replication-sp-base-port", nd.GroupReplicationSpBasePort, min_port_value, max_port_value) &&
		check_int("multiple-base-port", nd.MultipleBasePort, min_port_value, max_port_value) &&
		check_int("fan-in-base-port", nd.FanInReplicationBasePort, min_port_value, max_port_value) &&
		check_int("all-masters-base-port", nd.AllMastersReplicationBasePort, min_port_value, max_port_value) &&
		// check_int("galera-base-port", nd.GaleraBasePort, min_port_value, max_port_value) &&
		// check_int("pxc-base-port", nd.PxcBasePort, min_port_value, max_port_value) &&
		// check_int("ndb-base-port", nd.NdbBasePort, min_port_value, max_port_value) &&
		check_int("group-port-delta", nd.GroupPortDelta, 101, 299)
	check_int("mysqlx-port-delta", nd.MysqlXPortDelta, 2000, 15000)
	if !all_ints {
		return false
	}
	var no_conflicts bool
	no_conflicts = nd.MultipleBasePort != nd.GroupReplicationSpBasePort &&
		nd.MultipleBasePort != nd.GroupReplicationBasePort &&
		nd.MultipleBasePort != nd.MasterSlaveBasePort &&
		nd.MultipleBasePort != nd.FanInReplicationBasePort &&
		nd.MultipleBasePort != nd.AllMastersReplicationBasePort &&
		// nd.MultipleBasePort != nd.NdbBasePort &&
		// nd.MultipleBasePort != nd.GaleraBasePort &&
		// nd.MultipleBasePort != nd.PxcBasePort &&
		nd.MultiplePrefix != nd.GroupSpPrefix &&
		nd.MultiplePrefix != nd.GroupPrefix &&
		nd.MultiplePrefix != nd.MasterSlavePrefix &&
		nd.MultiplePrefix != nd.SandboxPrefix &&
		nd.MultiplePrefix != nd.FanInPrefix &&
		nd.MultiplePrefix != nd.AllMastersPrefix &&
		nd.MasterAbbr != nd.SlaveAbbr &&
		// nd.MultiplePrefix != nd.NdbPrefix &&
		// nd.MultiplePrefix != nd.GaleraPrefix &&
		// nd.MultiplePrefix != nd.PxcPrefix &&
		nd.SandboxHome != nd.SandboxBinary
	if !no_conflicts {
		fmt.Printf("Conflicts found in defaults values:\n")
		ShowDefaults(nd)
		return false
	}
	all_strings := nd.SandboxPrefix != "" &&
		nd.MasterSlavePrefix != "" &&
		nd.MasterName != "" &&
		nd.MasterAbbr != "" &&
		nd.NodePrefix != "" &&
		nd.SlavePrefix != "" &&
		nd.SlaveAbbr != "" &&
		nd.GroupPrefix != "" &&
		nd.GroupSpPrefix != "" &&
		nd.MultiplePrefix != "" &&
		nd.SandboxHome != "" &&
		nd.SandboxBinary != ""
	if !all_strings {
		fmt.Printf("One or more empty values found in defaults\n")
		ShowDefaults(nd)
		return false
	}
	versionList := common.VersionToList(common.CompatibleVersion)
	if !common.GreaterOrEqualVersion(nd.Version, versionList) {
		fmt.Printf("Provided defaults are for version %s. Current version is %s\n", nd.Version, common.CompatibleVersion)
		return false
	}
	return true
}

func RemoveDefaultsFile() {
	if common.FileExists(ConfigurationFile) {
		err := os.Remove(ConfigurationFile)
		if err != nil {
			common.Exitf(1, "%s", err)
		}
		fmt.Printf("#File %s removed\n", ConfigurationFile)
	} else {
		common.Exitf(1, "Configuration file %s not found", ConfigurationFile)
	}
}

func a_to_i(val string) int {
	numvalue, err := strconv.Atoi(val)
	if err != nil {
		common.Exitf(1, "Not a valid number: %s", val)
	}
	return numvalue
}

func UpdateDefaults(label, value string, store_defaults bool) {
	new_defaults := Defaults()
	switch label {
	case "version":
		new_defaults.Version = value
	case "sandbox-home":
		new_defaults.SandboxHome = value
	case "sandbox-binary":
		new_defaults.SandboxBinary = value
	case "use-sandbox-catalog":
		new_defaults.UseSandboxCatalog = common.TextToBool(value)
	//case "use-concurrency":
	//	new_defaults.UseConcurrency = common.TextToBool(value)
	case "master-slave-base-port":
		new_defaults.MasterSlaveBasePort = a_to_i(value)
	case "group-replication-base-port":
		new_defaults.GroupReplicationBasePort = a_to_i(value)
	case "group-replication-sp-base-port":
		new_defaults.GroupReplicationSpBasePort = a_to_i(value)
	case "multiple-base-port":
		new_defaults.MultipleBasePort = a_to_i(value)
	case "fan-in-base-port":
		new_defaults.FanInReplicationBasePort = a_to_i(value)
	case "all-masters-base-port":
		new_defaults.AllMastersReplicationBasePort = a_to_i(value)
	// case "ndb-base-port":
	//	 new_defaults.NdbBasePort = a_to_i(value)
	// case "galera-base-port":
	//	 new_defaults.GaleraBasePort = a_to_i(value)
	// case "pxc-base-port":
	//	 new_defaults.PxcBasePort = a_to_i(value)
	case "group-port-delta":
		new_defaults.GroupPortDelta = a_to_i(value)
	case "mysqlx-port-delta":
		new_defaults.MysqlXPortDelta = a_to_i(value)
	case "master-name":
		new_defaults.MasterName = value
	case "master-abbr":
		new_defaults.MasterAbbr = value
	case "node-prefix":
		new_defaults.NodePrefix = value
	case "slave-prefix":
		new_defaults.SlavePrefix = value
	case "slave-abbr":
		new_defaults.SlaveAbbr = value
	case "sandbox-prefix":
		new_defaults.SandboxPrefix = value
	case "master-slave-prefix":
		new_defaults.MasterSlavePrefix = value
	case "group-prefix":
		new_defaults.GroupPrefix = value
	case "group-sp-prefix":
		new_defaults.GroupSpPrefix = value
	case "multiple-prefix":
		new_defaults.MultiplePrefix = value
	case "fan-in-prefix":
		new_defaults.FanInPrefix = value
	case "all-masters-prefix":
		new_defaults.AllMastersPrefix = value
	case "reserved-ports":
		new_defaults.ReservedPorts = common.StringToIntSlice(value)
	// case "galera-prefix":
	// 	new_defaults.GaleraPrefix = value
	// case "pxc-prefix":
	// 	new_defaults.PxcPrefix = value
	// case "ndb-prefix":
	// 	new_defaults.NdbPrefix = value
	default:
		common.Exitf(1, "Unrecognized label %s", label)
	}
	if ValidateDefaults(new_defaults) {
		currentDefaults = new_defaults
		if store_defaults {
			WriteDefaultsFile(ConfigurationFile, Defaults())
			fmt.Printf("# Updated %s -> \"%s\"\n", label, value)
		}
	} else {
		common.Exitf(1, "Invalid defaults data %s : %s", label, value)
	}
}

func LoadConfiguration() {
	if !common.FileExists(ConfigurationFile) {
		// WriteDefaultsFile(ConfigurationFile, Defaults())
		return
	}
	new_defaults := ReadDefaultsFile(ConfigurationFile)
	if ValidateDefaults(new_defaults) {
		currentDefaults = new_defaults
	} else {
		fmt.Println(StarLine)
		fmt.Printf("Defaults file %s not validated.\n", ConfigurationFile)
		fmt.Println("Loading internal defaults")
		fmt.Println(StarLine)
		fmt.Println("")
		time.Sleep(1000 * time.Millisecond)
	}
}
