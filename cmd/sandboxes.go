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
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/spf13/cobra"
	"strings"
)

func ShowSandboxesFromCatalog(currentSandboxHome string, header bool) {
	sandboxList := defaults.ReadCatalog()
	if len(sandboxList) == 0 {
		return
	}
	template := "%-25s %-10s %-20s %5v %-25s %s \n"
	if header {
		fmt.Printf(template, "name", "version", "type", "nodes", "ports", "")
		fmt.Printf(template, "----", "-------", "-----", "-----", "-----", "")
	}
	for name, contents := range sandboxList {
		ports := "["
		for _, p := range contents.Port {
			ports += fmt.Sprintf("%d ", p)
		}
		ports += "]"
		extra := ""
		if !strings.HasPrefix(contents.Destination, currentSandboxHome) {
			extra = "(" + common.DirName(contents.Destination) + ")"
		}
		fmt.Printf(template, common.BaseName(name), contents.Version, contents.SBType, len(contents.Nodes), ports, extra)
	}
}

// Shows installed sandboxes
func ShowSandboxes(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	SandboxHome, _ := flags.GetString(defaults.SandboxHomeLabel)
	readCatalog, _ := flags.GetBool(defaults.CatalogLabel)
	useHeader, _ := flags.GetBool(defaults.HeaderLabel)
	if readCatalog {
		ShowSandboxesFromCatalog(SandboxHome, useHeader)
		return
	}
	sandboxList := common.GetInstalledSandboxes(SandboxHome)
	var dirs []string
	for _, sbinfo := range sandboxList {
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
				portText := ""
				for _, p := range sbd.Port {
					if portText != "" {
						portText += " "
					}
					portText += fmt.Sprintf("%d", p)
				}
				description = fmt.Sprintf("%-20s %10s [%s]", sbd.SBType, sbd.Version, portText)
			} else {
				var nodeDescr []common.SandboxDescription
				innerFiles := common.SandboxInfoToFileNames(common.GetInstalledSandboxes(SandboxHome + "/" + fname))
				for _, inner := range innerFiles {
					innerSbdesc := SandboxHome + "/" + fname + "/" + inner + "/sbdescription.json"
					if common.FileExists(innerSbdesc) {
						sdNode := common.ReadSandboxDescription(fmt.Sprintf("%s/%s/%s", SandboxHome, fname, inner))
						nodeDescr = append(nodeDescr, sdNode)
					}
				}
				ports := ""
				for _, nd := range nodeDescr {
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
			dirs = append(dirs, fmt.Sprintf("%-25s : %s %s", fname, description, locked))
		} else {
			locked := ""
			noClear := SandboxHome + "/" + fname + "/no_clear"
			noClearAll := SandboxHome + "/" + fname + "/no_clear_all"
			start := SandboxHome + "/" + fname + "/start"
			startAll := SandboxHome + "/" + fname + "/start_all"
			initializeSlaves := SandboxHome + "/" + fname + "/initialize_slaves"
			initializeNodes := SandboxHome + "/" + fname + "/initialize_nodes"
			if common.FileExists(noClear) || common.FileExists(noClearAll) {
				locked = "(LOCKED)"
			}
			if common.FileExists(startAll) {
				description = "multiple sandbox"
			}
			if common.FileExists(initializeSlaves) {
				description = "master-slave replication"
			}
			if common.FileExists(initializeNodes) {
				description = "group replication"
			}
			if common.FileExists(start) || common.FileExists(startAll) {
				dirs = append(dirs, fmt.Sprintf("%-20s : *%s* %s ", fname, description, locked))
			}
		}
	}
	if useHeader {
		//           1         2         3         4         5         6         7
		//  12345678901234567890123456789012345678901234567890123456789012345678901234567890
		//	fan_in_msb_5_7_21         : fan-in                   5.7.21 [14001 14002 14003]
		template := "%-25s   %-23s %-8s %s\n"
		fmt.Printf(template, "name", "type", "version", "ports")
		fmt.Printf(template, "----------------", "-------", "-------", "-----")
	}
	for _, dir := range dirs {
		fmt.Println(dir)
	}
}

var sandboxesCmd = &cobra.Command{
	Use:   "sandboxes",
	Short: "List installed sandboxes",
	Long: `Lists all sandboxes installed in $SANDBOX_HOME.
If sandboxes are installed in a different location, use --sandbox-home to 
indicate where to look.
Alternatively, using --catalog will list all sandboxes, regardless of where 
they were deployed.
`,
	Aliases: []string{"installed", "deployed"},
	Run:     ShowSandboxes,
}

func init() {
	rootCmd.AddCommand(sandboxesCmd)

	sandboxesCmd.Flags().BoolP(defaults.CatalogLabel, "", false, "Use sandboxes catalog instead of scanning directory")
	sandboxesCmd.Flags().BoolP(defaults.HeaderLabel, "", false, "Shows header with catalog output")
}
