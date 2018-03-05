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
	"regexp"
	"strconv"
	"strings"
)

type SandboxInfo struct {
	SandboxName string
	Locked bool
}

func SandboxInfoToFileNames(sb_list []SandboxInfo) (file_names []string) {
	for _, sbinfo := range sb_list {
		file_names = append(file_names, sbinfo.SandboxName)
	}
	return
}

func GetInstalledSandboxes(sandbox_home string) (installed_sandboxes []SandboxInfo) {
	if !DirExists(sandbox_home) {
		return
	}
	files, err := ioutil.ReadDir(sandbox_home)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
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
					installed_sandboxes = append(installed_sandboxes, SandboxInfo{ fname, true})
				} else {
					installed_sandboxes = append(installed_sandboxes, SandboxInfo{fname, false})
				}
			}
		}
	}
	return
}

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

func CheckOrigin(args []string) {
	if len(args) < 1 {
		fmt.Println("This command requires the MySQL version (x.xx.xx) as argument ")
		os.Exit(1)
	}
	if len(args) > 1 {
		fmt.Println("Extra argument detected. This command requires only the MySQL version (x.xx.xx) as argument ")
		os.Exit(1)
	}
	origin := args[0]
	if FileExists(origin) && strings.HasSuffix(origin, ".tar.gz") {
		fmt.Println("Tarball detected. - If you want to use a tarball to create a sandbox,")
		fmt.Println("you should first use the 'unpack' command")
		os.Exit(1)
	}

}

func CheckSandboxDir(sandbox_home string) {
	if !DirExists(sandbox_home) {
		fmt.Printf("Creating directory %s\n", sandbox_home)
		err := os.Mkdir(sandbox_home, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

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
