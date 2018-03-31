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

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func get_cmd_output(cmdText string) string {
	cmdList := strings.Split(cmdText, " ")
	command := cmdList[0]
	var args []string
	for n, arg := range cmdList {
		if n > 0 {
			args = append(args, arg)
		}
	}
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err = cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	slurp, _ := ioutil.ReadAll(stdout)
	stdout.Close()
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

	re_cmd := regexp.MustCompile(`{{([^}]+)}}`)
	re_flag := regexp.MustCompile(`(?sm)Global Flags:.*`)
	re_spaces := regexp.MustCompile(`(?m)^`)
	home := os.Getenv("HOME")
	re_home := regexp.MustCompile(home)
	for scanner.Scan() {
		line := scanner.Text()
		// Find a placeholder for a {{command}}
		findList := re_cmd.FindAllStringSubmatch(line, -1)
		if findList != nil {
			commandText := findList[0][1]
			// Run the command and gets its output
			out := get_cmd_output(commandText)
			// remove global flags from the output
			out = re_flag.ReplaceAllString(out, "")
			// Add 4 spaces to every line of the output
			out = re_spaces.ReplaceAllString(out, "    ")
			// Replace the literal $HOME value with its variable name
			out = re_home.ReplaceAllString(out, `$$HOME`)

			fmt.Printf("    $ %s\n", commandText)
			fmt.Printf("%s\n", out)
		} else {
			fmt.Printf("%s\n", line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
