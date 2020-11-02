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

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/data_load"
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
	dbName := args[0]
	archive, found := data_load.Archives[dbName]
	if found {
		fmt.Printf("Name:          %s\n", dbName)
		fmt.Printf("Description:   %s\n", archive.Description)
		fmt.Printf("URL:           %s\n", archive.Origin)
		fmt.Printf("File name:     %s\n", archive.FileName)
		fmt.Printf("Size:          %s\n", humanize.Bytes(archive.Size))
		fmt.Printf("internal dir   %s\n", archive.InternalDirectory)

	} else {
		return fmt.Errorf("archive %s not found", dbName)
	}
	return nil
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
)

func init() {
	rootCmd.AddCommand(dataLoadCmd)
	dataLoadCmd.AddCommand(dataLoadListCmd)
	dataLoadCmd.AddCommand(dataLoadShowCmd)
	dataLoadCmd.AddCommand(dataLoadGetCmd)

	dataLoadListCmd.Flags().BoolP(globals.FullInfoLabel, "", false, "Shows all archive details")
	dataLoadGetCmd.Flags().BoolP(globals.OverwriteLabel, "", false, "overwrite previously downloaded archive")
}
