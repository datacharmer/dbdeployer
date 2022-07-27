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

	"github.com/datacharmer/dbdeployer/common"
)

type VersionOptions struct {
	SandboxBinary string
	Flavor        string
	ByFlavor      bool
}

func validateVersionOptions(options VersionOptions) error {
	if options.SandboxBinary == "" {
		return fmt.Errorf("version options needs to include a non-empty SandboxBinary")
	}
	return nil
}

func ShowVersions(options VersionOptions) error {
	err := validateVersionOptions(options)
	if err != nil {
		return err
	}
	basedir := options.SandboxBinary
	flavor := options.Flavor
	byFlavor := options.ByFlavor

	var versionInfoList []common.VersionInfo
	var dirs []string
	var flavoredLists = make(map[string][]string)

	versionInfoList = common.GetVersionInfoFromDir(basedir)
	if byFlavor {
		for _, verInfo := range versionInfoList {
			flavoredLists[verInfo.Flavor] = append(flavoredLists[verInfo.Flavor], verInfo.Version)
		}
		count := 0
		for f, versions := range flavoredLists {
			listVersions(versions, basedir, f, count)
			count++
			fmt.Println("")
		}
	} else {
		for _, verInfo := range versionInfoList {
			if flavor == verInfo.Flavor || flavor == "" {
				dirs = append(dirs, verInfo.Version)
			}
		}
		listVersions(dirs, basedir, flavor, 0)
	}
	return nil
}

func listVersions(dirs []string, basedir, flavor string, iteration int) {
	maxWidth := 80
	maxLen := 0
	for _, dir := range dirs {
		if len(dir) > maxLen {
			maxLen = len(dir)
		}
	}
	var header string

	if basedir != "" && iteration == 0 {
		header = fmt.Sprintf("Basedir: %s\n", basedir)
	}
	if flavor != "" {
		header += fmt.Sprintf("(Flavor: %s)\n", flavor)
	}
	if header != "" {
		fmt.Printf("%s", header)
	}
	columns := int(maxWidth / (maxLen + 2))
	mask := fmt.Sprintf("%%-%ds", maxLen+2)
	count := 0
	for _, dir := range dirs {
		fmt.Printf(mask, dir)
		count += 1
		if count > columns {
			count = 0
			fmt.Println("")
		}
	}
	fmt.Println("")
}
