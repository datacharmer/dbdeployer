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

func ListTemplates (cmd *cobra.Command, args []string) {
	wanted := ""
	if len(args) > 0 {
		wanted = args[0]
	}

	found := false
	for name, tvar := range sandbox.AllTemplates {
		will_print := true
		if wanted != "" {
			if wanted != name {
				will_print = false
			}
		}
		if will_print {
			fmt.Printf("# %s\n", name)
			for k, v := range tvar {
				fmt.Printf("%-25s : %s\n", k, v.Description)
			}
			fmt.Println("")
			found = true
		}
	}
	if ! found {
		fmt.Printf("group %s not found\n", wanted)
		os.Exit(1)
	}
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [group]",
	Short: "list available templates",
	Long: ``,
	Run: ListTemplates,
}

func init() {
	templatesCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
