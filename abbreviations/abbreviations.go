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
		group  replication --topology=group

	Here, a command "dbdeployer group 8.0.4" becomes "dbdeployer replication --topology=group 8.0.4"

	It is also possible to set variables in the replacement.
	    sbdef --sandbox-directory={{.sb}} --port={{.port}}

	To use this abbreviation, we need to provide the values for 'sb' and 'port'
	dbdeployer sbdef:port=9000,sb=mysandbox single 8.0.4
	it will become  "dbdeployer --sandbox-directory=mysandbox --port=9000 single 8.0.4
*/

type argList []string
type AliasList map[string]argList

func debug_print(descr string, v interface{}) {
	if os.Getenv("DEBUG") != "" {
		fmt.Printf("%s : %v\n", descr, v)
	}
}

func LoadAbbreviations() {
	if os.Getenv("SKIP_ABBR") != "" {
		fmt.Printf("# Abbreviations suppressed by env variable SKIP_ABBR\n")
		return
	}
	var abbrev_file string = "abbreviations.txt"
	var new_args []string
	var abbreviations = make(AliasList)
	var variables = make(common.Smap)
	var verbose_abbr bool = true
	var debug_abbr bool = false
	var replacements_used bool = false
	if os.Getenv("SILENT_ABBR") != "" {
		verbose_abbr = false
	}
	if os.Getenv("DEBUG_ABBR") != "" {
		debug_abbr = true
	}
	user_defined_file := os.Getenv("DBDEPLOYER_ABBR_FILE")
	if user_defined_file != "" {
		abbrev_file = user_defined_file
	}
	if !common.FileExists(abbrev_file) {
		if debug_abbr {
			fmt.Printf("# File %s not found\n", abbrev_file)
		}
		return
	}
	abbr_lines := common.SlurpAsLines(abbrev_file)
	// Loads abbreviations from file
	for _, abbreviation := range abbr_lines {
		abbreviation = strings.TrimSpace(abbreviation)
		list := strings.Split(abbreviation, " ")
		abbr := list[0]
		is_comment, _ := regexp.MatchString(`^\s*#`, abbr)
		is_empty, _ := regexp.MatchString(`^\s*#\$`, abbr)
		if is_comment || is_empty {
			continue
		}
		var new_list argList
		for N, repl := range list {
			// Skips the first item, which is the abbreviation
			if N > 0 {
				new_list = append(new_list, repl)
			}
		}
		abbreviations[abbr] = new_list
	}
	// Loop through original arguments.
	// Replaces every occurrence of the abbreviation with its components
	debug_print("os.Args", os.Args)
	for _, arg := range os.Args {
		// An abbreviation may set variables
		// for example
		// myabbr:varname=var_value
		// myabbr:varname=var_value,other_var=other_value
		re := regexp.MustCompile(`(\w+)[-:](\S+)`)
		re_flag := regexp.MustCompile(`^-`)
		vars := re.FindStringSubmatch(arg)
		if re_flag.MatchString(arg) {
			new_args = append(new_args, arg)
			continue
		}
		if len(vars) > 0 {
			arg = vars[1]
			all_vars := vars[2]

			// Keys and values are separated by an equals (=) sign
			re = regexp.MustCompile(`(\w+)=(\w+)`)
			uvars := re.FindAllStringSubmatch(all_vars, -1)
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
					item = common.Tprintf(item, variables)
					// adds the replacement items to the new argument list
					replacement += " " + item
					new_args = append(new_args, item)
					replacements_used = true
				}
			}
			if verbose_abbr {
				fmt.Printf("# %s => %s\n", arg, replacement)
			}
		} else {
			// If there is no abbreviation for the current argument
			// it is added as it is.
			new_args = append(new_args, arg)
		}
	}
	debug_print("new_args", new_args)
	// Arguments replaced!
	if replacements_used {
		if debug_abbr {
			fmt.Printf("# Using file %s\n", abbrev_file)
		}
		os.Args = new_args
		if verbose_abbr {
			fmt.Printf("# %s\n", os.Args)
		}
	}
}
