// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2022 Giuseppe Maxia
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

package ops

import (
	"fmt"
	"os"
	"path"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
)

func ProcessBashCompletionEnabling(useRemote, runIt bool, remoteUrl, completionFile string) error {
	useLocal := completionFile != ""

	var bashCompletionScript string
	var bashCompletionScripts = []string{
		path.Join("/etc", "bash_completion"),
		path.Join("/usr", "local", "etc", "bash_completion"),
		path.Join("/etc", "profile.d", "bash_completion.sh"),
	}
	destinationDir := path.Join("/etc", "bash_completion.d")
	alternateDestinationDir := path.Join("/usr", "local", "etc", "bash_completion.d")
	if !common.DirExists(destinationDir) {
		if common.DirExists(alternateDestinationDir) {
			destinationDir = alternateDestinationDir
		} else {
			return fmt.Errorf("neither %s or %s found", destinationDir, alternateDestinationDir)
		}
	}

	for _, script := range bashCompletionScripts {
		if common.FileExists(script) {
			bashCompletionScript = script
			break
		}
	}
	if bashCompletionScript == "" {
		return fmt.Errorf("none of bash completion scripts found (%v)", bashCompletionScripts)
	}
	if completionFile == "" {
		completionFile = globals.CompletionFileValue
	}
	if useLocal && useRemote {
		return fmt.Errorf("only one of '--%s' or '--%s' should be used", globals.CompletionFileValue, globals.RemoteLabel)
	}
	if !useRemote {
		useLocal = true
	}
	if useLocal {
		defaultCompletionFile := path.Join(os.Getenv("PWD"), globals.CompletionFileValue)
		defaultSecondCompletionFile := path.Join(os.Getenv("PWD"), "docs", globals.CompletionFileValue)
		completionFile, _ = common.AbsolutePath(completionFile)
		if completionFile == defaultCompletionFile {
			if !common.FileExists(completionFile) {
				if common.FileExists(defaultSecondCompletionFile) {
					completionFile = defaultSecondCompletionFile
				}
			}
		}
	}
	if useRemote {
		if remoteUrl == "" {
			return fmt.Errorf("remote URL at '--%s' cannot be empty", globals.RemoteUrlLabel)
		}
		if common.FileExists(completionFile) {
			return fmt.Errorf(globals.ErrFileAlreadyExists, completionFile)
		}
		err := rest.DownloadFile(completionFile, remoteUrl, true, globals.MB)
		if err != nil {
			return fmt.Errorf("error downloading %s: %s", completionFile, err)
		}
		fmt.Printf("Download of file %s was successful\n", completionFile)
	}

	if !common.FileExists(completionFile) {
		return fmt.Errorf(globals.ErrFileNotFound, completionFile)
	}

	fmt.Printf("# completion file: %s\n", completionFile)
	bareCompletionFileName := common.BaseName(completionFile)
	destinationFile := path.Join(destinationDir, bareCompletionFileName)
	if common.FileExists(destinationFile) {
		// Get the checksum of both files, so we can skip the copy if they are already the same
		sourceChecksum, err := common.GetFileSha256(completionFile)
		if err != nil {
			return fmt.Errorf("error getting checksum from file %s", completionFile)
		}
		destChecksum, err := common.GetFileSha256(destinationFile)
		if err != nil {
			return fmt.Errorf("error getting checksum from file %s", destinationFile)
		}
		if sourceChecksum == destChecksum {
			fmt.Printf("Files '%s' and '%s' have the same checksum - Copy is not needed\n", completionFile, destinationFile)
			return nil
		}
	}

	if runIt {
		command := "cp"
		argsList := []string{completionFile, destinationDir}
		sudo := common.Which("sudo")
		if sudo != "" {
			command = sudo
			argsList = []string{"cp", completionFile, destinationDir}
		}
		fmt.Printf("# Running: sudo cp %s %s\n", completionFile, destinationDir)

		output, err := common.RunCmdWithArgs(command, argsList)
		if err != nil {
			fmt.Printf("%s\n", output)
			return fmt.Errorf("error copying bash completion file into %s: %s", destinationDir, err)
		}
		if !common.FileExists(destinationFile) {
			return fmt.Errorf("error after copying bash completion file: "+globals.ErrFileNotFound, destinationFile)
		}
		fmt.Printf("# File copied to %s\n", destinationFile)
	} else {

		fmt.Printf("# Run the command: sudo cp %s %s\n", completionFile, destinationDir)
	}
	fmt.Printf("# Run the command 'source %s'\n", bashCompletionScript)
	return nil
}
