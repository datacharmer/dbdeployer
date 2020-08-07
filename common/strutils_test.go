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
	home     string
	pwd      string
	value    string
	envVar   string
	expected string
}

func TestReplaceLiteralHome(t *testing.T) {
	saveHome := os.Getenv("HOME")
	savePWD := os.Getenv("PWD")
	groucho := "/home/Groucho"
	brothers := "/var/lib/MarxBrothers"
	var paths = []pathInfo{
		{groucho, brothers, "/home/Groucho/", "HOME", "$HOME/"},
		{groucho, brothers, "/home/Groucho/path1/path2", "HOME", "$HOME/path1/path2"},
		{groucho, brothers, "/home/Harpo/path1/path2", "HOME", "/home/Harpo/path1/path2"},
		{groucho, brothers, "/var/lib/MarxBrothers/path1/path2", "PWD", "$PWD/path1/path2"},
		{groucho, brothers, "/var/lib/MarxCousins/path1/path2", "PWD", "/var/lib/MarxCousins/path1/path2"},
		{groucho, "/", "/var/lib/MarxCousins/path1/path2", "PWD", "/var/lib/MarxCousins/path1/path2"},
		{groucho, "", "/var/lib/MarxCousins/path1/path2", "PWD", "/var/lib/MarxCousins/path1/path2"},
		{groucho, brothers, "$PWD/home/Groucho/path2", "HOME", "$PWD/home/Groucho/path2"},
	}
	for _, p := range paths {
		os.Setenv("HOME", p.home)
		os.Setenv("PWD", p.pwd)
		value := p.value
		envVar := p.envVar
		expected := p.expected
		canary := ReplaceLiteralEnvVar(value, envVar)
		if expected == canary {
			t.Logf("ok    %-35s %-10s =--> %-25s\n", value, "("+envVar+")", expected)
		} else {
			t.Logf("NOT OK - got %-35s %-10s =--> want %-25s\n", value, "("+envVar+")", expected)
			t.Fail()
		}
	}
	for _, p := range paths {
		os.Setenv("HOME", p.home)
		os.Setenv("PWD", p.pwd)
		value := p.expected
		envVar := p.envVar
		expected := p.value
		canary := ReplaceEnvVar(value, envVar)
		if expected == canary {
			t.Logf("ok    %-35s %-10s --=> %-25s\n", value, "("+envVar+")", expected)
		} else {
			t.Logf("NOT OK - got %-35s %-10s --=> want %-25s\n", value, "("+envVar+")", expected)
			t.Fail()
		}
	}
	os.Setenv("HOME", saveHome)
	os.Setenv("PWD", savePWD)
}

func TestTextToBool(t *testing.T) {
	var data = []testStringBool{
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

func describeBlank(s string) string {
	result := ""
	for _, c := range s {
		if result != "" {
			result += "-"
		}
		switch {
		case c == ' ':
			result += "space"
		case c == '\t':
			result += "tab"
		case c == '\n':
			result += "newline"
		case c >= '0' && c <= '9':
			result += "number"
		case c == '\r':
			result += "CR"
		case c >= 'a' && c <= 'z':
			result += "alpha"
		case c >= 0 && c < ' ':
			result += "nonprintable"
		default:
			result += fmt.Sprintf("'%c'", c)
		}
	}
	if result == "" {
		result = "empty"
	}
	return result
}

func TestIsEmptyOrBlank(t *testing.T) {
	var data = []testStringBool{
		{"", true},
		{" ", true},
		{"  ", true},
		{"   ", true},
		{"\t", true},
		{"\t\t", true},
		{"\t\t\t", true},
		{"\t \t", true},
		{" \t", true},
		{" \r", true},
		{" \r\n", true},
		{"\t\r\n", true},
		{"\r\n", true},
		{" \t\n", true},
		{"\n", true},
		{"\n ", true},
		{"\n\t ", true},
		{"\t\n", true},
		{"\t\t\n", true},
		{"\t \n", true},
		{"\t.\n", false},
		{"\tA\n", false},
		{"hello", false},
		{"0123", false},
		{"0a23", false},
		{",a!3", false},
		{",!", false},
	}
	for _, d := range data {
		result := IsEmptyOrBlank(d.input)
		compare.OkEqualBool(fmt.Sprintf("<%s> is empty or blank", describeBlank(d.input)), result, d.expected, t)
	}
}

func TestCoalesceString(t *testing.T) {
	type includeData struct {
		items    []string
		expected string
	}
	var data = []includeData{
		{[]string{"hello", "world"}, "hello"},
		{[]string{"", "hello", "world"}, "hello"},
		{[]string{"", "", "hello", "world"}, "hello"},
		{[]string{"", "", "", ""}, ""},
	}
	for _, d := range data {
		result := CoalesceString(d.items...)
		compare.OkEqualString(fmt.Sprintf("one of '%#v'", d.items), result, d.expected, t)
	}
}

func TestCoalesce(t *testing.T) {
	type includeData struct {
		items    []interface{}
		expected interface{}
	}
	var data = []includeData{
		{[]interface{}{"hello", "world"}, "hello"},
		{[]interface{}{nil, "hello", "world"}, "hello"},
		{[]interface{}{nil, nil, "hello", "world"}, "hello"},
		{[]interface{}{nil, nil, "", "world"}, ""},
		{[]interface{}{"", nil, 1, ""}, ""},
		{[]interface{}{nil, 1, ""}, 1},
		{[]interface{}{1, ""}, 1},
		{[]interface{}{"", 1, ""}, ""},
		{[]interface{}{nil, nil}, nil},
	}
	for _, d := range data {
		result := Coalesce(d.items...)
		if d.expected == nil {
			compare.OkIsNil(fmt.Sprintf("one of '%#v' (%#v)", d.items, result), result, t)
		} else {
			compare.OkIsNotNil(fmt.Sprintf("one of '%#v' (%#v)", d.items, result), result, t)
			compare.OkEqualInterface(fmt.Sprintf("one of '%#v' (%#v)", d.items, result), result, d.expected, t)
		}
	}
}

func TestSortVersionsSubset(t *testing.T) {

	type versionList struct {
		versions        []string
		wanted          string
		expectedVersion string
		expectedSize    int
	}
	var versionData = []versionList{
		{
			// homogeneous list
			versions:        []string{"abc5.7.8", "abc5.7.10", "abc5.7.7"},
			wanted:          "abc5.7",
			expectedVersion: "abc5.7.10",
		},
		{
			// Different prefix, but same short version
			versions:        []string{"abc5.7.9", "abc5.7.10", "abc5.7.8", "5.7.11"},
			wanted:          "abc5.7",
			expectedVersion: "5.7.11",
			expectedSize:    4,
		},
		{
			// mixed short versions
			versions:        []string{"abc5.7.7", "abc5.7.8", "abc5.7.10", "8.0.11"},
			wanted:          "abc5.7",
			expectedVersion: "abc5.7.10",
			expectedSize:    3,
		},
		{
			// same prefix, different short versions, request lower version
			versions:        []string{"abc8.0.23", "abc5.7.7", "abc5.7.8", "abc5.7.10", "8.0.11"},
			wanted:          "abc5.7",
			expectedVersion: "abc5.7.10",
			expectedSize:    3,
		},
		{
			// same prefix, different short versions, request higher version
			versions:        []string{"abc8.0.23", "abc5.7.7", "abc5.7.8", "abc5.7.10", "8.0.11"},
			wanted:          "abc8.0",
			expectedVersion: "abc8.0.23",
			expectedSize:    2,
		},
	}

	for _, vd := range versionData {
		sorted := SortVersionsSubset(vd.versions, vd.wanted)

		expectedSize := len(vd.versions)
		if vd.expectedSize > 0 {
			expectedSize = vd.expectedSize
		}
		fmt.Printf("%#v\n", sorted)
		compare.OkEqualInt("sort size", len(sorted), expectedSize, t)

		found := sorted[len(sorted)-1]
		compare.OkEqualString("found latest version", vd.expectedVersion, found, t)
	}
}

func TestOptionComponents(t *testing.T) {
	type optionData struct {
		input            string
		expectedValue    string
		expectedNegation bool
	}
	var optionInfo = []optionData{
		{"foo", "foo", false},
		{"!foo", "foo", true},
		{"no-foo", "foo", true},
		{"not-foo", "foo", true},
		{"boo-foo", "boo-foo", false},
	}

	for _, item := range optionInfo {
		value, negation := OptionComponents(item.input)
		if item.expectedValue == value {
			t.Logf("ok - %s", value)
		} else {
			t.Logf("not ok - expected '%s' - found '%s' ", item.expectedValue, value)
			t.Fail()
		}
		if item.expectedNegation == negation {
			t.Logf("ok - %s %v", item.input, negation)
		} else {
			t.Logf("not ok - input '%s' - expected '%v' - found '%v' ", item.input, item.expectedNegation, negation)
			t.Fail()
		}
	}
}
