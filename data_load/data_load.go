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

package data_load

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
	"github.com/datacharmer/dbdeployer/unpack"
)

type DataDefinition struct {
	Description       string   `json:"description,omitempty"`        // Description of the database
	Origin            string   `json:"origin"`                       // [Required] Where is the archive
	FileName          string   `json:"file_name"`                    // [Required] File name that we will get locally
	InternalDirectory string   `json:"internal-directory,omitempty"` // Optional internal directory of the archive
	ChangeDirectory   bool     `json:"change-dir,omitempty"`         // Whether we need to operate within the internal directory
	LoadCommands      []string `json:"load-commands"`                // [Required] Set of commands used to load the archive
	Size              uint64   `json:"size,omitempty"`               // Size of original archive
	Sha256            string   `json:"sha256,omitempty`              // SHA 256 checksum of the compressed archive
}

var defaultArchives = map[string]DataDefinition{
	"world": {
		Description:       "world database",
		Origin:            "https://downloads.mysql.com/docs/world-db.tar.gz",
		FileName:          "world-db.tar.gz",
		InternalDirectory: "world-db",
		LoadCommands:      []string{"$use < world-db/world.sql"},
		Size:              92707,
		Sha256:            "", // removed: CRC provided by MySQL is not reliable, as it changes often
	},
	"worldx": {
		Description:       "world_X database",
		Origin:            "https://downloads.mysql.com/docs/world_x-db.tar.gz",
		FileName:          "world_x-db.tar.gz",
		InternalDirectory: "world_x-db",
		LoadCommands:      []string{"$use < world_x-db/world_x.sql"},
		Size:              99295,
		Sha256:            "", // removed: CRC provided by MySQL is not reliable, as it changes often
	},
	"sakila": {
		Description:       "Sakila database",
		Origin:            "https://downloads.mysql.com/docs/sakila-db.tar.gz",
		FileName:          "sakila-db.tar.gz",
		InternalDirectory: "sakila-db",
		LoadCommands: []string{
			"$use < sakila-db/sakila-schema.sql",
			"$use < sakila-db/sakila-data.sql",
		},
		Size:   732126,
		Sha256: "", // removed: CRC provided by MySQL is not reliable, as it changes often
	},
	"employees": {
		Description:       "employee data (large dataset, includes data and test/verification suite)",
		Origin:            "https://github.com/datacharmer/test_db/releases/download/v1.0.7/test_db-1.0.7.tar.gz",
		FileName:          "test_db-1.0.7.tar.gz",
		InternalDirectory: "test_db",
		ChangeDirectory:   true,
		LoadCommands: []string{
			"$use < employees.sql",
		},
		Size:   35607473,
		Sha256: "c44c140f352f35d47fdb65df60f52b779ef552822fad6c4efcfa7b134c3faf84",
	},
	"menagerie": {
		Description:       "menagerie database",
		Origin:            "https://downloads.mysql.com/docs/menagerie-db.tar.gz",
		FileName:          "menagerie-db.tar.gz",
		InternalDirectory: "menagerie-db",
		LoadCommands: []string{
			"$use -e 'create database if not exists menagerie'",
			"$use menagerie < menagerie-db/cr_pet_tbl.sql",
			"$use menagerie < menagerie-db/cr_event_tbl.sql",
			"$use menagerie < menagerie-db/ins_puff_rec.sql",
			"$use menagerie -e 'set global local_infile=ON'",
			"$my sqlimport --local menagerie menagerie-db/pet.txt",
			"$my sqlimport --local menagerie menagerie-db/event.txt",
		},
		Size:   1990,
		Sha256: "", // removed: CRC provided by MySQL is not reliable, as it changes often
	},
}

func Archives() (map[string]DataDefinition, string) {

	if common.FileExists(defaults.ArchivesFile) {
		archiveText, err := common.SlurpAsBytes(defaults.ArchivesFile)
		if err != nil {
			fmt.Printf("WARNING: error reading archive file %s: %s\n", defaults.ArchivesFile, err)
			fmt.Printf("Using default archives\n")
			return defaultArchives, defaults.ArchivesFile
		}
		var newArchives map[string]DataDefinition
		err = json.Unmarshal(archiveText, &newArchives)
		if err != nil {
			fmt.Printf("WARNING: error decoding archives from file %s: %s\n", defaults.ArchivesFile, err)
			fmt.Printf("Using default archives\n")
			return defaultArchives, defaults.ArchivesFile
		}
		return newArchives, defaults.ArchivesFile
	} else {
		return defaultArchives, "defaults"
	}
}

// The JSON encoder by default converts "<" and ">" to their corresponding Unicode values
// Since this info is not used in HTML, we need to convert back from Unicode to plain text
func UnescapeJsonString(bytes []byte) string {
	result := strings.ReplaceAll(string(bytes), `\u003c`, "<")
	return strings.ReplaceAll(result, `\u003e`, ">")
}

func ArchivesAsJson() (string, error) {
	var archives map[string]DataDefinition
	archives, _ = Archives()
	result, err := json.MarshalIndent(archives, " ", " ")
	if err != nil {
		return "", err
	}

	return UnescapeJsonString(result), nil
}

func ListArchives() {
	archives, origin := Archives()
	if origin != "defaults" {
		fmt.Printf("Data load info from %s\n", origin)
	}
	for name, archive := range archives {
		fmt.Printf("%-30s %10s %s\n", name, humanize.Bytes(archive.Size), archive.Description)
	}
}

func LoadArchive(archiveName, sandboxName string, overwrite bool) error {

	archives, _ := Archives()
	archive, found := archives[archiveName]
	if !found {
		return fmt.Errorf("archive %s not found", archiveName)
	}
	sandboxHome := defaults.Defaults().SandboxHome

	sandboxPath := path.Join(sandboxHome, sandboxName)
	if !common.DirExists(sandboxPath) {
		return fmt.Errorf("sandbox %s not found", sandboxName)
	}
	ext := ""
	if strings.HasSuffix(archive.FileName, globals.TarGzExt) {
		ext = globals.TarGzExt
	} else {
		if strings.HasSuffix(archive.FileName, globals.GzExt) {
			ext = globals.GzExt
		}
	}

	useExecutable := path.Join(sandboxPath, "use")
	useMultiExecutable := path.Join(sandboxPath, "n1")
	myExecutable := path.Join(sandboxPath, "my")
	myMultiExecutable := path.Join(sandboxPath, defaults.Defaults().NodePrefix+"1", "my")
	myReplicationExecutable := path.Join(sandboxPath, defaults.Defaults().MasterName, "my")
	internalDir := path.Join(sandboxPath, archive.InternalDirectory)
	if !common.ExecExists(useExecutable) {
		if common.ExecExists(useMultiExecutable) {
			useExecutable = useMultiExecutable
		} else {
			return fmt.Errorf("executable %s not found", useExecutable)
		}
	}
	if !common.ExecExists(myExecutable) {
		if common.ExecExists(myMultiExecutable) {
			myExecutable = myMultiExecutable
		} else {
			if common.ExecExists(myReplicationExecutable) {
				myExecutable = myReplicationExecutable
			} else {
				return fmt.Errorf("executable %s not found", myExecutable)
			}
		}
	}
	if !overwrite && archive.InternalDirectory != "" && common.DirExists(path.Join(sandboxPath, internalDir)) {
		return fmt.Errorf("internal directory %s already exists", internalDir)
	}

	compressedFile := path.Join(sandboxPath, archive.FileName)
	if !overwrite && common.FileExists(compressedFile) {
		return fmt.Errorf(globals.ErrFileAlreadyExists, compressedFile)
	}
	fmt.Printf("downloading %s\n", archive.Origin)
	err := rest.DownloadFile(compressedFile, archive.Origin, true, globals.MB)
	if err != nil {
		return fmt.Errorf("error downloading archive %s: %s", archiveName, err)
	}
	if !common.FileExists(compressedFile) {
		return fmt.Errorf("file %s not found after downloading", compressedFile)
	}

	localChecksum, err := common.GetFileChecksum(compressedFile, "SHA256")
	if err != nil {
		return fmt.Errorf("error retrieving checksum for file %s", compressedFile)
	}
	if archive.Sha256 != "" && localChecksum != archive.Sha256 {
		return fmt.Errorf("the checksum of file %s doesn't match. Expected: %s - Found: %s", compressedFile, archive.Sha256, localChecksum)
	}

	fmt.Printf("Unpacking %s\n", compressedFile)
	switch ext {
	case globals.TarGzExt:
		err = unpack.UnpackTar(compressedFile, sandboxPath, unpack.VERBOSE)
	case globals.GzExt:
		err = unpack.GunzipFile(compressedFile, common.RemoveSuffix(compressedFile, `\.gz`), overwrite)
	default:
		return fmt.Errorf("unsupported file extension")
	}
	if err != nil {
		return fmt.Errorf("error unpacking file %s: %s", compressedFile, err)
	}

	if archive.InternalDirectory != "" && !common.DirExists(path.Join(sandboxPath, archive.InternalDirectory)) {
		return fmt.Errorf("internal directory %s not found after unpacking %s", archive.InternalDirectory, archiveName)
	}
	reUse := regexp.MustCompile(`\$use\b`)
	reMy := regexp.MustCompile(`\$my\b`)

	var loadCommands = []string{"#!/usr/bin/env bash", "set -x"}
	if archive.ChangeDirectory {
		loadCommands = append(loadCommands, fmt.Sprintf("cd %s/%s", sandboxPath, archive.InternalDirectory))
	} else {
		loadCommands = append(loadCommands, fmt.Sprintf("cd %s", sandboxPath))
	}
	for _, rawCommand := range archive.LoadCommands {
		command := reUse.ReplaceAllString(rawCommand, useExecutable)
		command = reMy.ReplaceAllString(command, myExecutable)
		loadCommands = append(loadCommands, command)
	}

	loadScript := path.Join(sandboxPath, "load_db.sh")
	err = common.WriteStrings(loadCommands, loadScript, "\n")
	if err != nil {
		return fmt.Errorf("error creating load script %s: %s", loadScript, err)
	}
	err = os.Chmod(loadScript, globals.ExecutableFileAttr)
	if err != nil {
		return fmt.Errorf("error changing attributes to load script %s: %s", loadScript, err)
	}
	fmt.Printf("Running %s\n", loadScript)
	_, err = common.RunCmd(loadScript)
	if err != nil {
		return fmt.Errorf("error running load script %s: %s", loadScript, err)
	}

	return nil
}
