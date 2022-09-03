// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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
	"crypto/md5"  // #nosec G501 need to compute legacy checksums
	"crypto/sha1" // #nosec G505 need to compute legacy checksums
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
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
	Host              string `json:"host,omitempty"`
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
	jsonString := string(b)
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
	f, err := os.Open(filename) // #nosec G304
	if err != nil {
		return globals.EmptyStrings, errors.Wrapf(err, "error opening file %s", filename)
	}
	defer f.Close() // #nosec G307

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
	// #nosec G304
	b, err := os.ReadFile(filename)
	if err != nil {
		return globals.EmptyBytes, errors.Wrapf(err, "error reading from file %s", filename)
	}
	return b, nil
}

// Writes a string slice into a file
// The file is created
func WriteStrings(lines []string, filename string, termination string) error {
	file, err := os.Create(filename) // #nosec G304
	if err != nil {
		return errors.Wrapf(err, "error creating file %s", filename)
	}
	defer file.Close() // #nosec G307

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

// Get a file checksum, choosing among MD5, SH1, SHA256, and SHA512
func GetFileChecksum(fileName, crcType string) (string, error) {
	var hasher hash.Hash
	switch strings.ToLower(crcType) {
	case "md5":
		hasher = md5.New() // #nosec G401 need to compute legacy checksums
	case "sha1":
		hasher = sha1.New() // #nosec G401 need to compute legacy checksums
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	default:
		return globals.EmptyString, fmt.Errorf("unsupported checksum type %s", crcType)
	}

	if hasher == nil {
		return globals.EmptyString, fmt.Errorf("unhandled checksum error")
	}
	f, err := os.Open(fileName) // #nosec G304
	if err != nil {
		return globals.EmptyString, err
	}
	defer f.Close() // #nosec G307
	_, err = io.Copy(hasher, f)
	if err != nil {
		return globals.EmptyString, err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func GetFileSha1(fileName string) (string, error) {
	return GetFileChecksum(fileName, "sha1")
}

func GetFileSha256(fileName string) (string, error) {
	return GetFileChecksum(fileName, "sha256")
}

func GetFileSha512(fileName string) (string, error) {
	return GetFileChecksum(fileName, "sha512")
}

func GetFileMd5(fileName string) (string, error) {
	return GetFileChecksum(fileName, "md5")
}

// append a string slice into an existing file
func AppendStrings(lines []string, filename string, termination string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend) // #nosec G304
	if err != nil {
		return err
	}
	defer file.Close() // #nosec G307

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

func CheckLibraries(basedir string) error {
	// This check is only needed on Linux
	if strings.ToLower(runtime.GOOS) != "linux" {
		return nil
	}

	var (
		missingAll = ""
		ldd        = Which("ldd")
		ldConfig   = Which("ldconfig")
		basedirLib = path.Join(basedir, "lib")
		mysqlBin   = path.Join(basedir, "bin", "mysql")
		mysqldBin  = path.Join(basedir, "bin", "mysqld")
	)

	// Make sure that the libraries from the expanded tarball are found
	_ = os.Setenv("LD_LIBRARY_PATH", fmt.Sprintf("%s:%s", basedirLib, os.Getenv("LD_LIBRARY_PATH")))

	// If `ldd` exists and the MySQL executables are found
	// we check the direct dependencies
	if ldd != "" && DirExists(basedir) && ExecExists(mysqlBin) && ExecExists(mysqldBin) {

		var missingMysql []string
		var missingMysqld []string

		// Gets the list of libraries for client and server executables
		mysqlLibs, err := runCmdCtrlArgsSimple(ldd, true, mysqlBin)
		if err != nil {
			if IsEnvSet("SBDEBUG") {
				fmt.Printf("# error checking ldd %s -> %s\n", mysqldBin, err)
			}
			return nil
		}
		mysqldLibs, err := runCmdCtrlArgsSimple(ldd, true, mysqldBin)
		if err != nil {
			if IsEnvSet("SBDEBUG") {
				fmt.Printf("# error checking ldd %s -> %s\n", mysqldBin, err)
			}
			return nil
		}

		// A library that is found in the system has the format
		//      library_name => /path/to/library_name
		reLibPath := regexp.MustCompile(`=>\s*/`)

		// An internal library has the format
		//      library_name (0x00007ffe1aceb000)
		reLibInternal := regexp.MustCompile(`\s+\(0x`)

		// We skip empty lines
		reEmpty := regexp.MustCompile(`^\s*$`)

		for _, lib := range strings.Split(mysqlLibs, "\n") {
			// If none of the known pattern apply, it's a not-found library
			if !reEmpty.MatchString(lib) && !reLibPath.MatchString(lib) && !reLibInternal.MatchString(lib) {
				missingMysql = append(missingMysql, lib)
			}
		}
		for _, lib := range strings.Split(mysqldLibs, "\n") {
			if !reEmpty.MatchString(lib) && !reLibPath.MatchString(lib) && !reLibInternal.MatchString(lib) {
				missingMysqld = append(missingMysqld, lib)
			}
		}

		// Add the missing libraries to the global results
		if len(missingMysql) > 0 {
			missingAll = fmt.Sprintf("client (%s): %v\n", mysqlBin, missingMysql)
		}
		if len(missingMysqld) > 0 {
			missingAll += fmt.Sprintf("\nserver (%s): %v\n", mysqldBin, missingMysqld)
		}
	}

	// "ldconfig -p" returns a list of installed libraries
	if ldConfig != "" {
		libraryList, err := runCmdCtrlArgsSimple(ldConfig, true, "-p")
		if err != nil {
			if IsEnvSet("SBDEBUG") {
				fmt.Printf("# error checking ldconfig -p -> %s\n", err)
			}
			return nil
		}

		var missing []string
		// We search for the libraries that are most likely to cause problems
		for _, library := range []string{"libaio", "libnuma", "libncurses", "libnsl"} {
			reLib := regexp.MustCompile(library)

			if !reLib.MatchString(libraryList) {
				missing = append(missing, library)
			}
		}
		if len(missing) > 0 {
			missingAll += fmt.Sprintf("global: %v\n", missing)
		}
	}
	if missingAll != "" {
		return fmt.Errorf("missing libraries will prevent MySQL from deploying correctly \n%s.\n"+
			"Use --%s to skip this check", missingAll, globals.SkipLibraryCheck)
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
	out, _, err := runCmdCtrlArgs(c, false, args...)
	return out, err
}

// Runs a command with arguments and output suppression
func RunCmdCtrlWithArgs(c string, args []string, silent bool) (string, error) {
	out, _, err := runCmdCtrlArgs(c, silent, args...)
	return out, err
}

func runCmdCtrlArgsSimple(c string, silent bool, args ...string) (string, error) {
	cmd := exec.Command(c, args...) // #nosec G204

	out, err := cmd.Output()

	if err != nil {
		CondPrintf("cmd:    %s\n", c)
		CondPrintf("err:    %s\n", err)
		CondPrintf("stdout: %s\n", out)
	} else {
		if !silent {
			fmt.Printf("%s", out)
		}
	}

	return string(out), err
}

func runCmdCtrlArgs(c string, silent bool, args ...string) (string, string, error) {
	cmd := exec.Command(c, args...) // #nosec G204
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", "", err
	}

	slurpErr, _ := io.ReadAll(stderr)
	slurpOut, _ := io.ReadAll(stdout)
	err = cmd.Wait()

	if err != nil {
		CondPrintf("cmd:    %s\n", c)
		CondPrintf("err:    %s\n", err)
		CondPrintf("stdout: %s\n", slurpOut)
		CondPrintf("stderr: %s\n", slurpErr)
	} else {
		if !silent {
			fmt.Printf("%s", slurpOut)
		}
	}

	return string(slurpOut), string(slurpErr), err
}

// Runs a command, with optional quiet output
func RunCmdCtrl(c string, silent bool) (string, error) {
	out, _, err := runCmdCtrlArgs(c, silent)
	return out, err
}

// Runs a command
func RunCmd(c string) (string, error) {
	out, _, err := runCmdCtrlArgs(c, false)
	return out, err
}

// Copies a file
func CopyFile(source, destination string) error {
	sourceFile, err := os.Stat(source)
	if err != nil {
		return errors.Wrapf(err, "error finding source file %s", source)
	}

	fileMode := sourceFile.Mode()
	from, err := os.Open(source) // #nosec G304
	if err != nil {
		return errors.Wrapf(err, "error opening source file %s", source)
	}
	defer from.Close() // #nosec G307

	to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, fileMode) // #nosec G304
	if err != nil {
		return errors.Wrapf(err, "error opening destination file %s", destination)
	}
	defer to.Close() // #nosec G307

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
	reTilde := regexp.MustCompile(`^~`)
	value = reTilde.ReplaceAllString(value, os.Getenv("HOME"))
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

// RmDirAll removes a directory with its contents, and exits if an error occurs
// Checks that the directory does not contain $HOME or $PWD
func RmdirAll(dirName string) {
	fullPath, err := AbsolutePath(dirName)
	if err != nil {
		Exitf(1, "error determining absolute path of %s", dirName)
	}
	fullHomePath, err := AbsolutePath(os.Getenv("HOME"))
	if err != nil {
		Exitf(1, "error determining absolute path of $HOME")
	}
	fullCurrentPath, err := AbsolutePath(os.Getenv("PWD"))
	if err != nil {
		Exitf(1, "error determining absolute path of $PWD")
	}
	if strings.HasPrefix(fullHomePath, fullPath) {
		Exitf(1, "attempt to delete a directory that contains the $HOME directory (%s)", fullHomePath)
	}
	if strings.HasPrefix(fullCurrentPath, fullPath) {
		Exitf(1, "attempt to delete a directory that contains the $PWD directory (%s)", fullCurrentPath)
	}
	err = os.RemoveAll(fullPath)
	ErrCheckExitf(err, 1, "error deep-removing directory %s\n%s\n", dirName, err)
}
