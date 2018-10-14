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
	"strings"

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
	LogDirectory      string   `json:"log-directory,omitempty"`
	CommandLine       string   `json:"command-line"`
}

type SandboxCatalog map[string]SandboxItem

const (
	timeout = 5
)

var enableCatalogManagement bool = true

func setLock(label string) bool {
	if !enableCatalogManagement {
		return true
	}
	lockFile := SandboxRegistryLock
	if !common.DirExists(ConfigurationDir) {
		common.Mkdir(ConfigurationDir)
	}
	elapsed := 0
	for common.FileExists(lockFile) {
		elapsed += 1
		time.Sleep(1000 * time.Millisecond)
		if elapsed > timeout {
			return false
		}
	}
	common.WriteString(label, lockFile)
	return true
}

func releaseLock() {
	if !enableCatalogManagement {
		return
	}
	lockFile := SandboxRegistryLock
	if common.FileExists(lockFile) {
		os.Remove(lockFile)
	}
}

func WriteCatalog(sc SandboxCatalog) {
	if !enableCatalogManagement {
		return
	}
	b, err := json.MarshalIndent(sc, " ", "\t")
	common.ErrCheckExitf(err, 1, "error encoding sandbox catalog: %s", err)
	jsonString := fmt.Sprintf("%s", b)
	filename := SandboxRegistry
	common.WriteString(jsonString, filename)
}

func ReadCatalog() (sc SandboxCatalog) {
	if !enableCatalogManagement {
		return
	}
	filename := SandboxRegistry
	if !common.FileExists(filename) {
		return
	}
	scBlob := common.SlurpAsBytes(filename)

	err := json.Unmarshal(scBlob, &sc)
	common.ErrCheckExitf(err, 1, "error decoding sandbox catalog: %s", err)
	return
}

func UpdateCatalog(sbName string, details SandboxItem) {
	details.DbDeployerVersion = common.VersionDef
	details.Timestamp = time.Now().Format(time.UnixDate)
	details.CommandLine = strings.Join(common.CommandLineArgs, " ")
	if !enableCatalogManagement {
		return
	}
	// fmt.Printf("+%s\n",sb_name)
	if setLock(sbName) {
		// fmt.Printf("+locked\n")
		current := ReadCatalog()
		if current == nil {
			current = make(SandboxCatalog)
		}
		current[sbName] = details
		WriteCatalog(current)
		releaseLock()
		// fmt.Printf("+unlocked\n")
	} else {
		fmt.Printf("%s\n", common.HashLine)
		fmt.Printf("# Could not get lock on %s\n", SandboxRegistryLock)
		fmt.Printf("%s\n", common.HashLine)
	}
}

func DeleteFromCatalog(sbName string) {
	if !enableCatalogManagement {
		return
	}
	if setLock(sbName) {
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
		delete(current, sbName)
		WriteCatalog(current)
		releaseLock()
	} else {
		fmt.Printf("%s\n", common.HashLine)
		fmt.Printf("# Could not get lock on %s\n", SandboxRegistryLock)
		fmt.Printf("%s\n", common.HashLine)
	}
}

func init() {
	if os.Getenv("SKIP_DBDEPLOYER_CATALOG") != "" {
		enableCatalogManagement = false
	}
}
