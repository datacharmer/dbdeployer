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
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/ops"
	"github.com/rogpeppe/go-internal/testscript"
)

// checkPorts is a testscript command that checks that the sandbox ports are as expected
func checkPorts(ts *testscript.TestScript, neg bool, args []string) {
	assertEqual[int](ts, len(args), 2, "no sandbox path and number of ports provided")
	sbDir := args[0]
	numPorts, err := strconv.Atoi(args[1])
	ts.Check(err)

	sbDescription, err := common.ReadSandboxDescription(sbDir)
	ts.Check(err)

	assertEqual[int](ts, len(sbDescription.Port), numPorts, "want ports: %d - got: %d", numPorts, len(sbDescription.Port))
}

// findErrorsInLogFile is a testscript command that finds ERROR strings inside a sandbox data directory
func findErrorsInLogFile(ts *testscript.TestScript, neg bool, args []string) {

	assertEqual[int](ts, len(args), 1, "no sandbox path provided")
	sbDir := args[0]
	dataDir := path.Join(sbDir, "data")
	logFile := path.Join(dataDir, "msandbox.err")
	assertDirExists(ts, dataDir, globals.ErrDirectoryNotFound, dataDir)

	assertFileExists(ts, logFile, globals.ErrFileNotFound, logFile)

	contents, err := ioutil.ReadFile(logFile) // #nosec G304
	ts.Check(err)
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
	assertGreater[int](ts, len(args), 1, "syntax: check_file directory_name file_name [file_name ...]")
	sbDir := args[0]

	for i := 1; i < len(args); i++ {
		f := path.Join(sbDir, args[i])
		if neg {
			assertFileNotExists(ts, f, "file %s found", f)
		}
		assertFileExists(ts, f, globals.ErrFileNotFound, f)
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
		ts.Check(err)
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
		"check_ports": checkPorts,

		// cleanup_at_end deletes a sandbox at the end of the test if it exists
		// invoke as:"cleanup_at_end /path/to/sandbox"
		"cleanup_at_end": cleanupAtEnd,

		// run_sql_in_sandbox runs a SQL query in a sandbox, and compares the result with an expected value
		// invoke as "run_sql_in_sandbox $sb_dir 'SQL query' value_to_compare"
		// Notice that the query must return a single value
		"run_sql_in_sandbox": runSqlInSandbox,
	}
}

// cleanupAtEnd is a testscript command that deletes a sandbox at the end of the test if it exists
// use as "cleanup_at_end sandbox_dir_path"
func cleanupAtEnd(ts *testscript.TestScript, neg bool, args []string) {
	assertEqual[int](ts, len(args), 1, "no sandbox path provided")
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

// runSqlInSandbox is a testscript command that runs a SQL query in a sandbox
// use as:
// run_sql_in_sandbox "query" wanted
func runSqlInSandbox(ts *testscript.TestScript, neg bool, args []string) {
	assertEqual[int](ts, len(args), 3, "syntax: run_sql_in_sandbox sandbox_dir 'query' wanted_value")
	sbDir := args[0]
	query := args[1]
	wanted := args[2]
	assertDirExists(ts, sbDir, globals.ErrDirectoryNotFound, sbDir)

	var strResult string
	if isANumber(wanted) {
		result, err := ops.RunSandboxQuery[int](sbDir, query)
		ts.Check(err)
		strResult = fmt.Sprintf("%d", result.(int))
	} else {
		result, err := ops.RunSandboxQuery[string](sbDir, query)
		ts.Check(err)
		strResult = result.(string)
	}

	assertEqual[string](ts, strResult, wanted, "got %s - want: %s", strResult, wanted)
}
