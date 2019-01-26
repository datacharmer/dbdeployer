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
	"github.com/datacharmer/dbdeployer/globals"
	"os"
	"strings"
	"sync"

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
var catalogMutex sync.Mutex

func isLocked() bool {
	return common.FileExists(SandboxRegistryLock)
}

func setLock(label string) error {
	if !enableCatalogManagement {
		return nil
	}
	if !common.DirExists(ConfigurationDir) {
		err := os.Mkdir(ConfigurationDir, globals.PublicDirectoryAttr)
		if err != nil {
			return fmt.Errorf("error making lock directory")
		}
	}
	if !common.FileExists(SandboxRegistry) {
		err := common.WriteString("{}", SandboxRegistry)
		if err != nil {
			return err
		}
	}
	elapsed := 0
	for isLocked() {
		elapsed += 1
		time.Sleep(1000 * time.Millisecond)
		if elapsed > timeout {
			return fmt.Errorf("timeout error for setLock")
		}
	}
	err := common.WriteString(label, SandboxRegistryLock)
	if err != nil {
		return err
	}
	catalogMutex.Lock()
	return nil
}

func releaseLock() error {
	if !enableCatalogManagement {
		return nil
	}
	if isLocked() {
		err := os.Remove(SandboxRegistryLock)
		if err != nil {
			return err
		}
	}
	catalogMutex.Unlock()
	return nil
}

func WriteCatalog(sc SandboxCatalog) error {
	if !enableCatalogManagement {
		return nil
	}
	byteBuf, err := json.MarshalIndent(sc, " ", "\t")
	common.ErrCheckExitf(err, 1, "error encoding sandbox catalog: %s", err)
	jsonString := fmt.Sprintf("%s", byteBuf)
	filename := SandboxRegistry
	return common.WriteString(jsonString, filename)
}

func ReadCatalog() (sc SandboxCatalog, err error) {
	if !enableCatalogManagement {
		return
	}
	filename := SandboxRegistry
	// if there is no catalog file (yet) we return an empty list
	if !common.FileExists(filename) {
		return sc, nil
	}
	fileStat, err := os.Stat(filename)
	if err != nil {
		return sc, err
	}
	if fileStat.Size() < 2 {
		return sc, err
	}
	scBlob, err := common.SlurpAsBytes(filename)
	if err != nil {
		return sc, err
	}

	err = json.Unmarshal(scBlob, &sc)
	if err != nil {
		if globals.UsingDbDeployer {
			fmt.Printf("error decoding catalog - Returning empty catalog: %s", err)
		}
		return SandboxCatalog{}, nil
	}
	return
}

func UpdateCatalog(sbName string, details SandboxItem) error {
	details.DbDeployerVersion = common.VersionDef
	details.Timestamp = time.Now().Format(time.UnixDate)
	details.CommandLine = strings.Join(common.CommandLineArgs, " ")
	if !enableCatalogManagement {
		return nil
	}
	err := setLock(sbName)
	if err == nil {
		current, err := ReadCatalog()
		if err != nil {
			err1 := releaseLock()
			if err1 != nil {
				panic(fmt.Sprintf("%s", err))
			}
			return err
		}
		if current == nil {
			current = make(SandboxCatalog)
		}
		current[sbName] = details
		err = WriteCatalog(current)
		err1 := releaseLock()
		if err1 != nil {
			panic(fmt.Sprintf("%s", err))
		}
		return err
	} else {
		common.CondPrintf("%s\n", globals.HashLine)
		common.CondPrintf("# UpdateCatalog Could not get lock on %s\n", SandboxRegistryLock)
		common.CondPrintf("%s\n", globals.HashLine)
		return fmt.Errorf("could not get lock on %s : %s", SandboxRegistryLock, err)
	}
}

func DeleteFromCatalog(sbName string) error {
	if !enableCatalogManagement {
		return nil
	}
	err := setLock(sbName)
	if err == nil {
		current, err := ReadCatalog()
		if err != nil {
			err1 := releaseLock()
			if err1 != nil {
				panic(fmt.Sprintf("%s", err))
			}
			return err
		}
		if current == nil {
			err1 := releaseLock()
			if err1 != nil {
				panic(fmt.Sprintf("%s", err))
			}
			return nil
		}
		delete(current, sbName)
		err = WriteCatalog(current)
		err1 := releaseLock()
		if err1 != nil {
			panic(fmt.Sprintf("%s", err))
		}
		return err
	} else {
		common.CondPrintf("%s\n", globals.HashLine)
		common.CondPrintf("# DeleteFromCatalog Could not get lock on %s\n", SandboxRegistryLock)
		common.CondPrintf("%s\n", globals.HashLine)
		return fmt.Errorf("could not get lock on %s: %s", SandboxRegistryLock, err)
	}
}

func init() {
	if common.IsEnvSet("SKIP_DBDEPLOYER_CATALOG") {
		enableCatalogManagement = false
	}
}
