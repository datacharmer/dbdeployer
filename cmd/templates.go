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
	//"fmt"

	"github.com/spf13/cobra"
)

// templatesCmd represents the templates command
var templatesCmd = &cobra.Command{
	Use:     "templates",
	Aliases: []string{"template", "tmpl", "templ"},
	Short:   "Admin operations on templates",
	Hidden:  false,
	Long: `The commands in this section show the templates used 
to create and manipulate sandboxes.
More commands (and flags) will follow to allow changing templates
either temporarily or permanently.`,
}

func init() {
	rootCmd.AddCommand(templatesCmd)

}
