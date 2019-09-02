// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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
	"path"

	"github.com/alexeyco/simpletable"
	"github.com/dustin/go-humanize/english"
	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

func showSandboxesFromCatalog(currentSandboxHome string, useFlavor, useHeader, useTable bool) {
	var sandboxList defaults.SandboxCatalog
	var err error
	sandboxList, err = defaults.ReadCatalog()

	common.ErrCheckExitf(err, 1, "error getting sandboxes from catalog: %s", err)
	if len(sandboxList) == 0 {
		return
	}

	table := simpletable.New()

	if useHeader {
		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Text: "name"},
				{Align: simpletable.AlignCenter, Text: "version"},
				{Align: simpletable.AlignCenter, Text: "type"},
				{Align: simpletable.AlignCenter, Text: "nodes"},
				{Align: simpletable.AlignCenter, Text: "ports"},
			},
		}
		if useFlavor {
			table.Header.Cells = append(table.Header.Cells,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: "flavor"},
			)
		}
	}
	for name, contents := range sandboxList {
		var cells []*simpletable.Cell
		ports := "["
		for _, p := range contents.Port {
			ports += fmt.Sprintf("%d ", p)
		}
		ports += "]"
		cells = append(cells, &simpletable.Cell{Text: common.BaseName(name)})
		cells = append(cells, &simpletable.Cell{Text: contents.Version})
		cells = append(cells, &simpletable.Cell{Text: contents.SBType})
		cells = append(cells, &simpletable.Cell{Text: fmt.Sprintf("%d", len(contents.Nodes))})
		cells = append(cells, &simpletable.Cell{Text: ports})

		if useFlavor {
			cells = append(cells, &simpletable.Cell{Text: contents.Flavor})
		}
		table.Body.Cells = append(table.Body.Cells, cells)
	}
	table.SetStyle(simpletable.StyleCompactLite)
	if useTable {
		table.SetStyle(simpletable.StyleRounded)
	}
	table.Println()
}

func getFullSandboxInfo(sandboxHome string) []common.SandboxInfo {
	var fullSandboxList []common.SandboxInfo
	simpleSandboxList, err := common.GetInstalledSandboxes(sandboxHome)
	if err != nil {
		return fullSandboxList
	}

	for _, sb := range simpleSandboxList {
		sbDescription := path.Join(sandboxHome, sb.SandboxName, globals.SandboxDescriptionName)
		var tempSandboxDesc common.SandboxDescription
		if common.FileExists(sbDescription) {
			tempSandboxDesc, _ = common.ReadSandboxDescription(path.Join(sandboxHome, sb.SandboxName))
		}
		// No description file was found
		// We try to get what we can. However, this should not happen unless users delete the description files
		slaveLabel := defaults.Defaults().SlavePrefix
		slavePlural := english.PluralWord(2, slaveLabel, "")
		initializeSlaves := "initialize_" + slavePlural
		if tempSandboxDesc.SBType == "" {
			tempSandboxDesc.SBType = globals.SbTypeSingle
			initSlaves := path.Join(sb.SandboxName, initializeSlaves)
			initNodes := path.Join(sb.SandboxName, globals.ScriptInitializeNodes)
			startAll := path.Join(sb.SandboxName, globals.ScriptStartAll)
			startAllExists := common.FileExists(startAll)
			initSlavesExists := common.FileExists(initSlaves)
			initNodesExists := common.FileExists(initNodes)
			if initSlavesExists {
				tempSandboxDesc.SBType = "master-slave"
			}
			if initNodesExists {
				tempSandboxDesc.SBType = "replication"
			}
			if startAllExists && !initNodesExists && !initSlavesExists {
				tempSandboxDesc.SBType = "multiple"
			}
			tempSandboxDesc.Version = "undetected"
			tempSandboxDesc.Flavor = "undetected"
			tempSandboxDesc.Port = []int{0}
		}
		fullSandboxList = append(fullSandboxList,
			common.SandboxInfo{SandboxName: sb.SandboxName, SandboxDesc: tempSandboxDesc, Locked: sb.Locked})
	}
	return fullSandboxList
}

// Shows installed sandboxes
func showSandboxes(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	SandboxHome, _ := flags.GetString(globals.SandboxHomeLabel)
	readCatalog, _ := flags.GetBool(globals.CatalogLabel)
	useHeader, _ := flags.GetBool(globals.HeaderLabel)
	useFlavor, _ := flags.GetBool(globals.FlavorLabel)
	useTable, _ := flags.GetBool(globals.TableLabel)
	useHost := false

	useFullInfo, _ := flags.GetBool(globals.FullInfoLabel)
	if useFullInfo {
		useHeader = true
		useFlavor = true
		useTable = true
		useHost = true
	}
	if readCatalog {
		showSandboxesFromCatalog(SandboxHome, useFlavor, useHeader, useTable)
		return
	}
	var sandboxList []common.SandboxInfo
	// If the sandbox directory hasn't been created yet, we start with an empty list
	if common.DirExists(SandboxHome) {
		sandboxList = getFullSandboxInfo(SandboxHome)
	}
	for _, sb := range sandboxList {
		if sb.SandboxDesc.Host != "" && sb.SandboxDesc.Host != globals.LocalHostIP {
			useHost = true
		}
	}
	if len(sandboxList) == 0 {
		return
	}
	table := simpletable.New()

	if useHeader {
		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Text: "name"},
				{Align: simpletable.AlignCenter, Text: "type"},
				{Align: simpletable.AlignCenter, Text: "version"},
			},
		}
		if useHost {
			table.Header.Cells = append(table.Header.Cells,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: "host"},
			)
		}
		table.Header.Cells = append(table.Header.Cells,
			&simpletable.Cell{Align: simpletable.AlignCenter, Text: "port"},
		)
		if useFlavor {
			table.Header.Cells = append(table.Header.Cells,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: "flavor"},
			)
		}
		if useTable {
			table.Header.Cells = append(table.Header.Cells,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: "nodes"},
			)
			table.Header.Cells = append(table.Header.Cells,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: "locked"},
			)
		}
	}
	for _, sb := range sandboxList {
		var cells []*simpletable.Cell
		sbName := common.BaseName(sb.SandboxName)
		if !useTable {
			numSpaces := 25 - len(sbName)
			if numSpaces > 0 {
				for N := 0; N < numSpaces; N++ {
					sbName += " "
				}
			}
			sbName += ": "
		}
		cells = append(cells, &simpletable.Cell{Text: sbName})
		cells = append(cells, &simpletable.Cell{Text: sb.SandboxDesc.SBType})
		cells = append(cells, &simpletable.Cell{Text: sb.SandboxDesc.Version})
		isLocked := ""
		if sb.Locked {
			isLocked = "(LOCKED)"
		}
		host := sb.SandboxDesc.Host
		if host == "" {
			host = globals.LocalHostIP
		}
		ports := "["
		for _, p := range sb.SandboxDesc.Port {
			ports += fmt.Sprintf("%d ", p)
		}
		ports += "]"
		if sb.Locked && !useTable {
			ports += " " + isLocked
		}
		if useHost {
			cells = append(cells, &simpletable.Cell{Text: host})
		}
		cells = append(cells, &simpletable.Cell{Text: ports})
		if useFlavor {
			cells = append(cells, &simpletable.Cell{Text: sb.SandboxDesc.Flavor})
		}
		if useTable {
			cells = append(cells, &simpletable.Cell{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", sb.SandboxDesc.Nodes)})
			cells = append(cells, &simpletable.Cell{Text: isLocked})
		}
		table.Body.Cells = append(table.Body.Cells, cells)
	}
	table.SetStyle(simpletable.StyleCompactLite)
	if useTable {
		table.SetStyle(simpletable.StyleRounded)
	}
	table.Println()
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
	Run:     showSandboxes,
}

func init() {
	rootCmd.AddCommand(sandboxesCmd)

	sandboxesCmd.Flags().BoolP(globals.CatalogLabel, "", false, "Use sandboxes catalog instead of scanning directory")
	sandboxesCmd.Flags().BoolP(globals.HeaderLabel, "", false, "Shows header with catalog output")
	sandboxesCmd.Flags().BoolP(globals.TableLabel, "", false, "Shows sandbox list as a table")
	sandboxesCmd.Flags().BoolP(globals.FlavorLabel, "", false, "Shows flavor in sandbox list")
	sandboxesCmd.Flags().BoolP(globals.FullInfoLabel, "", false, "Shows all info in table format")
}
