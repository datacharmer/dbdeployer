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
	"strings"
	"testing"
)

type pathInfo struct {
	value    string
	envVar   string
	expected string
}

func TestReplaceLiteralHome(t *testing.T) {
	saveHome := os.Getenv("HOME")
	savePWD := os.Getenv("PWD")
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
	os.Setenv("HOME", saveHome)
	os.Setenv("PWD", savePWD)
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
		compare.OkEqualInt("slice size", len(sd.expected), len(result), t)
		for N := 0; N < len(sd.expected); N++ {
			if N < len(result) {
				compare.OkEqualInt(fmt.Sprintf("slice element %d", N), sd.expected[N], result[N], t)
			}
		}
	}
}

func TestIntSliceToDottedString(t *testing.T) {
	type IntSliceToDottedStringData struct {
		input    []int
		expected string
	}
	var data = []IntSliceToDottedStringData{
		{[]int{5, 1, 67, 78}, "5.1.67.78"},
		{[]int{5, 1, 67}, "5.1.67"},
		{[]int{5, 67}, "5.67"},
		{[]int{-5, 67}, "-5.67"},
		{[]int{}, ""},
	}
	for _, d := range data {
		result := IntSliceToDottedString(d.input)
		if result == d.expected {
			t.Logf("ok - %+v => '%s' as expected", d.input, result)
		} else {
			t.Logf("not ok - %+v conversion - expected '%s' - found '%s'", d.input, d.expected, result)
			t.Fail()
		}
	}
}

func TestIsEnvSet(t *testing.T) {
	type envSet struct {
		input    string
		expected bool
	}
	var notExistingVar string = "MaryPoppinsLumpOfSugar"
	var justSetVar string = "dbdeployer"

	err := os.Setenv(justSetVar, "excellent")
	if err != nil {
		panic("Can't set an environment variable")
	}
	var data = []envSet{
		{notExistingVar, false},
		{globals.EmptyString, false},
		{justSetVar, true},
	}
	count := 0
	for _, e := range os.Environ() {
		vals := strings.Split(e, "=")
		if len(vals) > 1 && vals[1] != "" {
			data = append(data, envSet{vals[0], true})
		}
		count++
		if count > 5 {
			break
		}
	}
	for _, d := range data {
		result := IsEnvSet(d.input)
		if result == d.expected {
			t.Logf("ok - %s => '%v' as expected", d.input, result)
		} else {
			t.Logf("not ok - '%s' - expected '%v' - found '%v'", d.input, d.expected, result)
			t.Fail()
		}
	}
}

func TestIncludes(t *testing.T) {
	type includeData struct {
		mainStr   string
		searchStr string
		expected  bool
	}
	var data = []includeData{
		{"hello", "he", true},
		{"hello", "llo", true},
		{"hello", "hi", false},
		{"hello", `^llo`, false},
	}
	for _, d := range data {
		result := Includes(d.mainStr, d.searchStr)
		compare.OkEqualBool(fmt.Sprintf("'%s' includes '%s'", d.mainStr, d.searchStr), result, d.expected, t)
	}
}
