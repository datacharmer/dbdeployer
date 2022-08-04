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
	"os"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

// Describes the parameters needed after a known parameter
// such as a command keyword or an option
type RequiredInfo struct {
	HowMany       int    `json:"n"`                  // How many times this item should be used
	Name          string `json:"name"`               // Name of the item
	ReferenceData string `json:"ref_data,omitempty"` // Data used for multiple choice (if available)
}

// Defines what is needed for a command at the end of the known keywords
type ExportAnnotation struct {
	Arguments []RequiredInfo `json:"arguments,omitempty"` // Required or recommended arguments
	Options   []RequiredInfo `json:"options,omitempty"`   // Options needed for this command
}

// Describes a command option
type Option struct {
	Name     string `json:"name"`     // Name of the option
	Type     string `json:"type"`     // Type
	Usage    string `json:"usage"`    // Brief help on how to use this option
	Shortcut string `json:"shortcut"` // Option abbreviation
	Default  string `json:"default"`  // Default value
}

// Describes a command (or sub-command)
type Command struct {
	Version         string           `json:"version,omitempty"`     // dbdeployer version
	Name            string           `json:"name"`                  // Name of the command itself
	Use             string           `json:"use"`                   // How to use it
	Aliases         []string         `json:"aliases,omitempty"`     // Alternative names for this command
	Short           string           `json:"short,omitempty"`       // Short usage description
	Long            string           `json:"long,omitempty"`        // Long usage description
	Example         string           `json:"example,omitempty"`     // Usage examples, if any
	Breadcrumbs     []string         `json:"ancestors"`             // sequence of commands up to this (sub)command
	SubCommands     []Command        `json:"commands,omitempty"`    // Children of this (sub)command
	Options         []Option         `json:"flags,omitempty"`       // List of flags valid for this command
	NeedSubCommands bool             `json:"needs_sub_commands"`    // When the command alone cannot execute
	Annotations     ExportAnnotation `json:"annotations,omitempty"` // Arguments and constraints for this command
}

// Creates an annotation definition for needed arguments
// after a command
func makeExportArgs(name string, howMany int) string {
	return ExportAnnotationToJson(ExportAnnotation{
		Arguments: []RequiredInfo{
			{
				HowMany: howMany,
				Name:    name,
			},
		},
	})
}

// An annotation defining a string parameter of type sandbox template-group
var TemplateGroupExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: globals.ExportTemplateGroup},
		{HowMany: 1, Name: globals.ExportTemplateName},
	},
}

// An annotation defining a string parameter of type sandbox template-name
var TemplateNameExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: globals.ExportTemplateName},
	},
}

// An annotation defining what are the minimum required
// arguments for replication
var ReplicationExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: globals.ExportVersionDir},
	},
	Options: []RequiredInfo{
		{HowMany: 1, Name: globals.ExportTopology, ReferenceData: globals.ExportAllowedTopologies},
	},
}

// An annotation defining an argument as a generic string
var StringExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 1, Name: globals.ExportString},
	},
}

// An annotation defining two generic string arguments
var DoubleStringExport = ExportAnnotation{
	Arguments: []RequiredInfo{
		{HowMany: 2, Name: globals.ExportString},
	},
}

// Converts an annotation structure to a JSON string
func ExportAnnotationToJson(ea ExportAnnotation) string {
	b, err := json.MarshalIndent(ea, " ", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}

// Converts cobra.Command annotations intended for export into the
// original structure.
// The annotation is stored as
// map[string]string{ "export" : "string with JSON structure"}
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

// Converts a cobra.Command into a cmd.Command structure
func cobraToCommand(c *cobra.Command, ancestors []string, addVersion bool) Command {

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
	command.Breadcrumbs = append(command.Breadcrumbs, ancestors...)
	command.Breadcrumbs = append(command.Breadcrumbs, c.Name())
	if addVersion {
		// If this is the top command in the request, we want to report the version
		if command.Version == "" {
			command.Version = common.VersionDef
		}
	}
	for _, sc := range c.Commands() {
		subCommand := cobraToCommand(sc, command.Breadcrumbs, false)
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
			op.Shortcut = flag.Shorthand
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
	return string(b)
}

// Exports the whole dbdeployer structure to JSON
func ExportJson() string {
	return CommandToJson(Export())
}

// Returns the dbdeployer structure as a Command type
// This is used to create Go applications with a different
// user interface
func Export() Command {
	return cobraToCommand(rootCmd, []string{}, true)
}

// Export a given sub-command as JSON
func ExportJsonNamed(name string, subCommand string) string {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			if subCommand == "" {
				return CommandToJson(cobraToCommand(c, []string{rootCmd.Name()}, true))
			} else {
				for _, sc := range c.Commands() {
					if sc.Name() == subCommand {
						return CommandToJson(cobraToCommand(sc, []string{rootCmd.Name(), c.Name()}, true))
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
	outputGoesToTerminal := term.IsTerminal(int(os.Stdout.Fd()))
	forceOutputToTerminal, _ := c.Flags().GetBool(globals.ForceOutputToTermLabel)
	if outputGoesToTerminal && !forceOutputToTerminal {
		fmt.Println("No pager or pipe is used.")
		fmt.Println("The output of this command can be quite large.")
		fmt.Println("Please redirect to a pipe ")
		fmt.Println("(such as 'dbdeployer export | less' or 'dbdeployer export > struct.json')")
		fmt.Printf("or use the option --%s\n", globals.ForceOutputToTermLabel)
		os.Exit(0)
	}
	if len(args) > 0 {
		cmdString := args[0]
		subCmdString := ""
		if len(args) > 1 {
			subCmdString = args[1]
		}
		fmt.Println(ExportJsonNamed(cmdString, subCmdString))
	} else {
		fmt.Println(ExportJson())
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
	exportCmd.PersistentFlags().Bool(globals.ForceOutputToTermLabel, false, "display output to terminal regardless of pipes being used")
}
