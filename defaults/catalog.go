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
	"strings"
	"time"

	"github.com/nightlyone/lockfile"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
)

type SandboxItem struct {
	Origin            string   `json:"origin"`
	SBType            string   `json:"type"` // single multi master-slave group all-masters fan-in ndb pxc
	Version           string   `json:"version"`
	Flavor            string   `json:"flavor,omitempty"`
	Host              string   `json:"host,omitempty"`
	Port              []int    `json:"port"`
	Nodes             []string `json:"nodes"`
	Destination       string   `json:"destination"`
	DbDeployerVersion string   `json:"dbdeployer-version"`
	Timestamp         string   `json:"timestamp"`
	LogDirectory      string   `json:"log-directory,omitempty"`
	CommandLine       string   `json:"command-line"`
}

type SandboxCatalog map[string]SandboxItem

var enableCatalogManagement bool = true

// Timeout for waiting on concurrent requests
const lockTimeout = 1000 * time.Millisecond

// Writes the catalog on file
// This is an unsafe operation, which must be kept under a lock
func writeCatalog(sc SandboxCatalog) error {
	if !enableCatalogManagement {
		return nil
	}
	byteBuf, err := json.MarshalIndent(sc, " ", "\t")
	common.ErrCheckExitf(err, 1, "error encoding sandbox catalog: %s", err)
	jsonString := string(byteBuf)
	filename := SandboxRegistry
	return common.WriteString(jsonString, filename)
}

// Reads the catalog, making sure that there are no concurrent operations
func ReadCatalog() (sc SandboxCatalog, err error) {
	lock, err := setLock("reading")
	if err != nil {
		return
	}
	sc, err = unsafeReadCatalog()
	_ = lock.Unlock()
	return
}

// Reads the catalog, without waiting for a lock
func unsafeReadCatalog() (sc SandboxCatalog, err error) {
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

// Sets the catalog lock, waiting up to lockTimeout milliseconds
// if a concurrent operation is under way.
// This approach guarantees thread and inter-process safety
func setLock(label string) (lockfile.Lockfile, error) {
	lock, err := lockfile.New(SandboxRegistryLock)
	if err != nil {
		return lockfile.Lockfile(""), fmt.Errorf("could not establish lock file for %s: %s", label, err)
	}
	err = lock.TryLock()
	var elapsed time.Duration
	for err == lockfile.ErrBusy || err == lockfile.ErrNotExist {
		time.Sleep(3 * time.Millisecond)
		elapsed += 3
		if elapsed > lockTimeout {
			break
		}
		err = lock.TryLock()
	}
	if err != nil {
		return lockfile.Lockfile(""), fmt.Errorf("could not set lock for %s: %s", label, err)
	}
	return lock, nil
}

// Safe update of the catalog, protected by a lock
func UpdateCatalog(sbName string, details SandboxItem) error {
	details.DbDeployerVersion = common.VersionDef
	details.Timestamp = time.Now().Format(time.UnixDate)
	details.CommandLine = strings.Join(common.CommandLineArgs, " ")
	if !enableCatalogManagement {
		return nil
	}
	lock, err := setLock(sbName)
	if err != nil {
		return err
	}
	defer lock.Unlock()
	err = checkCatalog()
	if err != nil {
		return err
	}
	current, err := unsafeReadCatalog()
	if err != nil {
		return err
	}
	if current == nil {
		current = make(SandboxCatalog)
	}
	current[sbName] = details
	err = writeCatalog(current)
	return err
}

// Safe deletion of a catalog entry
func DeleteFromCatalog(sbName string) error {
	if !enableCatalogManagement {
		return nil
	}
	lock, err := setLock(sbName)
	if err != nil {
		return err
	}
	defer lock.Unlock()
	err = checkCatalog()
	if err != nil {
		return err
	}
	current, err := unsafeReadCatalog()
	if err != nil {
		return err
	}
	if current == nil {
		return nil
	}
	delete(current, sbName)
	err = writeCatalog(current)
	return err
}

// Check that the configuration directory exists and creates it if needed
// If no catalog exists, creates an empty one.
func checkCatalog() error {
	if !common.DirExists(ConfigurationDir) {
		err := os.Mkdir(ConfigurationDir, globals.PublicDirectoryAttr)
		if err != nil {
			return fmt.Errorf("error making lock directory %s", ConfigurationDir)
		}
	}
	if !common.FileExists(SandboxRegistry) {
		err := common.WriteString("{}", SandboxRegistry)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	if common.IsEnvSet("SKIP_DBDEPLOYER_CATALOG") {
		enableCatalogManagement = false
	}
}
