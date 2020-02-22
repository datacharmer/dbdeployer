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
	"fmt"
	"github.com/datacharmer/dbdeployer/compare"
	"github.com/datacharmer/dbdeployer/globals"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestLogDirName(t *testing.T) {
	type logDirTest struct {
		args     []string
		expected string
	}

	pid := fmt.Sprintf("%d", os.Getpid())
	var data = []logDirTest{
		{[]string{"single"}, "single_VERSION"},
		{[]string{"multiple"}, "multiple_VERSION"},
		{[]string{"replication"}, "replication_VERSION_master_slave"},
		{[]string{"--topology=master-slave", "replication"}, "replication_VERSION_master_slave"},
		{[]string{"replication", "--topology=group"}, "replication_VERSION_group"},
		{[]string{"replication", "--topology=group", "--single-primary"}, "replication_VERSION_group_sp"},
		{[]string{"replication", "--topology=all-masters"}, "replication_VERSION_all_masters"},
		{[]string{"replication", "--topology=fan-in"}, "replication_VERSION_fan_in"},
		{[]string{"replication", "--topology=UNUSED"}, "replication_VERSION_UNUSED"},
	}
	var versions = []string{"5.7.93", "8.0.94"}
	re := regexp.MustCompile(`VERSION`)
	for _, v := range versions {
		for _, d := range data {
			vname := VersionToName(v)
			expected := re.ReplaceAllString(d.expected, vname)
			CommandLineArgs = d.args
			CommandLineArgs = append(CommandLineArgs, v)
			result := LogDirName()
			compare.OkEqualString(fmt.Sprintf("Log dir name [%v]", CommandLineArgs), result, fmt.Sprintf("%s-%s", expected, pid), t)
		}
	}
}

func TestParseConfigFile(t *testing.T) {
	var sampleConfig = ConfigOptions{
		"label1": {
			KeyValue{Key: "one", Value: "1"},
			KeyValue{Key: "two", Value: "2"},
		},
		"label2": {
			{"abc", "hello"},
			{"def", "world"},
		},
	}
	var sampleConfigText1 = `
[label1]
one=1
two=2
[label2]
abc=hello
def=world
`
	var sampleConfigText2 = `
# sample comment
[label1]
one     =    1

# another sample comment

two     =   2

[label2]
abc = hello
def = world

`
	var sampleConfigFile = "/tmp/sample_config.ini"
	for _, sampleConfigText := range []string{sampleConfigText1, sampleConfigText2} {
		err := WriteString(sampleConfigText, sampleConfigFile)
		compare.OkIsNil("err for written sample file", err, t)
		readConfig, err := ParseConfigFile(sampleConfigFile)
		compare.OkIsNil("err for read sample file", err, t)
		for k := range sampleConfig {
			val, ok := readConfig[k]
			compare.OkEqualBool("key", ok, true, t)
			count := 0
			for _, item := range val {
				compare.OkEqualString(fmt.Sprintf("key: %s", k), sampleConfig[k][count].Key, item.Key, t)
				compare.OkEqualString(fmt.Sprintf("val: %s", k), sampleConfig[k][count].Value, item.Value, t)
				count++
			}
		}
	}
}

func TestExists(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("Could not get current file name")
	}
	exists := FileExists(currentFile)
	compare.OkEqualBool(fmt.Sprintf("file %s exists", currentFile), exists, true, t)
	dir := filepath.Dir(currentFile)
	if dir == "" {
		t.Skip("Could not get directory for current test file")
	}
	dir2 := DirName(currentFile)
	compare.OkEqualString("dirName", dir2, dir, t)
	bareName := filepath.Base(currentFile)
	bareName2 := BaseName(currentFile)
	compare.OkEqualString("baseName", bareName2, bareName, t)
	exists = DirExists(dir)
	compare.OkEqualBool(fmt.Sprintf("dir %s exists", dir), exists, true, t)
	upperDir := path.Join(".", "..")
	upperDirAbs, err := AbsolutePath(upperDir)
	if err != nil {
		t.Skip(fmt.Sprintf("Could not get absolute directory for %s", upperDir))
	}

	calculatedUpperDirAbs := filepath.Dir(dir)
	compare.OkEqualString("upper directory ./..", upperDirAbs, calculatedUpperDirAbs, t)
}

func TestWhich(t *testing.T) {
	goExec := Which("go")
	compare.OkMatchesString("[Which] Go executable is not an empty string", goExec, `.+`, t)
	exists := ExecExists(goExec)
	compare.OkEqualBool(fmt.Sprintf("Go executable '%s' exists", goExec), exists, true, t)
	goExec = FindInPath("go")
	compare.OkMatchesString("[FindInPath] Go executable is not an empty string", goExec, `.+`, t)
	exists = ExecExists(goExec)
	compare.OkEqualBool(fmt.Sprintf("Go executable '%s' exists", goExec), exists, true, t)
}

func TestSandboxDescription(t *testing.T) {
	pid := os.Getpid()
	var sd = SandboxDescription{
		Basedir:           fmt.Sprintf("basedir%d", pid),
		SBType:            "single",
		Version:           fmt.Sprintf("5.7.%d", pid),
		Nodes:             0,
		NodeNum:           0,
		DbDeployerVersion: VersionDef,
		CommandLine:       "",
		LogFile:           "",
	}
	descriptionDir := path.Join("/tmp", "test_sd")
	if !DirExists(descriptionDir) {
		err := os.Mkdir(descriptionDir, globals.PublicDirectoryAttr)
		if err != nil {
			t.Skip(fmt.Sprintf("could not create directory %s", descriptionDir))
		}
	}
	err := WriteSandboxDescription(descriptionDir, sd)
	compare.OkIsNil("err for write sandbox description", err, t)
	if err != nil {
		t.Skip(fmt.Sprintf("Can't write sandbox description into %s", descriptionDir))
	}
	newSd, err := ReadSandboxDescription(descriptionDir)
	compare.OkIsNil("err for read sandbox description", err, t)
	if err != nil {
		t.Skip(fmt.Sprintf("Can't read sandbox description from %s", descriptionDir))
	}
	compare.OkEqualString("basedir", newSd.Basedir, sd.Basedir, t)
	compare.OkEqualString("sb type", newSd.SBType, sd.SBType, t)
	compare.OkEqualString("version", newSd.Version, sd.Version, t)
	compare.OkMatchesString("timestamp is not empty", newSd.Timestamp, `.+`, t)
}

func TestWriteStrings(t *testing.T) {
	pid := os.Getpid()
	textFile := path.Join("/tmp", fmt.Sprintf("test%d", pid))

	copiedFile := textFile + "_copy"

	type DataWrite struct {
		termination string
		elements    int
		elements2   int
		expected    []string
		expected2   []string
	}

	lines := []string{"one", "two", "three"}
	appendLine := "four"

	data := []DataWrite{
		{"", 1, 1, []string{"onetwothree"}, []string{"onetwothreefour"}},
		{":", 1, 1, []string{"one:two:three:"}, []string{"one:two:three:four:"}},
		{"\n", 3, 4, []string{"one", "two", "three"}, []string{"one", "two", "three", "four"}},
	}

	for I, d := range data {
		err := WriteStrings(lines, textFile, d.termination)
		compare.OkIsNil(fmt.Sprintf("[%d] err writing lines text file", I), err, t)
		if err != nil {
			t.Skip(fmt.Sprintf("[%d] error writing lines to file %s", I, textFile))
		}
		err = CopyFile(textFile, copiedFile)
		compare.OkIsNil(fmt.Sprintf("[%d] err copying text file", I), err, t)
		if err != nil {
			t.Skip(fmt.Sprintf("[%d] error copying text file %s", I, textFile))
		}
		newLines, err := SlurpAsLines(copiedFile)
		compare.OkIsNil(fmt.Sprintf("[%d] err reading lines from text file", I), err, t)
		if err != nil {
			t.Skip(fmt.Sprintf("[ %d] error reading lines from file %s", I, copiedFile))
		}
		compare.OkEqualInt(fmt.Sprintf("[%d] read elements same as written elements", I), d.elements, len(newLines), t)
		compare.OkEqualStringSlices(t, newLines, d.expected)

		err = AppendStrings([]string{appendLine}, textFile, d.termination)
		compare.OkIsNil(fmt.Sprintf("[%d] err appending lines to text file", I), err, t)
		if err != nil {
			t.Skip(fmt.Sprintf("[ %d] error appending lines to text file %s: %+v", I, textFile, err))
		}
		newLines, err = SlurpAsLines(textFile)
		compare.OkIsNil(fmt.Sprintf("[%d] err reading lines from appended text file", I), err, t)
		if err != nil {
			t.Skip(fmt.Sprintf("[ %d] error reading lines from appended file %s", I, textFile))
		}
		compare.OkEqualInt(fmt.Sprintf("[%d] read appended elements same as written elements", I), d.elements2, len(newLines), t)
		compare.OkEqualStringSlices(t, newLines, d.expected2)

		allInOne := strings.Join(lines, d.termination)
		err = WriteString(allInOne, textFile)
		compare.OkIsNil(fmt.Sprintf("[%d] err writing string to text file", I), err, t)
		if err != nil {
			t.Skip(fmt.Sprintf("[%d] error writing string to file %s", I, textFile))
		}
		newText, err := SlurpAsString(textFile)
		compare.OkIsNil(fmt.Sprintf("[%d] err reading string from text file", I), err, t)
		if err != nil {
			t.Skip(fmt.Sprintf("[ %d] error reading string from file %s", I, textFile))
		}

		compare.OkEqualString(fmt.Sprintf("[%d] write/read string", I), allInOne, newText, t)
		buf, err := SlurpAsBytes(textFile)

		compare.OkIsNil(fmt.Sprintf("[%d] err reading bytes from text file", I), err, t)
		// t.Logf("%#v\n", buf)
		compare.OkEqualByteSlices(t, buf, []byte(newText))
	}
}

func createCommand(fileName string, command string) error {
	err := WriteString(command, fileName)
	if err != nil {
		return err
	}
	return os.Chmod(fileName, globals.ExecutableFileAttr)
}

func TestRunCmd(t *testing.T) {

	scriptText := `#!/bin/bash
value=$1
[ -z "$value" ] && value=noargs
if [ "$value" == "fail" ]
then
    echo -n "stout <$2>"
    echo -n 1>&2 "stderr [$2]"
    exit 1
fi
echo -n "You asked for $value, didn't you?"
`
	scriptName := path.Join("/tmp", "testcmd")

	// First, a failing test
	err := createCommand(scriptName, "#!/bin/bash\nexit 1")
	compare.OkIsNil("command creation err", err, t)
	if err != nil {
		t.Skip(fmt.Sprintf("error creating command: %s", err))
	}
	_, err = RunCmd(scriptName)
	compare.OkIsNotNil("[RunCmd] command execution expected err", err, t)
	if err == nil {
		t.Logf("[RunCmd] unexpected success of failing command")
		t.Fail()
	}

	err = createCommand(scriptName, scriptText)
	compare.OkIsNil("command creation err", err, t)
	if err != nil {
		t.Skip(fmt.Sprintf("error creating command: %s", err))
	}

	out, err := RunCmd(scriptName)
	compare.OkIsNil("[RunCmd] command execution err", err, t)
	if err != nil {
		t.Logf("[RunCmd] error executing command: %s", err)
		t.Fail()
	}
	compare.OkMatchesString("[RunCmd] command result", out, `noargs`, t)

	out, err = RunCmdCtrl(scriptName, true)
	compare.OkIsNil("[RunCmdCtrl] command execution err", err, t)
	if err != nil {
		t.Logf("[RunCmdCtrl] error executing command: %s", err)
		t.Fail()
	}
	compare.OkMatchesString("[RunCmdCtrl] command result", out, `noargs`, t)

	out, err = RunCmdWithArgs(scriptName, []string{"withArgs"})
	compare.OkIsNil("[RunCmdWithArgs] command execution err", err, t)
	if err != nil {
		t.Logf("[RunCmdWIthArgs] error executing command: %s", err)
		t.Fail()
	}
	compare.OkMatchesString("[RunCmdWithArgs] command result", out, `withArgs`, t)

	outText, errText, err := runCmdCtrlArgs(scriptName, false, []string{"fail", "both streams"}...)
	compare.OkMatchesString("[runCmdCtrlArgs] command result", errText, `\[both streams\]`, t)
	compare.OkMatchesString("[runCmdCtrlArgs] command result", outText, `<both streams>`, t)
	compare.OkIsNotNil("[runCmdCtrlArgs] command execution expected err", err, t)
}
