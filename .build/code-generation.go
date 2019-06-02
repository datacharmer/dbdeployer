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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/downloads"
	"github.com/datacharmer/dbdeployer/globals"
)

func getFileDate(filename string) string {
	file, err := os.Stat(filename)
	if err != nil {
		common.Exit(1, fmt.Sprintf("error getting timestamp for file %s (%s)", filename, err))
	}
	modifiedTime := file.ModTime()
	return modifiedTime.Format("2006-01-02")
}

func createVersionFile() {
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

func createTarballRegistry() {

	var tarballList downloads.TarballCollection
	sourceTarballList := "./downloads/tarball_list.json"
	tarballRegistryTemplate := "./.build/tarball_registry_template.txt"
	if !common.FileExists(tarballRegistryTemplate) {
		common.Exitf(1, globals.ErrFileNotFound, tarballRegistryTemplate)
	}
	if !common.FileExists(sourceTarballList) {
		common.Exitf(1, globals.ErrFileNotFound, sourceTarballList)
	}

	jsonText, err := common.SlurpAsBytes(sourceTarballList)
	if err != nil {
		common.Exitf(1, "error reading file %s: %s", sourceTarballList, err)
	}
	err = json.Unmarshal(jsonText, &tarballList)
	if err != nil {
		common.Exitf(1, "error decoding JSON file %s: %s", sourceTarballList, err)
	}
	registryTemplate, err := common.SlurpAsString(tarballRegistryTemplate)
	if err != nil {
		common.Exitf(1, "error reading file %s: %s", tarballRegistryTemplate, err)
	}
	if len(registryTemplate) < 10 {
		common.Exitf(1, "error reading file %s: 0 byte retrieved", tarballRegistryTemplate)
	}

	data := make(common.StringMap)
	data["Version"] = common.VersionDef
	data["Items"] = []common.StringMap{}
	// data["Timestamp"] = time.Now().Format("2006-01-02 15:04")
	for _, tb := range tarballList.Tarballs {
		tempItem := common.StringMap{
			"Name":            tb.Name,
			"Checksum":        tb.Checksum,
			"OperatingSystem": tb.OperatingSystem,
			"Url":             tb.Url,
			"Flavor":          tb.Flavor,
			"Minimal":         tb.Minimal,
			"Size":            tb.Size,
			"ShortVersion":    tb.ShortVersion,
			"Version":         tb.Version,
			"Notes":           "",
			"UpdatedBy":       "",
		}
		if tb.Notes != "" {
			tempItem["Notes"] = fmt.Sprintf(`Notes: "%s",`, tb.Notes)
		}
		if tb.UpdatedBy != "" {
			tempItem["UpdatedBy"] = fmt.Sprintf(`UpdatedBy: "%s",`, tb.UpdatedBy)
		}
		data["Items"] = append(data["Items"].([]common.StringMap), tempItem)
	}

	out, err := common.SafeTemplateFill("tarball", registryTemplate, data)
	if err != nil {
		common.Exitf(1, "error filling template: %s", err)
	}
	fmt.Printf("%s\n", out)
}

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("Syntax: code_generation {version|tarball} \n")
	}
	option := os.Args[1]
	switch option {
	case "version":
		createVersionFile()
	case "tarball":
		createTarballRegistry()
	default:
		fmt.Printf("Option %s not recognized. Use one of [version tarball]\n", option)
		os.Exit(1)
	}
}
