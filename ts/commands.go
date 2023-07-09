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
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/ops"
	"github.com/rogpeppe/go-internal/testscript"
)

// checkPorts is a testscript command that checks that the sandbox ports are as expected
func checkPorts(ts *testscript.TestScript, neg bool, args []string) {
	assertEqual(ts, len(args), 2, "no sandbox path and number of ports provided")
	sbDir := args[0]
	numPorts, err := strconv.Atoi(args[1])
	ts.Check(err)

	sbDescription, err := common.ReadSandboxDescription(sbDir)
	ts.Check(err)

	assertEqual(ts, len(sbDescription.Port), numPorts, "want ports: %d - got: %d", numPorts, len(sbDescription.Port))
}

// findErrorsInLogFile is a testscript command that finds ERROR strings inside a sandbox data directory
func findErrorsInLogFile(ts *testscript.TestScript, neg bool, args []string) {

	assertEqual(ts, len(args), 1, "no sandbox path provided")
	sbDir := args[0]
	dataDir := path.Join(sbDir, "data")
	logFile := path.Join(dataDir, "msandbox.err")
	assertDirExists(ts, dataDir, globals.ErrDirectoryNotFound, dataDir)

	assertFileExists(ts, logFile, globals.ErrFileNotFound, logFile)

	contents, err := os.ReadFile(logFile) // #nosec G304
	ts.Check(err)
	hasError := strings.Contains(string(contents), "ERROR")
	if neg && hasError {
		reLines := regexp.MustCompile(`(?sg)(^.*ERROR.*)`)
		errorLines := reLines.FindAll(contents, -1)
		ts.Fatalf("ERRORs found in %s (%s)\n", logFile, errorLines)
	}
	if !neg && !hasError {
		ts.Fatalf("ERRORs not found in %s\n", logFile)
	}
}

// checkFile is a testscript command that checks the existence of a list of files
// inside a directory
func checkFile(ts *testscript.TestScript, neg bool, args []string) {
	assertGreater(ts, len(args), 1, "syntax: check_file directory_name file_name [file_name ...]")
	sbDir := args[0]

	for i := 1; i < len(args); i++ {
		f := path.Join(sbDir, args[i])
		if neg {
			assertFileNotExists(ts, f, "file %s found", f)
		}
		assertFileExists(ts, f, globals.ErrFileNotFound, f)
	}
}

// checkExecutable is a testscript command that checks the existence of a list of executable files
// inside a directory
func checkExecutable(ts *testscript.TestScript, neg bool, args []string) {
	assertGreater(ts, len(args), 1, "syntax: check_exec directory_name file_name [file_name ...]")
	sbDir := args[0]

	for i := 1; i < len(args); i++ {
		f := path.Join(sbDir, args[i])
		if neg {
			assertFileNotExists(ts, f, "file %s found", f)
		}
		assertExecExists(ts, f, globals.ErrExecutableNotFound, f)
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

		// check_exec will check that a given list of executable files exists
		// invoke as "check_exec /path/to/sandbox file1 [file2 [file3 [file4]]]"
		// The command can be negated, i.e. it will succeed if the given files do not exist
		// "! check_file /path/to/sandbox file1 [file2 [file3 [file4]]]"
		"check_exec": checkExecutable,

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
		// invoke as "run_sql_in_sandbox $sb_dir 'SQL query' {eq|lt|le|gt|ge} value_to_compare "
		// Notice that the query must return a single value
		"run_sql_in_sandbox": runSqlInSandbox,

		// check_sandbox_manifest checks that all the files that should be in a sandbox are present
		// invoke as "check_sandbox_manifest $sb_dir sandbox_type"
		// sandbox_type is one of {single|replication|multiple|multi_source|group}
		"check_sandbox_manifest": checkSandboxManifest,
	}
}

// cleanupAtEnd is a testscript command that deletes a sandbox at the end of the test if it exists
// use as "cleanup_at_end sandbox_dir_path"
func cleanupAtEnd(ts *testscript.TestScript, neg bool, args []string) {
	assertEqual(ts, len(args), 1, "no sandbox path provided")
	sandboxDir := args[0]
	sandboxName := path.Base(sandboxDir)
	// testscript.Defer runs at the end of the current test
	ts.Defer(func() {
		if os.Getenv("ts_preserve") != "" {
			return
		}
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
// run_sql_in_sandbox "query" {eq|lt|le|gt|ge} wanted
func runSqlInSandbox(ts *testscript.TestScript, neg bool, args []string) {
	assertEqual(ts, len(args), 4, "syntax: run_sql_in_sandbox sandbox_dir 'query' {eq|lt|le|gt|ge} wanted_value")
	sbDir := args[0]
	query := args[1]
	operation := args[2]
	wanted := args[3]
	assertDirExists(ts, sbDir, globals.ErrDirectoryNotFound, sbDir)

	var strResult string
	if isANumber(wanted) {
		result, err := ops.RunSandboxQuery[int](sbDir, query, true)
		ts.Check(err)
		strResult = fmt.Sprintf("%d", result.(int))
	} else {
		result, err := ops.RunSandboxQuery[string](sbDir, query, true)
		ts.Check(err)
		strResult = result.(string)
	}

	switch strings.ToLower(operation) {
	case "eq", "=", "==":
		assertEqual(ts, strResult, wanted, "got %s - want: %s", strResult, wanted)
	case "ge", ">=":
		assertGreaterEqual(ts, strResult, wanted, "got %s - want: >= %s", strResult, wanted)
	case "gt", ">":
		assertGreater(ts, strResult, wanted, "got %s - want: > %s", strResult, wanted)
	case "le", "<=":
		assertGreaterEqual(ts, wanted, strResult, "got %s - want: <= %s", strResult, wanted)
	case "lt", "<":
		assertGreater(ts, wanted, strResult, "got %s - want: < %s", strResult, wanted)
	default:
		ts.Fatalf("unrecognized operation %s", operation)
	}
}

type manifest map[string][]string

func checkSandboxManifest(ts *testscript.TestScript, neg bool, args []string) {

	if len(args) < 2 {
		ts.Fatalf("syntax: check_sandbox_manifest directory sandbox-type")
	}
	dir := args[0]
	sbType := args[1]
	var err error
	manifest, ok := manifests[sbType]
	if !ok {
		ts.Fatalf("unrecognized sandbox type '%s'", sbType)
	}
	err = checkManifest(sbType, dir, manifest)
	if err != nil {
		ts.Fatalf("error checking sandbox manifest: %s", err)
	}
}

func checkManifest(sbType, dir string, manifest map[string][]string) error {
	sbDescription, err := common.ReadSandboxDescription(dir)

	if err != nil {
		return fmt.Errorf("error getting sandbox description from directory %s: %s", dir, err)
	}
	err = checkConditionalFiles(sbType, sbDescription.Version, dir)
	if err != nil {
		return fmt.Errorf("error checking conditional files in directory %s: %s", dir, err)
	}
	executables, ok := manifest["executable"]
	if ok {
		for _, fName := range executables {
			if testing.Verbose() {
				fmt.Printf("{manifest} ******[%s]**** %s\n", dir, fName)
			}
			if !common.ExecExists(path.Join(dir, fName)) {
				return fmt.Errorf("executable file %s not found in %s", fName, dir)
			}
		}
	}
	regularFiles, ok := manifest["regular"]
	if ok {
		for _, fName := range regularFiles {
			if !common.FileExists(path.Join(dir, fName)) {
				return fmt.Errorf("regular file %s not found in %s", fName, dir)
			}
		}
	}
	dirs, ok := manifest["directory"]
	if ok {
		for _, d := range dirs {
			if !common.DirExists(path.Join(dir, d)) {
				return fmt.Errorf("directory %s not found in %s", d, dir)
			}
		}
	}
	nodes, found := manifest["nodes"]
	if !found {
		return nil
	}
	for _, node := range nodes {
		err := checkManifest("single", path.Join(dir, node), manifests["single"])
		if err != nil {
			return err
		}
	}
	return nil
}

type conditionalFile struct {
	sandboxType string
	fileName    string
	check       func(string) bool
}

var manifestByVersion = map[string][]conditionalFile{
	"8.0.17": {
		{"single", "clone_connection.sql", common.FileExists},
		{"single", "clone_from", common.ExecExists},
	},
}

func checkConditionalFiles(sbType, version, dir string) error {
	for condVersion, condFile := range manifestByVersion {
		condVersionList, _ := common.VersionToList(condVersion)
		isAfter, _ := common.GreaterOrEqualVersion(version, condVersionList)
		for _, f := range condFile {
			if sbType != f.sandboxType {
				continue
			}
			if isAfter {
				if testing.Verbose() {
					fmt.Printf("{conditional manifest} ******[%s]**** %s\n", dir, f.fileName)
				}
				exists := f.check(path.Join(dir, f.fileName))
				if !exists {
					return fmt.Errorf("[current version: %s] file %s (>=%s) not found in sandbox %s", version, f.fileName, condVersion, dir)
				}
			}
		}
	}
	return nil
}

var manifests = map[string]manifest{
	"single": {
		"executable": {
			"add_option", "after_start", "clear", "init_db", "load_grants",
			"metadata", "my", "replicate_from", "restart", "send_kill",
			"show_binlog", "show_log", "show_relaylog", "start", "status", "stop",
			"sysbench", "sysbench_ready", "test_sb", "use", "wipe_and_restart",
		},
		"regular": {
			"start.log", "connection.conf", "connection.json",
			"connection.sql", "connection_super_user.conf", "connection_super_user.json",
			"grants.mysql", "my.sandbox.cnf", "sb_include", "sbdescription.json",
			"data/msandbox.err",
		},
		"directory":  {"data", "tmp"},
		"by_version": []string{"8.0.17", "clone_from", "clone_connection.sql"},
	},
	"replication": {
		"executable": {
			"check_slaves", "exec_all_slaves", "metadata_all",
			"s1", "start_all", "sysbench_ready", "use_all_masters",
			"clear_all", "initialize_slaves", "n1", "s2", "status_all", "test_replication", "use_all_slaves",
			"exec_all", "m", "n2", "replicate_from", "stop_all", "test_sb_all", "wipe_and_restart_all",
			"exec_all_masters", "n3", "restart_all", "send_kill_all", "sysbench", "use_all",
		},
		"directory": {"master", "node1", "node2"},
		"regular":   {"sbdescription.json"},
		"nodes":     []string{"master", "node1", "node2"},
	},
	"multiple": {
		"executable": {
			"metadata_all", "start_all", "sysbench_ready",
			"clear_all", "n1", "status_all",
			"exec_all", "n2", "replicate_from", "stop_all", "test_sb_all",
			"n3", "restart_all", "send_kill_all", "sysbench", "use_all",
		},
		"directory": {"node1", "node2", "node3"},
		"regular":   {"sbdescription.json"},
		"nodes":     []string{"node1", "node2", "node3"},
	},
	"group": {
		"executable": []string{
			"check_nodes", "exec_all_slaves", "metadata_all", "start_all", "sysbench_ready", "use_all_masters",
			"clear_all", "initialize_nodes", "n1", "status_all", "test_replication", "use_all_slaves",
			"exec_all", "n2", "replicate_from", "stop_all", "test_sb_all", "wipe_and_restart_all",
			"exec_all_masters", "n3", "restart_all", "send_kill_all", "sysbench", "use_all",
		},
		"directory": {"node1", "node2", "node3"},
		"regular":   {"sbdescription.json"},
		"nodes":     []string{"node1", "node2", "node3"},
	},
	"multi_source": {
		"executable": []string{
			"check_ms_nodes", "exec_all_slaves", "metadata_all", "start_all", "sysbench_ready", "use_all_masters",
			"clear_all", "initialize_ms_nodes", "n1", "status_all", "test_replication", "use_all_slaves",
			"exec_all", "n2", "replicate_from", "stop_all", "test_sb_all", "wipe_and_restart_all",
			"exec_all_masters", "n3", "restart_all", "send_kill_all", "sysbench", "use_all",
		},
		"directory": {"node1", "node2", "node3"},
		"regular":   {"sbdescription.json"},
		"nodes":     []string{"node1", "node2", "node3"},
	},
}
