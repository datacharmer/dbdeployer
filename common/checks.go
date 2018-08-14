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

var port_debug bool = os.Getenv("PORT_DEBUG") != ""

type PortMap map[int]bool

var MaxAllowedPort int = 64000

func SandboxInfoToFileNames(sb_list []SandboxInfo) (file_names []string) {
	for _, sbinfo := range sb_list {
		file_names = append(file_names, sbinfo.SandboxName)
	}
	return
}

// Gets a list of installed sandboxes from the $SANDBOX_HOME directory
func GetInstalledSandboxes(sandbox_home string) (installed_sandboxes []SandboxInfo) {
	if !DirExists(sandbox_home) {
		return
	}
	files, err := ioutil.ReadDir(sandbox_home)
	if err != nil {
		Exit(1, fmt.Sprintf("%s", err))
	}
	for _, f := range files {
		fname := f.Name()
		fmode := f.Mode()
		if fmode.IsDir() {
			sbdesc := sandbox_home + "/" + fname + "/sbdescription.json"
			start := sandbox_home + "/" + fname + "/start"
			start_all := sandbox_home + "/" + fname + "/start_all"
			no_clear := sandbox_home + "/" + fname + "/no_clear"
			no_clear_all := sandbox_home + "/" + fname + "/no_clear_all"
			if FileExists(sbdesc) || FileExists(start) || FileExists(start_all) {
				if FileExists(no_clear_all) || FileExists(no_clear) {
					installed_sandboxes = append(installed_sandboxes, SandboxInfo{fname, true})
				} else {
					installed_sandboxes = append(installed_sandboxes, SandboxInfo{fname, false})
				}
			}
		}
	}
	return
}

// Collects a list of used ports from deployed sandboxes
func GetInstalledPorts(sandbox_home string) []int {
	files := SandboxInfoToFileNames(GetInstalledSandboxes(sandbox_home))
	// If there is a file sbdescription.json in the top directory
	// it will be included in the reporting
	files = append(files, "")
	var port_collection []int
	var seen_ports = make(map[int]bool)
	for _, fname := range files {
		sbdesc := sandbox_home + "/" + fname + "/sbdescription.json"
		if FileExists(sbdesc) {
			sbd := ReadSandboxDescription(sandbox_home + "/" + fname)
			if sbd.Nodes == 0 {
				for _, p := range sbd.Port {
					if !seen_ports[p] {
						port_collection = append(port_collection, p)
						seen_ports[p] = true
					}
				}
			} else {
				var node_descr []SandboxDescription
				inner_files := SandboxInfoToFileNames(GetInstalledSandboxes(sandbox_home + "/" + fname))
				for _, inner := range inner_files {
					inner_sbdesc := sandbox_home + "/" + fname + "/" + inner + "/sbdescription.json"
					if FileExists(inner_sbdesc) {
						sd_node := ReadSandboxDescription(fmt.Sprintf("%s/%s/%s", sandbox_home, fname, inner))
						node_descr = append(node_descr, sd_node)
					}
				}
				for _, nd := range node_descr {
					for _, p := range nd.Port {
						if !seen_ports[p] {
							port_collection = append(port_collection, p)
							seen_ports[p] = true
						}
					}
				}
			}
		}
	}
	// fmt.Printf("%v\n",port_collection)
	return port_collection
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
	var finding_list = map[string]OSFinding{
		"libmysqlclient.so":            OSFinding{"lib", "linux", "mysql", true},
		"libperconaserverclient.so":    OSFinding{"lib", "linux", "percona", true},
		"libperconaserverclient.dylib": OSFinding{"lib", "darwin", "percona", true},
		"libmysqlclient.dylib":         OSFinding{"lib", "darwin", "mysql", true},
		"table.h":                      OSFinding{"sql", "source", "any", false},
		"mysqlprovision.zip":           OSFinding{"share/mysqlsh", "shell", "any", false},
	}
	wanted_os_found := false
	var found_list = make(map[string]OSFinding)
	var wanted_files []string
	for fname, rec := range finding_list {
		full_name := path.Join(basedir, rec.Dir, fname)
		if rec.OS == currentOs && rec.isBinary {
			wanted_files = append(wanted_files, path.Join(rec.Dir, fname))
		}
		if FileExists(full_name) {
			if rec.OS == currentOs && rec.isBinary {
				wanted_os_found = true
			}
			found_list[fname] = rec
		}
	}
	if !wanted_os_found {
		dash_line := strings.Repeat("-", 80)
		fmt.Println(dash_line)
		fmt.Printf("Looking for *%s* binaries\n", currentOs)
		fmt.Println(dash_line)
		if len(found_list) > 0 {
			fmt.Printf("# Found the following:\n")
		}
		for fname, rec := range found_list {
			full_name := path.Join(basedir, rec.Dir, fname)
			fmt.Printf("%-20s - tarball type: '%s' (flavor: %s)\n", full_name, rec.OS, rec.flavor)
			if rec.OS == "source" {
				fmt.Printf("THIS IS A SOURCE TARBALL. YOU NEED TO USE A *BINARY* TARBALL\n")
			}
			fmt.Println(dash_line)
		}
		Exit(1, fmt.Sprintf("Could not find any of the expected files for %s server: %s", currentOs, wanted_files), dash_line)
	}
}

// Checks the initial argument for a sandbox deployment
func CheckOrigin(args []string) {
	if len(args) < 1 {
		Exit(1, "This command requires the MySQL version (x.xx.xx) as argument ")
	}
	if len(args) > 1 {
		Exit(1, "Extra argument detected. This command requires only the MySQL version (x.xx.xx) as argument ")
	}
	origin := args[0]
	if FileExists(origin) && strings.HasSuffix(origin, ".tar.gz") {
		Exit(1,
			"Tarball detected. - If you want to use a tarball to create a sandbox,",
			"you should first use the 'unpack' command")
	}
}

// Creates a sandbox directory if it does not exist
func CheckSandboxDir(sandbox_home string) {
	if !DirExists(sandbox_home) {
		fmt.Printf("Creating directory %s\n", sandbox_home)
		err := os.Mkdir(sandbox_home, 0755)
		if err != nil {
			Exit(1, fmt.Sprintf("%s", err))
		}
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
func GreaterOrEqualVersion(version string, compared_to []int) bool {
	var cmajor, cminor, crev int = compared_to[0], compared_to[1], compared_to[2]
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
// requested_port.
// used_ports is a map of ports already used by other sandboxes.
// This function should not be used alone, but through FindFreePort
// Returns the first free port
func FindFreePortSingle(requested_port int, used_ports PortMap) int {
	found_port := 0
	candidate_port := requested_port
	for found_port == 0 {
		_, exists := used_ports[candidate_port]
		if exists {
			if port_debug {
				fmt.Printf("- port %d not free\n", candidate_port)
			}
		} else {
			found_port = candidate_port
		}
		candidate_port += 1
		if candidate_port > MaxAllowedPort {
			Exit(1, fmt.Sprintf("FATAL (FindFreePortSingle): Could not find a free port starting at %d.", requested_port),
				fmt.Sprintf("Maximum limit for port value (%d) reached", MaxAllowedPort))
		}
	}
	return found_port
}

// Finds the a range of how_many free ports available, starting at
// base_port.
// used_ports is a map of ports already used by other sandboxes.
// This function should not be used alone, but through FindFreePort
// Returns the first port of the requested range
func FindFreePortRange(base_port int, used_ports PortMap, how_many int) int {
	var found_port int = 0
	requested_port := base_port
	candidate_port := requested_port
	counter := 0
	for found_port == 0 {
		num_ports := 0
		for counter < how_many {
			_, exists := used_ports[candidate_port+counter]
			if exists {
				if port_debug {
					fmt.Printf("- port %d is not free\n", candidate_port+counter)
				}
				candidate_port += 1
				counter = 0
				num_ports = 0
				continue
			} else {
				if port_debug {
					fmt.Printf("+ port %d is free\n", candidate_port+counter)
				}
				num_ports += 1
			}
			counter++
			if candidate_port > MaxAllowedPort {
				Exit(1, fmt.Sprintf("FATAL (FindFreePortRange): Could not find a free range of %d ports starting at %d.", how_many, requested_port),
					fmt.Sprintf("Maximum limit for port value (%d) reached", MaxAllowedPort))
			}
		}
		if num_ports == how_many {
			found_port = candidate_port
		} else {
			Exit(1, "FATAL: FindFreePortRange should never reach this point",
				fmt.Sprintf("requested: %d - used: %v - candidate: %d", requested_port, used_ports, candidate_port))
		}
	}
	return found_port
}

// Finds the a range of how_many free ports available, starting at
// base_port.
// installed_ports is a slice of ports already used by other sandboxes.
// Calls either FindFreePortRange or FindFreePortSingle, depending on the
// amount of ports requested
// Returns the first port of the requested range
func FindFreePort(base_port int, installed_ports []int, how_many int) int {
	if port_debug {
		fmt.Printf("FindFreePort: requested: %d - used: %v - how_many: %d\n", base_port, installed_ports, how_many)
	}
	used_ports := make(PortMap)

	for _, p := range installed_ports {
		used_ports[p] = true
	}
	if how_many == 1 {
		return FindFreePortSingle(base_port, used_ports)
	}
	return FindFreePortRange(base_port, used_ports, how_many)
}
