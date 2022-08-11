// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2022 Giuseppe Maxia
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

package ops

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/importing"
)

type sandboxConnection struct {
	Host     string `json:"master_host"`
	Port     int    `json:"master_port"`
	User     string `json:"master_user"`
	Password string `json:"master_password"`
}

// getSandboxConnection finds the connection credentials in a given sandbox directory
func getSandboxConnection(sandboxPath string, asSuperUser bool) (sandboxConnection, error) {
	var sc sandboxConnection
	if !common.DirExists(sandboxPath) {
		return sc, fmt.Errorf(globals.ErrDirectoryNotFound, sandboxPath)
	}
	connectionFile := path.Join(sandboxPath, globals.ScriptConnectionJson)
	if asSuperUser {
		connectionFile = path.Join(sandboxPath, globals.ScriptConnectionSuperJson)
	}
	if !common.FileExists(connectionFile) {
		return sc, fmt.Errorf(globals.ErrFileNotFound, connectionFile)
	}

	text, err := os.ReadFile(connectionFile) // #nosec 304
	if err != nil {
		return sc, err
	}
	err = json.Unmarshal(text, &sc)
	return sc, err
}

// RunSandboxQuery runs a SQL query in a given sandbox directory
func RunSandboxQuery[T comparable](sandboxPath, query string, asSuperUser bool) (interface{}, error) {

	credentials, err := getSandboxConnection(sandboxPath, asSuperUser)
	if err != nil {
		return "", err
	}
	config := importing.ParamsToConfig(credentials.Host, credentials.User, credentials.Password, credentials.Port)
	db, err := importing.Connect(config)
	if err != nil {
		return "", err
	}
	var result T
	err = db.GetSingleResult(config, query, &result)
	return result, err
}
