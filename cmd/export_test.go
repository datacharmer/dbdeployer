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

package cmd

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/datacharmer/dbdeployer/compare"
)

func TestExport(t *testing.T) {

	rootCommand := Export()
	subCommands := len(rootCmd.Commands())
	compare.OkEqualInt("number of root sub-commands",
		len(rootCommand.SubCommands), subCommands, t)

	for _, c := range rootCmd.Commands() {
		command := cobraToCommand(c, false)
		subCommands := len(c.Commands())
		compare.OkEqualInt(fmt.Sprintf("number of %s sub-commands", c.Name()),
			len(command.SubCommands), subCommands, t)
		compare.OkNotEmptyString("version is set", command.Version, t)
	}

	toJson := ExportJson()

	var fromJson Command
	err := json.Unmarshal([]byte(toJson), &fromJson)
	compare.OkIsNil("json encoding err", err, t)

}
