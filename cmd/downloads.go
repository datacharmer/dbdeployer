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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/alexeyco/simpletable"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/downloads"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
)

func getOSWarning(tarball downloads.TarballDescription) string {
	currentOS := strings.ToLower(runtime.GOOS)
	tarballOS := strings.ToLower(tarball.OperatingSystem)
	if currentOS != tarballOS {
		return fmt.Sprintf("WARNING: Current OS is %s, but the tarball's OS is %s", currentOS, tarballOS)
	}
	return ""
}

func listRemoteTarballs(cmd *cobra.Command, args []string) {

	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	OS, _ := cmd.Flags().GetString(globals.OSLabel)
	OS = strings.ToLower(OS)
	flavor = strings.ToLower(flavor)
	showUrl, _ := cmd.Flags().GetBool(globals.ShowUrlLabel)
	if OS == "" {
		OS = strings.ToLower(runtime.GOOS)
	}
	if OS == "macos" || OS == "osx" {
		OS = "darwin"
	}
	table := simpletable.New()

	headerName := "name"
	if showUrl {
		headerName = "URL"
	}
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: headerName},
			{Align: simpletable.AlignCenter, Text: "OS"},
			{Align: simpletable.AlignRight, Text: "version"},
			{Align: simpletable.AlignCenter, Text: "flavor"},
			{Align: simpletable.AlignRight, Text: "size"},
			{Align: simpletable.AlignCenter, Text: "minimal"},
		},
	}

	notes := ""
	if downloads.TarballRegistryFileExist() {
		notes = fmt.Sprintf("[loaded from %s]", downloads.TarballFileRegistry)
	}
	fmt.Printf("Available tarballs %s\n", notes)
	for _, tb := range downloads.DefaultTarballRegistry.Tarballs {
		var cells []*simpletable.Cell
		minimalTag := ""
		if tb.Minimal {
			minimalTag = "Y"
		}
		if showUrl {
			cells = append(cells, &simpletable.Cell{Text: tb.Url})
		} else {
			cells = append(cells, &simpletable.Cell{Text: tb.Name})
		}
		cells = append(cells, &simpletable.Cell{Text: tb.OperatingSystem})
		cells = append(cells, &simpletable.Cell{Align: simpletable.AlignRight, Text: tb.Version})
		cells = append(cells, &simpletable.Cell{Text: tb.Flavor})
		cells = append(cells, &simpletable.Cell{Align: simpletable.AlignRight, Text: humanize.Bytes(uint64(tb.Size))})
		cells = append(cells, &simpletable.Cell{Text: minimalTag})
		if flavor == strings.ToLower(tb.Flavor) || flavor == "" || flavor == "all" {
			if OS == strings.ToLower(tb.OperatingSystem) || OS == "" || OS == "all" {
				table.Body.Cells = append(table.Body.Cells, cells)
			}
		}
	}
	table.SetStyle(simpletable.StyleCompactLite)
	table.Println()

}

var downloadedTarball string

func getRemoteTarball(cmd *cobra.Command, args []string) {

	if len(args) < 1 {
		common.Exit(1, "command 'get' requires a remote tarball name")
	}
	quiet, _ := cmd.Flags().GetBool(globals.QuietLabel)
	progressStep, _ := cmd.Flags().GetInt64(globals.ProgressStepLabel)
	dryRun, _ := cmd.Flags().GetBool(globals.DryRunLabel)

	var tarball downloads.TarballDescription
	var err error
	var fileName string

	tarball, err = downloads.FindTarballByName(args[0])
	if err != nil {
		common.Exitf(1, "%s", err)
	}

	fileName = tarball.Name
	absPath, err := common.AbsolutePath(fileName)
	if err != nil {
		common.Exitf(1, "%s", err)
	}
	if dryRun {
		displayTarball(tarball)
		return
	}
	if common.FileExists(absPath) {
		common.Exitf(1, globals.ErrFileAlreadyExists, absPath)
	}
	if !quiet {
		fmt.Printf("Downloading %s\n", tarball.Name)
	}
	err = rest.DownloadFile(absPath, tarball.Url, !quiet, progressStep)
	common.ErrCheckExitf(err, 1, "error getting remote file %s - %s", fileName, err)
	postDownloadOps(tarball, fileName, absPath)
	downloadedTarball = absPath
}

func displayTarball(tarball downloads.TarballDescription) {
	fmt.Printf("Name:          %s\n", tarball.Name)
	fmt.Printf("Short version: %s\n", tarball.ShortVersion)
	fmt.Printf("Version:       %s\n", tarball.Version)
	fmt.Printf("Flavor:        %s\n", tarball.Flavor)
	fmt.Printf("OS:            %s\n", tarball.OperatingSystem)
	fmt.Printf("URL:           %s\n", tarball.Url)
	fmt.Printf("Checksum:      %s\n", tarball.Checksum)
	fmt.Printf("Size:          %s\n", humanize.Bytes(uint64(tarball.Size)))
	if tarball.Notes != "" {
		fmt.Printf("Notes:         %s\n", tarball.Notes)
	}
}

func showRemoteTarball(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, "command 'show' requires a remote tarball name")
	}

	tarballName := args[0]
	tarball, err := downloads.FindTarballByName(tarballName)
	if err != nil {
		common.Exitf(1, "%s", err)
	}
	displayTarball(tarball)
}

func getRemoteTarballByVersion(cmd *cobra.Command, args []string) {

	if len(args) < 1 {
		common.Exit(1, "command 'get-by-version' requires at least a version")
	}
	quiet, _ := cmd.Flags().GetBool(globals.QuietLabel)
	minimal, _ := cmd.Flags().GetBool(globals.MinimalLabel)
	newest, _ := cmd.Flags().GetBool(globals.NewestLabel)
	dryRun, _ := cmd.Flags().GetBool(globals.DryRunLabel)
	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	OS, _ := cmd.Flags().GetString(globals.OSLabel)
	progressStep, _ := cmd.Flags().GetInt64(globals.ProgressStepLabel)

	if OS == "" {
		OS = runtime.GOOS
	}
	if flavor == "" {
		flavor = common.MySQLFlavor
	}
	var tarball downloads.TarballDescription
	var err error
	var fileName string
	version := args[0]

	tarball, err = downloads.FindTarballByVersionFlavorOS(version, flavor, OS, minimal, newest)
	if err != nil {
		common.Exitf(1, "%s", err)
	}

	fileName = tarball.Name
	absPath, err := common.AbsolutePath(fileName)
	if err != nil {
		common.Exitf(1, "%s", err)
	}
	if dryRun {
		fmt.Println("Would download:")
		fmt.Println("")
		displayTarball(tarball)
		return
	}
	if common.FileExists(absPath) {
		common.Exitf(1, globals.ErrFileAlreadyExists, absPath)
	}
	if !quiet {
		fmt.Printf("Downloading %s\n", tarball.Name)
	}
	err = rest.DownloadFile(absPath, tarball.Url, !quiet, progressStep)
	common.ErrCheckExitf(err, 1, "error getting remote file %s - %s", fileName, err)
	postDownloadOps(tarball, fileName, absPath)
}

func getUnpackRemoteTarball(cmd *cobra.Command, args []string) {
	deleteAfterUnpack, _ := cmd.Flags().GetBool(globals.DeleteAfterUnpackLabel)
	getRemoteTarball(cmd, args)

	unpackTarball(cmd, args)
	if deleteAfterUnpack {
		if downloadedTarball == "" {
			common.Exitf(1, "unhandled error. After unpack, the tarball to be deleted was not found")
		}
		err := os.Remove(downloadedTarball)
		common.ErrCheckExitf(err, 1, "error removing downloaded file %s - %s", downloadedTarball, err)
	}
}

func postDownloadOps(tarball downloads.TarballDescription, fileName, absPath string) {
	fmt.Printf("File %s downloaded\n", absPath)

	if tarball.Checksum == "" {
		fmt.Println("No checksum to compare")
	} else {
		err := downloads.CompareTarballChecksum(tarball, absPath)
		common.ErrCheckExitf(err, 1, "error comparing checksum for tarball %s - %s", fileName, err)
		fmt.Println("Checksum matches")
	}
	warningMsg := getOSWarning(tarball)
	if warningMsg != "" {
		fmt.Println(globals.HashLine)
		fmt.Println(warningMsg)
		fmt.Println(globals.HashLine)
	}
}

func exportTarballList(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, "command 'export' requires a file name")
	}
	addEmpty, _ := cmd.Flags().GetBool(globals.AddEmptyItemLabel)
	fileName := args[0]
	if path.Ext(fileName) == "" {
		fileName += ".json"
	}

	if common.FileExists(fileName) {
		common.Exitf(1, globals.ErrFileAlreadyExists, fileName)
	}

	tarballCollection := downloads.DefaultTarballRegistry
	if addEmpty {
		tarballCollection.Tarballs = append(tarballCollection.Tarballs, downloads.TarballDescription{
			Name:      "FillIt",
			Notes:     "Fill it",
			UpdatedBy: "Fill it",
		})
	}

	b, err := json.MarshalIndent(tarballCollection, " ", "\t")
	if err != nil {
		common.Exitf(1, "error encoding tarball list")
	}
	jsonString := fmt.Sprintf("%s", b)

	err = common.WriteString(jsonString, fileName)
	if err != nil {
		common.Exitf(1, "error writing tarball list to JSON file")
	}
	fmt.Printf("Tarball list exported to %s\n", fileName)
}

func resetTarballCollection(cmd *cobra.Command, args []string) {
	if downloads.TarballRegistryFileExist() {
		_, err := common.RunCmdWithArgs("rm", []string{"-f", downloads.TarballFileRegistry})
		if err != nil {
			common.Exitf(1, "error deleting %s: %s", downloads.TarballFileRegistry, err)
		}
		fmt.Printf("File %s removed\n", downloads.TarballFileRegistry)
	} else {
		fmt.Printf("No tarballs file found in %s\n", defaults.ConfigurationDir)
	}
}

func importTarballCollection(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, "command 'import' requires a file name")
	}
	fileName := args[0]

	var tarballCollection downloads.TarballCollection

	jsonText, err := common.SlurpAsBytes(fileName)
	if err != nil {
		common.Exitf(1, "error reading tarball list from %s", fileName)
	}

	err = json.Unmarshal(jsonText, &tarballCollection)
	if err != nil {
		common.Exitf(1, "error decoding tarball list from %s", fileName)
	}

	err = downloads.TarballFileInfoValidation(tarballCollection)
	if err != nil {
		common.Exitf(1, "error validating tarball list from %s\n %s", fileName, err)
	}
	err = downloads.WriteTarballFileInfo(tarballCollection)
	if err != nil {
		common.Exitf(1, "error writing tarball list: %s", err)
	}
	fmt.Printf("Tarball list imported from %s to %s\n", fileName, downloads.TarballFileRegistry)
}

var downloadsListCmd = &cobra.Command{
	Use:     "list [options]",
	Aliases: []string{"index"},
	Short:   "list remote tarballs",
	Long: `List remote tarballs.
By default it includes tarballs for current operating system.
Use '--OS=os_name' or '--OS=all' to change.
`,
	Run: listRemoteTarballs,
}

var downloadsGetCmd = &cobra.Command{
	Use:   "get tarball_name [options]",
	Short: "Downloads a remote tarball",
	Long:  ``,
	Run:   getRemoteTarball,
}

var downloadsGetUnpackCmd = &cobra.Command{
	Use:   "get-unpack tarball_name [options]",
	Short: "Downloads and unpacks a remote tarball",
	Long: `get-unpack downloads a tarball and then unpacks it, using the same
options available for "dbdeployer unpack".`,
	Run: getUnpackRemoteTarball,
}

var downloadsShowCmd = &cobra.Command{
	Use:     "show tarball_name",
	Aliases: []string{"display"},
	Short:   "Downloads a remote tarball",
	Long:    ``,
	Run:     showRemoteTarball,
}

var downloadsGetByVersionCmd = &cobra.Command{
	Use:   "get-by-version version [options]",
	Short: "Downloads a remote tarball",
	Long: `
Download a tarball identified by a combination of
version, flavor, operating system, and optionally its minimal state.
If you don't specify the Operating system, the current one will be assumed.
If the flavor is not specified, 'mysql' is assumed.
Use the option '--dry-run' to see what dbdeployer would download.
`,
	Example: `
$ dbdeployer downloads get-by-version 5.7 --newest --dry-run
$ dbdeployer downloads get-by-version 5.7 --newest --minimal --dry-run --OS=linux
$ dbdeployer downloads get-by-version 5.7 --newest
$ dbdeployer downloads get-by-version 8.0 --flavor=ndb
$ dbdeployer downloads get-by-version 5.7.26 --minimal
$ dbdeployer downloads get-by-version 5.7 --minimal
`,
	Run: getRemoteTarballByVersion,
}

var downloadsExportCmd = &cobra.Command{
	Use:   "export file-name [options]",
	Short: "Exports the list of tarballs to a file",
	Long:  ``,

	Run: exportTarballList,
}

var downloadsImportCmd = &cobra.Command{
	Use:   "import file-name [options]",
	Short: "Imports the list of tarballs from a file",
	Long:  ``,

	Run: importTarballCollection,
}

var downloadsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the custom list of tarballs and resume the defaults",
	Long:  ``,

	Run: resetTarballCollection,
}
var downloadsCmd = &cobra.Command{
	Use:   "downloads",
	Short: "Manages remote tarballs",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(downloadsCmd)
	downloadsCmd.AddCommand(downloadsListCmd)
	downloadsCmd.AddCommand(downloadsGetCmd)
	downloadsCmd.AddCommand(downloadsShowCmd)
	downloadsCmd.AddCommand(downloadsExportCmd)
	downloadsCmd.AddCommand(downloadsImportCmd)
	downloadsCmd.AddCommand(downloadsResetCmd)
	downloadsCmd.AddCommand(downloadsGetByVersionCmd)
	downloadsCmd.AddCommand(downloadsGetUnpackCmd)

	downloadsListCmd.Flags().BoolP(globals.ShowUrlLabel, "", false, "Show the URL")
	downloadsListCmd.Flags().String(globals.FlavorLabel, "", "Which flavor will be listed")
	downloadsListCmd.Flags().String(globals.OSLabel, "", "Which OS will be listed")

	downloadsGetByVersionCmd.Flags().BoolP(globals.NewestLabel, "", false, "Choose only the newest tarballs not yet downloaded")
	downloadsGetByVersionCmd.Flags().BoolP(globals.DryRunLabel, "", false, "Show what would be downloaded, but don't run it")
	downloadsGetByVersionCmd.Flags().BoolP(globals.MinimalLabel, "", false, "Choose only minimal tarballs")
	downloadsGetByVersionCmd.Flags().String(globals.FlavorLabel, "", "Choose only the given flavor")
	downloadsGetByVersionCmd.Flags().String(globals.OSLabel, "", "Choose only the given OS")
	downloadsGetByVersionCmd.Flags().BoolP(globals.QuietLabel, "", false, "Do not show download progress")
	downloadsGetByVersionCmd.Flags().Int64P(globals.ProgressStepLabel, "", globals.ProgressStepValue, "Progress interval")

	downloadsGetCmd.Flags().BoolP(globals.QuietLabel, "", false, "Do not show download progress")
	downloadsGetCmd.Flags().Int64P(globals.ProgressStepLabel, "", globals.ProgressStepValue, "Progress interval")
	downloadsGetCmd.Flags().BoolP(globals.DryRunLabel, "", false, "Show what would be downloaded, but don't run it")

	downloadsGetUnpackCmd.Flags().BoolP(globals.DeleteAfterUnpackLabel, "", false, "Delete the tarball after successful unpack")

	// downloadsGetUnpack needs the same flags that cmdUnpack has
	downloadsGetUnpackCmd.Flags().Int64P(globals.ProgressStepLabel, "", globals.ProgressStepValue, "Progress interval")
	downloadsGetUnpackCmd.PersistentFlags().Int(globals.VerbosityLabel, 1, "Level of verbosity during unpack (0=none, 2=maximum)")
	downloadsGetUnpackCmd.PersistentFlags().String(globals.UnpackVersionLabel, "", "which version is contained in the tarball")
	downloadsGetUnpackCmd.PersistentFlags().String(globals.PrefixLabel, "", "Prefix for the final expanded directory")
	downloadsGetUnpackCmd.PersistentFlags().Bool(globals.ShellLabel, false, "Unpack a shell tarball into the corresponding server directory")
	downloadsGetUnpackCmd.PersistentFlags().Bool(globals.OverwriteLabel, false, "Overwrite the destination directory if already exists")
	downloadsGetUnpackCmd.PersistentFlags().String(globals.TargetServerLabel, "", "Uses a different server to unpack a shell tarball")
	downloadsGetUnpackCmd.PersistentFlags().String(globals.FlavorLabel, "", "Defines the tarball flavor (MySQL, NDB, Percona Server, etc)")

	downloadsExportCmd.Flags().BoolP(globals.AddEmptyItemLabel, "", false, "Add an empty item to the tarballs list")
}
