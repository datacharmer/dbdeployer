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

	"github.com/datacharmer/dbdeployer/common"
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
			description := "single"
			sbdesc := SandboxHome + "/" + fname + "/sbdescription.json"
			if common.FileExists(sbdesc) {
				sbd := common.ReadSandboxDescription(SandboxHome + "/" + fname)
				if sbd.Nodes == 0 {
					description = fmt.Sprintf("%-15s %10s [%5d ]", sbd.SBType, sbd.Version, sbd.Port)
				} else {
					var node_descr []common.SandboxDescription
					if common.DirExists(SandboxHome + "/" + fname + "/master") {
						sd_master := common.ReadSandboxDescription(SandboxHome + "/" + fname + "/master")
						node_descr = append(node_descr, sd_master)
					}
					for node := 1; node <= sbd.Nodes; node++ {
						sd_node := common.ReadSandboxDescription(fmt.Sprintf("%s/%s/node%d", SandboxHome, fname, node))
						node_descr = append(node_descr, sd_node)
					}
					ports := "["
					for _,nd := range node_descr {
						ports += fmt.Sprintf(" %d", nd.Port)
					}
					ports += " ]"
					description = fmt.Sprintf("%-15s %10s %s", sbd.SBType, sbd.Version, ports)
				}
				dirs = append(dirs, fmt.Sprintf("%-20s : %s", fname, description))
			} else {
				start := SandboxHome + "/" + fname + "/start"
				start_all := SandboxHome + "/" + fname + "/start_all"
				initialize_slaves := SandboxHome + "/" + fname + "/initialize_slaves"
				initialize_nodes := SandboxHome + "/" + fname + "/initialize_nodes"
				if common.FileExists(start_all) {
					description = "multiple sandbox"
				}
				if common.FileExists(initialize_slaves) {
					description = "master-slave replication"
				}
				if common.FileExists(initialize_nodes) {
					description = "group replication"
				}
				if common.FileExists(start) || common.FileExists(start_all) {
					dirs = append(dirs, fmt.Sprintf("%-20s : %s", fname, description))
				}
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
