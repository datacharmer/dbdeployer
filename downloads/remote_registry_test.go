// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package downloads

import (
	"fmt"
	"strings"
	"testing"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/compare"
)

type boolMap map[bool]string
type VersionCollectionInfo struct {
	foundVersions         []string
	requestedShortVersion string
	expected              boolMap
}

func getShortVersion(s string) string {

	pieces := strings.Split(s, ".")
	return fmt.Sprintf("%s.%s", pieces[0], pieces[1])
}

func makeTarballCollection(info VersionCollectionInfo) []TarballDescription {
	var tbd []TarballDescription

	for _, v := range info.foundVersions {
		tbd = append(tbd, TarballDescription{
			Name:            "mysql-" + v,
			Checksum:        "",
			OperatingSystem: "linux",
			Url:             "",
			Flavor:          common.MySQLFlavor,
			Minimal:         false,
			Size:            0,
			ShortVersion:    getShortVersion(v),
			Version:         v,
			UpdatedBy:       "",
			Notes:           "",
		})
	}

	return tbd
}

func TestFindOrGuessTarballByVersionFlavorOS(t *testing.T) {

	var versionCollections = map[string][]string{
		"8.0":     []string{"8.0.19", "8.0.20", "8.0.22", "8.0.23"},
		"5.6":     []string{"5.6.31", "5.6.33"},
		"5.7":     []string{"5.7.31", "5.7.33"},
		"5.7-8.0": []string{"5.7.31", "5.7.33", "8.0.19", "8.0.20", "8.0.22", "8.0.23"},
	}
	var versionCollectionData = []VersionCollectionInfo{
		{
			foundVersions:         versionCollections["8.0"],
			requestedShortVersion: "8.0",
			expected:              boolMap{true: "8.0.24", false: "8.0.23"},
		},
		{
			foundVersions:         versionCollections["5.7-8.0"],
			requestedShortVersion: "8.0",
			expected:              boolMap{true: "8.0.24", false: "8.0.23"},
		},
		{
			foundVersions:         versionCollections["5.7-8.0"],
			requestedShortVersion: "5.7",
			expected:              boolMap{true: "5.7.34", false: "5.7.33"},
		},
		{
			foundVersions:         versionCollections["5.7"],
			requestedShortVersion: "5.7",
			expected:              boolMap{true: "5.7.34", false: "5.7.33"},
		},
		{
			foundVersions:         versionCollections["8.0"],
			requestedShortVersion: "5.7",
			expected:              boolMap{true: "", false: ""},
		},
		{
			foundVersions:         versionCollections["5.7"],
			requestedShortVersion: "8.0",
			expected:              boolMap{true: "", false: ""},
		},
		{
			foundVersions:         versionCollections["5.6"],
			requestedShortVersion: "8.0",
			expected:              boolMap{true: "", false: ""},
		},
		{
			foundVersions:         versionCollections["5.6"],
			requestedShortVersion: "5.7",
			expected:              boolMap{true: "", false: ""},
		},
		{
			foundVersions:         versionCollections["5.6"],
			requestedShortVersion: "5.6",
			expected:              boolMap{true: "", false: "5.6.33"},
		},
	}
	saveTarballCollection := DefaultTarballRegistry.Tarballs

	for _, data := range versionCollectionData {
		tbd := makeTarballCollection(data)
		DefaultTarballRegistry.Tarballs = tbd

		for guess, expected := range data.expected {

			tb, _ := FindOrGuessTarballByVersionFlavorOS(
				data.requestedShortVersion,
				common.MySQLFlavor,
				"linux", "amd64", false, !guess, guess)
			label := fmt.Sprintf("versions %s - requested '%s' - guess '%v'",
				data.foundVersions,
				data.requestedShortVersion,
				guess)
			compare.OkEqualString(
				label,
				tb.Version, expected, t)
		}

	}

	DefaultTarballRegistry.Tarballs = saveTarballCollection
}

func TestTarballRegistry(t *testing.T) {

	for _, tarball := range DefaultTarballRegistry.Tarballs {
		size, err := checkRemoteUrl(tarball.Url)
		if err != nil {
			t.Logf("not ok - tarball %s check failed: %s", tarball.Name, err)
			t.Fail()
		} else {
			t.Logf("ok - tarball %s found", tarball.Name)
			if size == 0 {
				t.Logf("not ok - size 0 for tarball %s", tarball.Name)
			}
		}
	}
}
