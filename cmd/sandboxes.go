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

	"github.com/datacharmer/dbdeployer/common"
	"github.com/spf13/cobra"
)

// Shows installed sandboxes
func ShowSandboxes(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	SandboxHome, _ := flags.GetString("sandbox-home")
	sandbox_list := common.GetInstalledSandboxes(SandboxHome)
	var dirs []string
	for _, sbinfo := range sandbox_list {
		//fname := f.Name()
		//fmode := f.Mode()
		//if fmode.IsDir() {
		fname := sbinfo.SandboxName
		description := "single"
		sbdesc := SandboxHome + "/" + fname + "/sbdescription.json"
		if common.FileExists(sbdesc) {
			sbd := common.ReadSandboxDescription(SandboxHome + "/" + fname)
			locked := ""
			if sbinfo.Locked {
				locked = "(LOCKED)"
			}
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
				inner_files := common.SandboxInfoToFileNames(common.GetInstalledSandboxes(SandboxHome + "/" + fname))
				for _, inner := range inner_files {
					inner_sbdesc := SandboxHome + "/" + fname + "/" + inner + "/sbdescription.json"
					if common.FileExists(inner_sbdesc) {
						sd_node := common.ReadSandboxDescription(fmt.Sprintf("%s/%s/%s", SandboxHome, fname, inner))
						node_descr = append(node_descr, sd_node)
					}
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
			dirs = append(dirs, fmt.Sprintf("%-20s : %s %s", fname, description, locked))
		} else {
			locked := ""
			no_clear := SandboxHome + "/" + fname + "/no_clear"
			no_clear_all := SandboxHome + "/" + fname + "/no_clear_all"
			start := SandboxHome + "/" + fname + "/start"
			start_all := SandboxHome + "/" + fname + "/start_all"
			initialize_slaves := SandboxHome + "/" + fname + "/initialize_slaves"
			initialize_nodes := SandboxHome + "/" + fname + "/initialize_nodes"
			if common.FileExists(no_clear) || common.FileExists(no_clear_all) {
				locked = "(LOCKED)"
			}
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
				dirs = append(dirs, fmt.Sprintf("%-20s : *%s* %s ", fname, description, locked))
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
	Long:    `Lists all sandboxes installed in $SANDBOX_HOME.`,
	Aliases: []string{"installed", "deployed"},
	Run:     ShowSandboxes,
}

func init() {
	rootCmd.AddCommand(sandboxesCmd)

	// sandboxesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
