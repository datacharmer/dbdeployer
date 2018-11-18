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

var cleanupActions = NewStack()

func IsEnvSet(envVar string) bool {
	if os.Getenv(envVar) != "" {
		return true
	}
	return false
}

// Given a path starting at the HOME directory
// returns a string where the literal value for $HOME
// is replaced by the string "$HOME"
func ReplaceLiteralHome(path string) string {
	return ReplaceLiteralEnvVar(path, "HOME")
}

func ReplaceLiteralEnvVar(name string, envVar string) string {
	value := os.Getenv(envVar)
	re := regexp.MustCompile(value)
	return re.ReplaceAllString(name, "$$"+envVar)
}

func ReplaceEnvVar(name string, envVar string) string {
	value := os.Getenv(envVar)
	re := regexp.MustCompile(`\$` + envVar + `\b`)
	return re.ReplaceAllString(name, value)
}

// Given a path with the variable "$HOME" at the start,
// returns a string with the value of HOME expanded
func ReplaceHomeVar(path string) string {
	return ReplaceEnvVar(path, "HOME")
}

func MakeCustomizedUuid(port, nodeNum int) (error, string) {
	reDigit := regexp.MustCompile(`\d`)
	group1 := fmt.Sprintf("%08d", port)
	group2 := fmt.Sprintf("%04d-%04d-%04d", nodeNum, nodeNum, nodeNum)
	group3 := fmt.Sprintf("%012d", port)
	//              12345678 1234 1234 1234 123456789012
	//    new_uuid="00000000-0000-0000-0000-000000000000"
	switch {
	case nodeNum > 0 && nodeNum <= 9:
		group2 = reDigit.ReplaceAllString(group2, fmt.Sprintf("%d", nodeNum))
		group3 = reDigit.ReplaceAllString(group3, fmt.Sprintf("%d", nodeNum))
	// Number greater than 10 make little sense for this purpose.
	// But we keep the rule so that a valid UUID will be formatted in any case.
	case nodeNum >= 10000 && nodeNum <= 99999:
		group2 = fmt.Sprintf("%04d-%04d-%04d", 0, int(nodeNum/10000), nodeNum-10000*int(nodeNum/10000))
	case nodeNum >= 100000 && nodeNum < 1000000:
		group2 = fmt.Sprintf("%04d-%04d-%04d", int(nodeNum/10000), 0, 0)
	case nodeNum >= 1000000:
		return fmt.Errorf("Node num out of boundaries: %d\n", nodeNum), ""
	}
	return nil, fmt.Sprintf("%s-%s-%s", group1, group2, group3)
}

func Includes(mainString, contained string) bool {
	re := regexp.MustCompile(contained)
	return re.MatchString(mainString)

}

// Given a list of version strings (in the format x.x.x)
// this function returns an ordered list, taking into account the
// components of the versions, so that 5.6.2 sorts lower than 5.6.11
// while a text sort would put 5.6.11 before 5.6.2
// If wanted is not empty, it will be interpreted as a short version to match.
// For example, when wanted is "5.6", only the versions starting with '5.6' will be considered.
func SortVersionsSubset(versions []string, wanted string) (sorted []string) {
	type versionList struct {
		text      string
		majMinRev []int
	}
	wantedMaj := 0
	wantedMin := 0
	if wanted != "" {
		verList := strings.Split(wanted, ".")
		if len(verList) > 0 {
			wantedMaj, _ = strconv.Atoi(verList[0])
		}
		if len(verList) > 1 {
			wantedMin, _ = strconv.Atoi(verList[1])
		}
	}
	var vlist []versionList
	for _, line := range versions {
		vl := VersionToList(line)
		rec := versionList{
			text:      line,
			majMinRev: vl,
		}
		if vl[0] > 0 {
			willAppend := true
			if wantedMaj > 0 {
				if vl[0] == wantedMaj {
					if vl[1] != wantedMin {
						willAppend = false
					}
				} else {
					willAppend = false
				}
			}
			if willAppend {
				vlist = append(vlist, rec)
			}
		}
	}
	sort.Slice(vlist, func(a, b int) bool {
		majA := vlist[a].majMinRev[0]
		minA := vlist[a].majMinRev[1]
		revA := vlist[a].majMinRev[2]
		majB := vlist[b].majMinRev[0]
		minB := vlist[b].majMinRev[1]
		revB := vlist[b].majMinRev[2]
		return majA < majB ||
			(majA == majB && minA < minB) ||
			(majA == majB && minA == minB && revA < revB)
	})
	for _, v := range vlist {
		sorted = append(sorted, v.text)
	}
	return
}

func SortVersions(versions []string) []string {
	return SortVersionsSubset(versions, "")
}

func LatestVersion(searchDir, pattern string) string {
	files, err := ioutil.ReadDir(searchDir)
	ErrCheckExitf(err, 1, "ERROR reading directory %s: %s", searchDir, err)
	var matchingList []string
	validPattern := regexp.MustCompile(`\d+\.\d+$`)
	if !validPattern.MatchString(pattern) {
		Exit(1, "Invalid pattern. Must be '#.#'")
	}
	re := regexp.MustCompile(fmt.Sprintf(`^%s\.\d+$`, pattern))
	for _, f := range files {
		fmode := f.Mode()
		if fmode.IsDir() && re.MatchString(f.Name()) {
			matchingList = append(matchingList, f.Name())
		}
	}
	sortedList := SortVersions(matchingList)
	if len(sortedList) > 0 {
		latest := sortedList[len(sortedList)-1]
		return latest
	}
	return ""
}

func Atoi(val string) int {
	num, err := strconv.Atoi(val)
	ErrCheckExitf(err, 1, fmt.Sprintf("Not a valid number: %s (%s)", val, err))
	return num
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

// Given a string containing comma-separated integers, returns an array of integers
func StringToIntSlice(val string) (numberList []int, err error) {
	list := strings.Split(val, ",")
	for _, item := range list {
		num, err := strconv.Atoi(strings.TrimSpace(item))
		if err != nil {
			return []int{}, fmt.Errorf("Value '%s' (part of %s) is not a number\n", item, val)
		}
		numberList = append(numberList, num)
	}
	return numberList, nil
}

// Given an array of integers, returns a string containing the numbers
// separated by a dot
func IntSliceToDottedString(val []int) string {
	result := ""
	for _, i := range val {
		if result != "" {
			result += "."
		}
		result += fmt.Sprintf("%d", i)
	}
	return result
}

// Adds an action to the list of clean-up operations
// to run before aborting the program
func AddToCleanupStack(cf CleanupFunc, funcName, arg string) {
	cleanupActions.Push(CleanupRec{f: cf, label: funcName, target: arg})
}

// Runs the cleanup actions (usually before Exit)
func RunCleanupActions() {
	if cleanupActions.Len() > 0 {
		fmt.Printf("# Pre-exit cleanup. \n")
	}
	count := 0
	for cleanupActions.Len() > 0 {
		count++
		cr := cleanupActions.Pop().(CleanupRec)
		fmt.Printf("#%d - Executing %s( %s)\n", count, cr.label, cr.target)
		cr.f(cr.target)
	}
}

// Checks the status of error variable and exit with custom message if it is not nil.
func ErrCheckExitf(err error, exitCode int, format string, args ...interface{}) {
	if err != nil {
		Exitf(exitCode, format, args...)
	}
}

// Exit with formatted message
// Runs cleanup actions before aborting the program
func Exitf(exitCode int, format string, args ...interface{}) {
	RunCleanupActions()
	fmt.Printf(format, args...)
	if !Includes(format, `\n`) {
		fmt.Println()
	}
	os.Exit(exitCode)
}

// Exit with custom set of messages
// Runs cleanup actions before aborting the program
func Exit(exitCode int, messages ...string) {
	RunCleanupActions()
	for _, msg := range messages {
		fmt.Printf("%s\n", msg)
	}
	os.Exit(exitCode)
}

func RemoveTrailingSlash(s string) string {
	re := regexp.MustCompile(`/$`)
	return re.ReplaceAllString(s, "")
}
