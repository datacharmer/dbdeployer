// Copyright © 2018 Giuseppe Maxia
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
	"fmt"
	"os"

	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func RunDescribeTemplate(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Argument required: template name")
		os.Exit(1)
	}
	requested := args[0]
	flags := cmd.Flags()
	complete_listing, _ := flags.GetBool("with-contents")
	DescribeTemplate(requested, complete_listing)
}

func GetTemplatesDescription(requested string, complete_listing bool) string {
	group, contents := FindTemplate(requested)
	out := ""
	out += fmt.Sprintf("# Collection    : %s\n", group)
	out += fmt.Sprintf("# Name          : %s\n", requested)
	out += fmt.Sprintf("# Description 	: %s\n", sandbox.AllTemplates[group][requested].Description)
	out += fmt.Sprintf("# Notes     	: %s\n", sandbox.AllTemplates[group][requested].Notes)
	out += fmt.Sprintf("# Length     	: %d\n", len(contents))
	if complete_listing {
		out += fmt.Sprintf("##START\n")
		out += fmt.Sprintf("%s\n", contents)
		out += fmt.Sprintf("##END\n\n")
	}
	return out
}

func DescribeTemplate(requested string, complete_listing bool) {
	fmt.Printf("%s", GetTemplatesDescription(requested, complete_listing))
}

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:     "describe",
	Aliases: []string{"descr", "structure", "struct"},
	Short:   "Describe a given template",
	Long:    ``,
	Run:     RunDescribeTemplate,
}

func init() {
	templatesCmd.AddCommand(describeCmd)

	describeCmd.Flags().BoolP("with-contents", "", false, "Shows complete structure and contents")
}
