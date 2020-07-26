// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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
	"sort"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/araddon/dateparse"
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

// Shows installed sandboxes
func showSandboxes(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	SandboxHome, _ := flags.GetString(globals.SandboxHomeLabel)
	readCatalog, _ := flags.GetBool(globals.CatalogLabel)
	useHeader, _ := flags.GetBool(globals.HeaderLabel)
	useFlavor, _ := flags.GetBool(globals.FlavorLabel)
	useTable, _ := flags.GetBool(globals.TableLabel)
	byDate, _ := flags.GetBool(globals.ByDateLabel)
	byVersion, _ := flags.GetBool(globals.ByVersionLabel)
	byFlavor, _ := flags.GetBool(globals.ByFlavorLabel)
	latest, _ := flags.GetBool(globals.LatestLabel)
	oldest, _ := flags.GetBool(globals.OldestLabel)
	useHost := false

	if oldest && latest {
		common.Exitf(1, "only one of '--%s' and '--%s' can be used", globals.OldestLabel, globals.LatestLabel)
	}
	if oldest || latest {
		byFlavor = false
		byVersion = false
	}
	if byVersion && byFlavor || byVersion && byDate || byDate && byFlavor {
		common.Exitf(1, "only one of '--%s', '--%s', and '--%s' can be used", globals.ByDateLabel, globals.ByFlavorLabel, globals.ByVersionLabel)
	}
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
	var sandboxList common.SandboxInfoList
	// If the sandbox directory hasn't been created yet, we start with an empty list
	if common.DirExists(SandboxHome) {
		if byDate || latest || oldest {
			var err error
			sandboxList, err = common.GetSandboxesByDate(SandboxHome)
			if err != nil {
				common.Exitf(1, "error during sandbox sorting by date: %s", err)
			}
		} else {
			sandboxList = common.GetFullSandboxInfo(SandboxHome)
		}
	}
	for _, sb := range sandboxList {
		if sb.SandboxDesc.Host != "" && sb.SandboxDesc.Host != globals.LocalHostIP {
			useHost = true
		}
	}
	if len(sandboxList) == 0 {
		return
	}

	if byVersion {
		sort.SliceStable(sandboxList, func(i, j int) bool {
			iVersionList, _ := common.VersionToList(sandboxList[i].SandboxDesc.Version)
			jVersionList, _ := common.VersionToList(sandboxList[j].SandboxDesc.Version)
			greater, _ := common.GreaterOrEqualVersionList(iVersionList, jVersionList)
			return greater
		})
	}
	if byFlavor {
		sort.SliceStable(sandboxList, func(i, j int) bool {
			return sandboxList[i].SandboxDesc.Flavor < sandboxList[j].SandboxDesc.Flavor
		})
	}
	if oldest {
		sandboxList = common.SandboxInfoList{sandboxList[0]}
	}
	if latest {
		sandboxList = common.SandboxInfoList{sandboxList[len(sandboxList)-1]}
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
		if byDate || latest || oldest || useFullInfo {
			table.Header.Cells = append(table.Header.Cells,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: "created"})
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
		if byDate || latest || oldest || useFullInfo {
			timestamp, _ := dateparse.ParseStrict(sb.SandboxDesc.Timestamp)
			//fmt.Printf("%s", timestamp.Format(time.RFC3339))
			//cells = append(cells, &simpletable.Cell{Text: sb.SandboxDesc.Timestamp})
			cells = append(cells, &simpletable.Cell{Text: timestamp.Format(time.RFC3339)})
			//cells = append(cells, &simpletable.Cell{Text: timestamp.Format("2006-01-02 15:04:05Z -07:00")})
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
	sandboxesCmd.Flags().BoolP(globals.ByDateLabel, "", false, "Show sandboxes in order of creation")
	sandboxesCmd.Flags().BoolP(globals.ByFlavorLabel, "", false, "Show sandboxes sorted by flavor")
	sandboxesCmd.Flags().BoolP(globals.ByVersionLabel, "", false, "Show sandboxes sorted by version")
	sandboxesCmd.Flags().BoolP(globals.LatestLabel, "", false, "Show only latest sandbox")
	sandboxesCmd.Flags().BoolP(globals.OldestLabel, "", false, "Show only oldest sandbox")
}
