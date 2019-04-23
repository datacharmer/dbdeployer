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
	"github.com/datacharmer/dbdeployer/globals"
	"testing"

	"github.com/datacharmer/dbdeployer/compare"
)

func TestExport(t *testing.T) {

	rootCommand := Export()
	subCommands := len(rootCmd.Commands())
	compare.OkEqualInt("number of root sub-commands",
		len(rootCommand.SubCommands), subCommands, t)

	for _, c := range rootCmd.Commands() {
		command := cobraToCommand(c, []string{}, true)
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

func TestExportStructure(t *testing.T) {
	rootCommand := Export()
	for _, c := range rootCommand.SubCommands {
		compare.OkEqualInt(fmt.Sprintf("%s breadcrumbs", c.Name), len(c.Breadcrumbs), 2, t)
		compare.OkEqualString(fmt.Sprintf("%s's immediate ancestor", c.Name), c.Breadcrumbs[0], rootCommand.Name, t)
		if len(c.SubCommands) > 0 {
			for _, c2 := range c.SubCommands {
				compare.OkEqualInt(fmt.Sprintf("%s's breadcrumbs", c2.Name), len(c2.Breadcrumbs), 3, t)
				compare.OkEqualString(fmt.Sprintf("%s's highest ancestor", c2.Name), c2.Breadcrumbs[0], rootCommand.Name, t)
				compare.OkEqualString(fmt.Sprintf("%s's immediate ancestor", c2.Name), c2.Breadcrumbs[1], c.Name, t)
				compare.OkEqualString(fmt.Sprintf("%s's latest breadcrumb", c2.Name), c2.Breadcrumbs[2], c2.Name, t)

				if len(c2.SubCommands) > 0 {
					for _, c3 := range c2.SubCommands {
						compare.OkEqualInt(fmt.Sprintf("%s's breadcrumbs", c3.Name), len(c3.Breadcrumbs), 4, t)
						compare.OkEqualString(fmt.Sprintf("%s's highest ancestor", c3.Name), c3.Breadcrumbs[0], rootCommand.Name, t)
						compare.OkEqualString(fmt.Sprintf("%s's high ancestor", c3.Name), c3.Breadcrumbs[1], c.Name, t)
						compare.OkEqualString(fmt.Sprintf("%s's immediate ancestor", c3.Name), c3.Breadcrumbs[2], c2.Name, t)
						compare.OkEqualString(fmt.Sprintf("%s's latest breadcrumb", c3.Name), c3.Breadcrumbs[3], c3.Name, t)

					}
				}
			}
		}
	}
	exportCapabilities := ExportJsonNamed("admin", "capabilities")
	var commandCapabilities Command
	err := json.Unmarshal([]byte(exportCapabilities), &commandCapabilities)
	compare.OkIsNil("JSON export to command", err, t)
	compare.OkEqualString("command name", commandCapabilities.Name, "capabilities", t)
	compare.OkEqualInt("command breadcrumbs", len(commandCapabilities.Breadcrumbs), 3, t)
	compare.OkMatchesString("command use", commandCapabilities.Use, `capabilities.*flavor.*version`, t)
}

func TestExportImport(t *testing.T) {

	type ExportImport struct {
		commandName         string
		subCommandName      string
		expectedName        string
		expectedAncestors   int
		expectedSubCommands int
		expectedArgument    string
	}
	var data = []ExportImport{
		{"admin", "", "admin",
			2, 4, ""},
		{"admin", "capabilities", "capabilities",
			3, 0, ""},
		{"admin", "lock", "lock",
			3, 0, globals.ExportSandboxDir},
		{"admin", "unlock", "unlock",
			3, 0, globals.ExportSandboxDir},
		{"admin", "upgrade", "upgrade",
			3, 0, globals.ExportSandboxDir},
		{"defaults", "templates", "templates",
			3, 6, ""},
		{"delete", "", "delete",
			2, 0, globals.ExportSandboxDir},
		{"delete-binaries", "", "delete-binaries",
			2, 0, globals.ExportVersionDir},
		{"unpack", "", "unpack",
			2, 0, globals.ExportString},
		{"deploy", "single", "single",
			3, 0, globals.ExportVersionDir},
		{"deploy", "replication", "replication",
			3, 0, globals.ExportVersionDir},
		{"deploy", "multiple", "multiple",
			3, 0, globals.ExportVersionDir},
		{"cookbook", "list", "list",
			3, 0, ""},
		{"cookbook", "create", "create",
			3, 0, globals.ExportCookbookName},
		{"cookbook", "show", "show",
			3, 0, globals.ExportCookbookName},
		{"global", "", "global",
			2, 7, ""},
		{"global", "test", "test",
			3, 0, ""},
		{"global", "test-replication", "test-replication",
			3, 0, ""},
		{"global", "status", "status",
			3, 0, ""},
		{"global", "start", "start",
			3, 0, ""},
		{"global", "restart", "restart",
			3, 0, ""},
		{"global", "stop", "stop",
			3, 0, ""},
		{"global", "use", "use",
			3, 0, globals.ExportString},
		{"sandboxes", "", "sandboxes",
			2, 0, ""},
		{"usage", "", "usage",
			2, 0, ""},
		{"versions", "", "versions",
			2, 0, ""},
	}

	for _, sample := range data {
		exportSample := ExportJsonNamed(sample.commandName, sample.subCommandName)
		var commandSample Command
		err := json.Unmarshal([]byte(exportSample), &commandSample)
		compare.OkIsNil(
			fmt.Sprintf("JSON export to command %s.%s", sample.subCommandName, sample.subCommandName), err, t)
		compare.OkEqualInt(
			fmt.Sprintf("command %s.%s breadcrumbs", sample.commandName, sample.subCommandName),
			len(commandSample.Breadcrumbs), sample.expectedAncestors, t)
		compare.OkEqualInt(
			fmt.Sprintf("command %s.%s subcommands", sample.commandName, sample.subCommandName),
			len(commandSample.SubCommands), sample.expectedSubCommands, t)
		foundArgument := ""
		if len(commandSample.Annotations.Arguments) > 0 {
			foundArgument = commandSample.Annotations.Arguments[0].Name
		}
		compare.OkEqualString(fmt.Sprintf("command %s.%s argument", sample.commandName, sample.subCommandName),
			foundArgument, sample.expectedArgument, t)
		compare.OkEqualString(fmt.Sprintf("command %s.%s name", sample.commandName, sample.subCommandName),
			commandSample.Name, sample.expectedName, t)
	}
}
