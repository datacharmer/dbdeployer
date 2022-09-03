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
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/datacharmer/dbdeployer/cmd"
	"github.com/datacharmer/dbdeployer/common"

	"github.com/rogpeppe/go-internal/testscript"
)

func preTest(t *testing.T, dirName string) []string {
	conditionalPrint("entering %s\n", t.Name())
	if dryRun {
		t.Skip("Dry Run")
	}
	if !common.DirExists("testdata") {
		t.Skip("no testdata found")
	}
	// Directories in testdata are created by the setup code in TestMain
	dirs, err := filepath.Glob("testdata/" + dirName + "/*")
	if err != nil {
		t.Skipf("no directories found in testdata/%s", dirName)
	}
	conditionalPrint("Directories: %v\n", dirs)
	return dirs
}

func testDbDeployer(t *testing.T, name string, parallel bool) {
	if parallel {
		t.Parallel()
	}
	dirs := preTest(t, name)
	for _, dir := range dirs {
		subTestName := path.Base(dir)
		conditionalPrint("entering %s/%s", t.Name(), subTestName)
		t.Run(subTestName, func(t *testing.T) {
			testscript.Run(t, testscript.Params{
				Dir:                 dir,
				Cmds:                customCommands(),
				Condition:           customConditions,
				Setup:               dbdeployerSetup(t, dir),
				RequireExplicitExec: true,
				TestWork:            os.Getenv("ts_preserve") != "",
			})
		})
	}
}

func TestFeature(t *testing.T) {
	testDbDeployer(t, "feature", true)
}

func TestReplication(t *testing.T) {
	testDbDeployer(t, "replication", true)
}

func TestMultiSource(t *testing.T) {
	testDbDeployer(t, "multi-source", false)
}

func TestGroup(t *testing.T) {
	testDbDeployer(t, "group", false)
}

func TestMain(m *testing.M) {

	identity := common.BaseName(os.Args[0])
	if identity == "ts.test" {
		preliminaryChecks()
	}
	conditionalPrint("TestMain: starting tests\n")
	exitCode := testscript.RunMain(m, map[string]func() int{
		"dbdeployer": cmd.Execute,
	})

	if identity == "ts.test" {
		if common.DirExists("testdata") && !dryRun {
			_ = os.RemoveAll("testdata")
		}
		os.Exit(exitCode)
	}
}
