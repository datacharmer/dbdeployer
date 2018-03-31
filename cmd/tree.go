// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2018 Giuseppe Maxia
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
	"strings"
	"github.com/spf13/cobra"
)

func ShowTree(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	api, _  := flags.GetBool("api")
	show_hidden, _  := flags.GetBool("show-hidden")
	if api {
		fmt.Println("This is the list of commands and modifiers available for dbdeployer")
		fmt.Println("")
		fmt.Println("{{dbdeployer --version}}")
		fmt.Println("# main")
		fmt.Println("{{dbdeployer -h }}")
		fmt.Println("")
		fmt.Println("{{dbdeployer tree }}")
	}
	traverse(rootCmd, "", 0, api, show_hidden)
}

func traverse(cmd *cobra.Command, parent string, level int, api, show_hidden bool) {
	subcommands := cmd.Commands()
	indent := strings.Repeat(" ", level*4) + "-"
	for _, c := range subcommands {
		hidden_flag := ""
		if c.Hidden || c.Name() == "help" {
			if show_hidden {
				hidden_flag = " (HIDDEN) "
			} else {
				continue
			}
		}
		size := len(c.Commands())
		if api {
			if size > 0  || level == 0 {
				fmt.Printf("\n##%s%s\n", parent + " " + c.Name(), hidden_flag)
			}
			fmt.Printf("{{dbdeployer%s %s -h}}\n", parent, c.Name())
		} else {
			fmt.Printf("%s %-20s%s\n", indent, c.Name(), hidden_flag)
		}
		if size > 0 {
			traverse(c, parent + " " + c.Name(), level + 1, api, show_hidden)
		}
	}
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "shows command tree",
	Long: `This command is only used to create API documentation. 
You can, however, use it to show the command structure at a glance.`,
	Hidden : true,
	Run: ShowTree,
}

func init() {
	rootCmd.AddCommand(treeCmd)
	treeCmd.PersistentFlags().Bool("api", false, "Writes API template")
	treeCmd.PersistentFlags().Bool("show-hidden", false, "Shows also hidden commands")
}
