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

package common

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/pkg/errors"

	"github.com/datacharmer/dbdeployer/globals"
)

type SandboxInfo struct {
	SandboxName string
	SandboxDesc SandboxDescription
	Locked      bool
}

type VersionInfo struct {
	Version string
	Flavor  string
}

var portDebug bool = IsEnvSet("PORT_DEBUG")

type PortMap map[int]bool

type SandboxInfoList []SandboxInfo

func GetFullSandboxInfo(sandboxHome string) SandboxInfoList {
	var fullSandboxList SandboxInfoList
	simpleSandboxList, err := GetInstalledSandboxes(sandboxHome)
	if err != nil {
		return fullSandboxList
	}

	for _, sb := range simpleSandboxList {
		sbDescription := path.Join(sandboxHome, sb.SandboxName, globals.SandboxDescriptionName)
		var tempSandboxDesc SandboxDescription
		if FileExists(sbDescription) {
			tempSandboxDesc, _ = ReadSandboxDescription(path.Join(sandboxHome, sb.SandboxName))
		}

		fullSandboxList = append(fullSandboxList,
			SandboxInfo{SandboxName: sb.SandboxName, SandboxDesc: tempSandboxDesc, Locked: sb.Locked})
	}
	return fullSandboxList
}

func GetSandboxesByDate(sandboxHome string) (SandboxInfoList, error) {
	sandboxList := GetFullSandboxInfo(sandboxHome)
	if len(sandboxList) == 0 {
		return sandboxList, nil
	}
	var errors = make(map[string]int)
	sort.SliceStable(sandboxList, func(i, j int) bool {
		iDate, err := dateparse.ParseStrict(sandboxList[i].SandboxDesc.Timestamp)
		if err != nil {
			if _, ok := errors[sandboxList[i].SandboxDesc.Timestamp]; !ok {
				errors[sandboxList[i].SandboxDesc.Timestamp] = 0
			}
			errors[sandboxList[i].SandboxDesc.Timestamp] += 1
			return false
		}
		jDate, err := dateparse.ParseStrict(sandboxList[j].SandboxDesc.Timestamp)
		if err != nil {
			if _, ok := errors[sandboxList[j].SandboxDesc.Timestamp]; !ok {
				errors[sandboxList[j].SandboxDesc.Timestamp] = 0
			}
			errors[sandboxList[j].SandboxDesc.Timestamp] += 1
			return false
		}
		return iDate.UnixNano() < jDate.UnixNano()
	})
	if len(errors) > 0 {
		return nil, fmt.Errorf("errors during date conversions %v", errors)
	}

	return sandboxList, nil
}

// Returns a list of inner sandboxes
func SandboxInfoToFileNames(sbList []SandboxInfo) (fileNames []string) {
	for _, sbinfo := range sbList {
		fileNames = append(fileNames, sbinfo.SandboxName)
	}
	return
}

// Returns the list of versions of a given flavor, or an empty list if no
// versions were available for that flavor
func GetFlavoredVersionsFromDir(basedir, flavor string) []string {
	var versions []string

	tempVersions := GetVersionInfoFromDir(basedir)
	for _, ver := range tempVersions {
		if ver.Flavor == flavor {
			versions = append(versions, ver.Version)
		}
	}

	return versions
}

func GetVersionInfoFromDir(basedir string) []VersionInfo {
	var versionInfos []VersionInfo

	tempVersions, _ := GetVersionsFromDir(basedir)
	for _, ver := range tempVersions {

		dirPath := path.Join(basedir, ver)
		flavorPath := path.Join(dirPath, "FLAVOR")
		var foundFlavor string = MySQLFlavor
		if FileExists(flavorPath) {
			foundFlavor, _ = SlurpAsString(flavorPath)
			foundFlavor = strings.TrimSpace(foundFlavor)
		} else {
			foundFlavor = DetectBinaryFlavor(dirPath)
		}
		versionInfos = append(versionInfos, VersionInfo{Version: ver, Flavor: foundFlavor})
	}
	return versionInfos
}

func GetCompatibleClientVersion(basedir, serverVersion string) (string, error) {
	var versionInfos []VersionInfo
	var versionList []string
	versionInfos = GetVersionInfoFromDir(basedir)

	// Clients of these flavors should work well with imported sandboxes
	compatibleFlavors := map[string]bool{
		MySQLFlavor:         true,
		PerconaServerFlavor: true,
	}
	serverVersionList, err := VersionToList(serverVersion)
	if err != nil {
		return globals.EmptyString, err
	}
	for _, v := range versionInfos {
		_, ok := compatibleFlavors[v.Flavor]
		if ok {
			versionList = append(versionList, v.Version)
		}
	}
	sortedVersionList := SortVersions(versionList)

	for _, v := range sortedVersionList {
		if v == serverVersion {
			return v, nil
		}
		isCompatible, err := GreaterOrEqualVersion(v, serverVersionList)
		if err != nil {
			return globals.EmptyString, err
		}
		if isCompatible {
			return v, nil
		}
	}

	// No suitable version so far.
	// We try to get the closest one in the same short version
	serverVersionDigits, err := VersionToList(serverVersion)
	if err != nil {
		return globals.EmptyString, err
	}
	shortVersion := fmt.Sprintf("%d.%d", serverVersionDigits[0], serverVersionDigits[1])
	latestVersion := GetLatestVersion(basedir, shortVersion, MySQLFlavor)
	if !strings.Contains(latestVersion, globals.VersionNotFound) {
		return latestVersion, nil
	}

	return globals.EmptyString, fmt.Errorf("no suitable client version found for version %s", serverVersion)
}

// Returns the list of versions available for deployment
func GetVersionsFromDir(basedir string) ([]string, error) {
	var dirs []string
	files, err := os.ReadDir(basedir)
	if err != nil {
		return dirs, fmt.Errorf("error reading directory %s: %s", basedir, err)
	}
	for _, f := range files {
		fname := f.Name()
		if f.IsDir() {
			mysqld := path.Join(basedir, fname, "bin", "mysqld")
			mysqldDebug := path.Join(basedir, fname, "bin", "mysqld-debug")
			tidb := path.Join(basedir, fname, "bin", "tidb-server")
			if FileExists(mysqld) || FileExists(mysqldDebug) || FileExists(tidb) {
				dirs = append(dirs, fname)
			}
		}
	}
	return dirs, nil
}

func GetAvailableVersions() ([]string, error) {
	basedir := os.Getenv("SANDBOX_BINARY")
	if basedir == "" {
		return []string{}, fmt.Errorf("variable SANDBOX_BINARY not set")
	}
	return GetVersionsFromDir(basedir)
}

// FindSandbox will find a wanted sandbox from a list of sandboxes
// It will get the sandbox by name, if it matches.
// If not, it will try to match by version, port, type, or flavor, provided
// that the wanted parameter is unique among the sandboxes.
// For example, if we ask for version 8.0.19 and only one sandbox uses such
// version, we return a match. If two or more sandboxes use that version,
// there is no match. The same applies for the other criteria.
//
// FOR FUTURE USAGE with `dbdeployer delete` or `dbdeployer run sandboxIdentifier`
func FindSandbox(sandboxList []SandboxInfo, wanted string) (SandboxInfo, error) {

	type match struct {
		occurrences int
		info        SandboxInfo
	}
	var versions = make(map[string]match)
	var flavors = make(map[string]match)
	var types = make(map[string]match)
	var ports = make(map[string]match)
	var setMatch = func(input map[string]match, s string, sb SandboxInfo) {
		item, ok := input[s]
		if ok {
			item.occurrences++
			item.info = SandboxInfo{}
		} else {
			item.occurrences = 1
			item.info = sb
		}
		input[s] = item
	}
	for _, sb := range sandboxList {
		if sb.SandboxName == wanted {
			return sb, nil
		}

		setMatch(versions, sb.SandboxDesc.Version, sb)

		setMatch(flavors, sb.SandboxDesc.Flavor, sb)

		setMatch(types, sb.SandboxDesc.SBType, sb)

		for _, port := range sb.SandboxDesc.Port {
			portStr := fmt.Sprintf("%d", port)
			setMatch(ports, portStr, sb)
		}
	}

	var checkMatch = func(input map[string]match) (SandboxInfo, bool) {
		for k, v := range input {
			if k == wanted && v.occurrences == 1 && v.info.SandboxName != "" {
				return v.info, true
			}
		}
		return SandboxInfo{}, false
	}

	allInput := []map[string]match{flavors, versions, types, ports}

	for _, input := range allInput {
		sb, found := checkMatch(input)
		if found {
			return sb, nil
		}
	}

	return SandboxInfo{}, fmt.Errorf("sandbox '%s' not found", wanted)
}

// Gets a list of installed sandboxes from the $SANDBOX_HOME directory
func GetInstalledSandboxes(sandboxHome string) (installedSandboxes SandboxInfoList, err error) {
	if !DirExists(sandboxHome) {
		return installedSandboxes, fmt.Errorf("directory SandboxHome not found")
	}
	files, err := os.ReadDir(sandboxHome)
	if err != nil {
		return installedSandboxes, err
	}
	for _, f := range files {
		fname := f.Name()
		if f.IsDir() {
			if fname == globals.ForbiddenDirName {
				continue
			}
			sbdesc := path.Join(sandboxHome, fname, globals.SandboxDescriptionName)
			start := path.Join(sandboxHome, fname, "start")
			startAll := path.Join(sandboxHome, fname, "start_all")
			noClear := path.Join(sandboxHome, fname, "no_clear")
			noClearAll := path.Join(sandboxHome, fname, "no_clear_all")
			if FileExists(sbdesc) || FileExists(start) || FileExists(startAll) {
				description, _ := ReadSandboxDescription(path.Join(sandboxHome, fname))
				if FileExists(noClearAll) || FileExists(noClear) {
					installedSandboxes = append(installedSandboxes, SandboxInfo{SandboxName: fname, Locked: true, SandboxDesc: description})
				} else {
					installedSandboxes = append(installedSandboxes, SandboxInfo{SandboxName: fname, Locked: false, SandboxDesc: description})
				}
			}
		}
	}
	return
}

// Collects a list of used ports from deployed sandboxes
func GetInstalledPorts(sandboxHome string) ([]int, error) {
	installedSandboxes, err := GetInstalledSandboxes(sandboxHome)
	if err != nil {
		return []int{}, err
	}

	files := SandboxInfoToFileNames(installedSandboxes)
	// If there is a file sbdescription.json in the top directory
	// it will be included in the reporting
	files = append(files, "")
	var portCollection []int
	var seenPorts = make(map[int]bool)
	for _, fname := range files {
		sbdesc := path.Join(sandboxHome, fname, globals.SandboxDescriptionName)
		if FileExists(sbdesc) {
			sbd, err := ReadSandboxDescription(path.Join(sandboxHome, fname))
			if err != nil {
				return []int{}, errors.Wrap(err, "error reading sandbox description")
			}
			for _, p := range sbd.Port {
				if !seenPorts[p] {
					portCollection = append(portCollection, p)
					seenPorts[p] = true
				}
			}
			if sbd.Nodes != 0 {
				var nodeDescr []SandboxDescription
				innerInstalledSandboxes, err := GetInstalledSandboxes(filepath.Join(sandboxHome, fname))
				if err != nil {
					return []int{}, err
				}
				innerFiles := SandboxInfoToFileNames(innerInstalledSandboxes)
				for _, inner := range innerFiles {
					innerSbdesc := path.Join(sandboxHome, fname, inner, globals.SandboxDescriptionName)
					if FileExists(innerSbdesc) {
						sdNode, err := ReadSandboxDescription(fmt.Sprintf("%s/%s/%s", sandboxHome, fname, inner))
						if err != nil {
							return []int{}, errors.Wrapf(err, "error reading inner sandbox description %s/%s/%s", sandboxHome, fname, inner)
						}
						nodeDescr = append(nodeDescr, sdNode)
					}
				}
				for _, nd := range nodeDescr {
					for _, p := range nd.Port {
						if !seenPorts[p] {
							portCollection = append(portCollection, p)
							seenPorts[p] = true
						}
					}
				}
			}
		}
	}
	// CondPrintf("%v\n",port_collection)
	return portCollection, nil
}

func CheckFlavorSupport(flavor string) error {
	for _, sf := range FlavorCompositionList {
		if sf.flavor == flavor {
			return nil
		}
	}
	return fmt.Errorf("flavor '%s' is not supported", flavor)
}

// Tries to detect the database flavor from files in the tarball directory
func DetectBinaryFlavor(basedir string) string {
	for _, fi := range FlavorCompositionList {
		var matches bool
		if fi.AllNeeded {
			matches = len(fi.elements) > 0
		}

		for _, element := range fi.elements {
			target := path.Join(basedir, element.dir, element.fileName)
			if fi.AllNeeded {
				if !FileExists(target) {
					matches = false
				}
			} else {
				if FileExists(target) {
					matches = true
				}
			}
		}
		if matches {
			return fi.flavor
		}
	}
	return MySQLFlavor
}

/*
Checks that the extracted tarball directory

	contains one or more files expected for the current
	operating system.
	It prevents simple errors like :
	* using a Linux tarball on a Mac or vice-versa
	* using a source or test tarball instead of a binaries one.
*/
func CheckTarballOperatingSystem(basedir string) error {
	currentOs := runtime.GOOS
	// CondPrintf("<%s>\n",currentOs)
	type OSFinding struct {
		Dir      string
		OS       string
		flavor   string
		isBinary bool
	}
	var findingList = map[string]OSFinding{
		globals.FnLibMySQLClientA:             {"lib", "linux", MySQLFlavor, true}, // 4.1 and old 5.0 releases
		globals.FnLibMySQLClientSo:            {"lib", "linux", MySQLFlavor, true},
		globals.FnLibMariadbClientSo:          {"lib", "linux", MariaDbFlavor, true},
		globals.FnLibMariadbClientDylib:       {"lib", "linux", MariaDbFlavor, true},
		globals.FnLibPerconaServerClientSo:    {"lib", "linux", PerconaServerFlavor, true},
		globals.FnLibPerconaServerClientDylib: {"lib", "darwin", PerconaServerFlavor, true},
		globals.FnLibMySQLClientDylib:         {"lib", "darwin", MySQLFlavor, true},
		globals.FnTiDbServer:                  {"bin", "any", TiDbFlavor, true},
		globals.FnTableH:                      {"sql", "source", "any", false},
		globals.FnMysqlProvisionZip:           {"share/mysqlsh", "shell", "any", false},
	}
	wantedOsFound := false
	var foundList = make(map[string]OSFinding)
	var wantedFiles []string
	for fname, rec := range findingList {
		fullName := path.Join(basedir, rec.Dir, fname)
		if rec.OS == currentOs && rec.isBinary {
			wantedFiles = append(wantedFiles, path.Join(rec.Dir, fname))
		}
		if FileExists(fullName) {
			// TODO: This is a workaround to make TiDB work. Later refinements may come.
			if (rec.OS == currentOs || rec.OS == "any") && rec.isBinary {
				wantedOsFound = true
			}
			foundList[fname] = rec
		}
	}
	if !wantedOsFound {
		fmt.Println(globals.DashLine)
		CondPrintf("Looking for *%s* binaries\n", currentOs)
		CondPrintln("At least one of the following was needed")
		for key, value := range findingList {
			if value.OS == currentOs {
				CondPrintf("\t%s/%s (%s)\n", value.Dir, key, value.flavor)
			}
		}
		fmt.Println(globals.DashLine)
		if len(foundList) > 0 {
			CondPrintf("# Found the following:\n")
		}
		for fname, rec := range foundList {
			fullName := path.Join(basedir, rec.Dir, fname)
			CondPrintf("%-20s - tarball type: '%s' (flavor: %s)\n", fullName, rec.OS, rec.flavor)
			if rec.OS == "source" {
				CondPrintf("THIS IS A SOURCE TARBALL. YOU NEED TO USE A *BINARY* TARBALL\n")
			}
			fmt.Println(globals.DashLine)
		}
		return fmt.Errorf("could not find any of the expected files for %s server: %s\n%s", currentOs, wantedFiles, globals.DashLine)
	}
	return nil
}

// Returns true if the file name has a recognized tarball extension
// for use with dbdeployer
func IsATarball(fileName string) bool {
	if strings.HasSuffix(fileName, ".tar.gz") ||
		strings.HasSuffix(fileName, ".tar.xz") {
		return true
	}
	return false
}

// Checks the initial argument for a sandbox deployment
func CheckOrigin(args []string) {
	if len(args) < 1 {
		Exit(1, "this command requires the MySQL version (x.xx[.xx]) as argument ")
	}
	if len(args) > 1 {
		Exit(1, "extra argument detected. This command requires only the MySQL version (x.xx[.xx]) as argument ")
	}
	origin := args[0]
	if FileExists(origin) && IsATarball(origin) {
		Exit(1,
			"tarball detected. - If you want to use a tarball to create a sandbox,",
			"you should first use the 'unpack' command")
	}
}

// Creates a sandbox directory if it does not exist
func CheckSandboxDir(sandboxHome string) error {
	if !DirExists(sandboxHome) {
		CondPrintf("Creating directory %s\n", sandboxHome)
		return os.Mkdir(sandboxHome, globals.PublicDirectoryAttr)
	}
	return nil
}

// Returns true if a given string looks contains a version
// number (major.minor.rev)
func IsVersion(version string) bool {
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)$`)
	return re.MatchString(version)
}

// Return true if a given string is a valid URL
func IsUrl(s string) bool {
	_, err := url.ParseRequestURI(s)
	validUrl := err == nil

	// We only consider HTTP addresses
	reHttp := regexp.MustCompile(`^(http|https)://`)
	if !reHttp.MatchString(s) {
		validUrl = false
	}

	// If the file is the only part of the address,
	// without a path, we consider it invalid.
	stripped := reHttp.ReplaceAllString(s, "")
	base := BaseName(stripped)
	if stripped == base {
		validUrl = false
	}
	return validUrl
}

// Returns true if a given string looks like an IPV4
func IsIPV4(ip string) bool {
	l := strings.Split(ip, ".")
	if len(l) != 4 {
		return false
	}
	for _, ns := range l {
		N, err := strconv.Atoi(ns)
		if err != nil {
			return false
		}
		if N < 0 || N > 255 {
			return false
		}
	}
	return true
}

// Gets three integers for a version string
// Converts "1.2.3" into []int{1, 2, 3}
func VersionToList(version string) ([]int, error) {
	// A valid version must be made of 3 integers
	re1 := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	// Also valid version is 3 numbers with a prefix
	re2 := regexp.MustCompile(`^[^.0-9-]+(\d+)\.(\d+)\.(\d+)$`)
	verList1 := re1.FindAllStringSubmatch(version, -1)
	verList2 := re2.FindAllStringSubmatch(version, -1)
	verList := verList1
	//CondPrintf("%#v\n", verList)
	if verList == nil {
		verList = verList2
	}
	if verList == nil {
		return []int{-1}, fmt.Errorf("required version format: x.x.xx - Got '%s'", version)
	}

	var intList = make([]int, 3)
	newCount := 0
	for N, item := range verList[0] {
		if N == 0 {
			continue
		}
		intVal, err := strconv.Atoi(item)
		if err != nil {
			return []int{-1}, fmt.Errorf("(%d) error converting %s (list: %+v) [%s] ", N, version, verList, err)
		}
		intList[newCount] = intVal
		newCount++
	}
	return intList, nil
}

// Converts a version string into a name.
// Replaces dots with underscores. "1.2.3" -> "1_2_3"
func VersionToName(version string) string {
	re := regexp.MustCompile(`\.`)
	name := re.ReplaceAllString(version, "_")
	return name
}

// VersionToPort converts a version string into a port number
// e.g. "5.6.33" -> 5633
// Note that if the conversion exceeds the maximum port (65,535), such as 10.10.x,
// the actual number will be reduced by 60,000
func VersionToPort(version string) (int, error) {
	verList, err := VersionToList(version)
	if err != nil {
		return -1, fmt.Errorf("error converting %s into a version", version)
	}
	major := verList[0]
	minor := verList[1]
	rev := verList[2]

	completeVersion := fmt.Sprintf("%d%d%02d", major, minor, rev)
	i, err := strconv.Atoi(completeVersion)
	if err != nil {
		return -1, fmt.Errorf("error converting %d%d%02d to version", major, minor, rev)
	}
	for i > globals.MaxAllowedPort {
		i -= globals.ReductionOnPortNumberOverflow
	}

	if i < globals.MinAllowedPort {
		i += globals.MinAllowedPort
	}
	return i, nil
}

// Checks if a version string is greater or equal a given numeric version
// "5.6.33" >= []int{5,7,0}  = false
// "5.7.21" >= []int{5,7,0}  = true
// "10.1.21" >= []int{5,7,0}  = false (!)
// Note: MariaDB versions are skipped. The function returns false for MariaDB 10+.
// So far (2018-02-19) this comparison holds, because MariaDB behaves like 5.5+ for
// the purposes of sandbox deployment
//
// DEPRECATED as of 1.18.0
// Use GreaterOrEqualVersionList and flavors instead
func GreaterOrEqualVersion(version string, comparedTo []int) (bool, error) {
	if len(comparedTo) != 3 {
		return false, errors.Wrapf(fmt.Errorf("invalid slice size: %v", comparedTo), "GreaterOrEqualVersion:")
	}
	var compMajor, compMinor, compRev int = comparedTo[0], comparedTo[1], comparedTo[2]
	verList, err := VersionToList(version)
	if err != nil {
		return false, errors.Wrapf(err, "VersionToList")
	}
	major := verList[0]
	if major < 0 {
		return false, errors.Wrapf(err, "major < 0")
	}
	minor := verList[1]
	rev := verList[2]

	// TODO: MariaDB 10.4 has changed behavior with regards to the above assumptions - Needs some more work
	if major == 10 {
		return false, nil
	}
	versionText := fmt.Sprintf("%02d%02d%02d", major, minor, rev)
	compareText := fmt.Sprintf("%02d%02d%02d", compMajor, compMinor, compRev)
	return versionText >= compareText, nil
}

// Checks if a version list is greater or equal a given numeric version
// []int{5,6,33} >= []int{5,7,0}  = false
// []int{5,7,21} >= []int{5,7,0}  = true
// []int{10,1,21} >= []int{5.7.0}  = true
// Note: Use this function in combination with flavors.
// Better yet, use common.HasCapability(flavors, feature, version)

func GreaterOrEqualVersionList(verList, comparedTo []int) (bool, error) {
	lenVerList := len(verList)
	lenCompareTo := len(comparedTo)

	if lenCompareTo < 1 {
		return false, fmt.Errorf("comparison version empty")
	}
	if lenVerList < 1 {
		return false, fmt.Errorf("requested version empty")
	}
	maxElements := lenVerList
	if lenCompareTo < maxElements {
		maxElements = lenCompareTo
	}
	versionText := ""
	compareText := ""

	for N := 0; N < maxElements; N++ {
		versionText += fmt.Sprintf("%05d", verList[N])
		compareText += fmt.Sprintf("%05d", comparedTo[N])
	}
	return versionText >= compareText, nil
}

// Finds the first free port available, starting at
// requestedPort.
// usedPorts is a map of ports already used by other sandboxes.
// This function should not be used alone, but through FindFreePort.
// Returns the first free port
func findFreePortSingle(requestedPort int, usedPorts PortMap) (int, error) {
	foundPort := 0
	candidatePort := requestedPort
	for foundPort == 0 {
		_, exists := usedPorts[candidatePort]
		if exists {
			if portDebug {
				CondPrintf("- port %d not free\n", candidatePort)
			}
		} else {
			foundPort = candidatePort
		}
		candidatePort += 1
		if candidatePort > globals.MaxAllowedPort {
			return -1,
				fmt.Errorf("FATAL (findFreePortSingle): Could not find a free port starting at %d.\n"+
					"Maximum limit for port value (%d) reached", requestedPort, globals.MaxAllowedPort)
		}
	}
	return foundPort, nil
}

// Finds the a range of howMany free ports available, starting at
// basePort.
// usedPorts is a map of ports already used by other sandboxes.
// This function should not be used alone, but through FindFreePort.
// Returns the first port of the requested range
func findFreePortRange(basePort int, usedPorts PortMap, howMany int) (int, error) {
	var foundPort int = 0
	requestedPort := basePort
	candidatePort := requestedPort
	counter := 0
	for foundPort == 0 {
		numPorts := 0
		for counter < howMany {
			_, exists := usedPorts[candidatePort+counter]
			if exists {
				if portDebug {
					CondPrintf("- port %d is not free\n", candidatePort+counter)
				}
				candidatePort += 1
				counter = 0
				numPorts = 0
				continue
			} else {
				if portDebug {
					CondPrintf("+ port %d is free\n", candidatePort+counter)
				}
				numPorts += 1
			}
			counter++
			if candidatePort > globals.MaxAllowedPort {
				return -1, fmt.Errorf("FATAL (findFreePortRange): \n"+
					"Could not find a free range of %d ports starting at %d.\n"+
					"Maximum limit for port value (%d) reached", howMany, requestedPort, globals.MaxAllowedPort)
			}
		}
		if numPorts == howMany {
			foundPort = candidatePort
		} else {
			return -1, fmt.Errorf("FATAL: findFreePortRange should never reach this point\n"+
				"requested: %d - used: %v - candidate: %d", requestedPort, usedPorts, candidatePort)
		}
	}
	return foundPort, nil
}

// Finds the a range of howMany free ports available, starting at
// basePort.
// installedPorts is a slice of ports already used by other sandboxes.
// Calls either findFreePortRange or findFreePortSingle, depending on the
// amount of ports requested.
// Returns the first port of the requested range
func FindFreePort(basePort int, installedPorts []int, howMany int) (int, error) {
	if portDebug {
		CondPrintf("FindFreePort: requested: %d - used: %v - howMany: %d\n", basePort, installedPorts, howMany)
	}
	usedPorts := make(PortMap)

	for _, p := range installedPorts {
		usedPorts[p] = true
	}
	if howMany == 1 {
		return findFreePortSingle(basePort, usedPorts)
	}
	return findFreePortRange(basePort, usedPorts, howMany)
}

// Checks whether a version string is equal or greater than the current compatible version
func IsCompatibleVersion(versionStr string) bool {
	if !IsVersion(versionStr) {
		return false
	}

	compVersionList, err := VersionToList(CompatibleVersion)
	if err != nil {
		return false
	}
	versionList, err := VersionToList(versionStr)
	if err != nil {
		return false
	}
	isGreater, err := GreaterOrEqualVersionList(versionList, compVersionList)
	if err != nil {
		return false
	}
	return isGreater
}

func GetLatestVersion(sandboxBinary string, wantedVersion, flavor string) string {
	return getSortedVersion(sandboxBinary, wantedVersion, flavor, -1)
}

func GetEarliestVersion(sandboxBinary string, wantedVersion, flavor string) string {
	return getSortedVersion(sandboxBinary, wantedVersion, flavor, 0)
}

func getSortedVersion(sandboxBinary, wantedVersion, flavor string, position int) string {
	if wantedVersion == "" {
		wantedVersion = os.Getenv("WANTED_VERSION")
	}
	versions := GetFlavoredVersionsFromDir(sandboxBinary, flavor)
	if len(versions) == 0 {
		return globals.VersionNotFound + "_" + flavor
	}

	sortedVersions := SortVersionsSubset(versions, wantedVersion)
	if len(sortedVersions) < 1 {
		return globals.VersionNotFound + "_" + flavor
	}
	if position == -1 {
		position = len(sortedVersions) - 1
	}
	latestVersion := sortedVersions[position]
	return latestVersion
}

func FindTarballInfo(fileName string) (flavor, version, shortVersion string, err error) {
	baseName := BaseName(fileName)
	flavor = DetectTarballFlavor(baseName)

	reVersion := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	versionList := reVersion.FindAllStringSubmatch(baseName, -1)
	if len(versionList) == 0 || len(versionList[0]) < 2 {
		return "", "", "", fmt.Errorf("error detecting version %v", versionList)
	}

	version = fmt.Sprintf("%s.%s.%s", versionList[0][1], versionList[0][2], versionList[0][3])
	shortVersion = fmt.Sprintf("%s.%s", versionList[0][1], versionList[0][2])
	return
}

// Tries to detect the database flavor from tarball name
func DetectTarballFlavor(tarballName string) string {
	flavor := ""
	flavorsRegexps := map[string]string{
		PerconaServerFlavor: `Percona-Server`,
		MariaDbFlavor:       `mariadb`,
		NdbFlavor:           `mysql-cluster`,
		TiDbFlavor:          `tidb`,
		PxcFlavor:           `Percona-XtraDB-Cluster`,
		MySQLShellFlavor:    `mysql-shell`,
		MySQLFlavor:         `mysql`,
	}

	// Flavors must be evaluated in order, or else
	// "mysql-cluster" may be detected as "mysql"
	flavorDetectionList := []string{
		PerconaServerFlavor,
		MariaDbFlavor,
		NdbFlavor,
		TiDbFlavor,
		PxcFlavor,
		MySQLShellFlavor,
		MySQLFlavor,
	}

	for _, key := range flavorDetectionList {
		value := flavorsRegexps[key]
		re := regexp.MustCompile(value)
		if re.MatchString(tarballName) {
			return key
		}
	}
	return flavor
}
