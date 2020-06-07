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
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
	"github.com/datacharmer/dbdeployer/unpack"
)

func updateDbDeployer(cmd *cobra.Command, args []string) {
	version := ""
	if len(args) > 0 {
		version = args[0]
	}
	flags := cmd.Flags()
	OS, _ := flags.GetString(globals.OSLabel)
	dryRun, _ := flags.GetBool(globals.DryRunLabel)
	verbose, _ := flags.GetBool(globals.VerboseLabel)
	newPath, _ := flags.GetString(globals.NewPathLabel)
	getDocs, _ := flags.GetBool(globals.DocsLabel)
	forceOldVersion, _ := flags.GetBool(globals.ForceOldVersionLabel)

	currentOS := strings.ToLower(runtime.GOOS)
	if OS == "" {
		OS = runtime.GOOS
	}
	OS = strings.ToLower(OS)
	if OS == "macos" || OS == "darwin" {
		OS = "osx"
	}
	if currentOS == "macos" || currentOS == "darwin" {
		currentOS = "osx"
	}
	if dryRun {
		verbose = true
	}
	release, err := rest.GetLatestRelease(version)
	common.ErrCheckExitf(err, 1, "error getting releases %s", err)

	targetDirectory := newPath
	programName := common.BaseName(os.Args[0])

	fullProgramName := common.Which(programName)
	fileInfo, err := os.Stat(fullProgramName)
	if err != nil {
		common.Exitf(1, "error retrieving file info for %s", fullProgramName)
	}
	filePermissions := fileInfo.Mode().Perm()
	dbdeployerPath := common.DirName(common.Which(programName))
	if targetDirectory == "" {
		targetDirectory = dbdeployerPath
	}
	if !common.DirExists(targetDirectory) {
		common.Exitf(1, globals.ErrDirectoryNotFound, targetDirectory)
	}

	tag := release.TagName
	reV := regexp.MustCompile(`^v`)
	tag = reV.ReplaceAllString(tag, "")
	tagList, err := common.VersionToList(tag)
	common.ErrCheckExitf(err, 1, "error converting tag %s to version list", tag)
	if tag == common.VersionDef && !forceOldVersion {
		common.Exit(0,
			fmt.Sprintf("download version (%s) is the same as the current version ", tag),
			fmt.Sprintf("Option --%s was not used\n", globals.ForceOldVersionLabel),
			"Download canceled",
		)
	}
	foundOldVersion, err := common.GreaterOrEqualVersion(common.VersionDef, tagList)
	common.ErrCheckExitf(err, 1, "error comparing remote tag %s to dbdeployer version %s", tag, common.VersionDef)
	if foundOldVersion && !forceOldVersion {
		common.Exit(0,
			fmt.Sprintf("download version (%s) is older than current version (%s) ", tag, common.VersionDef),
			fmt.Sprintf("Option --%s was not used\n", globals.ForceOldVersionLabel),
			"Download canceled",
		)
	}

	docsLabel := ""
	if getDocs {
		docsLabel = "-docs"
	}

	fileName := fmt.Sprintf("dbdeployer-%s%s.%s", tag, docsLabel, OS)
	tarballName := fileName + ".tar.gz"
	signatureName := tarballName + ".sha256"

	fileUrl := ""
	signatureUrl := ""

	if verbose {
		fmt.Printf("Remote version:      %s\n", tag)
		fmt.Printf("Remote file:         %s\n", tarballName)
		fmt.Printf("OS:                  %s\n", OS)
		fmt.Printf("dbdeployer location: %s\n", dbdeployerPath)
		fmt.Println()
		fmt.Printf("%s\n", globals.DashLine)
		fmt.Printf("Release : %s\n", release.Name)
		fmt.Printf("Date    : %s\n", release.PublishedAt)
		fmt.Printf("%s\n", release.Body)
		fmt.Printf("%s\n", globals.DashLine)
	}

	for _, asset := range release.Assets {
		chosenLabel := ""
		if signatureName == asset.Name {
			signatureUrl = asset.BrowserDownloadURL
			if verbose {
				fmt.Printf("\t%s (%s) [CHOSEN CRC]\n", asset.Name, humanize.Bytes(uint64(asset.Size)))
			}
		}
		if tarballName == asset.Name {
			fileUrl = asset.BrowserDownloadURL
			chosenLabel = " [CHOSEN]"
		}
		if verbose {
			fmt.Printf("\t%s (%s)%s\n", asset.Name, humanize.Bytes(uint64(asset.Size)), chosenLabel)
		}
	}

	if fileUrl == "" {
		common.Exitf(1, "file %s not found in release", tarballName)
	}
	if dryRun {
		fmt.Printf("Download %s\n", fileUrl)
		if currentOS == OS {
			fmt.Printf("save to %s/%s\n", targetDirectory, programName)
		}
		return
	}
	if common.FileExists(tarballName) {
		err = os.Remove(tarballName)
		common.ErrCheckExitf(err, 1, "error removing old copy of %s", tarballName)
	}
	err = rest.DownloadFile(tarballName, fileUrl, true, globals.MB)
	common.ErrCheckExitf(err, 1, "error downloading %s", tarballName)
	if !common.FileExists(tarballName) {
		common.Exitf(1, "tarball %s not found after download", tarballName)
	}
	if common.FileExists(fileName) {
		err = os.Remove(fileName)
		common.ErrCheckExitf(err, 1, "error removing old copy of %s", fileName)
	}
	err = unpack.UnpackTar(tarballName, os.Getenv("PWD"), unpack.VERBOSE)
	common.ErrCheckExitf(err, 1, "error unpacking %s", tarballName)
	if !common.FileExists(fileName) {
		common.Exitf(1, globals.ErrFileNotFound, fileName)
	}
	if verbose {
		fmt.Printf("File %s extracted from %s\n", fileName, tarballName)
	}
	if signatureUrl == "" {
		fmt.Printf("*** WARNING *** No SHA256 checksum found for %s\n", tarballName)
	} else {
		err = rest.DownloadFile(signatureName, signatureUrl, true, globals.MB)
		common.ErrCheckExitf(err, 1, "error downloading %s", signatureName)
		signature, err := common.SlurpAsBytes(signatureName)
		common.ErrCheckExitf(err, 1, "error reading from %s", signatureName)
		reSignature := regexp.MustCompile(`^(\S+)\s+(\S+)`)
		signatureList := reSignature.FindAllSubmatch(signature, -1)
		if len(signatureList) == 0 || len(signatureList[0]) == 0 {
			common.Exitf(1, "signature not found in %s", signatureName)
		}
		checksum := signatureList[0][1]
		checksumFileName := signatureList[0][2]
		if string(checksumFileName) != tarballName {
			common.Exitf(1, "wanted signature for %s but got %s", tarballName, checksumFileName)
		}
		calculatedChecksum, err := common.GetFileChecksum(tarballName, "sha256")
		common.ErrCheckExitf(err, 1, "error calculating checksum for %s: %s", tarballName, err)
		if string(checksum) != calculatedChecksum {
			common.Exitf(1, "wanted checksum for %s: %s but got %s", tarballName, checksum, calculatedChecksum)
		}
		fmt.Printf("checksum for %s matches\n", tarballName)
		_ = os.Remove(signatureName)
	}
	_ = os.Remove(tarballName)
	if verbose {
		fmt.Printf("File %s removed\n", tarballName)
	}

	// Give the new file the same attributes of the existing dbdeployer executable
	err = os.Chmod(fileName, filePermissions)
	common.ErrCheckExitf(err, 1, "error changing attributes of %s", fileName)
	if currentOS != OS && targetDirectory == dbdeployerPath {
		fmt.Printf("OS of the remote file (%s) different from current OS (%s)\n", OS, currentOS)
		fmt.Printf("Won't overwrite current dbdeployer executable.\n")
		fmt.Printf("The downloaded file is %s\n", fileName)
		return
	}
	out, err := common.RunCmdCtrlWithArgs("mv", []string{fileName, path.Join(targetDirectory, programName)}, false)
	if err != nil {
		fmt.Printf("%s\n", out)
		common.Exitf(1, "error moving %s to %s/%s", fileName, targetDirectory, programName)
	}
	if verbose {
		fmt.Printf("File %s moved to %s\n", programName, targetDirectory)
	}
	if targetDirectory == "." {
		currentDir, err := os.Getwd()
		if err != nil {
			currentDir = os.Getenv("PWD")
		}
		if currentDir == "" {
			common.Exitf(1, "error getting current working directory")
		}
		targetDirectory = currentDir
	}
	_, err = common.RunCmdCtrlWithArgs(path.Join(targetDirectory, programName), []string{"--version"}, false)
	if err != nil {
		common.Exitf(1, "error running  %s/%s :%s", targetDirectory, programName, err)
	}
}

var updateCmd = &cobra.Command{
	Use:   "update [version]",
	Short: "Gets dbdeployer newest version",
	Long:  `Updates dbdeployer in place using the latest version (or one of your choice)`,
	Example: `
$ dbdeployer update
# gets the latest release, overwrites current dbdeployer binaries 

$ dbdeployer update --dry-run
# shows what it will do, but does not do it

$ dbdeployer update --new-path=$PWD
# downloads the latest executable into the current directory

$ dbdeployer update v1.34.0 --force-old-version
# downloads dbdeployer 1.34.0 and replace the current one
# (WARNING: a version older than 1.36.0 won't support updating)
`,
	Run: updateDbDeployer,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	setPflag(updateCmd, globals.NewPathLabel, "", "", "", "Download updated dbdeployer into a different path", false)
	setPflag(updateCmd, globals.OSLabel, "", "", "", "Gets the executable for this Operating system", false)
	updateCmd.Flags().BoolP(globals.DryRunLabel, "", false, "Show what would happen, but don't execute it")
	updateCmd.Flags().BoolP(globals.VerboseLabel, "", false, "Gives more info")
	updateCmd.Flags().BoolP(globals.ForceOldVersionLabel, "", false, "Force download of older version")
	updateCmd.Flags().BoolP(globals.DocsLabel, "", false, "Gets the docs version of the executable")
}
