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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/datacharmer/dbdeployer/globals"
)

type SandboxUser struct {
	Description string `json:"description"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Privileges  string `json:"privileges"`
}

type SandboxDescription struct {
	Basedir           string `json:"basedir"`
	ClientBasedir     string `json:"client_basedir,omitempty"`
	SBType            string `json:"type"` // single multi master-slave group
	Version           string `json:"version"`
	Flavor            string `json:"flavor,omitempty"`
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

var CommandLineArgs []string

// Returns the name of the log directory
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

// Reads a MySQL configuration file and returns its structured contents
func ParseConfigFile(filename string) (ConfigOptions, error) {
	config := make(ConfigOptions)
	lines, err := SlurpAsLines(filename)
	if err != nil {
		return ConfigOptions{}, errors.Wrapf(err, "error reading configuration file %s", filename)
	}
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
	return config, nil
}

// Writes the description of a sandbox in the appropriate directory
func WriteSandboxDescription(destination string, sd SandboxDescription) error {
	sd.DbDeployerVersion = VersionDef
	sd.Timestamp = time.Now().Format(time.UnixDate)
	sd.CommandLine = strings.Join(CommandLineArgs, " ")
	b, err := json.MarshalIndent(sd, " ", "\t")
	if err != nil {
		return errors.Wrapf(err, "error encoding sandbox description")
	}
	jsonString := fmt.Sprintf("%s", b)
	filename := path.Join(destination, globals.SandboxDescriptionName)
	return WriteString(jsonString, filename)
}

// Reads sandbox description from a given directory
func ReadSandboxDescription(sandboxDirectory string) (SandboxDescription, error) {
	filename := path.Join(sandboxDirectory, globals.SandboxDescriptionName)
	if !FileExists(filename) {
		return SandboxDescription{}, errors.Wrapf(fmt.Errorf("file not found %s", filename), "Sandbox description file not found")
	}
	stat, err := os.Stat(filename)
	if err != nil {
		return SandboxDescription{}, errors.Wrapf(err, "error getting stats for file %s", filename)
	}
	if stat.Size() == 0 {
		return SandboxDescription{}, errors.Wrapf(fmt.Errorf("empty description"), "empty sandbox description %s", filename)
	}
	sbBlob, err := SlurpAsBytes(filename)
	if err != nil {
		return SandboxDescription{}, errors.Wrapf(err, "error reading from file %s", filename)
	}
	var sd SandboxDescription

	err = json.Unmarshal(sbBlob, &sd)
	if err != nil {
		return SandboxDescription{}, errors.Wrapf(err, "error decoding sandbox description")
	}
	return sd, nil
}

// Reads a file and returns its lines as a string slice
func SlurpAsLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return globals.EmptyStrings, errors.Wrapf(err, "error opening file %s", filename)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return globals.EmptyStrings, errors.Wrapf(err, "error reading file %s", filename)
	}
	return lines, nil
}

// Reads a file and returns its contents as a single string
func SlurpAsString(filename string) (string, error) {
	b, err := SlurpAsBytes(filename)
	var str string
	if err == nil {
		str = string(b)
	}
	return str, errors.Wrapf(err, "SlurpAsString")
}

// reads a file and returns its contents as a byte slice
func SlurpAsBytes(filename string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return globals.EmptyBytes, errors.Wrapf(err, "error reading from file %s", filename)
	}
	return b, nil
}

// Writes a string slice into a file
// The file is created
func WriteStrings(lines []string, filename string, termination string) error {
	file, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "error creating file %s", filename)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		//fmt.Fprintln(w, line+termination)
		N, err := fmt.Fprintf(w, "%s", line+termination)
		if err != nil {
			return errors.Wrapf(err, "error writing to file %s ", filename)
		}
		if N == 0 {
			return nil
		}
	}
	return w.Flush()
}

// append a string slice into an existing file
func AppendStrings(lines []string, filename string, termination string) error {
	// file, err := os.Open(filename)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		N, err := fmt.Fprint(w, line+termination)
		if err != nil {
			return errors.Wrapf(err, "error writing to file %s ", filename)
		}
		if N == 0 {
			return nil
		}
	}
	return w.Flush()
}

// append a string into an existing file
func WriteString(line string, filename string) error {
	return WriteStrings([]string{line}, filename, "")
}

// returns true if a given file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// returns true if a given directory exists
func DirExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	filemode := f.Mode()
	return filemode.IsDir()
}

func GlobalTempDir() string {
	globalTmpDir := os.Getenv("TMPDIR")
	if globalTmpDir == "" {
		globalTmpDir = "/tmp"
	}
	return globalTmpDir
}

// Returns the full path of an executable, or an empty string if the executable is not found
func Which(filename string) string {
	filePath, err := exec.LookPath(filename)
	if err == nil {
		return filePath
	}
	return globals.EmptyString
}

func CheckPrerequisites(label string, neededExecutables []string) error {
	missingExecutables := []string{}
	for _, executable := range neededExecutables {
		execPath := Which(executable)
		if execPath == "" {
			missingExecutables = append(missingExecutables, executable)
		}
	}
	if len(missingExecutables) > 0 {
		return fmt.Errorf("[%s] missing executables: %v", label, missingExecutables)
	}
	return nil
}

// returns true if a given executable exists
func ExecExists(filename string) bool {
	_, err := exec.LookPath(filename)
	return err == nil
}

// Same as Which
func FindInPath(filename string) string {
	filePath, _ := exec.LookPath(filename)
	return filePath
}

// Runs a command with arguments
func RunCmdWithArgs(c string, args []string) (string, error) {
	cmd := exec.Command(c, args...)
	var out []byte
	var err error
	out, err = cmd.Output()
	if err != nil {
		CondPrintf("err: %s\n", err)
		CondPrintf("cmd: %s %s\n", c, args)
		CondPrintf("stdout: %s\n", out)
	} else {
		CondPrintf("%s", out)
	}
	return string(out), err
}

// Runs a command with arguments and output suppression
func RunCmdCtrlWithArgs(c string, args []string, silent bool) (string, error) {
	cmd := exec.Command(c, args...)
	var out []byte
	var err error
	out, err = cmd.Output()
	if err != nil {
		CondPrintf("err: %s\n", err)
		CondPrintf("cmd: %s %s\n", c, args)
		CondPrintf("stdout: %s\n", out)
	} else {
		if !silent {
			fmt.Printf("%s", out)
		}
	}
	return string(out), err
}

// Runs a command, with optional quiet output
func RunCmdCtrl(c string, silent bool) (string, error) {
	cmd := exec.Command(c, "")
	var out []byte
	var err error
	out, err = cmd.Output()
	if err != nil {
		CondPrintf("err: %s\n", err)
		CondPrintf("cmd: %s\n", c)
		CondPrintf("stdout: %s\n", out)
	} else {
		if !silent {
			fmt.Printf("%s", out)
		}
	}
	return string(out), err
}

// Runs a command
func RunCmd(c string) (string, error) {
	return RunCmdCtrl(c, false)
}

// Copies a file
func CopyFile(source, destination string) error {
	sourceFile, err := os.Stat(source)
	if err != nil {
		return errors.Wrapf(err, "error finding source file %s", source)
	}

	fileMode := sourceFile.Mode()
	from, err := os.Open(source)
	if err != nil {
		return errors.Wrapf(err, "error opening source file %s", source)
	}
	defer from.Close()

	to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, fileMode) // 0666)
	if err != nil {
		return errors.Wrapf(err, "error opening destination file %s", destination)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return errors.Wrapf(err, "error copying from source %s to destination file %s", source, destination)
	}
	return nil
}

// Returns the base name of a file
func BaseName(filename string) string {
	return filepath.Base(filename)
}

// Returns the directory name of a file
func DirName(filename string) string {
	return filepath.Dir(filename)
}

// Returns the absolute path of a file
func AbsolutePath(value string) (string, error) {
	filename, err := filepath.Abs(value)
	if err != nil {
		return "", errors.Wrapf(err, "error getting absolute path for %s", value)
	}
	return filename, nil
}

// ------------------------------------------------------------------------------------
// The functions below this point are intended only for use with a command line client,
// and may not be suitable for other client types
// ------------------------------------------------------------------------------------

// Creates a directory, and exits if an error occurs
func Mkdir(dirName string) {
	err := os.Mkdir(dirName, globals.PublicDirectoryAttr)
	ErrCheckExitf(err, 1, "error creating directory %s\n%s\n", dirName, err)
}

// Removes a directory, and exits if an error occurs
func Rmdir(dirName string) {
	err := os.Remove(dirName)
	ErrCheckExitf(err, 1, "error removing directory %s\n%s\n", dirName, err)
}

// Removes a directory with its contents, and exits if an error occurs
func RmdirAll(dirName string) {
	err := os.RemoveAll(dirName)
	ErrCheckExitf(err, 1, "error deep-removing directory %s\n%s\n", dirName, err)
}
