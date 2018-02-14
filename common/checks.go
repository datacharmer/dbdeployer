package common

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
)

func GetInstalledSandboxes(sandbox_home string) []string {
	files, err := ioutil.ReadDir(sandbox_home)
	if err != nil {
		fmt.Printf("%s",err)
		os.Exit(1)
	}
	var installed_sandboxes []string
	for _, f := range files {
		fname := f.Name()
		fmode := f.Mode()
		if fmode.IsDir() {
			sbdesc := sandbox_home + "/" + fname + "/sbdescription.json"
			if FileExists(sbdesc) {
				installed_sandboxes = append( installed_sandboxes, fname)

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
