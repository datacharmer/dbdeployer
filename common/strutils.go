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
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type CleanupFunc func(target string)
type CleanupRec struct {
	target string
	label  string
	f      CleanupFunc
}

type CleanupStack []CleanupRec

// var cleanup_actions CleanupStack
var cleanup_actions = new(Stack)

// Given a path starting at the HOME directory
// returns a string where the literal value for $HOME
// is replaced by the string "$HOME"
func ReplaceLiteralHome(path string) string {
	// home := os.Getenv("HOME")
	// re := regexp.MustCompile(`^` + home)
	// return re.ReplaceAllString(path, "$$HOME")
	return ReplaceLiteralEnvVar(path, "HOME")
}

func ReplaceLiteralEnvVar(name string, env_var string) string {
	value := os.Getenv(env_var)
	re := regexp.MustCompile(value)
	return re.ReplaceAllString(name, "$$"+env_var)
}

func ReplaceEnvVar(name string, env_var string) string {
	value := os.Getenv(env_var)
	re := regexp.MustCompile(`\$` + env_var + `\b`)
	return re.ReplaceAllString(name, value)
}

// Given a path with the variable "$HOME" at the start,
// returns a string with the value of HOME expanded
func ReplaceHomeVar(path string) string {
	// home := os.Getenv("HOME")
	// re := regexp.MustCompile(`^\$HOME\b`)
	//return re.ReplaceAllString(path, home)
	return ReplaceEnvVar(path, "HOME")
}

func MakeCustomizedUuid(port, node_num int) string {
	re_digit := regexp.MustCompile(`\d`)
	group1 := fmt.Sprintf("%08d", port)
	group2 := fmt.Sprintf("%04d-%04d-%04d", node_num, node_num, node_num)
	group3 := fmt.Sprintf("%012d", port)
	//              12345678 1234 1234 1234 123456789012
	//    new_uuid="00000000-0000-0000-0000-000000000000"
	switch {
	case node_num > 0 && node_num <= 9:
		group2 = re_digit.ReplaceAllString(group2, fmt.Sprintf("%d", node_num))
		group3 = re_digit.ReplaceAllString(group3, fmt.Sprintf("%d", node_num))
	// Number greater than 10 make little sense for this purpose.
	// But we keep the rule so that a valid UUID will be formatted in any case.
	case node_num >= 10000 && node_num <= 99999:
		group2 = fmt.Sprintf("%04d-%04d-%04d", 0, int(node_num/10000), node_num-10000*int(node_num/10000))
	case node_num >= 100000:
		group2 = fmt.Sprintf("%04d-%04d-%04d", int(node_num/10000), 0, 0)
	case node_num >= 1000000:
		fmt.Printf("Node num out of boundaries: %d\n", node_num)
		os.Exit(1)
	}
	return fmt.Sprintf("%s-%s-%s", group1, group2, group3)
}

func Includes(main_string, contained string) bool {
	re := regexp.MustCompile(contained)
	return re.MatchString(main_string)

}

// Given a list of version strings (in the format x.x.x)
// this function returns an ordered list, taking into account the
// components of the versions, so that 5.6.2 sorts lower than 5.6.11
// while a text sort would put 5.6.11 before 5.6.2
func SortVersions(versions []string) (sorted []string) {
	type version_list struct {
		text        string
		maj_min_rev []int
	}
	var vlist []version_list
	for _, line := range versions {
		vl := VersionToList(line)
		rec := version_list{
			text:        line,
			maj_min_rev: vl,
		}
		if vl[0] > 0 {
			vlist = append(vlist, rec)
		}
	}
	sort.Slice(vlist, func(a, b int) bool {
		maj_a := vlist[a].maj_min_rev[0]
		min_a := vlist[a].maj_min_rev[1]
		rev_a := vlist[a].maj_min_rev[2]
		maj_b := vlist[b].maj_min_rev[0]
		min_b := vlist[b].maj_min_rev[1]
		rev_b := vlist[b].maj_min_rev[2]
		return maj_a < maj_b ||
			(maj_a == maj_b && min_a < min_b) ||
			(maj_a == maj_b && min_a == min_b && rev_a < rev_b)
	})
	for _, v := range vlist {
		sorted = append(sorted, v.text)
	}
	return
}

func LatestVersion(search_dir, pattern string) string {
	files, err := ioutil.ReadDir(search_dir)
	// fmt.Printf("<%s> <%s>\n",search_dir, pattern)
	ErrCheckExitf(err, 1, "ERROR reading directory %s: %s", search_dir, err)
	var matching_list []string
	valid_pattern := regexp.MustCompile(`\d+\.\d+$`)
	if !valid_pattern.MatchString(pattern) {
		Exit(1, "Invalid pattern. Must be '#.#'")
	}
	re := regexp.MustCompile(fmt.Sprintf(`^%s\.\d+$`, pattern))
	for _, f := range files {
		fmode := f.Mode()
		// fmt.Printf("<%s> %#v\n",f.Name(), fmode.IsDir())
		if fmode.IsDir() && re.MatchString(f.Name()) {
			matching_list = append(matching_list, f.Name())
		}
	}
	// fmt.Printf("%s\n",matching_list)
	sorted_list := SortVersions(matching_list)
	// fmt.Printf("%s\n",sorted_list)
	if len(sorted_list) > 0 {
		latest := sorted_list[len(sorted_list)-1]
		return latest
	}
	return ""
}

func Atoi(val string) int {
	numvalue, err := strconv.Atoi(val)
	ErrCheckExitf(err, 1, fmt.Sprintf("Not a valid number: %s (%s)", val, err))
	return numvalue
}

func TextToBool(value string) (result bool) {
	value = strings.ToLower(value)
	switch value {
	case "yes":
		result = true
	case "true":
		result = true
	case "1":
		result = true
	default:
		result = false
	}
	return
}

func StringToIntSlice(val string) (num_list []int) {
	list := strings.Split(val, ",")
	for _, item := range list {
		num_list = append(num_list, Atoi(strings.TrimSpace(item)))
	}
	return num_list
}

// Adds an action to the list of clean-up operations
// to run before aborting the program
func AddToCleanupStack(cf CleanupFunc, func_name, arg string) {
	cleanup_actions.Push(CleanupRec{f: cf, label: func_name, target: arg})
}

// Runs the cleanup actions (usually before Exit)
func RunCleanupActions() {
	if cleanup_actions.Len() > 0 {
		fmt.Printf("# Pre-exit cleanup. \n")
	}
	count := 0
	for cleanup_actions.Len() > 0 {
		count++
		cr := cleanup_actions.Pop().(CleanupRec)
		fmt.Printf("#%d - Executing %s( %s)\n", count, cr.label, cr.target)
		cr.f(cr.target)
	}
}

// Checks the status of error variable and exit with custom message if it is not nil.
func ErrCheckExitf(err error, exit_code int, format string, args ...interface{}) {
	if err != nil {
		Exitf(exit_code, format, args...)
	}
}

// Exit with formatted message
// Runs cleanup actions before aborting the program
func Exitf(exit_code int, format string, args ...interface{}) {
	RunCleanupActions()
	fmt.Printf(format, args...)
	if !Includes(format, `\n`) {
		fmt.Println()
	}
	os.Exit(exit_code)
}

// Exit with custom set of messages
// Runs cleanup actions before aborting the program
func Exit(exit_code int, messages ...string) {
	RunCleanupActions()
	for _, msg := range messages {
		fmt.Printf("%s\n", msg)
	}
	os.Exit(exit_code)
}

func RemoveTrailingSlash(s string) string {
	re := regexp.MustCompile(`/$`)
	return re.ReplaceAllString(s, "")
}
