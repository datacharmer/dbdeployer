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
	"regexp"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/rogpeppe/go-internal/testscript"
	"golang.org/x/exp/constraints"
)

func isANumber(s string) bool {
	return regexp.MustCompile(`^[0-9]+$`).MatchString(s)
}

func assertEqual[T comparable](ts *testscript.TestScript, a, b T, msg string, args ...interface{}) {
	if a != b {
		ts.Fatalf(msg, args...)
	}
}

func assertFileExists(ts *testscript.TestScript, fName string, msg string, args ...interface{}) {
	if !common.FileExists(fName) {
		ts.Fatalf(msg, args...)
	}
}

func assertFileNotExists(ts *testscript.TestScript, fName string, msg string, args ...interface{}) {
	if common.FileExists(fName) {
		ts.Fatalf(msg, args...)
	}
}

func assertDirExists(ts *testscript.TestScript, dir string, msg string, args ...interface{}) {
	if !common.DirExists(dir) {
		ts.Fatalf(msg, args...)
	}
}

func assertExecExists(ts *testscript.TestScript, dir string, msg string, args ...interface{}) {
	if !common.ExecExists(dir) {
		ts.Fatalf(msg, args...)
	}
}

func assertGreater[T constraints.Ordered](ts *testscript.TestScript, a, b T, msg string, args ...interface{}) {
	if a <= b {
		ts.Fatalf(msg, args...)
	}
}

func assertGreaterEqual[T constraints.Ordered](ts *testscript.TestScript, a, b T, msg string, args ...interface{}) {
	if a < b {
		ts.Fatalf(msg, args...)
	}
}
