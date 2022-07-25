// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2022 Giuseppe Maxia
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

package ts

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/rogpeppe/go-internal/testscript"
)

// checkPorts is a testscript command that checks that the sandbox ports are as expected
func checkPorts(ts *testscript.TestScript, neg bool, args []string) {

	if len(args) < 2 {
		ts.Fatalf("no sandbox path and number of ports provided")
	}
	sbDir := args[0]
	numPorts, err := strconv.Atoi(args[1])
	if err != nil {
		ts.Fatalf("error converting text '%s' to number: %s", args[1], err)
	}

	sbDescription, err := common.ReadSandboxDescription(sbDir)
	if err != nil {
		ts.Fatalf("error reading description file from %s: %s", sbDir, err)
	}

	if len(sbDescription.Port) != numPorts {
		ts.Fatalf("sandbox '%s': wanted %d ports - got %d", path.Base(sbDir), numPorts, len(sbDescription.Port))
	}

}

// findErrorsInLogFile is a testscript command that finds ERROR strings inside a sandbox data directory
func findErrorsInLogFile(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 1 {
		ts.Fatalf("no sandbox path provided")
	}
	sbDir := args[0]
	dataDir := path.Join(sbDir, "data")
	logFile := path.Join(dataDir, "msandbox.err")
	if !common.DirExists(dataDir) {
		ts.Fatalf("sandbox data dir %s not found", dataDir)
	}
	if !common.FileExists(logFile) {
		ts.Fatalf("file %s not found", logFile)
	}

	contents, err := ioutil.ReadFile(logFile) // #nosec G304
	if err != nil {
		ts.Fatalf("%s", err)
	}
	hasError := strings.Contains(string(contents), "ERROR")
	if neg && hasError {
		ts.Fatalf("ERRORs found in %s\n", logFile)
	}
	if !neg && !hasError {
		ts.Fatalf("ERRORs not found in %s\n", logFile)
	}
}

// checkFile is a testscript command that checks the existence of a list of files
// inside a directory
func checkFile(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 1 {
		ts.Fatalf("no sandbox path provided")
	}
	sbDir := args[0]

	for i := 1; i < len(args); i++ {
		f := path.Join(sbDir, args[i])
		exists := common.FileExists(f)

		if neg && exists {
			ts.Fatalf("file %s found", f)
		}
		if !exists {
			ts.Fatalf("file %s not found", f)
		}
	}
}

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
		}
	}
	time.Sleep(time.Duration(duration) * time.Second)
}

func customCommands() map[string]func(ts *testscript.TestScript, neg bool, args []string) {
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
		"check_ports":    checkPorts,
		"cleanup_at_end": cleanupAtEnd,
	}
}

// cleanupAtEnd is a testscript command that deletes a sandbox at the end of the test if it exists
func cleanupAtEnd(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 1 {
		return
	}
	sandboxDir := args[0]
	sandboxName := path.Base(sandboxDir)
	// testscript.Defer runs at the end of the current test
	ts.Defer(func() {
		if !common.DirExists(sandboxDir) {
			return
		}
		cmd := exec.Command("dbdeployer", "delete", sandboxName) // #nosec G204

		msg, err := cmd.Output()
		if err != nil {
			// Here we are already in a test failure, or else the regular cleanup would kick in
			// We just want to inform the user that also the emergency cleanup has issues
			ts.Logf("error deleting sandbox '%s': %s - %s", sandboxName, msg, err)
		}
	})
}
