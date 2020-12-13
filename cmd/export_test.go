// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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
	"github.com/datacharmer/dbdeployer/common"
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
}

// This test checks the expected structure and arguments
// for most dbdeployer commands
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
		{
			commandName:         "admin",
			subCommandName:      "",
			expectedName:        "admin",
			expectedAncestors:   2,
			expectedSubCommands: 6,
			expectedArgument:    "",
		},
		{
			commandName:         "admin",
			subCommandName:      "capabilities",
			expectedName:        "capabilities",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "admin",
			subCommandName:      "lock",
			expectedName:        "lock",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportSandboxDir,
		},
		{
			commandName:         "admin",
			subCommandName:      "unlock",
			expectedName:        "unlock",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportSandboxDir,
		},
		{
			commandName:         "admin",
			subCommandName:      "upgrade",
			expectedName:        "upgrade",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportSandboxDir,
		},
		{
			commandName:         "defaults",
			subCommandName:      "templates",
			expectedName:        "templates",
			expectedAncestors:   3,
			expectedSubCommands: 6,
			expectedArgument:    "",
		},
		{
			commandName:         "delete",
			subCommandName:      "",
			expectedName:        "delete",
			expectedAncestors:   2,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportSandboxDir,
		},
		{
			commandName:         "delete-binaries",
			subCommandName:      "",
			expectedName:        "delete-binaries",
			expectedAncestors:   2,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportVersionDir,
		},
		{
			commandName:         "unpack",
			subCommandName:      "",
			expectedName:        "unpack",
			expectedAncestors:   2,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportString,
		},
		{
			commandName:         "deploy",
			subCommandName:      "",
			expectedName:        "deploy",
			expectedAncestors:   2,
			expectedSubCommands: 3,
			expectedArgument:    "",
		},
		{
			commandName:         "deploy",
			subCommandName:      "single",
			expectedName:        "single",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportVersionDir,
		},
		{
			commandName:         "deploy",
			subCommandName:      "replication",
			expectedName:        "replication",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportVersionDir,
		},
		{
			commandName:         "deploy",
			subCommandName:      "multiple",
			expectedName:        "multiple",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportVersionDir,
		},
		{
			commandName:         "export",
			subCommandName:      "",
			expectedName:        "export",
			expectedAncestors:   2,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "cookbook",
			subCommandName:      "list",
			expectedName:        "list",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "cookbook",
			subCommandName:      "create",
			expectedName:        "create",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportCookbookName,
		},
		{
			commandName:         "cookbook",
			subCommandName:      "show",
			expectedName:        "show",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportCookbookName,
		},
		{
			commandName:         "global",
			subCommandName:      "",
			expectedName:        "global",
			expectedAncestors:   2,
			expectedSubCommands: 9,
			expectedArgument:    "",
		},
		{
			commandName:         "global",
			subCommandName:      "test",
			expectedName:        "test",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "global",
			subCommandName:      "test-replication",
			expectedName:        "test-replication",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "global",
			subCommandName:      "status",
			expectedName:        "status",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "global",
			subCommandName:      "start",
			expectedName:        "start",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "global",
			subCommandName:      "restart",
			expectedName:        "restart",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "global",
			subCommandName:      "stop",
			expectedName:        "stop",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "global",
			subCommandName:      "use",
			expectedName:        "use",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportString,
		},
		{
			commandName:         "remote",
			subCommandName:      "",
			expectedName:        "remote",
			expectedAncestors:   2,
			expectedSubCommands: 2,
			expectedArgument:    "",
		},
		{
			commandName:         "remote",
			subCommandName:      "list",
			expectedName:        "list",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "remote",
			subCommandName:      "download",
			expectedName:        "download",
			expectedAncestors:   3,
			expectedSubCommands: 0,
			expectedArgument:    globals.ExportString,
		},
		{
			commandName:         "sandboxes",
			subCommandName:      "",
			expectedName:        "sandboxes",
			expectedAncestors:   2,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "usage",
			subCommandName:      "",
			expectedName:        "usage",
			expectedAncestors:   2,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
		{
			commandName:         "versions",
			subCommandName:      "",
			expectedName:        "versions",
			expectedAncestors:   2,
			expectedSubCommands: 0,
			expectedArgument:    "",
		},
	}

	for _, sample := range data {
		exportSample := ExportJsonNamed(sample.commandName, sample.subCommandName)
		var commandSample Command
		err := json.Unmarshal([]byte(exportSample), &commandSample)
		if err != nil {
			panic(fmt.Sprintf("ERROR unmarshalling JSON structure - Can't continue: %s", err))
		}
		testName := ""
		if sample.subCommandName == "" {
			testName = sample.commandName
		} else {
			testName = fmt.Sprintf("%s.%s", sample.commandName, sample.subCommandName)
		}

		foundArgument := ""
		if len(commandSample.Annotations.Arguments) > 0 {
			foundArgument = commandSample.Annotations.Arguments[0].Name
		}
		testExpectedBreadcrumbs := fmt.Sprintf("command %s breadcrumbs", testName)
		testExpectedSubCommands := fmt.Sprintf("command %s subcommands", testName)
		testExpectedName := fmt.Sprintf("command %s name", testName)
		testExpectedVersion := fmt.Sprintf("command %s version", testName)
		testArgument := fmt.Sprintf("command %s argument", testName)

		compare.OkIsNil(fmt.Sprintf("JSON export string to command %s", testName), err, t)
		compare.OkEqualInt(testExpectedBreadcrumbs, len(commandSample.Breadcrumbs), sample.expectedAncestors, t)
		compare.OkEqualInt(testExpectedSubCommands, len(commandSample.SubCommands), sample.expectedSubCommands, t)
		compare.OkEqualString(testArgument, foundArgument, sample.expectedArgument, t)

		// The name of the exported command should be either the main command (when the sub-command is empty)
		// or the name of the sub-command, when used
		compare.OkEqualString(testExpectedName, commandSample.Name, sample.expectedName, t)

		// Every exported structure at the top should have the current version
		compare.OkEqualString(testExpectedVersion, commandSample.Version, common.VersionDef, t)

		if len(commandSample.SubCommands) > 0 {
			subCommand := commandSample.SubCommands[0]
			// The version of an exported sub-command should be empty. Only the top exported command should
			// have the version
			compare.OkEqualString("sub-command version", subCommand.Version, "", t)
		}
	}
}
