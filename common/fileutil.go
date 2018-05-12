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
	"bufio"
	//"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

type SandboxUser struct {
	Description string `json:"description"`
	Username string `json:"username"`
	Password string `json:"password"`
	Privileges string `json:"privileges"`
}

type SandboxDescription struct {
	Basedir string `json:"basedir"`
	SBType  string `json:"type"` // single multi master-slave group
	Version string `json:"version"`
	Port    []int  `json:"port"`
	Nodes   int    `json:"nodes"`
	NodeNum int    `json:"node_num"`
	DbDeployerVersion string `json:"dbdeployer-version"`
	Timestamp string `json:"timestamp"`
	CommandLine string `json:"command-line"`
}

type KeyValue struct {
	Key   string
	Value string
}

type ConfigOptions map[string][]KeyValue

var CommandLineArgs string

func ParseConfigFile(filename string) ConfigOptions {
	config := make(ConfigOptions)
	lines := SlurpAsLines(filename)
	re_comment := regexp.MustCompile(`^\s*#`)
	re_empty := regexp.MustCompile(`^\s*$`)
	re_header := regexp.MustCompile(`\[\s*(\w+)\s*\]`)
	re_k_v := regexp.MustCompile(`(\S+)\s*=\s*(.*)`)
	current_header := ""
	for _, line := range lines {
		if re_comment.MatchString(line) || re_empty.MatchString(line) {
			continue
		}
		headerList := re_header.FindAllStringSubmatch(line, -1)
		if headerList != nil {
			header := headerList[0][1]
			current_header = header
		}
		kvList := re_k_v.FindAllStringSubmatch(line, -1)
		if kvList != nil {
			kv := KeyValue{
				Key:   kvList[0][1],
				Value: kvList[0][2],
			}
			config[current_header] = append(config[current_header], kv)
		}
	}
	/*for header, kvList := range config {
		fmt.Printf("%s \n", header)
		for N, kv := range kvList {
			fmt.Printf("%d %s : %s \n", N, kv.key, kv.value)
		}
		fmt.Printf("\n")
	}*/
	return config
}

func WriteSandboxDescription(destination string, sd SandboxDescription) {
	sd.DbDeployerVersion = VersionDef
	sd.Timestamp = time.Now().Format(time.UnixDate)
	sd.CommandLine = CommandLineArgs
	b, err := json.MarshalIndent(sd, " ", "\t")
	if err != nil {
		Exit(1, fmt.Sprintf("error encoding sandbox description: %s", err))
	}
	json_string := fmt.Sprintf("%s", b)
	filename := destination + "/sbdescription.json"
	WriteString(json_string, filename)
}

func ReadSandboxDescription(sandbox_directory string) (sd SandboxDescription) {
	filename := sandbox_directory + "/sbdescription.json"
	sb_blob := SlurpAsBytes(filename)

	err := json.Unmarshal(sb_blob, &sd)
	if err != nil {
		Exit(1, fmt.Sprintf("error decoding sandbox description: %s", err))
	}
	return
}

func SlurpAsLines(filename string) []string {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		Exit(1, fmt.Sprintf("%s", err))
	}
	return lines
}

func SlurpAsString(filename string) string {
	b := SlurpAsBytes(filename)
	str := string(b)
	return str
}

func SlurpAsBytes(filename string) []byte {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return b
}

func WriteStrings(lines []string, filename string, termination string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line+termination)
	}
	return w.Flush()
}

func AppendStrings(lines []string, filename string, termination string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line+termination)
	}
	return w.Flush()
}

func WriteString(line string, filename string) error {
	return WriteStrings([]string{line}, filename, "")
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func DirExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	filemode := f.Mode()
	return filemode.IsDir()
}

func Which(filename string) string {
	path, err := exec.LookPath(filename)
	if err == nil {
		return path
	}
	return ""
}

func ExecExists(filename string) bool {
	_, err := exec.LookPath(filename)
	return err == nil
}

func FindInPath(filename string) string {
	path, _ := exec.LookPath(filename)
    return path
}

func Run_cmd_with_args(c string, args []string) (error, string) {
	cmd := exec.Command(c, args...)
	//var out bytes.Buffer
	//var stderr bytes.Buffer
	//cmd.Stdout = &out
	//cmd.Stderr = &stderr
	var out []byte
	var err error
	out, err = cmd.Output()
	if err != nil {
		fmt.Printf("err: %s\n", err)
		fmt.Printf("cmd: %#v\n", cmd)
		fmt.Printf("stdout: %s\n", out)
		//fmt.Printf("stderr: %s\n", stderr.String())
		// os.Exit(1)
	} else {
		//fmt.Printf("%s", out.String())
		fmt.Printf("%s", out)
	}
	return err, string(out)
}

func Run_cmd_ctrl(c string, silent bool) (error, string) {
	//cmd := exec.Command(c, args...)
	cmd := exec.Command(c, "")
	//var out bytes.Buffer
	//var stderr bytes.Buffer
	//cmd.Stdout = &out
	//cmd.Stderr = &stderr

	//err := cmd.Run()
	var out []byte
	var err error
	out, err = cmd.Output()
	if err != nil {
		fmt.Printf("err: %s\n", err)
		fmt.Printf("cmd: %#v\n", cmd)
		fmt.Printf("stdout: %s\n", out)
		//fmt.Printf("stdout: %s\n", out.String())
		//fmt.Printf("stderr: %s\n", stderr.String())
		// os.Exit(1)
	} else {
		if !silent {
			//fmt.Printf("%s", out.String())
			fmt.Printf("%s", out)
		}
	}
	return err, string(out)
}

func Run_cmd(c string) (error, string) {
	return Run_cmd_ctrl(c, false)
}

func CopyFile(source, destination string) {
	sfile, err := os.Stat(source)
	if err != nil {
		log.Fatal(err)
	}
	fmode := sfile.Mode()
	from, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}
	defer from.Close()

	to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, fmode) // 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
	}
}

func BaseName(filename string) string {
	return filepath.Base(filename)
}

func DirName(filename string) string {
	return filepath.Dir(filename)
}

func Mkdir(dir_name string) {
	err := os.Mkdir(dir_name, 0755)
	if err != nil {
		fmt.Printf("Error creating directory %s\n%s\n", dir_name, err)
	}
}
