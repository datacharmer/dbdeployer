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

package main

import (
	"bufio"
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func getCmdOutput(cmdText string) string {
	cmdList := strings.Split(cmdText, " ")
	command := cmdList[0]
	var args []string
	for n, arg := range cmdList {
		if n > 0 {
			args = append(args, arg)
		}
	}
	// #nosec G204
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err = cmd.Start(); err != nil {
		common.Exitf(1, "# ERROR: %s", err)
	}
	slurp, _ := ioutil.ReadAll(stdout)
	_ = stdout.Close()
	return fmt.Sprintf("%s", slurp)
}

/*
	Reads the README template and replaces the commands indicated
	as {{command argument}} with their output.
	This allows us to produce a file README.md with the output
	from the current release.

	Use as:
	./mkreadme/make_readme < ./mkreadme/readme_template.md > README.md

*/
func main() {
	// Gets input from stdin
	scanner := bufio.NewScanner(os.Stdin)

	reVersion := regexp.MustCompile(`{{\.Version}}`)
	reDate := regexp.MustCompile(`{{\.Date}}`)
	reCmd := regexp.MustCompile(`{{([^}]+)}}`)
	reFlag := regexp.MustCompile(`(?sm)Global Flags:.*`)
	reSpaces := regexp.MustCompile(`(?m)^`)
	home := os.Getenv("HOME")
	reHome := regexp.MustCompile(home)
	timeFormat := "02-Jan-2006 15:04 MST"
	timestamp := time.Now().UTC().Format(timeFormat)
	// An user defined timestamp could be used instead of the generated one.
	if os.Getenv("DBDEPLOYER_TIMESTAMP") != "" {
		timestamp = os.Getenv("DBDEPLOYER_TIMESTAMP")
	}
	for scanner.Scan() {
		line := scanner.Text()
		// Replacement for version and date must occur BEFORE
		// we search for commands, as the regexp for commands would
		// match version and date as well.
		line = reVersion.ReplaceAllString(line, common.VersionDef)
		line = reDate.ReplaceAllString(line, timestamp)
		// Find a placeholder for a {{command}}
		findList := reCmd.FindAllStringSubmatch(line, -1)
		if findList != nil {
			commandText := findList[0][1]
			// Run the command and gets its output
			out := getCmdOutput(commandText)
			// remove global flags from the output
			out = reFlag.ReplaceAllString(out, "")
			// Add 4 spaces to every line of the output
			out = reSpaces.ReplaceAllString(out, "    ")
			// Replace the literal $HOME value with its variable name
			out = reHome.ReplaceAllString(out, `$$HOME`)

			fmt.Printf("    $ %s\n", commandText)
			fmt.Printf("%s\n", out)
		} else {
			fmt.Printf("%s\n", line)
		}
	}

	if err := scanner.Err(); err != nil {
		common.Exitf(1, "# ERROR: %s", err)
	}
}
