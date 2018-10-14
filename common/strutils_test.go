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
	"os"
	"testing"
)

type pathInfo struct {
	value    string
	envVar   string
	expected string
}

func TestReplaceLiteralHome(t *testing.T) {
	os.Setenv("HOME", "/home/Groucho")
	os.Setenv("PWD", "/var/lib/MarxBrothers")
	var paths = []pathInfo{
		{"/home/Groucho/", "HOME", "$HOME/"},
		{"/home/Groucho/path1/path2", "HOME", "$HOME/path1/path2"},
		{"/home/Harpo/path1/path2", "HOME", "/home/Harpo/path1/path2"},
		{"/var/lib/MarxBrothers/path1/path2", "PWD", "$PWD/path1/path2"},
		{"/var/lib/MarxCousins/path1/path2", "PWD", "/var/lib/MarxCousins/path1/path2"},
	}
	for _, p := range paths {
		value := p.value
		envVar := p.envVar
		expected := p.expected
		canary := ReplaceLiteralEnvVar(value, envVar)
		if expected == canary {
			t.Logf("ok    %-35s %-10s =--> %-25s\n", value, "("+envVar+")", expected)
		} else {
			t.Logf("NOT OK %-35s %-10s =--> %-25s\n", value, "("+envVar+")", expected)
			t.Fail()
		}
	}
	for _, p := range paths {
		value := p.expected
		envVar := p.envVar
		expected := p.value
		canary := ReplaceEnvVar(value, envVar)
		if expected == canary {
			t.Logf("ok    %-35s %-10s --=> %-25s\n", value, "("+envVar+")", expected)
		} else {
			t.Logf("NOT OK %-35s %-10s --=> %-25s\n", value, "("+envVar+")", expected)
			t.Fail()
		}
	}
}

func TestTextToBool(t *testing.T) {
	type textBoolData struct {
		input    string
		expected bool
	}
	var data = []textBoolData{
		{"yes", true},
		{"no", false},
		{"true", true},
		{"True", true},
		{"false", false},
		{"False", false},
		{"1", true},
		{"0", false},
		{"unexpected", false},
	}
	for _, tb := range data {
		if tb.expected == TextToBool(tb.input) {
			t.Logf("ok - value '%s' translated to %v\n", tb.input, tb.expected)
		} else {
			t.Logf("ok - value '%s' does not translate to %v\n", tb.input, tb.expected)
			t.Fail()
		}
	}
}

func TestRemoveTrailingSlash(t *testing.T) {
	type trailingSlashData struct {
		input    string
		expected string
	}
	var data = []trailingSlashData{
		{"one/", "one"},
		{"one//", "one/"},
		{"one", "one"},
		{"/", ""},
		{"one/one/", "one/one"},
		{"//", "/"},
		{"", ""},
	}
	for _, ts := range data {
		result := RemoveTrailingSlash(ts.input)
		if result == ts.expected {
			t.Logf("ok - value '%s' becomes '%s'\n", ts.input, ts.expected)
		} else {
			t.Logf("not ok - value '%s' not translated correctly. Expected: '%s' - found '%s'\n",
				ts.input, ts.expected, result)
			t.Fail()
		}
	}
}

func TestStringToIntSlice(t *testing.T) {
	type StringToSliceData struct {
		input    string
		expected []int
	}
	var data = []StringToSliceData{
		{"1 2000 3839 6783 -1", []int{}},
		{"1;2000;3839;6783;-1", []int{}},
		{"1,2000,3839,6783,-a", []int{}},
		{"1,2000,3839,6783,-1", []int{1, 2000, 3839, 6783, -1}},
		{"2, 2001, 3840, 6784, -2", []int{2, 2001, 3840, 6784, -2}},
		{" 3, 2002, 3841, 6785, -3 ", []int{3, 2002, 3841, 6785, -3}},
	}

	for _, sd := range data {
		result, _ := StringToIntSlice(sd.input)
		okEqualInt("slice size", len(sd.expected), len(result), t)
		for N := 0; N < len(sd.expected); N++ {
			if N < len(result) {
				okEqualInt(fmt.Sprintf("slice element %d", N), sd.expected[N], result[N], t)
			}
		}
	}
}
