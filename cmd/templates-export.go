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

	"github.com/datacharmer/dbdeployer/common"
	"github.com/spf13/cobra"
)

func ExportTemplates(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println("The export command requires two arguments: group_name and file_name")
		fmt.Println("If group_name is 'all', it will export all groups")
		os.Exit(1)
	}
	wanted := args[0]
	file_name := args[1]
	if wanted == "all" {
		wanted = ""
	}
	if common.FileExists(file_name) {
		fmt.Printf("# File <%s> already exists\n", file_name)
		os.Exit(1)
	}
	templates := GetTemplatesList(wanted)
	out := ""
	for _, template := range templates {
		out += GetTemplatesDescription(template.Name, true)
	}
	err := common.WriteString(out, file_name)
	if err == nil {
		fmt.Printf("Exported to file %s\n", file_name)
	} else {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export group_name file_name",
	Short: "Exports all templates to a file",
	//Args: cobra.MinimumNArgs(2),
	//Args: cobra.ExactArgs(2),
	Long: `Exports a group of templates to a given file`,
	Run:  ExportTemplates,
}

func init() {
	templatesCmd.AddCommand(exportCmd)

	// exportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
