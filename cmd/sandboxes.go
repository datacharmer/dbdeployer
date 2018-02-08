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

func GetInstalledPorts(sandbox_home string) []int {
	files, err := ioutil.ReadDir(sandbox_home)
	if err != nil {
		log.Fatal(err)
	}
	var port_collection []int
	for _, f := range files {
		fname := f.Name()
		fmode := f.Mode()
		if fmode.IsDir() {
			sbdesc := sandbox_home + "/" + fname + "/sbdescription.json"
			if common.FileExists(sbdesc) {
				sbd := common.ReadSandboxDescription(sandbox_home + "/" + fname)
				if sbd.Nodes == 0 {
					for _, p := range sbd.Port {
						port_collection = append(port_collection, p)
					}
				} else {
					var node_descr []common.SandboxDescription
					if common.DirExists(sandbox_home + "/" + fname + "/master") {
						sd_master := common.ReadSandboxDescription(sandbox_home + "/" + fname + "/master")
						node_descr = append(node_descr, sd_master)
					}
					for node := 1; node <= sbd.Nodes; node++ {
						sd_node := common.ReadSandboxDescription(fmt.Sprintf("%s/%s/node%d", sandbox_home, fname, node))
						node_descr = append(node_descr, sd_node)
					}
					for _, nd := range node_descr {
						for _, p := range nd.Port {
							port_collection = append(port_collection, p)
						}
					}
				}
			}
		}
	}
	return port_collection
}

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
					port_text := ""
					for _, p := range sbd.Port {
						if port_text != "" {
							port_text += " "
						}
						port_text += fmt.Sprintf("%d", p)
					}
					description = fmt.Sprintf("%-20s %10s [%s]", sbd.SBType, sbd.Version, port_text)
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
					ports := ""
					for _, nd := range node_descr {
						for _, p := range nd.Port {
							if ports != "" {
								ports += " "
							}
							ports += fmt.Sprintf("%d", p)
						}
					}
					//ports += " ]"
					description = fmt.Sprintf("%-20s %10s [%s]", sbd.SBType, sbd.Version, ports)
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

	// sandboxesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
