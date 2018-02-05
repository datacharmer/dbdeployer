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

	"github.com/spf13/cobra"
	"github.com/datacharmer/dbdeployer/sandbox"
)

func DescribeTemplate (cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Argument required: template name")
		os.Exit(1)
	}
	requested := args[0]
	name, contents := FindTemplate(requested)
	fmt.Printf("# Collection    : %s\n", name)
	fmt.Printf("# Name          : %s\n", requested)
	fmt.Printf("# Description 	: %s\n", sandbox.AllTemplates[name][requested].Description)
	fmt.Printf("# Notes     	: %s\n", sandbox.AllTemplates[name][requested].Notes)
	fmt.Printf("# Length     	: %d\n", len(contents))
}

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a given template",
	Long: ``,
	Run: DescribeTemplate,
}

func init() {
	templatesCmd.AddCommand(describeCmd)

}
