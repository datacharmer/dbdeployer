// Copyright © 2017-2018 Giuseppe Maxia
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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type SandboxDescription struct {
	Basedir string `json:"basedir"`
	SBType  string `json:"type"` // single multi master-slave group
	Version string `json:"version"`
	Port    []int  `json:"port"`
	Nodes   int    `json:"nodes"`
	NodeNum int    `json:"node_num"`
}

func WriteSandboxDescription(destination string, sd SandboxDescription) {
	b, err := json.MarshalIndent(sd, " ", "\t")
	if err != nil {
		fmt.Println("error encoding sandbox description: ", err)
		os.Exit(1)
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
		fmt.Println("error decoding sandbox description: ", err)
		os.Exit(1)
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

func WriteStrings(lines []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func WriteString(line string, filename string) error {
	return WriteStrings([]string{line}, filename)
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

func Run_cmd_with_args(c string, args []string) (error, string) {
	cmd := exec.Command(c, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("err: %s\n", err)
		fmt.Printf("cmd: %#v\n", cmd)
		fmt.Printf("stdout: %s\n", out.String())
		fmt.Printf("stderr: %s\n", stderr.String())
		// os.Exit(1)
	} else {
		fmt.Printf("%s", out.String())
	}
	return err, out.String()
}

func Run_cmd_ctrl(c string, silent bool) (error, string) {
	//cmd := exec.Command(c, args...)
	cmd := exec.Command(c, "")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("err: %s\n", err)
		fmt.Printf("cmd: %#v\n", cmd)
		fmt.Printf("stdout: %s\n", out.String())
		fmt.Printf("stderr: %s\n", stderr.String())
		// os.Exit(1)
	} else {
		if !silent {
			fmt.Printf("%s", out.String())
		}
	}
	return err, out.String()
}

func Run_cmd(c string) (error, string) {
	return Run_cmd_ctrl(c, false)
}
