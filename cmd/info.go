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

package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
)

func displayDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1,
			"'defaults' requires a label",
			"Example: dbdeployer info defaults master-slave-base-port")
	}
	label := args[0]
	defaultsMap := defaults.DefaultsToMap()
	value, ok := defaultsMap[label]
	if ok {
		fmt.Println(value)
	} else {
		fmt.Printf("# ERROR: field %s not found in defaults\n", label)
	}
}

func displayAllVersions(basedir, wantedVersion, flavor string) {
	result := ""

	reShortVersion := regexp.MustCompile(`(\d+\.\d+)`)
	wantedVersionList := reShortVersion.FindAllStringSubmatch(wantedVersion, -1)

	wantedShortVersion := ""
	if len(wantedVersionList) > 0 && len(wantedVersionList[0]) > 0 {
		wantedShortVersion = wantedVersionList[0][1]
	}
	var versionInfoList []common.VersionInfo = common.GetVersionInfoFromDir(basedir)
	for _, verInfo := range versionInfoList {
		versionList, err := common.VersionToList(verInfo.Version)
		if err != nil {
			common.Exitf(1, "error retrieving version list from %s", verInfo.Version)
		}
		shortVersion := fmt.Sprintf("%d.%d", versionList[0], versionList[1])
		if wantedShortVersion == shortVersion || strings.ToLower(wantedVersion) == "all" {
			if verInfo.Flavor == flavor {
				if result != "" {
					result += " "
				}
				result += verInfo.Version
			}
		}
	}
	if result != "" {
		fmt.Println(result)
	}
}

func displayVersion(cmd *cobra.Command, args []string) {
	wantedVersion := ""
	allVersions := ""
	if len(args) > 0 {
		wantedVersion = args[0]
	}

	reNotFound := regexp.MustCompile(globals.VersionNotFound)
	if len(args) > 1 {
		allVersions = args[1]
	}
	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	showEarliest, _ := cmd.Flags().GetBool(globals.EarliestLabel)
	if flavor == "" {
		flavor = common.MySQLFlavor
	}
	if strings.ToLower(allVersions) == "all" {
		basedir, err := getAbsolutePathFromFlag(cmd, "sandbox-binary")
		common.ErrCheckExitf(err, 1, "error getting absolute path for 'sandbox-binary'")
		displayAllVersions(basedir, wantedVersion, flavor)
	} else {
		if strings.ToLower(wantedVersion) == "all" {
			result := ""
			for _, v := range globals.SupportedAllVersions {
				latest := common.GetLatestVersion(defaults.Defaults().SandboxBinary, v, flavor)
				if !reNotFound.MatchString(latest) {
					if result != "" {
						result += " "
					}
					result += latest
				}
			}
			if result != "" {
				fmt.Println(result)
			}
		} else {
			var result string
			if showEarliest {
				result = common.GetEarliestVersion(defaults.Defaults().SandboxBinary, wantedVersion, flavor)
			} else {
				result = common.GetLatestVersion(defaults.Defaults().SandboxBinary, wantedVersion, flavor)
			}
			if !reNotFound.MatchString(result) {
				fmt.Println(result)

			}
		}
	}
}

func displayReleaseStats(release rest.DbdeployerRelease, total int64) int64 {
	var relTotal int64
	fmt.Printf("%s\n", globals.DashLine)
	fmt.Printf("Release:        %-30s %s\n", release.Name, release.PublishedAt)
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "sha256") {
			continue
		}
		fmt.Printf("\t%-50s %5d\n", asset.Name, asset.DownloadCount)
		total += asset.DownloadCount
		relTotal += asset.DownloadCount
	}
	fmt.Println()
	fmt.Printf("\t%-50s %5d\n", "total", relTotal)
	return total
}

func displayRelease(release rest.DbdeployerRelease) {
	fmt.Printf("%s\n", globals.DashLine)
	fmt.Printf("Remote version: %s\n", release.TagName)
	fmt.Printf("Release:        %s\n", release.Name)
	fmt.Printf("Date:           %s\n", release.PublishedAt)
	fmt.Printf("%s\n", release.Body)
	fmt.Printf("%s\n", globals.DashLine)
	for _, asset := range release.Assets {
		fmt.Printf("\t%s (%s)\n", asset.Name, humanize.Bytes(uint64(asset.Size)))
	}
}

func displayReleases(cmd *cobra.Command, args []string) {
	limit, _ := cmd.Flags().GetInt(globals.LimitLabel)
	raw, _ := cmd.Flags().GetBool(globals.RawLabel)
	stats, _ := cmd.Flags().GetBool(globals.StatsLabel)
	tag := ""
	var releases []rest.DbdeployerRelease
	var err error
	if len(args) > 0 {
		tag = args[0]
	}

	if tag == "" {
		releases, err = rest.GetReleases()
		if err != nil {
			common.Exitf(1, "error retrieving releases: %s", err)
		}
	} else {
		release, err := rest.GetLatestRelease(tag)
		if err != nil {
			common.Exitf(1, "error retrieving release %s: %s", tag, err)
		}
		releases = append(releases, release)
	}
	if stats && raw {
		common.Exitf(1, "only one of flags --%s or --%s is allowed", globals.StatsLabel, globals.RawLabel)
	}
	if stats {
		var total int64
		for i, r := range releases {
			done := i + 1
			total = displayReleaseStats(r, total)
			if limit > 0 && done >= limit {
				break
			}
		}
		fmt.Printf("%s\n", globals.DashLine)
		fmt.Printf("TOTAL downloawds: %d\n", total)
		return
	}
	if raw {
		jsonText, err := json.MarshalIndent(releases, " ", " ")
		if err != nil {
			common.Exitf(1, "error encoding JSON releases : %s", err)
		}
		fmt.Printf("%s\n", jsonText)
		return
	}
	for i, r := range releases {
		done := i + 1
		displayRelease(r)
		fmt.Println()
		if limit > 0 && done >= limit {
			break
		}
	}
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Shows information about dbdeployer environment samples",
	Long:  `Shows current information about defaults and environment.`,
}

var infoDefaultsCmd = &cobra.Command{
	Use: "defaults field-name",

	Short: "displays a defaults value",
	Example: `
	$ dbdeployer info defaults master-slave-base-port 
`,
	Long:        `Displays one field of the defaults.`,
	Run:         displayDefaults,
	Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
}

var infoReleaseCmd = &cobra.Command{
	Use: "releases [tag]",

	Short: "displays info on releases, or a given release",
	Example: `
	$ dbdeployer info releases
	$ dbdeployer info releases v1.35.0
	$ dbdeployer info releases latest
`,
	Long: `Displays info on all the available releases, or a specific one`,
	Run:  displayReleases,
}

var infoVersionCmd = &cobra.Command{
	Use: "version [short-version|all] [all]",

	Short: "displays the latest version available",
	Example: `
    # Shows the latest version available
    $ dbdeployer info version
    8.0.16

    # shows the latest version belonging to 5.7
    $ dbdeployer info version 5.7
    5.7.26

    # shows the latest version for every short version
    $ dbdeployer info version all
    5.0.96 5.1.73 5.5.53 5.6.41 5.7.26 8.0.16

    # shows all the versions for a given short version
    $ dbdeployer info version 8.0 all
    8.0.11 8.0.12 8.0.13 8.0.14 8.0.15 8.0.16
`,
	Long: `Displays the latest version available for deployment.
If a short version is indicated (such as 5.7, or 8.0), only the versions belonging to that short
version are searched.
If "all" is indicated after the short version, displays all versions belonging to that short version.
`,
	Run:         displayVersion,
	Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.AddCommand(infoDefaultsCmd)
	infoCmd.AddCommand(infoVersionCmd)
	infoCmd.AddCommand(infoReleaseCmd)
	setPflag(infoCmd, globals.FlavorLabel, "", "", "", "For which flavor this info is", false)
	infoCmd.PersistentFlags().Bool(globals.EarliestLabel, false, "Return the earliest version")
	infoReleaseCmd.PersistentFlags().Int(globals.LimitLabel, 3, "Limit number of releases to show (0 = unlimited)")
	infoReleaseCmd.PersistentFlags().Bool(globals.RawLabel, false, "Show the original data")
	infoReleaseCmd.PersistentFlags().Bool(globals.StatsLabel, false, "Show downloads statistics")
}
