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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"
)

type SandboxUser struct {
	Description string `json:"description"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Privileges  string `json:"privileges"`
}

type SandboxDescription struct {
	Basedir           string `json:"basedir"`
	SBType            string `json:"type"` // single multi master-slave group
	Version           string `json:"version"`
	Port              []int  `json:"port"`
	Nodes             int    `json:"nodes"`
	NodeNum           int    `json:"node_num"`
	DbDeployerVersion string `json:"dbdeployer-version"`
	Timestamp         string `json:"timestamp"`
	CommandLine       string `json:"command-line"`
	LogFile           string `json:"log-file,omitempty"`
}

type KeyValue struct {
	Key   string
	Value string
}

type ConfigOptions map[string][]KeyValue

const PublicDirectoryAttr = 0755
const SandboxDescriptionName = "sbdescription.json"

var CommandLineArgs []string

func LogDirName() string {
	logDirName := ""
	topology := ""
	nameQualifier := ""
	useReplication := false
	reTopology := regexp.MustCompile(`^--topology\s*=\s*(\S+)`)
	reSinglePrimary := regexp.MustCompile(`^--single-primary`)
	reUnwantedChars := regexp.MustCompile(`[- ./()\[\]]`)
	for _, arg := range CommandLineArgs {
		if arg == "dbdeployer" || arg == "deploy" {
			continue
		}
		if arg == "replication" {
			if topology == "" {
				topology = "master-slave"
			}
			useReplication = true
		}
		if Includes(arg, `^--`) {
			findTopology := reTopology.FindAllStringSubmatch(arg, -1)
			if len(findTopology) > 0 {
				topology = findTopology[0][1]
			}
			if reSinglePrimary.MatchString(arg) {
				nameQualifier = "sp"
			}
		} else {
			if logDirName != "" {
				logDirName += "_"
			}
			logDirName += arg
		}
	}
	if !useReplication {
		topology = ""
		nameQualifier = ""
	}
	if topology != "" {
		logDirName += "_" + topology
	}
	if nameQualifier != "" {
		logDirName += "_" + nameQualifier
	}
	PID := os.Getpid()
	logDirName = reUnwantedChars.ReplaceAllString(logDirName, "_")
	logDirName = fmt.Sprintf("%s-%d", logDirName, PID)
	return logDirName
}

func ParseConfigFile(filename string) ConfigOptions {
	config := make(ConfigOptions)
	lines := SlurpAsLines(filename)
	reComment := regexp.MustCompile(`^\s*#`)
	reEmpty := regexp.MustCompile(`^\s*$`)
	reHeader := regexp.MustCompile(`\[\s*(\w+)\s*\]`)
	reKeyValue := regexp.MustCompile(`(\S+)\s*=\s*(.*)`)
	currentHeader := ""
	for _, line := range lines {
		if reComment.MatchString(line) || reEmpty.MatchString(line) {
			continue
		}
		headerList := reHeader.FindAllStringSubmatch(line, -1)
		if headerList != nil {
			header := headerList[0][1]
			currentHeader = header
		}
		kvList := reKeyValue.FindAllStringSubmatch(line, -1)
		if kvList != nil {
			kv := KeyValue{
				Key:   kvList[0][1],
				Value: kvList[0][2],
			}
			config[currentHeader] = append(config[currentHeader], kv)
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

func WriteSandboxDescription(destination string, sd SandboxDescription) error {
	sd.DbDeployerVersion = VersionDef
	sd.Timestamp = time.Now().Format(time.UnixDate)
	sd.CommandLine = strings.Join(CommandLineArgs, " ")
	b, err := json.MarshalIndent(sd, " ", "\t")
	ErrCheckExitf(err, 1, "error encoding sandbox description: %s", err)
	jsonString := fmt.Sprintf("%s", b)
	filename := path.Join(destination, SandboxDescriptionName)
	return WriteString(jsonString, filename)
}

func ReadSandboxDescription(sandboxDirectory string) (sd SandboxDescription) {
	filename := path.Join(sandboxDirectory, SandboxDescriptionName)
	sbBlob := SlurpAsBytes(filename)

	err := json.Unmarshal(sbBlob, &sd)
	ErrCheckExitf(err, 1, "error decoding sandbox description: %s", err)
	return
}

func SlurpAsLines(filename string) []string {
	f, err := os.Open(filename)
	ErrCheckExitf(err, 1, "error opening file %s: %s", filename, err)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		Exitf(1, "%s", err)
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
	ErrCheckExitf(err, 1, "error reading from file %s: %s", filename, err)
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
		//fmt.Fprintln(w, line+termination)
		fmt.Fprintf(w, "%s", line+termination)
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

func RunCmdWithArgs(c string, args []string) (error, string) {
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
		fmt.Printf("cmd: %s %s\n", c, args)
		fmt.Printf("stdout: %s\n", out)
		//fmt.Printf("stderr: %s\n", stderr.String())
		// os.Exit(1)
	} else {
		//fmt.Printf("%s", out.String())
		fmt.Printf("%s", out)
	}
	return err, string(out)
}

func RunCmdCtrl(c string, silent bool) (error, string) {
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
		fmt.Printf("cmd: %s\n", c)
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

func RunCmd(c string) (error, string) {
	return RunCmdCtrl(c, false)
}

func CopyFile(source, destination string) {
	sourceFile, err := os.Stat(source)
	ErrCheckExitf(err, 1, "error finding source file %s: %s", source, err)
	fileMode := sourceFile.Mode()
	from, err := os.Open(source)
	ErrCheckExitf(err, 1, "error opening source file %s: %s", source, err)
	defer from.Close()

	to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, fileMode) // 0666)
	ErrCheckExitf(err, 1, "error opening destination file %s: %s", destination, err)
	defer to.Close()

	_, err = io.Copy(to, from)
	ErrCheckExitf(err, 1, "error copying from source %s to destination file %s: %s", source, destination, err)
}

// These lines should:
// 1) Do not exit. Return errors instead.
// 2) Wrappers for 1 single line will be removed to use errors instead

// func BaseName(filename string) string {
// 	return filepath.Base(filename)
// }
//
// func DirName(filename string) string {
// 	return filepath.Dir(filename)
// }

// func AbsolutePath(value string) string {
// 	filename, err := filepath.Abs(value)
// 	ErrCheckExitf(err, 1, "error getting absolute path for %s", value)
// 	return filename
// }

// func Mkdir(dirName string) {
// 	err := os.Mkdir(dirName, PublicDirectoryAttr)
// 	ErrCheckExitf(err, 1, "error creating directory %s\n%s\n", dirName, err)
// }

// func Rmdir(dirName string) {
// 	err := os.Remove(dirName)
// 	ErrCheckExitf(err, 1, "error removing directory %s\n%s\n", dirName, err)
// }
//
// func RmdirAll(dirName string) {
// 	err := os.RemoveAll(dirName)
// 	ErrCheckExitf(err, 1, "error deep-removing directory %s\n%s\n", dirName, err)
// }
