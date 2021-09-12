// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"
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
	DateAdded       string `json:"date_added,omitempty"`
}

type TarballDescriptionByName []TarballDescription
type TarballDescriptionByDate []TarballDescription
type TarballDescriptionByVersion []TarballDescription
type TarballDescriptionByShortVersion []TarballDescription

func (tb TarballDescriptionByDate) Less(i, j int) bool {
	dateI, errI := dateparse.ParseAny(tb[i].DateAdded)
	dateJ, errJ := dateparse.ParseAny(tb[j].DateAdded)
	if errI != nil || errJ != nil {
		return tb[i].DateAdded < tb[j].DateAdded
	}
	return dateI.UnixNano() < dateJ.UnixNano()
}

func (tb TarballDescriptionByDate) Len() int {
	return len(tb)
}

func (tb TarballDescriptionByDate) Swap(i, j int) {
	tb[i], tb[j] = tb[j], tb[i]
}

func (tb TarballDescriptionByName) Len() int {
	return len(tb)
}

func (tb TarballDescriptionByName) Swap(i, j int) {
	tb[i], tb[j] = tb[j], tb[i]
}

func (tb TarballDescriptionByName) Less(i, j int) bool {
	return tb[i].Name < tb[j].Name
}

func (tb TarballDescriptionByVersion) Len() int {
	return len(tb)
}

func (tb TarballDescriptionByVersion) Swap(i, j int) {
	tb[i], tb[j] = tb[j], tb[i]

}
func (tb TarballDescriptionByVersion) Less(i, j int) bool {
	iVersionList, _ := common.VersionToList(tb[i].Version)
	jVersionList, _ := common.VersionToList(tb[j].Version)
	greater, _ := common.GreaterOrEqualVersionList(iVersionList, jVersionList)
	return !greater
}

func (tb TarballDescriptionByShortVersion) Len() int {
	return len(tb)
}

func (tb TarballDescriptionByShortVersion) Swap(i, j int) {
	tb[i], tb[j] = tb[j], tb[i]

}

func (tb TarballDescriptionByShortVersion) Less(i, j int) bool {
	iVersionList, _ := common.VersionToList(tb[i].ShortVersion)
	jVersionList, _ := common.VersionToList(tb[j].ShortVersion)
	greater, _ := common.GreaterOrEqualVersionList(iVersionList, jVersionList)
	return !greater
}

type TarballCollection struct {
	DbdeployerVersion string
	UpdatedOn         string `json:"updated_on,omitempty"`
	Tarballs          []TarballDescription
}

func SortedTarballList(tbl []TarballDescription, ByField string) []TarballDescription {
	switch ByField {
	case "version":
		sort.Stable(TarballDescriptionByVersion(tbl))
	case "short":
		sort.Stable(TarballDescriptionByShortVersion(tbl))
	case "date":
		sort.Stable(TarballDescriptionByDate(tbl))
	case "name":
		sort.Stable(TarballDescriptionByName(tbl))
	default:
		sort.Stable(TarballDescriptionByName(tbl))
	}
	return tbl
}

func TarballTree(tbl []TarballDescription) map[string][]TarballDescription {
	tbl = SortedTarballList(tbl, "short")

	var tarballTree = make(map[string][]TarballDescription)
	for _, tb := range tbl {
		_, seen := tarballTree[tb.ShortVersion]
		if !seen {
			tarballTree[tb.ShortVersion] = []TarballDescription{}
		}
		tarballTree[tb.ShortVersion] = append(tarballTree[tb.ShortVersion], tb)
	}
	return tarballTree
}

func FindTarballByUrl(tarballUrl string) (TarballDescription, error) {
	for _, tb := range DefaultTarballRegistry.Tarballs {
		if tb.Url == tarballUrl {
			return tb, nil
		}
	}
	return TarballDescription{}, fmt.Errorf("tarball with Url %s not found", tarballUrl)
}

func FindTarballByName(tarballName string) (TarballDescription, error) {
	for _, tb := range DefaultTarballRegistry.Tarballs {
		if tb.Name == tarballName {
			return tb, nil
		}
	}
	return TarballDescription{}, fmt.Errorf("tarball with name %s not found", tarballName)
}
func DeleteTarball(tarballName string) ([]TarballDescription, error) {
	var newList []TarballDescription
	found := false
	for _, tb := range DefaultTarballRegistry.Tarballs {
		if tb.Name == tarballName {
			found = true
		} else {
			newList = append(newList, tb)
		}
	}
	if !found {
		return nil, fmt.Errorf("tarball %s not found", tarballName)
	}
	return newList, nil
}

func CompareTarballChecksum(tarball TarballDescription, fileName string) error {
	if tarball.Checksum == "" {
		return nil
	}
	reCRC := regexp.MustCompile(`(MD5|SHA1|SHA256|SHA512)\s*:\s*(\S+)`)
	crcList := reCRC.FindAllStringSubmatch(tarball.Checksum, -1)

	if len(crcList) < 1 || len(crcList[0]) < 2 {
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
	localChecksum, err := common.GetFileChecksum(fileName, crcType)
	if err != nil {
		return err
	}
	if localChecksum != crcText {
		return fmt.Errorf("unmatched checksum: expected '%s' but found '%s'", crcText, localChecksum)
	}
	// fmt.Printf("MATCHED %s\n",localChecksum)
	return nil
}

func FindTarballByVersionFlavorOS(version, flavor, OS string, minimal, newest bool) (TarballDescription, error) {
	return FindOrGuessTarballByVersionFlavorOS(version, flavor, OS, minimal, newest, false)
}

func FindOrGuessTarballByVersionFlavorOS(version, flavor, OS string, minimal, newest, guess bool) (TarballDescription, error) {
	flavor = strings.ToLower(flavor)
	OS = strings.ToLower(OS)
	if OS == "osx" || OS == "macos" || OS == "os x" {
		OS = "darwin"
	}
	if guess {
		minimal = false
	}
	var tbd []TarballDescription
	newestVersionList := []int{0, 0, 0}
	for _, tb := range DefaultTarballRegistry.Tarballs {
		if (tb.Version == version || tb.ShortVersion == version) &&
			strings.ToLower(tb.Flavor) == flavor &&
			strings.ToLower(tb.OperatingSystem) == OS &&
			(!minimal || minimal == tb.Minimal) {

			if guess {
				if !isAllowedForGuessing(tb.ShortVersion) {
					return TarballDescription{}, fmt.Errorf("can only guess versions %s ", allowedGuessVersions)
				}
			}
			tbd = append(tbd, tb)
			greatest, err := common.GreaterOrEqualVersion(tb.Version, newestVersionList)
			if err == nil && greatest {
				versionList, err := common.VersionToList(tb.Version)
				if err == nil {
					newestVersionList = versionList
				}
			}
		}
	}

	if newestVersionList[0] == 0 {
		return TarballDescription{}, fmt.Errorf("error detecting latest version")
	}
	newestVersion := fmt.Sprintf("%d.%d.%d", newestVersionList[0], newestVersionList[1], newestVersionList[2])

	if guess && len(tbd) > 0 {

		newest = true
		OS := strings.ToLower(tbd[0].OperatingSystem)
		if OS == "linux" {
			minimal = true
		}
		rev := newestVersionList[2] + 1
		newVersion := fmt.Sprintf("%d.%d.%d", newestVersionList[0], newestVersionList[1], rev)

		shortVersion := tbd[0].ShortVersion
		ext := "tar.gz"
		if OS == "linux" && shortVersion == "8.0" {
			ext = "tar.xz"
		}
		minimalData := ""

		if minimal {
			minimalData = "-minimal"
		}
		data := common.StringMap{"Version": newVersion, "Ext": ext, "Minimal": minimalData}

		fileNameTemplate := ""
		switch OS {
		case "linux":
			fileNameTemplate = defaults.Defaults().DownloadNameLinux
		case "darwin":
			fileNameTemplate = defaults.Defaults().DownloadNameMacOs
		}
		name, err := common.SafeTemplateFill("", fileNameTemplate, data)
		if err != nil {
			return TarballDescription{}, fmt.Errorf("[guess version] error filling new download name %s", err)
		}
		downloadUrl := fmt.Sprintf("%s-%s/%s", defaults.Defaults().DownloadUrl, shortVersion, name)
		tbd = append(tbd, TarballDescription{
			Name:            name,
			Checksum:        "",
			OperatingSystem: OS,
			Url:             downloadUrl,
			Flavor:          flavor,
			Minimal:         minimal,
			Size:            0,
			ShortVersion:    shortVersion,
			Version:         newVersion,
			UpdatedBy:       "",
			Notes:           "guessed",
		})
		newestVersion = newVersion
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
				if tb.Version != newestVersion {
					continue
				}
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

func checkConfigurationDir() error {
	if !common.DirExists(defaults.ConfigurationDir) {
		return os.Mkdir(defaults.ConfigurationDir, globals.PublicDirectoryAttr)
	}
	return nil
}

func WriteTarballFileInfo(collection TarballCollection) error {
	err := CheckTarballList(collection.Tarballs)
	if err != nil {
		return fmt.Errorf("[write tarball file info] tarball list check failed : %s", err)
	}
	text, err := json.MarshalIndent(collection, " ", " ")
	if err != nil {
		return err
	}
	err = checkConfigurationDir()
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

func GetTarballInfo(fileName string, description TarballDescription) (TarballDescription, error) {
	crc, err := common.GetFileSha512(fileName)
	if err != nil {
		return TarballDescription{}, err
	}
	description.Checksum = fmt.Sprintf("SHA512:%s", crc)
	stat, err := os.Stat(fileName)
	if err != nil {
		return TarballDescription{}, err
	}
	description.Size = stat.Size()

	flavor, version, shortVersion, err := common.FindTarballInfo(fileName)
	if err != nil {
		return TarballDescription{}, err
	}
	if description.Version == "" {
		description.Version = version
	}
	if description.ShortVersion == "" {
		description.ShortVersion = shortVersion
	}
	if description.Flavor == "" {
		description.Flavor = flavor
	}
	if description.OperatingSystem == "" {
		description.OperatingSystem = strings.Title(runtime.GOOS)
	}
	description.Name = common.BaseName(fileName)

	return description, nil
}

func checkRemoteUrl(remoteUrl string) (int64, error) {
	// #nosec G107
	resp, err := http.Get(remoteUrl)
	if err != nil {
		return 0, fmt.Errorf("[checkRemoteUrl] error getting %s: %s", remoteUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("[checkRemoteUrl] received code %d ", resp.StatusCode)
	}
	var size int64 = 0
	for key := range resp.Header {
		if key == "Content-Length" && len(resp.Header[key]) > 0 {
			size, _ = strconv.ParseInt(resp.Header[key][0], 10, 0)
		}
	}
	return size, nil
}

func CheckTarballList(tarballList []TarballDescription) error {
	uniqueNames := make(map[string]bool)
	uniqueCombinations := make(map[string]bool)
	for _, tb := range tarballList {
		key := fmt.Sprintf("%s-%s-%s-%v", tb.OperatingSystem, tb.Flavor, tb.Version, tb.Minimal)

		// Makes sure that we don't have duplicate names in the list
		_, seen := uniqueNames[tb.Name]
		if seen {
			return fmt.Errorf("tarball name %s listed more than once", tb.Name)
		}
		uniqueNames[tb.Name] = true

		// Makes sure that we don't have duplicate combinations of OS+Flavor+Version+Minimal in the list
		_, seen = uniqueCombinations[key]
		if seen {
			return fmt.Errorf("tarball with OS %s, flavor %s, version %s, and minimal %v listed more than once",
				tb.OperatingSystem, tb.Flavor, tb.Version, tb.Minimal)
		}
		uniqueCombinations[key] = true
	}
	return nil
}
