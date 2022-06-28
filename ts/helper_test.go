package ts

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/downloads"
	"github.com/rogpeppe/go-internal/testscript"
)

var groupVersions = []string{"5.7", "8.0"}

// sleep is a testscript command that pauses the execution for the required number of seconds
func sleep(ts *testscript.TestScript, neg bool, args []string) {
	duration := 0
	var err error
	if len(args) == 0 {
		duration = 1
	} else {
		duration, err = strconv.Atoi(args[0])
		if err != nil {
			ts.Fatalf("invalid number provided: '%s'", args[0])
			return
		}
	}
	time.Sleep(time.Duration(duration) * time.Second)
}

// checkFile is a testscript command that checks the existence of a list of files
// inside a directory
func checkFile(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 1 {
		ts.Fatalf("no sandbox path provided")
		return
	}
	sbDir := args[0]

	for i := 1; i < len(args); i++ {
		f := path.Join(sbDir, args[i])
		exists := fileExists(f)

		if neg && exists {
			ts.Fatalf("file %s found", f)
		}
		if !exists {
			ts.Fatalf("file %s not found", f)
		}
	}
}

// checkPorts is a testscript command that checks that the sandbox ports are as expected
func checkPorts(ts *testscript.TestScript, neg bool, args []string) {

	portAdjustment80 := map[string]int{
		"single":               1,
		"master-slave":         3,
		"multiple":             3,
		"group-multi-primary":  3,
		"group-single-primary": 3,
	}
	if len(args) < 2 {
		ts.Fatalf("no sandbox path provided and number of ports provided")
		return
	}
	sbDir := args[0]
	numPorts, err := strconv.Atoi(args[1])
	if err != nil {
		ts.Fatalf("error converting text '%s' to number: %s", args[1], err)
		return
	}
	//sbDesc := path.Join(sbDir, "sbdescription.json")
	//if !fileExists(sbDesc) {
	//	ts.Fatalf("file %s not found", sbDesc)
	//	return
	//}
	sbDescription, err := common.ReadSandboxDescription(sbDir)
	if err != nil {
		ts.Fatalf("error reading description file from %s: %s", sbDir, err)
		return
	}
	isGreater, err := common.GreaterOrEqualVersion(sbDescription.Version, []int{8, 0, 11})
	if err != nil {
		ts.Fatalf("error comparing version '%s': %s", sbDescription.Version, err)
		return
	}
	if isGreater {
		morePorts, ok := portAdjustment80[sbDescription.SBType]
		if !ok {
			ts.Fatalf("error recognizing the type of sandbox '%s': %s", path.Base(sbDir), sbDescription.SBType)
			return
		}
		numPorts += morePorts
	}
	if len(sbDescription.Port) != numPorts {
		ts.Fatalf("sandbox '%s': wanted %d ports - got %d", path.Base(sbDir), numPorts, len(sbDescription.Port))
		return
	}

}

// findErrorsInLogFile is a testscript command that finds ERROR strings inside a sandbox data directory
func findErrorsInLogFile(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 1 {
		ts.Fatalf("no sandbox path provided")
		return
	}
	sbDir := args[0]
	dataDir := path.Join(sbDir, "data")
	logFile := path.Join(dataDir, "msandbox.err")
	if !dirExists(dataDir) {
		ts.Fatalf("sandbox data dir %s not found", dataDir)
		return
	}
	if !fileExists(logFile) {
		ts.Fatalf("file %s not found", logFile)
		return
	}

	contents, err := ioutil.ReadFile(logFile)
	if err != nil {
		ts.Fatalf("%s", err)
		return
	}
	hasError := strings.Contains(string(contents), "ERROR")
	if neg && hasError {
		ts.Fatalf("ERRORs found in %s\n", logFile)
		return
	}
	if !neg && !hasError {
		ts.Fatalf("ERRORs not found in %s\n", logFile)
		return
	}
}

// dirExists reports whether a given directory exists
func dirExists(filename string) bool {
	f, err := os.Stat(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	fileMode := f.Mode()
	return fileMode.IsDir()
}

// fileExists reports whether a given file exists
func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !errors.Is(err, fs.ErrNotExist)
}

// buildTests takes all the files from templateDir and populates several data directories
// Each directory is named with the combination of the bare name of the template file + the label
// for example, from the data directory "testdata, file "single.tmpl", and label "8_0_29" we get the file
// "single_8_0_29.txt" under "testdata/single"
func buildTests(templateDir, dataDir, label string, data map[string]string) error {

	for _, needed := range []string{"DbVersion", "DbPathVer", "Home", "TmpDir"} {
		neededTxt, ok := data[needed]
		if !ok {
			return fmt.Errorf("[buildTests] the data must contain a '%s' element", needed)
		}
		if neededTxt == "" {
			return fmt.Errorf("[buildTests] the element '%s' in data is empty", needed)
		}
	}

	homeDir := data["Home"]
	if !dirExists(homeDir) {
		return fmt.Errorf("[buildTests] home directory '%s' not found", homeDir)
	}

	tmpDir := data["TmpDir"]
	if !dirExists(tmpDir) {
		return fmt.Errorf("[buildTests] temp directory '%s' not found", tmpDir)
	}

	//err := setEnvironment(data["DbVersion"])
	//if err != nil {
	//	return fmt.Errorf("[buildTests] error setting environment for  %s: %s", data["DbVersion"], err)
	//}

	if !dirExists(dataDir) {
		err := os.Mkdir(dataDir, 0755)
		if err != nil {
			return fmt.Errorf("[buildTests] error creating directory %s: %s", dataDir, err)
		}
	}
	files, err := filepath.Glob(templateDir + "/*.tmpl")

	if err != nil {
		return fmt.Errorf("[buildTests] error retrieving template files: %s", err)
	}
fileLoop:
	for _, f := range files {
		fName := strings.Replace(path.Base(f), ".tmpl", "", 1)
		if strings.HasPrefix(fName, "group") {
			groupEnabled := false
			for _, v := range groupVersions {
				if strings.HasPrefix(data["DbVersion"], v) {
					groupEnabled = true
				}
			}
			if !groupEnabled {
				continue fileLoop
			}
		}

		contents, err := ioutil.ReadFile(f)
		if err != nil {
			return fmt.Errorf("[buildTests] error reading file %s: %s", f, err)
		}

		subDataDir := path.Join(dataDir, label)
		if !dirExists(subDataDir) {
			err := os.Mkdir(subDataDir, 0755)
			if err != nil {
				return fmt.Errorf("[buildTests] error creating directory %s: %s", subDataDir, err)
			}
		}
		processTemplate := template.Must(template.New(label).Parse(string(contents)))
		buf := &bytes.Buffer{}

		if err := processTemplate.Execute(buf, data); err != nil {
			return fmt.Errorf("[buildTests] error processing template from %s: %s", f, err)
		}
		testName := path.Join(subDataDir, fName+"_"+label+".txt")
		err = ioutil.WriteFile(testName, buf.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("[buildTests] error writing text file %s: %s", testName, err)
		}

	}
	return nil
}

func setEnvironment(version string) error {

	curDir := os.Getenv("PWD")
	homeDir := os.Getenv("HOME")
	if !dirExists(homeDir) {
		return fmt.Errorf("home directory '%s' not found", homeDir)
	}
	sandboxDir := path.Join(homeDir, "sandboxes")
	binaryDir := path.Join(homeDir, "opt", "mysql")
	for _, dir := range []string{sandboxDir, binaryDir} {
		if !dirExists(dir) {
			err := os.Mkdir(dir, 0755)
			if err != nil {
				return fmt.Errorf("error creating directory '%s': %s", dir, err)
			}
		}
	}
	minimal := false
	if strings.EqualFold(runtime.GOOS, "linux") {
		minimal = true
	}
	os.Chdir(homeDir)
	defer os.Chdir(curDir)
	tarball, err := downloads.FindOrGuessTarballByVersionFlavorOS(version, common.MySQLFlavor, runtime.GOOS, minimal, true, false)
	if err != nil {
		return fmt.Errorf("error getting version %s (%s-%s)[minimal: %v - newest: true - guess: false]: %s",
			version, common.MySQLFlavor, runtime.GOOS, minimal, err)
	}
	if !fileExists(tarball.Name) {
		return fmt.Errorf("downloaded file %s not found", tarball.Name)
	}

	return nil
}

func defineCommands() map[string]func(ts *testscript.TestScript, neg bool, args []string) {
	return map[string]func(ts *testscript.TestScript, neg bool, args []string){
		// find_errors will check that the error log in a sandbox contains the string ERROR
		// invoke as "find_errors /path/to/sandbox"
		// The command can be negated, i.e. it will succeed if the log does not contain the string ERROR
		// "! find_errors /path/to/sandbox"
		"find_errors": findErrorsInLogFile,

		// check_file will check that a given list of files exists
		// invoke as "check_file /path/to/sandbox file1 [file2 [file3 [file4]]]"
		// The command can be negated, i.e. it will succeed if the given files do not exist
		// "! check_file /path/to/sandbox file1 [file2 [file3 [file4]]]"
		"check_file": checkFile,

		// sleep will pause execution for the required number of seconds
		// Invoke as "sleep 3"
		// If no number is passed, it pauses for 1 second
		"sleep": sleep,

		// check_ports will check that the number of ports expected for a given sandbox correspond to the ones
		// found in sbdescription.json
		// Invoke as "check_ports /path/to/sandbox 3"
		"check_ports": checkPorts,
	}
}
