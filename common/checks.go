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

package common

import (
	"fmt"
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

const lineLength = 80

var (
	portDebug bool   = os.Getenv("PORT_DEBUG") != ""
	DashLine  string = strings.Repeat("-", lineLength)
	StarLine  string = strings.Repeat("*", lineLength)
	HashLine  string = strings.Repeat("#", lineLength)
)

type PortMap map[int]bool

var MaxAllowedPort int = 64000

func SandboxInfoToFileNames(sbList []SandboxInfo) (fileNames []string) {
	for _, sbinfo := range sbList {
		fileNames = append(fileNames, sbinfo.SandboxName)
	}
	return
}

// Gets a list of installed sandboxes from the $SANDBOX_HOME directory
func GetInstalledSandboxes(sandboxHome string) (installedSandboxes []SandboxInfo) {
	if !DirExists(sandboxHome) {
		return
	}
	files, err := ioutil.ReadDir(sandboxHome)
	ErrCheckExitf(err, 1, "%s", err)
	for _, f := range files {
		fname := f.Name()
		fmode := f.Mode()
		if fmode.IsDir() {
			sbdesc := sandboxHome + "/" + fname + "/sbdescription.json"
			start := sandboxHome + "/" + fname + "/start"
			startAll := sandboxHome + "/" + fname + "/start_all"
			noClear := sandboxHome + "/" + fname + "/no_clear"
			noClearAll := sandboxHome + "/" + fname + "/no_clear_all"
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
func GetInstalledPorts(sandboxHome string) []int {
	files := SandboxInfoToFileNames(GetInstalledSandboxes(sandboxHome))
	// If there is a file sbdescription.json in the top directory
	// it will be included in the reporting
	files = append(files, "")
	var portCollection []int
	var seenPorts = make(map[int]bool)
	for _, fname := range files {
		sbdesc := sandboxHome + "/" + fname + "/sbdescription.json"
		if FileExists(sbdesc) {
			sbd := ReadSandboxDescription(sandboxHome + "/" + fname)
			if sbd.Nodes == 0 {
				for _, p := range sbd.Port {
					if !seenPorts[p] {
						portCollection = append(portCollection, p)
						seenPorts[p] = true
					}
				}
			} else {
				var nodeDescr []SandboxDescription
				innerFiles := SandboxInfoToFileNames(GetInstalledSandboxes(sandboxHome + "/" + fname))
				for _, inner := range innerFiles {
					innerSbdesc := sandboxHome + "/" + fname + "/" + inner + "/sbdescription.json"
					if FileExists(innerSbdesc) {
						sdNode := ReadSandboxDescription(fmt.Sprintf("%s/%s/%s", sandboxHome, fname, inner))
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
	// fmt.Printf("%v\n",port_collection)
	return portCollection
}

/* Checks that the extracted tarball directory
   contains one or more files expected for the current
   operating system.
   It prevents simple errors like :
   * using a Linux tarball on a Mac or vice-versa
   * using a source or test tarball instead of a binaries one.
*/
func CheckTarballOperatingSystem(basedir string) {
	currentOs := runtime.GOOS
	// fmt.Printf("<%s>\n",currentOs)
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
		fmt.Println(DashLine)
		fmt.Printf("Looking for *%s* binaries\n", currentOs)
		fmt.Println(DashLine)
		if len(foundList) > 0 {
			fmt.Printf("# Found the following:\n")
		}
		for fname, rec := range foundList {
			fullName := path.Join(basedir, rec.Dir, fname)
			fmt.Printf("%-20s - tarball type: '%s' (flavor: %s)\n", fullName, rec.OS, rec.flavor)
			if rec.OS == "source" {
				fmt.Printf("THIS IS A SOURCE TARBALL. YOU NEED TO USE A *BINARY* TARBALL\n")
			}
			fmt.Println(DashLine)
		}
		Exitf(1, "Could not find any of the expected files for %s server: %s\n%s\n", currentOs, wantedFiles, DashLine)
	}
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
		Exit(1, "This command requires the MySQL version (x.xx[.xx]) as argument ")
	}
	if len(args) > 1 {
		Exit(1, "Extra argument detected. This command requires only the MySQL version (x.xx[.xx]) as argument ")
	}
	origin := args[0]
	if FileExists(origin) && IsATarball(origin) {
		Exit(1,
			"Tarball detected. - If you want to use a tarball to create a sandbox,",
			"you should first use the 'unpack' command")
	}
}

// Creates a sandbox directory if it does not exist
func CheckSandboxDir(sandboxHome string) {
	if !DirExists(sandboxHome) {
		fmt.Printf("Creating directory %s\n", sandboxHome)
		Mkdir(sandboxHome)
	}
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

// Gets three integers for a version string
// Converts "1.2.3" into []int{1, 2, 3}
func VersionToList(version string) []int {
	// A valid version must be made of 3 integers
	re1 := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	// Also valid version is 3 numbers with a prefix
	re2 := regexp.MustCompile(`^[^.0-9-]+(\d+)\.(\d+)\.(\d+)$`)
	verList1 := re1.FindAllStringSubmatch(version, -1)
	verList2 := re2.FindAllStringSubmatch(version, -1)
	verList := verList1
	//fmt.Printf("%#v\n", verList)
	if verList == nil {
		verList = verList2
	}
	if verList == nil {
		fmt.Printf("Required version format: x.x.xx - Got '%s'\n", version)
		return []int{-1}
		//os.Exit(1)
	}

	major, err1 := strconv.Atoi(verList[0][1])
	minor, err2 := strconv.Atoi(verList[0][2])
	rev, err3 := strconv.Atoi(verList[0][3])
	if err1 != nil || err2 != nil || err3 != nil {
		return []int{-1}
	}
	return []int{major, minor, rev}
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
func VersionToPort(version string) int {
	verList := VersionToList(version)
	major := verList[0]
	if major < 0 {
		return -1
	}
	minor := verList[1]
	rev := verList[2]
	//if major < 0 || minor < 0 || rev < 0 {
	//	return -1
	//}
	completeVersion := fmt.Sprintf("%d%d%02d", major, minor, rev)
	// fmt.Println(completeVersion)
	i, err := strconv.Atoi(completeVersion)
	if err == nil {
		return i
	}
	return -1
}

// Checks if a version string is greater or equal a given numeric version
// "5.6.33" >= []{5.7.0}  = false
// "5.7.21" >= []{5.7.0}  = true
// "10.1.21" >= []{5.7.0}  = false (!)
// Note: MariaDB versions are skipped. The function returns false for MariaDB 10+
// So far (2018-02-19) this comparison holds, because MariaDB behaves like 5.5+ for
// the purposes of sandbox deployment
func GreaterOrEqualVersion(version string, comparedTo []int) bool {
	var cmajor, cminor, crev int = comparedTo[0], comparedTo[1], comparedTo[2]
	verList := VersionToList(version)
	major := verList[0]
	if major < 0 {
		return false
	}
	minor := verList[1]
	rev := verList[2]

	if major == 10 {
		return false
	}
	sversion := fmt.Sprintf("%02d%02d%02d", major, minor, rev)
	scompare := fmt.Sprintf("%02d%02d%02d", cmajor, cminor, crev)
	// fmt.Printf("<%s><%s>\n", sversion, scompare)
	return sversion >= scompare
}

// Finds the first free port available, starting at
// requestedPort.
// usedPorts is a map of ports already used by other sandboxes.
// This function should not be used alone, but through FindFreePort
// Returns the first free port
func FindFreePortSingle(requestedPort int, usedPorts PortMap) int {
	foundPort := 0
	candidatePort := requestedPort
	for foundPort == 0 {
		_, exists := usedPorts[candidatePort]
		if exists {
			if portDebug {
				fmt.Printf("- port %d not free\n", candidatePort)
			}
		} else {
			foundPort = candidatePort
		}
		candidatePort += 1
		if candidatePort > MaxAllowedPort {
			Exit(1, fmt.Sprintf("FATAL (FindFreePortSingle): Could not find a free port starting at %d.", requestedPort),
				fmt.Sprintf("Maximum limit for port value (%d) reached", MaxAllowedPort))
		}
	}
	return foundPort
}

// Finds the a range of howMany free ports available, starting at
// basePort.
// usedPorts is a map of ports already used by other sandboxes.
// This function should not be used alone, but through FindFreePort
// Returns the first port of the requested range
func FindFreePortRange(basePort int, usedPorts PortMap, howMany int) int {
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
					fmt.Printf("- port %d is not free\n", candidatePort+counter)
				}
				candidatePort += 1
				counter = 0
				numPorts = 0
				continue
			} else {
				if portDebug {
					fmt.Printf("+ port %d is free\n", candidatePort+counter)
				}
				numPorts += 1
			}
			counter++
			if candidatePort > MaxAllowedPort {
				Exit(1, fmt.Sprintf("FATAL (FindFreePortRange): Could not find a free range of %d ports starting at %d.", howMany, requestedPort),
					fmt.Sprintf("Maximum limit for port value (%d) reached", MaxAllowedPort))
			}
		}
		if numPorts == howMany {
			foundPort = candidatePort
		} else {
			Exit(1, "FATAL: FindFreePortRange should never reach this point",
				fmt.Sprintf("requested: %d - used: %v - candidate: %d", requestedPort, usedPorts, candidatePort))
		}
	}
	return foundPort
}

// Finds the a range of howMany free ports available, starting at
// basePort.
// installedPorts is a slice of ports already used by other sandboxes.
// Calls either FindFreePortRange or FindFreePortSingle, depending on the
// amount of ports requested
// Returns the first port of the requested range
func FindFreePort(basePort int, installedPorts []int, howMany int) int {
	if portDebug {
		fmt.Printf("FindFreePort: requested: %d - used: %v - howMany: %d\n", basePort, installedPorts, howMany)
	}
	usedPorts := make(PortMap)

	for _, p := range installedPorts {
		usedPorts[p] = true
	}
	if howMany == 1 {
		return FindFreePortSingle(basePort, usedPorts)
	}
	return FindFreePortRange(basePort, usedPorts, howMany)
}
