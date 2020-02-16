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
	"testing"
)

type testStringBool struct {
	input    string
	expected bool
}

func TestIsVersion(t *testing.T) {
	var data = []testStringBool{
		{"1.2.3", true},
		{"abc1.2.3", true},
		{"1.2", false},
		{"1.2.3abc", false},
		{"abc1.2", false},
		{"11.22.30", true},
	}
	for _, tv := range data {
		result := IsVersion(tv.input)
		compare.OkEqualBool(fmt.Sprintf("is version: %s", tv.input), tv.expected, result, t)
	}
}

func TestIsATarball(t *testing.T) {
	var data = []testStringBool{
		{"dummy.tar.gz", true},
		{"dummy.tar.xz", true},
		{"dummy.targz", false},
		{"dummy.tar", false},
		{"dummy.gz", false},
		{"dummy.xz", false},
	}
	for _, tv := range data {
		result := IsATarball(tv.input)
		compare.OkEqualBool(fmt.Sprintf("is a tarball: %s", tv.input), tv.expected, result, t)
	}
}

func TestIsAUrl(t *testing.T) {
	var data = []testStringBool{
		{"dummy.tar.gz", false},
		{"dummy.tar.xz", false},
		{"http://dummy.tar.xz", false},
		{"ftp://dummy.tar.xz", false},
		{"http://example.com/dummy.tar.xz", true},
		{"ssh://example.com/dummy.tar.xz", false},
		{"https://example.com/dummy.tar.xz", true},
		{"https://example.com/some/path/dummy.tar.xz", true},
	}
	for _, dataItem := range data {
		result := IsUrl(dataItem.input)
		compare.OkEqualBool(fmt.Sprintf("is a url: %s", dataItem.input), dataItem.expected, result, t)
	}
}

func TestFindFreePort(t *testing.T) {
	type testFreePort struct {
		usedPorts []int
		basePort  int
		howMany   int
		expected  int
	}
	var data = []testFreePort{
		{usedPorts: []int{}, basePort: 5000, howMany: 1, expected: 5000},
		{usedPorts: []int{4999, 5001}, basePort: 5000, howMany: 1, expected: 5000},
		{usedPorts: []int{3306, 1186, 33060, 33062}, basePort: 5000, howMany: 1, expected: 5000},
		{usedPorts: []int{3306, 1186, 33060, 33062, 5000}, basePort: 5000, howMany: 1, expected: 5001},
		{usedPorts: []int{5000, 5001, 5002}, basePort: 5000, howMany: 1, expected: 5003},
		{usedPorts: []int{5000, 5001, 5002}, basePort: 5000, howMany: 3, expected: 5003},
		{usedPorts: []int{5000, 5001, 5002, 5005}, basePort: 5000, howMany: 3, expected: 5006},
		{usedPorts: []int{5000, 5001, 5002, 5006, 5007, 5008}, basePort: 5000, howMany: 3, expected: 5003},
		{usedPorts: []int{5000, 5001, 5002, 5006, 5007, 5008}, basePort: 5000, howMany: 4, expected: 5009},
	}
	for _, d := range data {
		result, err := FindFreePort(d.basePort, d.usedPorts, d.howMany)
		compare.OkIsNil("FindFreePort result", err, t)
		compare.OkEqualInt(fmt.Sprintf("Free ports %v : %d:%d", d.usedPorts, d.basePort, d.howMany), d.expected, result, t)
	}
}

func TestFindSandbox(t *testing.T) {
	type testFIndSb struct {
		sample      []SandboxInfo
		matching    []string
		nonMatching []string
		wanted      string
	}

	var sampleSbInfo1 = []SandboxInfo{
		{
			SandboxName: "one",
			SandboxDesc: SandboxDescription{
				SBType:  "single",
				Version: "8.0.19",
				Flavor:  "mysql",
				Port:    []int{8019, 18019},
			},
		},
		{
			SandboxName: "two",
			SandboxDesc: SandboxDescription{
				SBType:  "replication",
				Version: "8.0.18",
				Flavor:  "mysql",
				Port:    []int{20819, 30819, 20820, 30820, 20821, 30821},
			},
		},
	}

	var sampleSbInfo2 = []SandboxInfo{
		{
			SandboxName: "three",
			SandboxDesc: SandboxDescription{
				SBType:  "single",
				Version: "8.0.19",
				Flavor:  "mysql",
				Port:    []int{8019, 18019},
			},
		},
		{
			SandboxName: "four",
			SandboxDesc: SandboxDescription{
				SBType:  "single",
				Version: "8.0.18",
				Flavor:  "mysql",
				Port:    []int{8018, 18018},
			},
		},
		{
			SandboxName: "five",
			SandboxDesc: SandboxDescription{
				SBType:  "single",
				Version: "8.0.18",
				Flavor:  "percona",
				Port:    []int{8018, 18018},
			},
		},
	}
	var data = []testFIndSb{
		{
			sample:      sampleSbInfo1,
			wanted:      "one",
			matching:    []string{"one", "single", "8.0.19", "8019"},
			nonMatching: []string{"mysql"},
		},
		{
			sample:      sampleSbInfo1,
			wanted:      "two",
			matching:    []string{"two", "replication", "8.0.18", "20819"},
			nonMatching: []string{"mysql"},
		},
		{
			sample:      sampleSbInfo2,
			wanted:      "three",
			matching:    []string{"three", "8.0.19", "8019"},
			nonMatching: []string{"mysql", "single"},
		},
		{
			sample:      sampleSbInfo2,
			wanted:      "four",
			matching:    []string{"four"},
			nonMatching: []string{"mysql", "single", "8.0.18", "8018"},
		},
		{
			sample:      sampleSbInfo2,
			wanted:      "five",
			matching:    []string{"five", "percona"},
			nonMatching: []string{"single", "8.0.18", "8018"},
		},
	}
	for _, info := range data {
		for _, m := range info.matching {
			result, err := FindSandbox(info.sample, m)
			if err != nil {
				t.Logf("not ok - %s : %s", m, err)
				t.Fail()
			}
			if result.SandboxName == info.wanted {
				t.Logf("ok - param '%s' - got '%s' as expected", m, info.wanted)
			} else {
				t.Logf("not ok - not matching (%s): got '%s' instead of '%s'", m, result.SandboxName, info.wanted)
				t.Fail()
			}
		}
		for _, nm := range info.nonMatching {
			result, err := FindSandbox(info.sample, nm)
			if err == nil {
				t.Logf("not ok - expected failure but got success %s", nm)
				t.Fail()
			} else {
				t.Logf("ok - search for '%s' failed as expected", nm)
			}
			if result.SandboxName != "" {
				t.Logf("not ok - expected failure but got success %s (%s)", nm, result.SandboxName)
				t.Fail()
			}
		}
	}
}
