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
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

func sandboxesUsingBinariesDir(basedir, binariesDir string) ([]string, error) {
	var sandboxes []string
	var sandboxList defaults.SandboxCatalog
	var err error
	sandboxList, err = defaults.ReadCatalog()
	if err != nil {
		return nil, fmt.Errorf("error getting sandboxes from catalog: %s", err)
	}
	fullPath := path.Join(basedir, binariesDir)
	for _, sb := range sandboxList {
		if sb.Origin == fullPath {
			sandboxes = append(sandboxes, sb.Destination)
		}
	}
	return sandboxes, nil
}

func DeleteBinaries(basedir, binariesDir string, confirm bool) (deleted bool, err error) {
	fullPath := path.Join(basedir, binariesDir)
	if !common.DirExists(fullPath) {
		return false, fmt.Errorf(globals.ErrDirectoryNotFound, fullPath)
	}

	sandboxesUsingBinaries, err := sandboxesUsingBinariesDir(basedir, binariesDir)
	if err != nil {
		return false, fmt.Errorf("error detecting sandboxes using binaries: %s", err)
	}
	if len(sandboxesUsingBinaries) > 0 {
		return false, fmt.Errorf("\nbinaries directory %s is used by the following deployments:\n%s",
			fullPath, strings.Join(sandboxesUsingBinaries, "\n"))
	}
	if confirm {
		common.CondPrintf("Do you want to delete %s? y/[N] ", binariesDir)
		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()
		if err != nil {
			fmt.Println(err)
		} else {
			answer := string(line)
			if answer == "y" || answer == "Y" {
				fmt.Println("Proceeding with deletion")
			} else {
				fmt.Println("Deletion skipped at user request")
				return false, nil
			}
		}
	}
	_, err = common.RunCmdWithArgs("rm", []string{"-rf", fullPath})
	if err != nil {
		return false, fmt.Errorf("error removing %s", fullPath)
	}
	if common.DirExists(fullPath) {
		return false, fmt.Errorf("directory %s was not removed - Reason unknown", fullPath)
	}
	return true, nil
}
