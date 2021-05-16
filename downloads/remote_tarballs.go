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
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"jaytaylor.com/html2text"
)

type TarballDef struct {
	Flavor      string
	NameInFile  string
	NameInUrl   string
	DownloadDir string
	Versions    []string
}
type TarballType string

const (
	// Recognized Operating systems

	OsLinux = "Linux"
	OsMacOs = "MacOs"

	userAgentLinux  = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:59.0) Gecko/20100101 Firefox/59.0"
	userAgentMacOs  = "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (K HTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"
	basePageUrl     = "https://dev.mysql.com/downloads"
	baseDownloadUrl = "https://dev.mysql.com/get/Downloads"

	// Tarball types

	TtMysql   TarballType = "mysql"
	TtCluster TarballType = "cluster"
	TtShell   TarballType = "shell"
)

var downloadsSettings = map[TarballType]TarballDef{
	TtCluster: {
		Flavor:      "ndb",
		NameInFile:  "mysql-cluster",
		NameInUrl:   "cluster",
		DownloadDir: "MySQL-Cluster-VERSION",
		Versions:    []string{"7.6", "8.0"},
	},
	TtMysql: {
		Flavor:      "mysql",
		NameInFile:  "mysql",
		NameInUrl:   "mysql",
		DownloadDir: "MySQL-VERSION",
		Versions:    []string{"5.7", "8.0"},
	},
	TtShell: {
		Flavor:      "shell",
		NameInFile:  "mysql-shell",
		NameInUrl:   "shell",
		DownloadDir: "MySQL-Shell-VERSION",
		Versions:    []string{"8.0"},
	},
}

var userAgents = map[string]string{
	OsLinux: userAgentLinux,
	OsMacOs: userAgentMacOs,
}
var internalOsName = map[string]string{
	OsLinux: "Linux",
	OsMacOs: "Darwin",
}
var osNormalize = map[string]string{
	OsLinux:  OsLinux,
	"linux":  OsLinux,
	OsMacOs:  OsMacOs,
	"macos":  OsMacOs,
	"macOs":  OsMacOs,
	"macOS":  OsMacOs,
	"macOSX": OsMacOs,
	"Macos":  OsMacOs,
	"osx":    OsMacOs,
	"OSX":    OsMacOs,
	"osX":    OsMacOs,
	"darwin": OsMacOs,
	"Darwin": OsMacOs,
}

func validateTarballRequest(tarballType TarballType, version string, OS string) (string, error) {
	osText, ok := osNormalize[OS]
	if !ok {
		return "", fmt.Errorf("unrecognized OS %s: it must be one of [%s, %s] ", OS, OsMacOs, OsLinux)
	}
	OS = osText
	_, foundOs := userAgents[OS]
	if !foundOs {
		return "", fmt.Errorf("unrecognized OS %s: it must be one of [%s, %s] ", OS, OsMacOs, OsLinux)
	}
	settings, foundSettings := downloadsSettings[tarballType]
	if !foundSettings {
		return "", fmt.Errorf("unrecognized tarball type %s: it must be one of [%s, %s, %s] ", OS, TtMysql, TtCluster, TtShell)
	}
	acceptedVersion := false
	for _, v := range settings.Versions {
		if v == version {
			acceptedVersion = true
		}
	}
	if !acceptedVersion {
		return "", fmt.Errorf("version '%s' is not accepted for tarball type %s: it must be one of [%v] ", version, tarballType, settings.Versions)
	}
	return OS, nil
}

func GetRemoteTarballList(tarballType TarballType, version, OS string, withSize bool) ([]TarballDescription, error) {
	var err error
	OS, err = validateTarballRequest(tarballType, version, OS)
	if err != nil {
		return nil, err
	}

	pageUrl := basePageUrl + "/" + downloadsSettings[tarballType].NameInUrl + "/" + version + ".html"

	userAgent, found := userAgents[OS]
	if !found {
		return nil, fmt.Errorf("unrecognized OS %s: it must be one of [%s, %s] ", OS, OsMacOs, OsLinux)
	}

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet, pageUrl, nil)
	req.Header.Add("User-Agent", userAgent)

	response, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("GET %s: %s", pageUrl, err)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}
	htmlText := string(body)

	text, err := html2text.FromString(htmlText, html2text.Options{PrettyTables: true})
	if err != nil {
		return nil, err
	}

	var result []TarballDescription
	/*
		reLine := regexp.MustCompile(fmt.Sprintf(
			`\(` +                                   // An open parenthesis
				`(` +                                // start capture
				downloads[tarballType].NameInFile +  // the tarball type (mysql | cluster | shell)
				`-\d+\.\d+\.\d+` +                   // A version
				`[^\)]+` +                           // more characters except a closed parenthesis
				`x86.64`                             // x86-64 or x86_64
				`[^\)]+` +                           // more characters except a closed parenthesis
				`z)` +                               // a "z" (as in .gz or .xz) followed by a end capture
				`\)` +                               // a closed parenthesis
				`.*?MD5:` +                          // any character up to "MD5:"
				`\s*` +                              // optional spaces
				`(\w+)`))                            // Capture a string of alphanumeric characters (the checksum)
	*/
	reLine := regexp.MustCompile(fmt.Sprintf(`\((%s-\d+\.\d+\.\d+[^\)]+x86.64[^\)]+z)\).*?MD5:\s*(\w+)`, downloadsSettings[tarballType].NameInFile))
	matches := reLine.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no %s tarballs found for %s", tarballType, version)
	}

	reVersions := regexp.MustCompile(`((\d+\.\d+)\.\d+)`)
	for _, m := range matches {
		versionList := reVersions.FindAllStringSubmatch(m[1], 1)
		if len(versionList) == 0 || len(versionList[0]) < 2 {
			return nil, fmt.Errorf("error extracting version from file name '%s'", m[1])
		}
		longVersion := versionList[0][1]
		shortVersion := versionList[0][2]

		tbd := TarballDescription{
			Name:            m[1],
			Checksum:        "MD5:" + m[2],
			OperatingSystem: internalOsName[OS],
			ShortVersion:    shortVersion,
			Version:         longVersion,
			Flavor:          downloadsSettings[tarballType].Flavor,
			Minimal:         strings.Contains(m[1], "minimal"),
			Url: baseDownloadUrl + "/" +
				strings.Replace(downloadsSettings[tarballType].DownloadDir, "VERSION", version, 1) +
				"/" + m[1],
		}
		if withSize {
			tbd.Size, _ = checkRemoteUrl(tbd.Url)
		}
		result = append(result, tbd)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no tarballs found")
	}
	return result, nil
}
