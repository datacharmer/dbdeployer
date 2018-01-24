// Copyright © 2017-2018 Giuseppe Maxia
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
	"log"

	"dbdeployer/common"
	"github.com/spf13/cobra"
	"io/ioutil"
)

// Shows installed sandboxes
func ShowSandboxes(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	SandboxHome, _ := flags.GetString("sandbox-home")
	files, err := ioutil.ReadDir(SandboxHome)
	if err != nil {
		log.Fatal(err)
	}
	var dirs []string
	for _, f := range files {
		fname := f.Name()
		fmode := f.Mode()
		if fmode.IsDir() {
			start := SandboxHome + "/" + fname + "/start"
			start_all := SandboxHome + "/" + fname + "/start_all"
			if common.FileExists(start) || common.FileExists(start_all) {
				dirs = append(dirs, fname)
			}
		}
	}
	for _, dir := range dirs {
		fmt.Println(dir)
	}
}

// sandboxesCmd represents the sandboxes command
var sandboxesCmd = &cobra.Command{
	Use:     "sandboxes",
	Short:   "List installed sandboxes",
	Long:    ``,
	Aliases: []string{"installed", "deployed"},
	Run:     ShowSandboxes,
}

func init() {
	rootCmd.AddCommand(sandboxesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sandboxesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sandboxesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
