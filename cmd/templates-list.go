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
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
	"os"
)

type TemplateInfo struct {
	Group       string
	Name        string
	Description string
}

func GetTemplatesList(wanted string) (tlist []TemplateInfo) {
	found := false
	for group_name, tvar := range sandbox.AllTemplates {
		will_include := true
		//fmt.Printf("[%s]\n", group_name)
		if wanted != "" {
			if wanted != group_name {
				will_include = false
			}
		}
		var td TemplateInfo
		if will_include {
			for k, v := range tvar {
				td.Description = v.Description
				td.Group = group_name
				td.Name = k
				tlist = append(tlist, td)
				found = true
			}
		}
	}
	if !found {
		fmt.Printf("group %s not found\n", wanted)
		os.Exit(1)
	}
	return
}

func ListTemplates(cmd *cobra.Command, args []string) {
	wanted := ""
	if len(args) > 0 {
		wanted = args[0]
	}
	flags := cmd.Flags()
	simple_list, _ := flags.GetBool("simple")

	templates := GetTemplatesList(wanted)
	for _, template := range templates {
		if simple_list {
			fmt.Printf("%-13s %-25s\n", "["+template.Group+"]", template.Name)
		} else {
			fmt.Printf("%-13s %-25s : %s\n", "["+template.Group+"]", template.Name, template.Description)
		}
	}
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [group]",
	Short: "list available templates",
	Long:  ``,
	Run:   ListTemplates,
}

func init() {
	templatesCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("simple", "s", false, "Shows only the template names, without description")
}
