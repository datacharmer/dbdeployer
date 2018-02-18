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
	Version                    string `json:"version"`
	SandboxHome                string `json:"sandbox-home"`
	SandboxBinary              string `json:"sandbox-binary"`
	MasterSlaveBasePort        int    `json:"master-slave-base-port"`
	GroupReplicationBasePort   int    `json:"group-replication-base-port"`
	GroupReplicationSpBasePort int    `json:"group-replication-sp-base-port"`
	MultipleBasePort           int    `json:"multiple-base-port"`
	GroupPortDelta             int    `json:"group-port-delta"`
	SandboxPrefix              string `json:"sandbox-prefix"`
	MasterSlavePrefix          string `json:"master-slave-prefix"`
	GroupPrefix                string `json:"group-prefix"`
	GroupSpPrefix              string `json:"group-sp-prefix"`
	MultiplePrefix             string `json:"multiple-prefix"`
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
	StarLine                string = strings.Repeat("*", LineLength)
	DashLine                string = strings.Repeat("-", LineLength)
	HashLine                string = strings.Repeat("#", LineLength)

	factoryDefaults = DbdeployerDefaults{
		Version:                    common.CompatibleVersion,
		SandboxHome:                home_dir + "/sandboxes",
		SandboxBinary:              home_dir + "/opt/mysql",
		MasterSlaveBasePort:        11000,
		GroupReplicationBasePort:   12000,
		GroupReplicationSpBasePort: 13000,
		MultipleBasePort:           16000,
		GroupPortDelta:             125,
		SandboxPrefix:              "msb_",
		MasterSlavePrefix:          "rsandbox_",
		GroupPrefix:                "group_msb_",
		GroupSpPrefix:              "group_sp_msb_",
		MultiplePrefix:             "multi_msb_",
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
	if common.FileExists(ConfigurationFile) {
		fmt.Printf("# Configuration file: %s\n", ConfigurationFile)
	} else {
		fmt.Println("# Internal values:")
	}
	b, err := json.MarshalIndent(defaults, " ", "\t")
	if err != nil {
		fmt.Println("error encoding defaults: ", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", b)
}

func WriteDefaultsFile(filename string, defaults DbdeployerDefaults) {
	defaults_dir := common.DirName(filename)
	if !common.DirExists(defaults_dir) {
		common.Mkdir(defaults_dir)
	}
	b, err := json.MarshalIndent(defaults, " ", "\t")
	if err != nil {
		fmt.Println("error encoding defaults: ", err)
		os.Exit(1)
	}
	json_string := fmt.Sprintf("%s", b)
	common.WriteString(json_string, filename)
}

func ReadDefaultsFile(filename string) (defaults DbdeployerDefaults) {
	defaults_blob := common.SlurpAsBytes(filename)

	err := json.Unmarshal(defaults_blob, &defaults)
	if err != nil {
		fmt.Println("error decoding defaults: ", err)
		os.Exit(1)
	}
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
		check_int("group-port-delta", nd.GroupPortDelta, 101, 299)
	if !all_ints {
		return false
	}
	var no_conflicts bool
	no_conflicts = nd.MultipleBasePort != nd.GroupReplicationSpBasePort &&
		nd.MultipleBasePort != nd.GroupReplicationBasePort &&
		nd.MultipleBasePort != nd.MasterSlaveBasePort &&
		nd.MultiplePrefix != nd.GroupSpPrefix &&
		nd.MultiplePrefix != nd.GroupPrefix &&
		nd.MultiplePrefix != nd.MasterSlavePrefix &&
		nd.MultiplePrefix != nd.SandboxPrefix &&
		nd.SandboxHome != nd.SandboxBinary
	if !no_conflicts {
		fmt.Printf("Conflicts found in defaults values:\n")
		ShowDefaults(nd)
		return false
	}
	all_strings := nd.SandboxPrefix != "" &&
		nd.MasterSlavePrefix != "" &&
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
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		fmt.Printf("#File %s removed\n", ConfigurationFile)
	} else {
		fmt.Printf("Configuration file %s not found\n", ConfigurationFile)
		os.Exit(1)
	}
}

func a_to_i(val string) int {
	numvalue, err := strconv.Atoi(val)
	if err != nil {
		fmt.Printf("Not a valid number: %s\n", val)
		os.Exit(1)
	}
	return numvalue
}

func UpdateDefaults(label, value string) {
	new_defaults := Defaults()
	switch label {
	case "version":
		new_defaults.Version = value
	case "sandbox-home":
		new_defaults.SandboxHome = value
	case "sandbox-binary":
		new_defaults.SandboxBinary = value
	case "master-slave-base-port":
		new_defaults.MasterSlaveBasePort = a_to_i(value)
	case "group-replication-base-port":
		new_defaults.GroupReplicationBasePort = a_to_i(value)
	case "group-replication-sp-base-port":
		new_defaults.GroupReplicationSpBasePort = a_to_i(value)
	case "multiple-base-port":
		new_defaults.MultipleBasePort = a_to_i(value)
	case "group-port-delta":
		new_defaults.GroupPortDelta = a_to_i(value)
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
	default:
		fmt.Printf("Unrecognized label %s\n", label)
		os.Exit(1)
	}
	if ValidateDefaults(new_defaults) {
		currentDefaults = new_defaults
		WriteDefaultsFile(ConfigurationFile, Defaults())
	}
	fmt.Printf("# Updated %s -> \"%s\"\n", label, value)
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
