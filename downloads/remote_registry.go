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
package downloads

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

type TarballDescription struct {
	Name            string `json:"name"`
	Checksum        string `json:"checksum,omitempty"`
	OperatingSystem string `json:"OS"`
	Url             string `json:"url"`
	Flavor          string `json:"flavor"`
	Minimal         bool   `json:"minimal"`
	Size            int64  `json:"size"`
	ShortVersion    string `json:"short_version"`
	Version         string `json:"version"`
	UpdatedBy       string `json:"updated_by,omitempty"`
	Notes           string `json:"notes,omitempty"`
}

type TarballCollection struct {
	DbdeployerVersion string
	Tarballs          []TarballDescription
}

func FindTarballByName(tarballName string) (TarballDescription, error) {
	for _, tb := range DefaultTarballRegistry.Tarballs {
		if tb.Name == tarballName {
			return tb, nil
		}
	}
	return TarballDescription{}, fmt.Errorf("tarball with name %s not found", tarballName)
}

func CompareTarballChecksum(tarball TarballDescription, fileName string) error {

	if tarball.Checksum == "" {
		return nil
	}
	reCRC := regexp.MustCompile(`(MD5|SHA1|SHA256|SHA512)\s*:\s*(\S+)`)
	crcList := reCRC.FindAllStringSubmatch(tarball.Checksum, -1)

	if len(crcList) < 1 || len(crcList[0]) < 1 {
		return fmt.Errorf("not a valid CRC pattern found. Expected: (MD5|SHA1|SHA256|SHA512):CHECKSUM_STRING")
	}

	crcType := crcList[0][1]
	crcText := crcList[0][2]

	if crcType == "" {
		return fmt.Errorf("no CRC type detected in checksum field for %s", tarball.Name)
	}
	if crcText == "" {
		return fmt.Errorf("no CRC detected in checksum field for %s", tarball.Name)
	}
	var hasher hash.Hash
	switch crcType {
	case "MD5":
		hasher = md5.New()
	case "SHA1":
		hasher = sha1.New()
	case "SHA256":
		hasher = sha256.New()
	case "SHA512":
		hasher = sha512.New()
	}

	if hasher == nil {
		return fmt.Errorf("can't establish a hash function for %s", crcType)
	}
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(hasher, f)
	if err != nil {
		return err
	}
	localChecksum := hex.EncodeToString(hasher.Sum(nil))
	if localChecksum != crcText {
		return fmt.Errorf("unmatched checksum: expected '%s' but found '%s'", crcText, localChecksum)
	}
	// fmt.Printf("MATCHED %s\n",localChecksum)
	return nil
}

func FindTarballByVersionFlavorOS(version, flavor, OS string, minimal, newest bool) (TarballDescription, error) {
	flavor = strings.ToLower(flavor)
	OS = strings.ToLower(OS)
	if OS == "osx" || OS == "macos" || OS == "os x" {
		OS = "darwin"
	}
	var tbd []TarballDescription
	for _, tb := range DefaultTarballRegistry.Tarballs {
		if (tb.Version == version || tb.ShortVersion == version) &&
			strings.ToLower(tb.Flavor) == flavor &&
			strings.ToLower(tb.OperatingSystem) == OS &&
			(!minimal || minimal == tb.Minimal) {
			tbd = append(tbd, tb)
		}
	}

	if len(tbd) == 1 {
		return tbd[0], nil
	}

	if len(tbd) > 1 {
		if newest {
			var newestTarball TarballDescription = tbd[0]
			greaterVL, err := common.VersionToList(newestTarball.Version)
			if err != nil {
				return TarballDescription{}, fmt.Errorf("could not establish the version for %s", newestTarball.Name)
			}

			for _, tb := range tbd {
				if tb.Name != newestTarball.Name && tb.Version == newestTarball.Version {
					return TarballDescription{}, fmt.Errorf("tarballs %s and %s have the same version - Get the one you want by name",
						tb.Name, newestTarball.Name)
				}
				currentVL, err := common.VersionToList(tb.Version)
				if err != nil {
					return TarballDescription{}, fmt.Errorf("could not establish the version for %s", tb.Name)
				}
				isBigger, err := common.GreaterOrEqualVersionList(currentVL, greaterVL)
				if err != nil {
					return TarballDescription{}, fmt.Errorf("%s", err)
				}
				if isBigger {
					greaterVL = currentVL
					newestTarball = tb
				}
			}
			return newestTarball, nil
		}
		names := ""
		for _, tb := range tbd {
			names += " " + tb.Name
		}
		return TarballDescription{}, fmt.Errorf("more than one tarballs found with current search criteria (%s).\n"+
			"Get it by name instead (or use --%s)", names, globals.NewestLabel)
	}

	return TarballDescription{}, fmt.Errorf("tarball with version %s, flavor %s, OS %s not found", version, flavor, OS)
}

const tarballRegistryName string = "tarball-list.json"

var TarballFileRegistry string = path.Join(defaults.ConfigurationDir, tarballRegistryName)

func TarballRegistryFileExist() bool {
	return common.FileExists(TarballFileRegistry)
}

func ReadTarballFileCount() int {
	// If there is no file, return an empty collection
	if !TarballRegistryFileExist() {
		return 0
	}
	rfc, err := ReadTarballFileInfo()
	if err != nil {
		return 0
	}
	return len(rfc.Tarballs)
}

func ReadTarballFileInfo() (collection TarballCollection, err error) {
	// If there is no file, return an empty collection
	if !TarballRegistryFileExist() {
		return collection, nil
	}
	text, err := common.SlurpAsBytes(TarballFileRegistry)
	if err != nil {
		return TarballCollection{}, err
	}
	err = json.Unmarshal(text, &collection)
	return collection, err
}

func LoadTarballFileInfo() error {

	collection, err := ReadTarballFileInfo()
	if err != nil {
		return err
	}
	err = TarballFileInfoValidation(collection)
	if err != nil {
		return err
	}
	DefaultTarballRegistry = collection
	return nil
}

func WriteTarballFileInfo(collection TarballCollection) error {
	text, err := json.MarshalIndent(collection, " ", " ")
	if err != nil {
		return err
	}
	return common.WriteString(string(text), TarballFileRegistry)
}

func TarballFileInfoValidation(collection TarballCollection) error {
	type tarballError struct {
		Name  string
		Issue string
	}
	var tarballErrorList []tarballError

	if collection.DbdeployerVersion == "" {
		tarballErrorList = append(tarballErrorList, tarballError{"collection version", "dbdeployer version not set"})
	}
	if !common.IsCompatibleVersion(collection.DbdeployerVersion) {
		tarballErrorList = append(tarballErrorList, tarballError{
			"collection version",
			fmt.Sprintf("dbdeployer version '%s' not compatible with current '%s'",
				collection.DbdeployerVersion, common.CompatibleVersion)})
	}
	for _, tb := range collection.Tarballs {
		if tb.Name == "" {
			tarballErrorList = append(tarballErrorList, tarballError{"No Name", "name is missing"})
		}
		if tb.Url == "" {
			tarballErrorList = append(tarballErrorList, tarballError{tb.Name, "Url is missing"})
		}
		if tb.ShortVersion == "" {
			tarballErrorList = append(tarballErrorList, tarballError{tb.Name, "short version is missing"})
		}
		if tb.Version == "" {
			tarballErrorList = append(tarballErrorList, tarballError{tb.Name, "version is missing"})
		}
		// TODO: validate the checksum type and the corresponding checksum length
		if tb.Checksum == "" && tb.Flavor != "tidb" {
			tarballErrorList = append(tarballErrorList, tarballError{tb.Name, "checksum is missing"})
		}
		if tb.OperatingSystem == "" {
			tarballErrorList = append(tarballErrorList, tarballError{tb.Name, "operating system is missing"})
		}
	}
	if len(tarballErrorList) > 0 {
		errorBytes, err := json.MarshalIndent(tarballErrorList, " ", " ")
		if err != nil {
			return fmt.Errorf("%v", tarballErrorList)
		}
		return fmt.Errorf("validation errors\n%s", string(errorBytes))
	}
	return nil
}
