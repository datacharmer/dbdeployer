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
	"os"
	"regexp"
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
