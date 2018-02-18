package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func GetInstalledSandboxes(sandbox_home string) []string {
	files, err := ioutil.ReadDir(sandbox_home)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	var installed_sandboxes []string
	for _, f := range files {
		fname := f.Name()
		fmode := f.Mode()
		if fmode.IsDir() {
			sbdesc := sandbox_home + "/" + fname + "/sbdescription.json"
			if FileExists(sbdesc) {
				installed_sandboxes = append(installed_sandboxes, fname)

			}
		}
	}
	return installed_sandboxes
}

func GetInstalledPorts(sandbox_home string) []int {
	files := GetInstalledSandboxes(sandbox_home)
	// If there is a file sbdescription.json in the top directory
	// it will be included in the reporting
	files = append(files, "")
	var port_collection []int
	var seen_ports = make(map[int]bool)
	for _, fname := range files {
		//fname := f.Name()
		// fmode := f.Mode()
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
				if DirExists(sandbox_home + "/" + fname + "/master") {
					sd_master := ReadSandboxDescription(sandbox_home + "/" + fname + "/master")
					node_descr = append(node_descr, sd_master)
				}
				for node := 1; node <= sbd.Nodes; node++ {
					sd_node := ReadSandboxDescription(fmt.Sprintf("%s/%s/node%d", sandbox_home, fname, node))
					node_descr = append(node_descr, sd_node)
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
		fmt.Println("Required version format: x.x.xx")
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

func VersionToName(version string) string {
	re := regexp.MustCompile(`\.`)
	name := re.ReplaceAllString(version, "_")
	return name
}

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
