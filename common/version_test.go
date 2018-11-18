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
	"github.com/datacharmer/dbdeployer/compare"
	"testing"
)

type versionPort struct {
	version string
	port    int
}

type versionPair struct {
	version     string
	versionList []int
	expected    bool
}

type uuidComponent struct {
	port     int
	nodeNum  int
	expected string
}

func TestVersionToPort(t *testing.T) {
	//t.Parallel()
	var versions = []versionPort{
		{"", -1},            // FAIL: Empty string
		{"5.0.A", -1},       // FAIL: Hexadecimal number
		{"5.0.-9", -1},      // FAIL: Negative revision
		{"-5.0.9", -1},      // FAIL: Negative major version
		{"5.-1.9", -1},      // FAIL: Negative minor version
		{"5096", -1},        // FAIL: No separators
		{"50.96", -1},       // FAIL: Not enough separators
		{"dummy", -1},       // FAIL: Not numbers
		{"5.0.96.2", -1},    // FAIL: Too many components
		{"5.0.96", 5096},    // OK: 5.0
		{"5.1.72", 5172},    // OK: 5.1
		{"5.5.55", 5555},    // OK: 5.5
		{"ps5.7.20", 5720},  // OK: 5.7 with prefix
		{"5.7.21", 5721},    // OK: 5.7
		{"8.0.0", 8000},     // OK: 8.0
		{"8.0.4", 8004},     // OK: 8.0
		{"8.0.04", 8004},    // OK: 8.0
		{"ma10.2.1", 10201}, // OK: 10.2 with prefix
	}
	//t.Logf("Name: %s\n", t.Name())
	for _, vp := range versions {
		version := vp.version
		expected := vp.port
		port := VersionToPort(version)
		//t.Logf("+%s\n", version)
		if expected == port {
			t.Logf("ok     %-10s => %5d\n", version, port)
		} else {
			t.Logf("NOT OK %-10s => %5d\n", version, port)
			t.Fail()
		}
	}
}

func TestGreaterOrEqualVersion(t *testing.T) {

	var versions = []versionPair{
		{"5.0.0", []int{5, 6, 0}, false},
		{"8.0.0", []int{5, 6, 0}, true},
		{"ps5.7.5", []int{5, 7, 0}, true},
		{"10.0.1", []int{5, 6, 0}, false},
	}
	for _, v := range versions {
		result := GreaterOrEqualVersion(v.version, v.versionList)
		if v.expected == result {
			t.Logf("ok     %-10s => %v %v \n", v.version, v.versionList, result)
		} else {
			t.Logf("NOT OK %-10s => %v %v \n", v.version, v.versionList, result)
			t.Fail()
		}
	}
}

func TestCustomUuid(t *testing.T) {
	var uuidSamples = []uuidComponent{
		//                            12345678 1234 1234 1234 123456789012
		//                           "00000000-0000-0000-0000-000000000000"
		{5000, 0, "00005000-0000-0000-0000-000000005000"},
		{15000, 0, "00015000-0000-0000-0000-000000015000"},
		{15000, 1, "00015000-1111-1111-1111-111111111111"},
		{25000, 2, "00025000-2222-2222-2222-222222222222"},
		{12987, 7, "00012987-7777-7777-7777-777777777777"},
		{20742, 1, "00020742-1111-1111-1111-111111111111"},
		{8004, 0, "00008004-0000-0000-0000-000000008004"},
		{8004, 11, "00008004-0011-0011-0011-000000008004"},
		{8004, 3452, "00008004-3452-3452-3452-000000008004"},
		{8004, 18976, "00008004-0000-0001-8976-000000008004"},
		{6000, 35281, "00006000-0000-0003-5281-000000006000"},
		{6000, 235281, "00006000-0023-0000-0000-000000006000"},
	}
	for _, sample := range uuidSamples {
		_, newUuid := MakeCustomizedUuid(sample.port, sample.nodeNum)
		if newUuid == sample.expected {
			t.Logf("ok     %5d %6d => %s \n", sample.port, sample.nodeNum, newUuid)
		} else {
			t.Logf("NOT OK %5d %6d => <%#v> (expected: <%#v>) \n", sample.port, sample.nodeNum, newUuid, sample.expected)
			t.Fail()
		}
	}
	err, newUuid := MakeCustomizedUuid(5000, 10000001)
	compare.OkEqualString("over boundaries node", newUuid, "", t)
	compare.OkIsNotNil("over boundaries node", err, t)
}

type expectedData struct {
	index int
	value string
}
type sortVersionData struct {
	data     []string
	expected []expectedData
}

func checkSortVersion(t *testing.T, sortData sortVersionData) {
	sorted := SortVersions(sortData.data)
	for _, exp := range sortData.expected {
		if exp.value == sorted[exp.index] {
			t.Logf("ok - element %d = '%s'\n", exp.index, exp.value)
		} else {
			t.Logf("not ok - out of position element %d - Expected: '%s' - Found: '%s'\n",
				exp.index, exp.value, sorted[exp.index])
			t.Fail()
		}
	}
}

func TestSortVersions(t *testing.T) {

	var sortData = []sortVersionData{
		{
			data: []string{"5.0.1", "5.0.11", "5.0.9", "5.0.6", "5.0.10"},
			expected: []expectedData{
				{0, "5.0.1"},
				{4, "5.0.11"},
			},
		},
		{
			data: []string{"8.0.11", "8.0.1", "5.0.9", "5.0.6", "5.0.10"},
			expected: []expectedData{
				{0, "5.0.6"},
				{4, "8.0.11"},
			},
		},
		{
			data: []string{"10.0.2", "8.0.1", "5.1.5", "5.0.6", "5.0.10"},
			expected: []expectedData{
				{0, "5.0.6"},
				{4, "10.0.2"},
			},
		},
		{
			data: []string{"ps8.0.2", "ps8.0.1", "ps5.1.5", "ps5.0.6", "ps5.0.10"},
			expected: []expectedData{
				{0, "ps5.0.6"},
				{4, "ps8.0.2"},
			},
		},
	}

	for _, sd := range sortData {
		checkSortVersion(t, sd)
	}
}
