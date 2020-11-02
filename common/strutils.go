// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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

	"github.com/datacharmer/dbdeployer/globals"
)

type CleanupFunc func(target string)
type CleanupRec struct {
	target string
	label  string
	f      CleanupFunc
}

var cleanupActions = NewStack()

func CondPrintf(format string, args ...interface{}) {
	if globals.UsingDbDeployer {
		fmt.Printf(format, args...)
	}
}

func CondPrintln(args ...interface{}) {
	if globals.UsingDbDeployer {
		fmt.Println(args...)
	}
}

// Checks whether a given environment variable is set
func IsEnvSet(envVar string) bool {
	return os.Getenv(envVar) != ""
}

// Given a path starting at the HOME directory
// returns a string where the literal value for $HOME
// is replaced by the string "$HOME"
func ReplaceLiteralHome(path string) string {
	return ReplaceLiteralEnvVar(path, "HOME")
}

// Replaces the literal value of an environment variable with its name
// for example, if "$HOME" resolves to "/home/goofy" the string "/home/goofy/some/path" would become "$HOME/some/path"
func ReplaceLiteralEnvVar(name string, envVar string) string {
	value := os.Getenv(envVar)
	// If the environment variable is empty, we return the initial name
	if value == "" {
		return name
	}
	// If the current location is at the top of the directory tree, we don't want to do any replacement
	if value == "/" {
		return name
	}
	// If there is already a variable in the name, no further replacement is needed
	if strings.Contains(name, "$") {
		return name
	}
	re := regexp.MustCompile(`^` + value)
	return re.ReplaceAllString(name, "$$"+envVar)
}

// Replaces the environment variable `envVar` with its value
// for example, if "$HOME" resolves to "/home/goofy" the string "$HOME/some/path" would become "/home/goofy/some/path"
func ReplaceEnvVar(name string, envVar string) string {
	value := os.Getenv(envVar)
	if value == "" || value == "/" {
		return name
	}
	re := regexp.MustCompile(`\$` + envVar + `\b`)
	return re.ReplaceAllString(name, value)
}

// Given a path with the variable "$HOME" at the start,
// returns a string with the value of HOME expanded
func ReplaceHomeVar(path string) string {
	return ReplaceEnvVar(path, "HOME")
}

// Creates a "human readable" and predictable UUID using some
// pre-defined elements.
// Used to replace the random UUID created by MySQL 5.6+,
// with the purpose of returning easy to read server identifiers that
// can be processed visually.
func MakeCustomizedUuid(port, nodeNum int) (string, error) {
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
		return "", fmt.Errorf("node num out of boundaries: %d", nodeNum)
	}
	return fmt.Sprintf("%s-%s-%s", group1, group2, group3), nil
}

// Return true is `contained` is a sub-string of `mainString`
func Includes(mainString, contained string) bool {
	re := regexp.MustCompile(contained)
	return re.MatchString(mainString)
}

// IsEmptyOrBlank returns true if the given string is empty
// or contains only spaces
// It also returns true for a string that contains spaces AND A NEWLINE
// empty:              ""
// empty with newline: "\n"
// only spaces:        "  "
// spaces and tabs:    "  \t"
// spaces and newline: " \n"
func IsEmptyOrBlank(s string) bool {
	if s == "" {
		return true
	}
	reEmpty := regexp.MustCompile(`^\s*$`)
	return reEmpty.MatchString(s)
}

// CoalesceString returns the first string that it is not empty
// or an empty string if all items are empty
func CoalesceString(items ...string) string {
	for _, item := range items {
		if !IsEmptyOrBlank(item) {
			return item
		}
	}
	return items[0]
}

// Coalesce returns the first object that is not empty
func Coalesce(items ...interface{}) interface{} {
	for _, item := range items {
		if item != nil {
			return item
		}
	}
	return nil
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
	reNonNumeric := regexp.MustCompile(`^\D+`)
	wanted = reNonNumeric.ReplaceAllString(wanted, "")
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
		vl, err := VersionToList(line)
		if err != nil {
			CondPrintf("%s\n", err)
			return versions
		}
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

// Returns true if the input value is either of "true", "yes", "1"
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
// Example: an input of "1,2,3" returns []int{1, 2, 3}
func StringToIntSlice(val string) (numberList []int, err error) {
	list := strings.Split(val, ",")
	for _, item := range list {
		num, err := strconv.Atoi(strings.TrimSpace(item))
		if err != nil {
			return []int{}, fmt.Errorf("value '%s' (part of %s) is not a number", item, val)
		}
		numberList = append(numberList, num)
	}
	return numberList, nil
}

// Given an array of integers, returns a string containing the numbers
// separated by a given string.
// For example: an input of []int{1, 2, 3}, "#" returns "1#2#3"
func IntSliceToSeparatedString(val []int, separator string) string {
	result := ""
	for _, i := range val {
		if result != "" {
			result += separator
		}
		result += fmt.Sprintf("%d", i)
	}
	return result
}

// Given an array of integers, returns a string containing the numbers
// separated by a dot.
// For example: an input of []int{1, 2, 3} returns "1.2.3"
func IntSliceToDottedString(val []int) string {
	return IntSliceToSeparatedString(val, ".")
}

// Removes a slash (if any) at the end of a given string
func RemoveTrailingSlash(s string) string {
	re := regexp.MustCompile(`/$`)
	return re.ReplaceAllString(s, "")
}

func RemoveSuffix(s, suffix string) string {
	re := regexp.MustCompile(suffix + `$`)
	return re.ReplaceAllString(s, "")
}

// ------------------------------------------------------------------------------------
// The functions below this point are intended only for use with a command line client,
// and may not be suitable for other client types
// ------------------------------------------------------------------------------------

// Converts a string to an integer. Exits on error
func Atoi(val string) int {
	num, err := strconv.Atoi(val)
	ErrCheckExitf(err, 1, fmt.Sprintf("Not a valid number: %s (%s)", val, err))
	return num
}

// Adds an action to the list of clean-up operations
// to run before aborting the program
func AddToCleanupStack(cf CleanupFunc, funcName, arg string) {
	cleanupActions.Push(CleanupRec{f: cf, label: funcName, target: arg})
}

// Runs the cleanup actions (usually before Exit)
func RunCleanupActions() {
	if cleanupActions.Len() > 0 {
		CondPrintf("# Pre-exit cleanup. \n")
	}
	count := 0
	for cleanupActions.Len() > 0 {
		count++
		cr := cleanupActions.Pop().(CleanupRec)
		CondPrintf("#%d - Executing %s( %s)\n", count, cr.label, cr.target)
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
	CondPrintf(format, args...)
	if !Includes(format, `\n`) {
		CondPrintln()
	}
	os.Exit(exitCode)
}

// Exit with custom set of messages
// Runs cleanup actions before aborting the program
func Exit(exitCode int, messages ...string) {
	RunCleanupActions()
	for _, msg := range messages {
		CondPrintf("%s\n", msg)
	}
	os.Exit(exitCode)
}

// Returns the latest version among the ones found in a Sandbox binary directory
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

func OptionComponents(s string) (value string, negation bool) {
	reNegation := regexp.MustCompile(`^(?:!|no-|not-)(\S+)`)
	valueList := reNegation.FindAllStringSubmatch(s, -1)
	if len(valueList) == 0 || len(valueList[0]) == 0 {
		return s, false
	}
	return valueList[0][1], true
}

func OptionCompare(option, value string) bool {
	optionValue, negation := OptionComponents(option)
	matches := optionValue == value
	if negation {
		matches = !matches
	}
	return matches
}
