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
	"encoding/json"
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/data_load"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

func listArchives(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	full, _ := flags.GetBool(globals.FullInfoLabel)
	if full {
		result, err := data_load.ArchivesAsJson()
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", result)
	} else {
		data_load.ListArchives()
	}
	return nil
}

func loadArchive(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("command 'get' requires a database name and a destination sandbox")
	}
	overwrite, _ := cmd.Flags().GetBool(globals.OverwriteLabel)
	return data_load.LoadArchive(args[0], args[1], overwrite)
}

func showArchive(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("command 'show' requires a database name")
	}
	flags := cmd.Flags()
	fullInfo, _ := flags.GetBool(globals.FullInfoLabel)
	dbName := args[0]
	archives, origin := data_load.Archives()
	archive, found := archives[dbName]
	if found {
		if fullInfo {
			result, err := json.MarshalIndent(archive, " ", " ")
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", data_load.UnescapeJsonString(result))
		} else {
			if origin != "defaults" {
				fmt.Printf("Data load info from: %s\n", origin)
			}
			fmt.Printf("Name:          %s\n", dbName)
			fmt.Printf("Description:   %s\n", archive.Description)
			fmt.Printf("URL:           %s\n", archive.Origin)
			fmt.Printf("File name:     %s\n", archive.FileName)
			fmt.Printf("Size:          %s\n", humanize.Bytes(archive.Size))
			fmt.Printf("internal dir   %s\n", archive.InternalDirectory)
			fmt.Printf("Loading commands:\n")
			delta := 1
			if archive.ChangeDirectory {
				fmt.Printf("\t%2d cd %s\n", delta, dbName)
				delta = 2
			}
			for i, line := range archive.LoadCommands {
				fmt.Printf("\t%2d %s\n", i+delta, line)
			}
		}

	} else {
		return fmt.Errorf("archive %s not found", dbName)
	}
	return nil
}

func exportArchives(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("command 'export' requires a file name")
	}
	destFile := args[0]
	if common.FileExists(destFile) {
		return fmt.Errorf(globals.ErrFileAlreadyExists, destFile)
	}
	archives, err := data_load.ArchivesAsJson()
	if err != nil {
		return err
	}

	err = common.WriteString(archives, destFile)
	if err != nil {
		return err
	}
	fmt.Printf("data load info saved to %s\n", destFile)
	return nil
}

func importArchives(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("command 'import' requires a file name")
	}
	importFile := args[0]
	if !common.FileExists(importFile) {
		return fmt.Errorf(globals.ErrFileNotFound, importFile)
	}
	jsonText, err := common.SlurpAsBytes(importFile)
	if err != nil {
		return err
	}
	_, err = json.Marshal(jsonText)
	if err != nil {
		return err
	}

	err = common.WriteString(string(jsonText), defaults.ArchivesFile)
	if err != nil {
		return err
	}
	fmt.Printf("data load info recorded at %s\n", defaults.ArchivesFile)
	return nil
}

func resetArchives(cmd *cobra.Command, args []string) error {
	if common.FileExists(defaults.ArchivesFile) {
		err := os.Remove(defaults.ArchivesFile)
		if err != nil {
			return fmt.Errorf("error removing %s: %s", defaults.ArchivesFile, err)
		}
		fmt.Printf("file %s removed\n", defaults.ArchivesFile)
		return nil
	}
	return fmt.Errorf("no archives file found in configuration directory")
}

var (
	dataLoadCmd = &cobra.Command{
		Use:     "data-load",
		Short:   "tasks related to dbdeployer data loading",
		Aliases: []string{"load-data"},
		Long:    `Runs commands related to the database loading`,
	}
	dataLoadListCmd = &cobra.Command{
		Use:   "list",
		Short: "list databases available for loading",
		Long:  `List databases available for loading`,
		RunE:  listArchives,
	}
	dataLoadShowCmd = &cobra.Command{
		Use:   "show archive-name",
		Short: "show details of a database",
		Long:  "show details of a database",
		RunE:  showArchive,
	}
	dataLoadGetCmd = &cobra.Command{
		Use:   "get archive-name sandbox-name",
		Short: "Loads an archived database into a sandbox",
		Long:  "Loads an archived database into a sandbox",
		RunE:  loadArchive,
	}
	dataLoadExportCmd = &cobra.Command{
		Use:   "export file-name",
		Short: "Saves the archives details into a file",
		Long:  "Saves the archives details into a file",
		RunE:  exportArchives,
	}
	dataLoadImportCmd = &cobra.Command{
		Use:   "import file-name",
		Short: "Imports the archives details from a file",
		Long: `
Imports modified archives from a JSON file.
In the archive specification, the strings "$use" and "$my"
will be expanded to the relative scripts in the target sandbox directory.
`,
		RunE: importArchives,
	}
	dataLoadResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Resets the archives to their default values",
		Long:  "Resets the archives to their default values",
		RunE:  resetArchives,
	}
)

func init() {
	rootCmd.AddCommand(dataLoadCmd)
	dataLoadCmd.AddCommand(dataLoadListCmd)
	dataLoadCmd.AddCommand(dataLoadShowCmd)
	dataLoadCmd.AddCommand(dataLoadGetCmd)
	dataLoadCmd.AddCommand(dataLoadExportCmd)
	dataLoadCmd.AddCommand(dataLoadImportCmd)
	dataLoadCmd.AddCommand(dataLoadResetCmd)

	dataLoadListCmd.Flags().BoolP(globals.FullInfoLabel, "", false, "Shows all archive details")
	dataLoadShowCmd.Flags().BoolP(globals.FullInfoLabel, "", false, "Shows all archive details")
	dataLoadGetCmd.Flags().BoolP(globals.OverwriteLabel, "", false, "overwrite previously downloaded archive")
}
