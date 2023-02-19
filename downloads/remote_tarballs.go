// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
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
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/datacharmer/dbdeployer/common"
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

	preferredUserAgentLinux           = "dbdeployer/XXXX (X11; Linux; rv:XXXX) dbdeployer/XXXX"
	preferredUserAgentMacOs           = "dbdeployer/XXXX (Macintosh; Mac OS; rv:XXXX) dbdeployer/XXXX"
	alternativeUserAgentLinuxTemplate = "Mozilla/5.0 (X11; Linux ; rv:XXXX.0) Gecko/20100101 Firefox/XXXX.0"
	alternativeUserAgentMacOsTemplate = "Mozilla/5.0 (Macintosh; Mac OS; rv:XXXX.0) Gecko/20100101 Firefox/XXXX.0"
	basePageUrl                       = "https://dev.mysql.com/downloads"
	baseDownloadUrl                   = "https://dev.mysql.com/get/Downloads"

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

var (
	userAgentLinux = strings.Replace(preferredUserAgentLinux, "XXXX", common.VersionDef, -1)
	userAgentMacOs = strings.Replace(preferredUserAgentMacOs, "XXXX", common.VersionDef, -1)

	thisYear = fmt.Sprintf("%d", time.Now().Year())
	// Alternative user agents
	alternativeUserAgentMacOs = strings.Replace(alternativeUserAgentMacOsTemplate, "XXXX", thisYear, -1)
	alternativeUserAgentLinux = strings.Replace(alternativeUserAgentLinuxTemplate, "XXXX", thisYear, -1)
)

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
	if _, foundOs := userAgents[osText]; !foundOs {
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

func GetRemoteTarballList(tarballType TarballType, version, OS string, withSize, alternativeUserAgent bool) ([]TarballDescription, error) {
	var err error
	OS, err = validateTarballRequest(tarballType, version, OS)
	if err != nil {
		return nil, err
	}

	pageUrl := basePageUrl + "/" + downloadsSettings[tarballType].NameInUrl + "/" + version + ".html"

	if alternativeUserAgent {
		userAgents[OsLinux] = alternativeUserAgentLinux
		userAgents[OsMacOs] = alternativeUserAgentMacOs
	}

	userAgent, foundOs := userAgents[OS]
	if !foundOs {
		return nil, fmt.Errorf("unrecognized OS %s: it must be one of [%s, %s] ", OS, OsMacOs, OsLinux)
	}
	if common.IsEnvSet("SBDEBUG") {
		fmt.Printf("user agent: %s\n", userAgent)
	}

	client := &http.Client{}

	var matches [][]string
	var result []TarballDescription
	reVersions := regexp.MustCompile(`((\d+\.\d+)\.\d+)`)
	/*
		reLine := regexp.MustCompile(fmt.Sprintf(
			`\(` +                                   // An open parenthesis
				`(` +                                // start capture
				downloads[tarballType].NameInFile +  // the tarball type (mysql | cluster | shell)
				`-\d+\.\d+\.\d+` +                   // A version
				`[^\)]+` +                           // more characters except a closed parenthesis
				`\S+64`                             // x86-64 or x86_64 or amd64
				`[^\)]+` +                           // more characters except a closed parenthesis
				`z)` +                               // a "z" (as in .gz or .xz) followed by a end capture
				`\)` +                               // a closed parenthesis
				`.*?MD5:` +                          // any character up to "MD5:"
				`\s*` +                              // optional spaces
				`(\w+)`))                            // Capture a string of alphanumeric characters (the checksum)
	*/
	reLine := regexp.MustCompile(fmt.Sprintf(`\((%s-\d+\.\d+\.\d+[^\)]+-(\S+64)[^\)]+z)\).*?MD5:\s*(\w+)`, downloadsSettings[tarballType].NameInFile))
	req, _ := http.NewRequest(http.MethodGet, pageUrl, nil)
	req.Header.Add("User-Agent", userAgent)

	response, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("GET %s: %s", pageUrl, err)
	}

	body, err := io.ReadAll(response.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("[GetRemoteTarballList] error closing response body: %s", err)
		}
	}(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}
	htmlText := string(body)

	text, err := html2text.FromString(htmlText, html2text.Options{PrettyTables: true})
	if err != nil {
		return nil, err
	}

	matches = reLine.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		if common.IsEnvSet("SBDEBUG") {
			fmt.Printf("With user agent '%s'\n", userAgent)
			fmt.Printf("Text retrieved from server: %s\n", basePageUrl)
			fmt.Println(text)
		} else {
			return nil, fmt.Errorf("no %s tarballs found for %s - set environment variable 'SBDBUG' to see failure details", tarballType, version)
		}
		return nil, fmt.Errorf("no %s tarballs found for %s", tarballType, version)
	}

	for _, m := range matches {
		versionList := reVersions.FindAllStringSubmatch(m[1], 1)
		if len(versionList) == 0 || len(versionList[0]) < 2 {
			return nil, fmt.Errorf("error extracting version from file name '%s'", m[1])
		}
		longVersion := versionList[0][1]
		shortVersion := versionList[0][2]

		tbd := TarballDescription{
			Name:            m[1],
			Checksum:        "MD5:" + m[3],
			OperatingSystem: internalOsName[OS],
			Arch:            archNormalize(m[2]),
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

func archNormalize(s string) string {
	if s == "x86_64" {
		return "amd64"
	}
	return s
}
