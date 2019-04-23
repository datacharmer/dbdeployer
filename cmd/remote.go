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
	"regexp"
	"runtime"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
	"github.com/spf13/cobra"
)

func listRemoteFiles(cmd *cobra.Command, args []string) {
	index, err := rest.GetRemoteIndex()
	common.ErrCheckExitf(err, 1, "error getting remote index: %s", err)

	fmt.Printf("Files available in %s\n", rest.IndexUrl())
	for k, v := range index {
		fmt.Printf("%s -> %+v\n", k, v)
	}
}

func downloadFile(cmd *cobra.Command, args []string) {

	if len(args) < 1 {
		common.Exit(1, "command 'download' requires a version [and optionally a file-name]")
	}
	currentOs := runtime.GOOS
	skipOsCheckVar := "DBDEPLOYER_FORCE_REMOTE_GET"
	skipOsCheck := os.Getenv(skipOsCheckVar) != ""
	if currentOs != "linux" {
		if skipOsCheck {
			fmt.Printf("Variable '%s' was set. Forcing download on OS '%s'\n", skipOsCheckVar, currentOs)
			fmt.Println(globals.HashLine)
			fmt.Printf("# WARNING: the binaries being downloaded WILL NOT WORK on the current OS.\n")
			fmt.Println(globals.HashLine)
		} else {
			common.Exit(0, "The binaries retrieved by this command are only for Linux",
				fmt.Sprintf("Set the environment variable '%s' to force a download on a non-Linux host", skipOsCheckVar))
		}
	}

	version := args[0]

	fileName := version

	if len(args) > 1 {
		fileName = args[1]
	}

	absPath, err := common.AbsolutePath(fileName)
	if err != nil {
		common.Exitf(1, "%s", err)
	}
	match, _ := regexp.MatchString(`^(?:mysql-)?\d+\.\d+$`, version)
	if match {
		common.Exitf(1, " short version detected (%s). The version should have 3 numbers (#.#.#)", version)
	}
	match, _ = regexp.MatchString(`^(?:mysql-)?\d+\.\d+\.\d+$`, version)
	if match {
		version += globals.TarXzExt
	}

	match, _ = regexp.MatchString(`\d+\.\d+\.\d+$`, absPath)
	if match {
		absPath += globals.TarXzExt
	}

	if common.FileExists(absPath) {
		common.Exitf(1, globals.ErrFileAlreadyExists, absPath)
	}
	err = rest.DownloadFile(absPath, rest.FileUrl(version))
	common.ErrCheckExitf(err, 1, "error getting remote file %s - %s", version, err)
	fmt.Printf("File %s downloaded\n", absPath)
}

var remoteDownloadCmd = &cobra.Command{
	Use:         "download version [file-name]",
	Aliases:     []string{"get"},
	Short:       "download a remote tarball into a local file",
	Long:        `If no file name is given, the file name will be <version>.tar.xz`,
	Run:         downloadFile,
	Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
}

var remoteListCmd = &cobra.Command{
	Use:     "list [version]",
	Aliases: []string{"index"},
	Short:   "list remote tarballs",
	Long:    ``,
	Run:     listRemoteFiles,
}

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manages remote tarballs",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(remoteDownloadCmd)
	remoteCmd.AddCommand(remoteListCmd)
}
