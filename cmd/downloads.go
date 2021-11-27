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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/downloads"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

func getOSWarning(tarball downloads.TarballDescription) string {
	currentOS := strings.ToLower(runtime.GOOS)
	tarballOS := strings.ToLower(tarball.OperatingSystem)
	if currentOS != tarballOS {
		return fmt.Sprintf("WARNING: Current OS is %s, but the tarball's OS is %s", currentOS, tarballOS)
	}
	return ""
}

func treeRemoteTarballs(cmd *cobra.Command, args []string) {
	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	OS, _ := cmd.Flags().GetString(globals.OSLabel)
	version, _ := cmd.Flags().GetString(globals.VersionLabel)
	maxItemsPerVersion, _ := cmd.Flags().GetInt(globals.MaxItemsLabel)
	showUrl, _ := cmd.Flags().GetBool(globals.ShowUrlLabel)
	OS = strings.ToLower(OS)
	flavor = strings.ToLower(flavor)
	if OS == "" {
		OS = strings.ToLower(runtime.GOOS)
	}
	if OS == "macos" || OS == "osx" {
		OS = "darwin"
	}
	var tbList []downloads.TarballDescription
	for _, tb := range downloads.DefaultTarballRegistry.Tarballs {

		if !strings.EqualFold(OS, tb.OperatingSystem) {
			continue
		}
		if flavor != tb.Flavor {
			continue
		}
		if !strings.HasPrefix(tb.Version, version) {
			continue
		}
		tbList = append(tbList, tb)
	}

	tree := downloads.TarballTree(tbList)

	var index []string
	for v := range tree {
		index = append(index, v)
	}
	sort.Strings(index)

	table := simpletable.New()

	headerName := "name"
	if showUrl {
		headerName = "URL"
	}
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Vers"},
			{Align: simpletable.AlignCenter, Text: headerName},
			{Align: simpletable.AlignRight, Text: "version"},
			{Align: simpletable.AlignRight, Text: "size"},
			{Align: simpletable.AlignCenter, Text: "minimal"},
		},
	}
	var emptyCells []*simpletable.Cell
	for i := 0; i < 5; i++ {
		emptyCells = append(emptyCells, &simpletable.Cell{Text: ""})
	}
	for _, v := range index {
		list := tree[v]

		shown := false
		list = downloads.SortedTarballList(list, "version")
		size := len(list)

		maxItems := maxItemsPerVersion
		if maxItems == 0 {
			maxItems = len(list)
		}
		for i, tb := range list {
			remaining := size - i
			if remaining > maxItems {
				continue
			}
			minimalTag := ""
			if tb.Minimal {
				minimalTag = "Y"
			}
			var cells []*simpletable.Cell
			if !shown {
				cells = append(cells, &simpletable.Cell{Text: v})
				shown = true
			} else {
				cells = append(cells, &simpletable.Cell{Text: ""})
			}
			if showUrl {
				cells = append(cells, &simpletable.Cell{Text: tb.Url})
			} else {
				cells = append(cells, &simpletable.Cell{Text: tb.Name})
			}
			cells = append(cells, &simpletable.Cell{Align: simpletable.AlignRight, Text: tb.Version})
			cells = append(cells, &simpletable.Cell{Align: simpletable.AlignRight, Text: humanize.Bytes(uint64(tb.Size))})
			cells = append(cells, &simpletable.Cell{Text: minimalTag})
			table.Body.Cells = append(table.Body.Cells, cells)
		}
		table.Body.Cells = append(table.Body.Cells, emptyCells)
	}
	table.SetStyle(simpletable.StyleCompactLite)
	table.Println()
}

func listRemoteTarballs(cmd *cobra.Command, args []string) {

	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	OS, _ := cmd.Flags().GetString(globals.OSLabel)
	version, _ := cmd.Flags().GetString(globals.VersionLabel)
	sortBy, _ := cmd.Flags().GetString(globals.SortByLabel)
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
	fmt.Printf("Available tarballs %s (%s)\n", notes, downloads.DefaultTarballRegistry.UpdatedOn)
	tarballList := downloads.DefaultTarballRegistry.Tarballs
	tarballList = downloads.SortedTarballList(tarballList, sortBy)
	for _, tb := range tarballList {
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
		if version != "" && version != "all" && version != tb.Version && version != tb.ShortVersion {
			continue
		}
		if flavor != "" && flavor != "all" && flavor != strings.ToLower(tb.Flavor) {
			continue
		}
		if OS != "" && strings.ToLower(OS) != "all" && OS != strings.ToLower(tb.OperatingSystem) {
			continue
		}
		table.Body.Cells = append(table.Body.Cells, cells)
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
	wantedOs, _ := cmd.Flags().GetString(globals.OSLabel)

	var tarball downloads.TarballDescription
	var err error
	var fileName string

	wanted := args[0]

	if common.IsUrl(wanted) {
		storedTarball, err := downloads.FindTarballByUrl(wanted)
		if err == nil {
			tarball = storedTarball
		} else {
			name := filepath.Base(wanted)
			lowerName := strings.ToLower(name)
			flavor, version, shortVersion, err := common.FindTarballInfo(name)
			if err != nil {
				common.Exitf(1, "%s", err)
			}
			tarballOs := runtime.GOOS
			if wantedOs != "" {
				tarballOs = wantedOs
			} else {
				switch {
				case strings.Contains(lowerName, "linux"):
					tarballOs = "linux"
				case strings.Contains(lowerName, "macos") || strings.Contains(lowerName, "darwin"):
					tarballOs = "darwin"
				default:
					common.Exitf(1, "unable to determine the operating system of tarball '%s'", name)
				}
			}
			tarball = downloads.TarballDescription{
				Name:            name,
				Url:             wanted,
				Flavor:          flavor,
				ShortVersion:    shortVersion,
				Version:         version,
				OperatingSystem: tarballOs,
			}
		}
	} else {
		tarball, err = downloads.FindTarballByName(wanted)
		if err != nil {
			common.Exitf(1, "%s", err)
		}
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
	if tarball.DateAdded != "" {
		fmt.Printf("Added on:      %s\n", tarball.DateAdded)
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
	guessLatest, _ := cmd.Flags().GetBool(globals.GuessLatestLabel)
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

	tarball, err = downloads.FindOrGuessTarballByVersionFlavorOS(version, flavor, OS, minimal, newest, guessLatest)
	if err != nil {
		common.Exit(1, fmt.Sprintf("Error getting version %s (%s-%s)[minimal: %v - newest: %v - guess: %v]: %s",
			version, flavor, OS, minimal, newest, guessLatest, err),
			fmt.Sprintf("Try using 'dbdeployer downloads add-remote mysql %s %s'", version, OS))
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
	if err != nil {
		if tarball.Notes == "guessed" {
			common.Exitf(1, "Guessed %s file not ready for download : %s", tarball.Name, err)
		}
	}
	common.ErrCheckExitf(err, 1, "error getting remote file %s - %s", fileName, err)
	postDownloadOps(tarball, fileName, absPath)
}

func getUnpackRemoteTarball(cmd *cobra.Command, args []string) {
	deleteAfterUnpack, _ := cmd.Flags().GetBool(globals.DeleteAfterUnpackLabel)
	getRemoteTarball(cmd, args)

	// unpack flags are passed from current command and resolved in the called function
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
			Checksum:  "Fill it",
			Notes:     "Fill it",
			UpdatedBy: "Fill it",
		})
	}

	b, err := json.MarshalIndent(tarballCollection, " ", "\t")
	if err != nil {
		common.Exitf(1, "error encoding tarball list")
	}
	jsonString := string(b)

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

	if fileName == "remote-github" || fileName == "remote-tarballs" {
		fileName = defaults.Defaults().RemoteTarballUrl
	}

	if common.IsUrl(fileName) {
		fileUrl := fileName
		re := regexp.MustCompile(`^(http|https)://`)
		fileName = common.BaseName(re.ReplaceAllString(fileUrl, ""))
		if common.FileExists(fileName) {
			common.Exitf(1, "file %s already exists", fileName)
		}
		fmt.Printf("Downloading tarball list from %s\n", fileUrl)
		err := rest.DownloadFile(fileName, fileUrl, true, globals.ProgressStepValue)
		if err != nil {
			common.Exitf(1, "error downloading tarball list to JSON file")
		}
	}
	if !common.FileExists(fileName) {
		common.Exitf(1, globals.ErrFileNotFound, fileName)
	}
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

func addRemoteTarballToCollection(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	if len(args) < 3 {
		common.Exit(1, "command 'add' requires tarball type, short version, and operating system")
	}
	tarballType := downloads.TarballType(args[0])

	list, err := downloads.GetRemoteTarballList(tarballType, args[1], args[2], true)

	if err != nil {
		common.Exitf(1, "error getting remote tarball description: %s", err)
	}

	var tarballCollection = downloads.DefaultTarballRegistry
	minimal, _ := flags.GetBool(globals.MinimalLabel)
	overwrite, _ := flags.GetBool(globals.OverwriteLabel)
	var added []downloads.TarballDescription
	for _, tb := range list {
		if minimal && !tb.Minimal {
			continue
		}
		existingTarball, err := downloads.FindTarballByName(tb.Name)
		if err == nil {
			if overwrite {
				var newList []downloads.TarballDescription
				newList, err = downloads.DeleteTarball(tb.Name)
				if err != nil {
					common.Exitf(1, "error removing tarball %s from list", tb.Name)
				}
				tarballCollection.Tarballs = newList
			} else {
				displayTarball(existingTarball)
				fmt.Println()
				common.Exitf(1, "tarball %s already in the list", tb.Name)
			}
		}

		tb.DateAdded = time.Now().Format("2006-01-02 15:04")
		tb.Notes = fmt.Sprintf("added with version %s", common.VersionDef)
		tarballCollection.Tarballs = append(tarballCollection.Tarballs, tb)
		added = append(added, tb)
	}
	if len(added) == 0 {
		common.Exitf(1, "no tarballs found to add")
	}
	err = downloads.WriteTarballFileInfo(tarballCollection)
	if err != nil {
		common.Exitf(1, "error writing tarball list: %s", err)
	}
	fmt.Printf("Tarball below added to %s\n", downloads.TarballFileRegistry)
	for _, tb := range added {
		displayTarball(tb)
	}
}

func addTarballToCollection(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	if len(args) < 1 {
		common.Exit(1, "command 'add' requires a tarball name")
	}
	var err error
	fileName := args[0]
	fileName, err = common.AbsolutePath(fileName)
	if err != nil {
		common.Exitf(1, "error detecting absolute path of %s: %s", fileName, err)
	}

	if !common.FileExists(fileName) {
		common.Exitf(1, globals.ErrFileNotFound, fileName)
	}
	var tarballCollection = downloads.DefaultTarballRegistry

	baseName := common.BaseName(fileName)

	OS, _ := flags.GetString(globals.OSLabel)
	flavor, _ := flags.GetString(globals.FlavorLabel)
	tarballUrl, _ := flags.GetString(globals.UrlLabel)
	version, _ := flags.GetString(globals.VersionLabel)
	shortVersion, _ := flags.GetString(globals.ShortVersionLabel)
	minimal, _ := flags.GetBool(globals.MinimalLabel)
	overwrite, _ := flags.GetBool(globals.OverwriteLabel)
	existingTarball, err := downloads.FindTarballByName(baseName)
	if err == nil {
		if overwrite {
			var newList []downloads.TarballDescription
			newList, err = downloads.DeleteTarball(baseName)
			if err != nil {
				common.Exitf(1, "error removing tarball %s from list", baseName)
			}
			tarballCollection.Tarballs = newList
		} else {
			displayTarball(existingTarball)
			fmt.Println()
			common.Exitf(1, "tarball %s already in the list", baseName)
		}
	}
	var tarballDesc = downloads.TarballDescription{
		OperatingSystem: OS,
		Minimal:         minimal,
		ShortVersion:    shortVersion,
		Version:         version,
		Flavor:          flavor,
		Url:             tarballUrl,
		Notes:           fmt.Sprintf("added with version %s", common.VersionDef),
		DateAdded:       time.Now().Format("2006-01-02 15:04"),
	}
	tarballDesc, err = downloads.GetTarballInfo(fileName, tarballDesc)
	if err != nil {
		common.Exitf(1, "error collecting tarball info: %s", err)
	}

	tarballCollection.Tarballs = append(tarballCollection.Tarballs, tarballDesc)

	err = downloads.WriteTarballFileInfo(tarballCollection)
	if err != nil {
		common.Exitf(1, "error writing tarball list: %s", err)
	}
	fmt.Printf("Tarball below added to %s\n", downloads.TarballFileRegistry)
	displayTarball(tarballDesc)
}

func addTarballToCollectionFromStdin(cmd *cobra.Command, args []string) {
	var err error
	var tarballCollection = downloads.DefaultTarballRegistry
	overwrite, _ := cmd.Flags().GetBool(globals.OverwriteLabel)
	scanner := bufio.NewScanner(os.Stdin)

	text := ""
	for scanner.Scan() {
		text += scanner.Text()
	}
	var tarballDesc downloads.TarballDescription
	err = json.Unmarshal([]byte(text), &tarballDesc)
	if err != nil {
		common.Exitf(1, "error decoding JSON item:%s", err)
	}
	existingTarball, err := downloads.FindTarballByName(tarballDesc.Name)
	if err == nil {
		if overwrite {
			var newList []downloads.TarballDescription
			newList, err = downloads.DeleteTarball(tarballDesc.Name)
			if err != nil {
				common.Exitf(1, "error removing tarball %s from list", tarballDesc.Name)
			}
			tarballCollection.Tarballs = newList
		} else {
			displayTarball(existingTarball)
			fmt.Println()
			common.Exitf(1, "tarball %s already in the list", tarballDesc.Name)
		}
	}
	tarballDesc.Notes = fmt.Sprintf("added with version %s", common.VersionDef)
	tarballCollection.Tarballs = append(tarballCollection.Tarballs, tarballDesc)

	err = downloads.WriteTarballFileInfo(tarballCollection)
	if err != nil {
		common.Exitf(1, "error writing tarball list: %s", err)
	}
	fmt.Printf("Tarball below added to %s\n", downloads.TarballFileRegistry)
	displayTarball(tarballDesc)
}

func removeTarballFromCollection(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, "command 'delete' requires a tarball name")
	}
	var err error
	fileName := args[0]

	var tarballCollection = downloads.DefaultTarballRegistry

	var newList []downloads.TarballDescription
	newList, err = downloads.DeleteTarball(fileName)
	if err != nil {
		common.Exitf(1, "error removing tarball %s from list", fileName)
	}
	tarballCollection.Tarballs = newList

	err = downloads.WriteTarballFileInfo(tarballCollection)
	if err != nil {
		common.Exitf(1, "error writing updated tarball list: %s", err)
	}
	fmt.Printf("Tarball '%s' deleted from %s\n", fileName, downloads.TarballFileRegistry)
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

var downloadsTreeCmd = &cobra.Command{
	Use:   "tree [options]",
	Short: "Display a tree by version of remote tarballs",
	Long: `Display a tree by version of remote tarballs.
By default it includes tarballs for current operating system.
The flag '--flavor' is mandatory for this command.

Use '--OS=os_name' to change.
Use '--max-items' to display fewer or more items per version.
`,
	Run: treeRemoteTarballs,
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
	Use:   "import {file-name | URL}",
	Short: "Imports the list of tarballs from a file or URL",
	Long: `
Imports the list of tarballs from a file or a URL.
If the argument is "remote-github" or "remote-tarballs", dbdeployer will get the file from
its Github repository.
(See: dbdeployer info defaults remote-tarball-url)
`,

	Run: importTarballCollection,
}

var downloadsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the custom list of tarballs and resume the defaults",
	Long:  ``,

	Run: resetTarballCollection,
}

var downloadsAddCmd = &cobra.Command{
	Use:   "add tarball_name",
	Short: "Adds a tarball to the list",
	Long:  ``,

	Run: addTarballToCollection,
}

// downloadsDeleteCmd is only intended for internal maintenance
var downloadsDeleteCmd = &cobra.Command{
	Use:     "delete tarball_name",
	Aliases: []string{"remove"},
	Hidden:  true,
	Short:   "Removes a tarball from the list",
	Long:    ``,

	Run: removeTarballFromCollection,
}

var downloadsAddRemoteCmd = &cobra.Command{
	Use:   "add-remote tarball-type short-version operating-system",
	Short: "Adds a tarball to the list, by searching MySQL downloads site ",
	Long: `This command can add a tarball by searching the MySQL site for one of these
tarball types: mysql, cluster, shell`,

	Run: addRemoteTarballToCollection,
}

var downloadsAddStdinCmd = &cobra.Command{
	Use:    "add-stdin ",
	Short:  "Adds a tarball to the list from standard input",
	Long:   ``,
	Hidden: true,

	Run: addTarballToCollectionFromStdin,
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
	downloadsCmd.AddCommand(downloadsAddCmd)
	downloadsCmd.AddCommand(downloadsAddRemoteCmd)
	downloadsCmd.AddCommand(downloadsAddStdinCmd)
	downloadsCmd.AddCommand(downloadsDeleteCmd)
	downloadsCmd.AddCommand(downloadsTreeCmd)

	downloadsListCmd.Flags().BoolP(globals.ShowUrlLabel, "", false, "Show the URL")
	downloadsListCmd.Flags().String(globals.FlavorLabel, "", "Which flavor will be listed")
	downloadsListCmd.Flags().String(globals.OSLabel, "", "Which OS will be listed")
	downloadsListCmd.Flags().String(globals.VersionLabel, "", "Which version will be listed")
	downloadsListCmd.Flags().String(globals.SortByLabel, "name", "Sort by field {name/date/version}")

	downloadsTreeCmd.Flags().String(globals.FlavorLabel, "", "Which flavor will be listed")
	downloadsTreeCmd.Flags().BoolP(globals.ShowUrlLabel, "", false, "Show the URL")
	downloadsTreeCmd.Flags().String(globals.OSLabel, "", "Which OS will be listed")
	downloadsTreeCmd.Flags().String(globals.VersionLabel, "", "Which version will be listed")
	downloadsTreeCmd.Flags().IntP(globals.MaxItemsLabel, "", 3, "Show a maximum of items for each Short version (0 = ALL)")
	_ = downloadsTreeCmd.MarkFlagRequired(globals.FlavorLabel)

	downloadsGetByVersionCmd.Flags().BoolP(globals.NewestLabel, "", false, "Choose only the newest tarballs not yet downloaded")
	downloadsGetByVersionCmd.Flags().BoolP(globals.DryRunLabel, "", false, "Show what would be downloaded, but don't run it")
	downloadsGetByVersionCmd.Flags().BoolP(globals.MinimalLabel, "", false, "Choose only minimal tarballs")
	downloadsGetByVersionCmd.Flags().String(globals.FlavorLabel, "", "Choose only the given flavor")
	downloadsGetByVersionCmd.Flags().String(globals.OSLabel, "", "Choose only the given OS")
	downloadsGetByVersionCmd.Flags().BoolP(globals.QuietLabel, "", false, "Do not show download progress")
	downloadsGetByVersionCmd.Flags().BoolP(globals.GuessLatestLabel, "", false, "Guess the latest version (highest version w/ increased revision number)")
	downloadsGetByVersionCmd.Flags().Int64P(globals.ProgressStepLabel, "", globals.ProgressStepValue, "Progress interval")

	downloadsGetCmd.Flags().BoolP(globals.QuietLabel, "", false, "Do not show download progress")
	downloadsGetCmd.Flags().Int64P(globals.ProgressStepLabel, "", globals.ProgressStepValue, "Progress interval")
	downloadsGetCmd.Flags().BoolP(globals.DryRunLabel, "", false, "Show what would be downloaded, but don't run it")
	downloadsGetCmd.Flags().String(globals.OSLabel, "", "Set the OS of the tarball")

	downloadsGetUnpackCmd.Flags().BoolP(globals.DeleteAfterUnpackLabel, "", false, "Delete the tarball after successful unpack")
	downloadsGetUnpackCmd.Flags().Bool(globals.DryRunLabel, false, "Show Get operations, but do not run them")

	downloadsAddCmd.Flags().String(globals.OSLabel, "", "Define the tarball OS (default: current OS)")
	downloadsAddCmd.Flags().String(globals.FlavorLabel, "", "Define the tarball flavor")
	downloadsAddCmd.Flags().String(globals.VersionLabel, "", "Define the tarball version")
	downloadsAddCmd.Flags().String(globals.ShortVersionLabel, "", "Define the tarball short version")
	downloadsAddCmd.Flags().String(globals.UrlLabel, "", "Define the tarball URL")
	downloadsAddCmd.Flags().BoolP(globals.MinimalLabel, "", false, "Define whether the tarball is a minimal one")
	downloadsAddCmd.Flags().BoolP(globals.OverwriteLabel, "", false, "Overwrite existing entry")
	_ = downloadsAddCmd.MarkFlagRequired(globals.UrlLabel)
	_ = downloadsAddCmd.MarkFlagRequired(globals.OSLabel)

	downloadsAddRemoteCmd.Flags().BoolP(globals.OverwriteLabel, "", false, "Overwrite existing entry")
	downloadsAddRemoteCmd.Flags().BoolP(globals.MinimalLabel, "", false, "Define whether the wanted tarball is a minimal one")

	downloadsAddStdinCmd.Flags().BoolP(globals.OverwriteLabel, "", false, "Overwrite existing entry")

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
