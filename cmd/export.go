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
	"golang.org/x/crypto/ssh/terminal"
	"os"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Describes the parameters needed after a known parameter
// such as a command keyword or an option
type RequiredInfo struct {
	HowMany       int    `json:"n"`
	Name          string `json:"name"`
	ReferenceData string `json:"ref_data,omitempty"`
}

// Defines what is needed for a command at the end of the known keywords
type ExportAnnotation struct {
	Arguments []RequiredInfo `json:"arguments"`
	Options   []RequiredInfo `json:"options"`
}

// Describes a command option
type Option struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Usage   string `json:"usage"`
	Short   string `json:"short"`
	Default string `json:"default"`
}

// Describes a command (or sub-command)
type Command struct {
	Version string   `json:"version,omitempty"`
	Name    string   `json:"name"`
	Use     string   `json:"use"`
	Aliases []string `json:"aliases,omitempty"`
	Short   string   `json:"short,omitempty"`
	Long    string   `json:"long,omitempty"`
	Example string   `json:"example,omitempty"`

	SubCommands     []Command        `json:"commands,omitempty"`
	Options         []Option         `json:"flags,omitempty"`
	NeedSubCommands bool             `json:"needs_sub_commands"`
	Annotations     ExportAnnotation `json:"annotations,omitempty"`
}

// An annotation defining a command that requires a version directory
// (such as 5.7.25) as an argument
var DeployExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: "version-dir"},
	},
}

// An annotation for commands that require a sandbox directory
// (such as msb_5_7_25) as an argument
var SandboxDirExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: "sandbox-dir"},
	},
}

// An annotation defining two parameters of type sandbox directory
var DoubleSandboxDirExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 2, Name: "sandbox-dir"},
	},
}

// An annotation defining a string parameter of type sandbox cookbook-name
var CookbookNameExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: "cookbook-name"},
	},
}

// An annotation defining a string parameter of type sandbox template-group
var TemplateGroupExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: "template-group"},
		{HowMany: 1, Name: "string"},
	},
}

// An annotation defining a string parameter of type sandbox template-name
var TemplateNameExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: "template-name"},
	},
}

// An annotation defining what are the minimum required
// arguments for replication
var ReplicationExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: "version-dir"},
	},
	Options: []RequiredInfo{
		{HowMany: 1, Name: "topology", ReferenceData: "AllowedTopologies"},
	},
}

// An annotation defining an argument as a generic string
var StringExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: "string"},
	},
}

// An annotation defining two generic string arguments
var DoubleStringExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 2, Name: "string"},
	},
}

// An annotation defining an argument as a generic integer
var IntegerExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: "integer"},
	},
}

// Converts an annotation structure to a JSON string
func ExportAnnotationToJson(ea ExportAnnotation) string {
	b, err := json.MarshalIndent(ea, " ", "\t")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", b)
}

func jsonToExportAnnotation(annotations map[string]string) ExportAnnotation {
	var ea ExportAnnotation
	for k, v := range annotations {
		if k == "export" {
			err := json.Unmarshal([]byte(v), &ea)
			if err != nil {
				return ExportAnnotation{}
			}
		}
	}
	return ea
}

func cobraToCommand(c *cobra.Command, recursive bool) Command {

	var command Command
	command.Name = c.Name()
	command.Use = c.Use
	command.Short = c.Short
	command.Long = c.Long
	command.Example = c.Example
	command.Aliases = c.Aliases
	command.Version = c.Version
	command.Annotations = jsonToExportAnnotation(c.Annotations)
	command.NeedSubCommands = c.Run == nil
	if !recursive {
		// If this is the top command in the request, we want to report the version
		if command.Version == "" {
			command.Version = common.VersionDef
		}
	}
	for _, sc := range c.Commands() {
		subCommand := cobraToCommand(sc, true)
		command.SubCommands = append(command.SubCommands, subCommand)
	}

	var options []Option
	seenFlags := make(map[string]bool)
	visitFunc := func(flag *pflag.Flag) {
		var op Option
		_, seen := seenFlags[flag.Name]
		if !seen {
			op.Name = flag.Name
			op.Usage = flag.Usage
			op.Type = flag.Value.Type()
			op.Short = flag.Shorthand
			op.Default = flag.DefValue
			options = append(options, op)
			seenFlags[op.Name] = true
		}
	}
	if c.HasAvailableFlags() {
		c.Flags().VisitAll(visitFunc)
	}
	if c.HasAvailableLocalFlags() {
		c.LocalFlags().VisitAll(visitFunc)
	}
	// if c.HasAvailableInheritedFlags() {
	// 	c.InheritedFlags().VisitAll(visitFunc)
	// }
	if c.HasAvailablePersistentFlags() {
		c.PersistentFlags().VisitAll(visitFunc)
	}
	command.Options = options

	return command
}

// Converts a Command structure to JSON
func CommandToJson(c Command) string {
	b, err := json.MarshalIndent(c, " ", "\t")
	if err != nil {
		return "unencoded structure"
	}
	return fmt.Sprintf("%s", b)
}

// Exports the whole dbdeployer structure to JSON
func ExportJson() string {
	return CommandToJson(Export())
}

// Returns the dbdeployer structure as a Command type
// This is used to create Go applications with a different
// user interface
func Export() Command {
	return cobraToCommand(rootCmd, false)
}

// Export a given sub-command as JSON
func ExportJsonNamed(name string, subCommand string) string {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			if subCommand == "" {
				return CommandToJson(cobraToCommand(c, false))
			} else {
				for _, sc := range c.Commands() {
					if sc.Name() == subCommand {
						return CommandToJson(cobraToCommand(sc, false))
					}
				}
				return fmt.Sprintf(`{"ERROR": "Command '%s.%s' not found"}`, name, subCommand)
			}
		}
	}
	return fmt.Sprintf(`{"ERROR": "Command '%s' not found"}`, name)
}

func runExport(c *cobra.Command, args []string) {
	// Detects if output goes to terminal and warns about usage
	outputGoesToTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	forceOutputToTerminal, _ := c.Flags().GetBool("force-output-to-terminal")
	if outputGoesToTerminal && !forceOutputToTerminal {
		fmt.Printf("No pager or pipe is used.\n")
		fmt.Printf("The output of this command can be quite large.\n")
		fmt.Printf("Please redirect to a pipe \n")
		fmt.Printf("(such as 'dbdeployer export | less' or 'dbdeployer export > struct.json')\n")
		fmt.Printf("or use the option --force-output-to-terminal\n")
		os.Exit(0)
	}
	if len(args) > 0 {
		cmdString := args[0]
		subCmdString := ""
		if len(args) > 1 {
			subCmdString = args[1]
		}
		fmt.Printf("%s\n", ExportJsonNamed(cmdString, subCmdString))
	} else {
		fmt.Printf("%s\n", ExportJson())
	}
}

var (
	exportCmd = &cobra.Command{
		Use:     "export [command [sub-command]] [ > filename ] [ | command ] ",
		Short:   "Exports the command structure in JSON format",
		Aliases: []string{"dump"},
		Long: `Exports the command line structure, with examples and flags, to a JSON structure.
If a command is given, only the structure of that command and below will be exported.
Given the length of the output, it is recommended to pipe it to a file or to another command.
`,
		Run: runExport,
	}
)

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.PersistentFlags().Bool("force-output-to-terminal", false, "display output to terminal regardless of pipes being used")
}
