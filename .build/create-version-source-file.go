// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2018 Giuseppe Maxia
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

package main

import (
	//"bufio"
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"os"
	"strings"
	"time"
)

func getFileDate(filename string) string {
	file, err := os.Stat(filename)
	if err != nil {
		common.Exit(1, fmt.Sprintf("error getting timestamp for file %s (%s)", filename, err))
	}
	modifiedTime := file.ModTime()
	return modifiedTime.Format("2006-01-02")
}

func main() {
	templatesDir := ".build"
	if !common.DirExists(templatesDir) {
		err := os.Chdir("..")
		common.ErrCheckExitf(err, 1, "error changing directory to %s/..", templatesDir)
		if !common.DirExists(templatesDir) {
			common.Exit(1, "Directory .build/ not found")
		}
	}
	versionDestFile := "common/version.go"

	if !common.FileExists(versionDestFile) {
		common.Exit(1, fmt.Sprintf("File %s not found", versionDestFile))
	}
	version_template := templatesDir + "/version_template.txt"
	template, err := common.SlurpAsString(version_template)
	common.ErrCheckExitf(err, 1, "error reading version template")

	versionFile := templatesDir + "/VERSION"
	versionText, err := common.SlurpAsString(versionFile)
	common.ErrCheckExitf(err, 1, "error reading version file")

	version := strings.TrimSpace(versionText)
	versionDate := getFileDate(versionFile)

	compatibleVersionFile := templatesDir + "/COMPATIBLE_VERSION"
	compatibleVersionText, err := common.SlurpAsString(compatibleVersionFile)
	common.ErrCheckExitf(err, 1, "error reading compatible version file")
	compatibleVersion := strings.TrimSpace(compatibleVersionText)
	compatibleVersionDate := getFileDate(compatibleVersionFile)

	var data = common.StringMap{
		"Version":               version,
		"VersionDate":           versionDate,
		"CompatibleVersion":     compatibleVersion,
		"CompatibleVersionDate": compatibleVersionDate,
		"Timestamp":             time.Now().Format("2006-01-02 15:04"),
	}
	versionCode, err := common.SafeTemplateFill("create-version-source-file", template, data)
	common.ErrCheckExitf(err, 1, "error filling version code template %s", err)

	err = common.WriteString(versionCode, versionDestFile)
	common.ErrCheckExitf(err, 1, "error writing version code file %s", versionDestFile)
}
