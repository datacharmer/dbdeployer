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
	"github.com/datacharmer/dbdeployer/globals"
	"path"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/spf13/cobra"
	"strings"
)

func ShowSandboxesFromCatalog(currentSandboxHome string, header bool) {
	sandboxList, err := defaults.ReadCatalog()
	common.ErrCheckExitf(err, 1, "error getting sandboxes from catalog: %s", err)
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
	SandboxHome, _ := flags.GetString(globals.SandboxHomeLabel)
	readCatalog, _ := flags.GetBool(globals.CatalogLabel)
	useHeader, _ := flags.GetBool(globals.HeaderLabel)
	if readCatalog {
		ShowSandboxesFromCatalog(SandboxHome, useHeader)
		return
	}
	var sandboxList []common.SandboxInfo
	var err error
	// If the sandbox directory hasn't been created yet, we start with an empty list
	if common.DirExists(SandboxHome) {
		sandboxList, err = common.GetInstalledSandboxes(SandboxHome)
		common.ErrCheckExitf(err, 1, globals.ErrRetrievingSandboxList, err)
	}
	var dirs []string
	for _, sandboxInfo := range sandboxList {
		fileName := sandboxInfo.SandboxName
		description := "single"
		sbDesc := path.Join(SandboxHome, fileName, globals.SandboxDescriptionName)
		if common.FileExists(sbDesc) {
			sbd, err := common.ReadSandboxDescription(path.Join(SandboxHome, fileName))
			common.ErrCheckExitf(err, 1, "error reading sandbox description from %s", fileName)
			locked := ""
			if sandboxInfo.Locked {
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
				var nodeDescriptions []common.SandboxDescription
				innerSandboxList, err := common.GetInstalledSandboxes(path.Join(SandboxHome, fileName))
				common.ErrCheckExitf(err, 1, globals.ErrRetrievingSandboxList, err)
				innerFiles := common.SandboxInfoToFileNames(innerSandboxList)
				for _, inner := range innerFiles {
					innerSbDesc := path.Join(SandboxHome, fileName, inner, globals.SandboxDescriptionName)
					if common.FileExists(innerSbDesc) {
						sbNode, err := common.ReadSandboxDescription(fmt.Sprintf("%s/%s/%s", SandboxHome, fileName, inner))
						common.ErrCheckExitf(err, 1, "error reading sandbox description from %s/%s/%s", SandboxHome, fileName, inner)
						nodeDescriptions = append(nodeDescriptions, sbNode)
					}
				}
				ports := ""
				for _, nd := range nodeDescriptions {
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
			dirs = append(dirs, fmt.Sprintf("%-25s : %s %s", fileName, description, locked))
		} else {
			locked := ""
			noClear := path.Join(SandboxHome, fileName, globals.ScriptNoClear)
			noClearAll := path.Join(SandboxHome, fileName, globals.ScriptNoClearAll)
			start := path.Join(SandboxHome, fileName, globals.ScriptStart)
			startAll := path.Join(SandboxHome, fileName, globals.ScriptStartAll)
			initializeSlaves := path.Join(SandboxHome, fileName, globals.ScriptInitializeSlaves)
			initializeNodes := path.Join(SandboxHome, fileName, globals.ScriptInitializeNodes)
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
				dirs = append(dirs, fmt.Sprintf("%-20s : *%s* %s ", fileName, description, locked))
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

	sandboxesCmd.Flags().BoolP(globals.CatalogLabel, "", false, "Use sandboxes catalog instead of scanning directory")
	sandboxesCmd.Flags().BoolP(globals.HeaderLabel, "", false, "Shows header with catalog output")
}
