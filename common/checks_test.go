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
	"testing"
)

func TestIsVersion(t *testing.T) {
	type testVersion struct {
		candidate string
		expected  bool
	}
	var data = []testVersion{
		{"1.2.3", true},
		{"abc1.2.3", true},
		{"1.2", false},
		{"1.2.3abc", false},
		{"abc1.2", false},
		{"11.22.30", true},
	}
	for _, tv := range data {
		result := IsVersion(tv.candidate)
		okEqualBool(fmt.Sprintf("is version: %s", tv.candidate), tv.expected, result, t)
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
		{usedPorts: []int{3306, 1186, 33060}, basePort: 5000, howMany: 1, expected: 5000},
		{usedPorts: []int{3306, 1186, 33060, 5000}, basePort: 5000, howMany: 1, expected: 5001},
		{usedPorts: []int{5000, 5001, 5002}, basePort: 5000, howMany: 1, expected: 5003},
		{usedPorts: []int{5000, 5001, 5002}, basePort: 5000, howMany: 3, expected: 5003},
		{usedPorts: []int{5000, 5001, 5002, 5005}, basePort: 5000, howMany: 3, expected: 5006},
		{usedPorts: []int{5000, 5001, 5002, 5006, 5007, 5008}, basePort: 5000, howMany: 3, expected: 5003},
		{usedPorts: []int{5000, 5001, 5002, 5006, 5007, 5008}, basePort: 5000, howMany: 4, expected: 5009},
	}
	for _, d := range data {
		result := FindFreePort(d.basePort, d.usedPorts, d.howMany)
		okEqualInt(fmt.Sprintf("Free ports %v : %d:%d", d.usedPorts, d.basePort, d.howMany), d.expected, result, t)
	}
}
