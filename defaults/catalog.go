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
	"fmt"
	"os"
	//"strings"
	"encoding/json"
	"github.com/datacharmer/dbdeployer/common"
	"time"
)

type SandboxItem struct {
	Origin            string   `json:"origin"`
	SBType            string   `json:"type"` // single multi master-slave group all-masters fan-in
	Version           string   `json:"version"`
	Port              []int    `json:"port"`
	Nodes             []string `json:"nodes"`
	Destination       string   `json:"destination"`
	DbDeployerVersion string   `json:"dbdeployer-version"`
	Timestamp         string   `json:"timestamp"`
	CommandLine       string   `json:"command-line"`
}

type SandboxCatalog map[string]SandboxItem

const (
	timeout = 5
)

var enable_catalog_management bool = true

func setLock(label string) bool {
	if !enable_catalog_management {
		return true
	}
	lock_file := SandboxRegistryLock
	if !common.DirExists(ConfigurationDir) {
		common.Mkdir(ConfigurationDir)
	}
	elapsed := 0
	for common.FileExists(lock_file) {
		elapsed += 1
		time.Sleep(1000 * time.Millisecond)
		if elapsed > timeout {
			return false
		}
	}
	common.WriteString(label, lock_file)
	return true
}

func releaseLock() {
	if !enable_catalog_management {
		return
	}
	lock_file := SandboxRegistryLock
	if common.FileExists(lock_file) {
		os.Remove(lock_file)
	}
}

func WriteCatalog(sc SandboxCatalog) {
	if !enable_catalog_management {
		return
	}
	b, err := json.MarshalIndent(sc, " ", "\t")
	if err != nil {
		common.Exitf(1, "error encoding sandbox catalog: %s", err)
	}
	json_string := fmt.Sprintf("%s", b)
	filename := SandboxRegistry
	common.WriteString(json_string, filename)
}

func ReadCatalog() (sc SandboxCatalog) {
	if !enable_catalog_management {
		return
	}
	filename := SandboxRegistry
	if !common.FileExists(filename) {
		return
	}
	sc_blob := common.SlurpAsBytes(filename)

	err := json.Unmarshal(sc_blob, &sc)
	if err != nil {
		common.Exitf(1, "error decoding sandbox catalog: %s", err)
	}
	return
}

func UpdateCatalog(sb_name string, details SandboxItem) {
	details.DbDeployerVersion = common.VersionDef
	details.Timestamp = time.Now().Format(time.UnixDate)
	details.CommandLine = common.CommandLineArgs
	if !enable_catalog_management {
		return
	}
	// fmt.Printf("+%s\n",sb_name)
	if setLock(sb_name) {
		// fmt.Printf("+locked\n")
		current := ReadCatalog()
		if current == nil {
			current = make(SandboxCatalog)
		}
		current[sb_name] = details
		WriteCatalog(current)
		releaseLock()
		// fmt.Printf("+unlocked\n")
	} else {
		fmt.Printf("%s\n", HashLine)
		fmt.Printf("# Could not get lock on %s\n", SandboxRegistryLock)
		fmt.Printf("%s\n", HashLine)
	}
}

func DeleteFromCatalog(sb_name string) {
	if !enable_catalog_management {
		return
	}
	if setLock(sb_name) {
		current := ReadCatalog()
		defer releaseLock()
		if current == nil {
			return
		}
		//for name, _ := range current {
		//	if strings.HasPrefix(name, sb_name) {
		//		delete(current, name)
		//	}
		//}
		delete(current, sb_name)
		WriteCatalog(current)
		releaseLock()
	} else {
		fmt.Printf("%s\n", HashLine)
		fmt.Printf("# Could not get lock on %s\n", SandboxRegistryLock)
		fmt.Printf("%s\n", HashLine)
	}
}

func init() {
	if os.Getenv("SKIP_DBDEPLOYER_CATALOG") != "" {
		enable_catalog_management = false
	}
}
