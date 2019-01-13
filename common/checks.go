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

package common

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type SandboxInfo struct {
	SandboxName string
	Locked      bool
}

var portDebug bool = IsEnvSet("PORT_DEBUG")

type PortMap map[int]bool

// Returns a list of inner sandboxes
func SandboxInfoToFileNames(sbList []SandboxInfo) (fileNames []string) {
	for _, sbinfo := range sbList {
		fileNames = append(fileNames, sbinfo.SandboxName)
	}
	return
}

// Returns the list of versions available for deployment
func GetVersionsFromDir(basedir string) ([]string, error) {
	var dirs []string
	files, err := ioutil.ReadDir(basedir)
	if err != nil {
		return dirs, fmt.Errorf("error reading directory %s: %s", basedir, err)
	}
	for _, f := range files {
		fname := f.Name()
		fmode := f.Mode()
		//CondPrintf("%#v\n", fmode)
		if fmode.IsDir() {
			//fmt.Println(fname)
			mysqld := path.Join(basedir, fname, "bin", "mysqld")
			if FileExists(mysqld) {
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

// Gets a list of installed sandboxes from the $SANDBOX_HOME directory
func GetInstalledSandboxes(sandboxHome string) (installedSandboxes []SandboxInfo, err error) {
	if !DirExists(sandboxHome) {
		return installedSandboxes, fmt.Errorf("directory SandboxHome not found")
	}
	files, err := ioutil.ReadDir(sandboxHome)
	if err != nil {
		return installedSandboxes, err
	}
	for _, f := range files {
		fname := f.Name()
		fmode := f.Mode()
		if fmode.IsDir() {
			if fname == globals.ForbiddenDirName {
				continue
			}
			sbdesc := path.Join(sandboxHome, fname, globals.SandboxDescriptionName)
			start := path.Join(sandboxHome, fname, "start")
			startAll := path.Join(sandboxHome, fname, "start_all")
			noClear := path.Join(sandboxHome, fname, "no_clear")
			noClearAll := path.Join(sandboxHome, fname, "no_clear_all")
			if FileExists(sbdesc) || FileExists(start) || FileExists(startAll) {
				if FileExists(noClearAll) || FileExists(noClear) {
					installedSandboxes = append(installedSandboxes, SandboxInfo{SandboxName: fname, Locked: true})
				} else {
					installedSandboxes = append(installedSandboxes, SandboxInfo{SandboxName: fname, Locked: false})
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
			if sbd.Nodes == 0 {
				for _, p := range sbd.Port {
					if !seenPorts[p] {
						portCollection = append(portCollection, p)
						seenPorts[p] = true
					}
				}
			} else {
				var nodeDescr []SandboxDescription
				innerInstalledSandboxes, err := GetInstalledSandboxes(sandboxHome)
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

/* Checks that the extracted tarball directory
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
		"libmysqlclient.a":             {"lib", "linux", "mysql", true}, // 4.1 and old 5.0 releases
		"libmysqlclient.so":            {"lib", "linux", "mysql", true},
		"libperconaserverclient.so":    {"lib", "linux", "percona", true},
		"libperconaserverclient.dylib": {"lib", "darwin", "percona", true},
		"libmysqlclient.dylib":         {"lib", "darwin", "mysql", true},
		"table.h":                      {"sql", "source", "any", false},
		"mysqlprovision.zip":           {"share/mysqlsh", "shell", "any", false},
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
			if rec.OS == currentOs && rec.isBinary {
				wantedOsFound = true
			}
			foundList[fname] = rec
		}
	}
	if !wantedOsFound {
		fmt.Println(globals.DashLine)
		CondPrintf("Looking for *%s* binaries\n", currentOs)
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
		return fmt.Errorf("could not find any of the expected files for %s server: %s\n%s\n", currentOs, wantedFiles, globals.DashLine)
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
	re1 := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)$`)
	if re1.MatchString(version) {
		return true
	}
	return false
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

// Converts a version string into a port number
// e.g. "5.6.33" -> 5633
func VersionToPort(version string) (int, error) {
	verList, err := VersionToList(version)
	if err != nil {
		return -1, fmt.Errorf("error converting %s into a version", version)
	}
	major := verList[0]
	minor := verList[1]
	rev := verList[2]
	//if major < 0 || minor < 0 || rev < 0 {
	//	return -1
	//}
	completeVersion := fmt.Sprintf("%d%d%02d", major, minor, rev)
	// fmt.Println(completeVersion)
	i, err := strconv.Atoi(completeVersion)
	if err != nil {
		return -1, fmt.Errorf("error converting %d%d%02d to version", major, minor, rev)
	}
	return i, nil
}

// Checks if a version string is greater or equal a given numeric version
// "5.6.33" >= []{5.7.0}  = false
// "5.7.21" >= []{5.7.0}  = true
// "10.1.21" >= []{5.7.0}  = false (!)
// Note: MariaDB versions are skipped. The function returns false for MariaDB 10+.
// So far (2018-02-19) this comparison holds, because MariaDB behaves like 5.5+ for
// the purposes of sandbox deployment
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
