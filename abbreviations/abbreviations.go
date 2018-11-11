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

package abbreviations

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"os"
	"regexp"
	"strings"
)

/*
	This package implements custom abbreviations.
	It looks for a file "abbreviations.txt" and treats every line
	as an abbreviation followed by its replacement.
	Then, it looks at the command line arguments.
	If an argument matches an abbreviation, it will be replaced by the replacement items.
	For example, the file contains this line:
		sbs sandboxes

	when the user types "dbdeployer sbs", it will be replaced with "dbdeployer sandboxes"

	A more interesting example:
		groupr  replication --topology=group

	Here, a command "dbdeployer groupr 8.0.4" becomes "dbdeployer deploy replication --topology=group 8.0.4"

	It is also possible to set variables in the replacement.
	    sbdef --sandbox-directory={{.sb}} --port={{.port}}

	To use this abbreviation, we need to provide the values for 'sb' and 'port'
	dbdeployer deploy sbdef:port=9000,sb=mysandbox single 8.0.4
	it will become  "dbdeployer deploy --sandbox-directory=mysandbox --port=9000 single 8.0.4
*/

type argList []string
type aliasLists map[string]argList

var debugAbbr bool = false
var abbrevFile string = "abbreviations.txt"
var userDefinedFile string = os.Getenv("DBDEPLOYER_ABBR_FILE")

func showArgs(args argList) {
	for N, arg := range args {
		if debugAbbr {
			fmt.Printf("%d <<%s>>\n", N, arg)
		}
	}
}

func debugPrint(descr string, v interface{}) {
	if debugAbbr {
		fmt.Printf("%s : %v\n", descr, v)
	}
}

func LoadAbbreviations() {
	if common.IsEnvSet("SKIP_ABBR") {
		fmt.Printf("# Abbreviations suppressed by env variable SKIP_ABBR\n")
		return
	}
	var newArgs []string
	var abbreviations = make(aliasLists)
	var variables = make(common.StringMap)
	var verboseAbbr bool = true
	var replacementsUsed bool = false
	if common.IsEnvSet("SILENT_ABBR") {
		verboseAbbr = false
	}
	if userDefinedFile != "" {
		abbrevFile = userDefinedFile
	}
	if !common.FileExists(abbrevFile) {
		if debugAbbr {
			fmt.Printf("# File %s not found\n", abbrevFile)
		}
		return
	}
	abbrLines := common.SlurpAsLines(abbrevFile)
	// Loads abbreviations from file
	for _, abbreviation := range abbrLines {
		abbreviation = strings.TrimSpace(abbreviation)
		list := strings.Split(abbreviation, " ")
		abbr := list[0]
		isComment, _ := regexp.MatchString(`^\s*#`, abbr)
		isEmpty, _ := regexp.MatchString(`^\s*#\$`, abbr)
		if isComment || isEmpty {
			continue
		}
		var newList argList
		for N, repl := range list {
			// Skips the first item, which is the abbreviation
			if N > 0 {
				newList = append(newList, repl)
			}
		}
		abbreviations[abbr] = newList
	}
	// Loop through original arguments.
	// Replaces every occurrence of the abbreviation with its components
	debugPrint("os.Args", os.Args)
	showArgs(os.Args)
	for _, arg := range os.Args {
		// An abbreviation may set variables
		// for example
		// myabbr:varname=var_value
		// myabbr:varname=var_value,other_var=other_value
		re := regexp.MustCompile(`(\w+)[-:](\S+)`)
		reFlag := regexp.MustCompile(`^-`)
		vars := re.FindStringSubmatch(arg)
		if reFlag.MatchString(arg) {
			newArgs = append(newArgs, arg)
			continue
		}
		if len(vars) > 0 {
			arg = vars[1]
			allVars := vars[2]

			// Keys and values are separated by an equals (=) sign
			re = regexp.MustCompile(`(\w+)=(\w+)`)
			uvars := re.FindAllStringSubmatch(allVars, -1)
			for _, vgroup := range uvars {
				variables[string(vgroup[1])] = string(vgroup[2])
			}
		}
		// If exists an abbreviation for the current argument
		if abbreviations[arg] != nil {
			replacement := ""
			for _, item := range abbreviations[arg] {
				if item != "" {
					// Replaces possible vars with their value
					item = common.TemplateFill(item, variables)
					// adds the replacement items to the new argument list
					replacement += " " + item
					newArgs = append(newArgs, item)
					replacementsUsed = true
				}
			}
			if verboseAbbr {
				fmt.Printf("# %s => %s\n", arg, replacement)
			}
		} else {
			// If there is no abbreviation for the current argument
			// it is added as it is.
			newArgs = append(newArgs, arg)
		}
	}
	debugPrint("new_args", newArgs)
	// Arguments replaced!
	if replacementsUsed {
		if debugAbbr {
			fmt.Printf("# Using file %s\n", abbrevFile)
		}
		os.Args = newArgs
		if verboseAbbr {
			fmt.Printf("# %s\n", os.Args)
		}
	}
	for _, arg := range os.Args {
		common.CommandLineArgs = append(common.CommandLineArgs, arg)
	}
}

func init() {
	if common.IsEnvSet("DEBUG_ABBR") {
		debugAbbr = true
	}
}
