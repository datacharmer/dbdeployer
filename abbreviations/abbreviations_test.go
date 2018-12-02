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

package abbreviations

import (
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/compare"
	"os"
	"testing"
)

func TestLoadAbbreviations(t *testing.T) {
	userDefinedFile = "/tmp/abbreviations.txt"
	type abbrData struct {
		commandLine []string
		expected    []string
	}
	abbreviations := []string{
		"groupr deploy replication --topology=group --concurrent",
		"groupsp deploy replication --topology=group --single-primary --concurrent",
		"sbs sandboxes --header",
		"msbdef --sandbox-directory={{.sb}} --base-port={{.port}}",
	}
	err := common.WriteStrings(abbreviations, userDefinedFile, "\n")
	compare.OkIsNil("writing abbreviations", err, t)
	if err != nil {
		t.Fatalf("can't write abbreviation file %s", userDefinedFile)
	}
	saveArgs := os.Args
	var data = []abbrData{
		{
			commandLine: []string{"groupr", "8.0"},
			expected:    []string{"deploy", "replication", "--topology=group", "--concurrent", "8.0"},
		},
		{
			commandLine: []string{"groupsp", "8.0"},
			expected:    []string{"deploy", "replication", "--topology=group", "--single-primary", "--concurrent", "8.0"},
		},
		{
			commandLine: []string{"sbs"},
			expected:    []string{"sandboxes", "--header"},
		},
		{
			commandLine: []string{"deploy", "replication", "msbdef:port=9999,sb=DummyDir", "5.7.22"},
			expected:    []string{"deploy", "replication", "--sandbox-directory=DummyDir", "--base-port=9999", "5.7.22"},
		},
	}
	for _, d := range data {
		os.Args = d.commandLine
		LoadAbbreviations()
		compare.OkEqualStringSlices(t, os.Args, d.expected)
	}

	os.Args = saveArgs
	compare.OkEqualStringSlices(t, os.Args, saveArgs)
}
