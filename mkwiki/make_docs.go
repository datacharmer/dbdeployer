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
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/datacharmer/dbdeployer/common"
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
	Reads the documentation template and replaces the commands indicated
	as {{command argument}} with their output.
	This allows us to produce updated wiki pages with the output
	from the current release.

	Use as:
	./mkreadme/make_docs < ./mkreadme/wiki_template.markdown
	mv -f *.md /path/to/dbdeployer.wiki/

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

	fullDocName := "full_doc.markdown"
	homeFile, err := os.Create("Home.md")
	if err != nil {
		fmt.Printf("ERROR opening file Home.md: %s\n", err)
		os.Exit(1)
	}
	fullDoc, err := os.Create(fullDocName)
	if err != nil {
		fmt.Printf("ERROR opening file %s: %s\n", fullDocName, err)
		os.Exit(1)
	}
	defer homeFile.Close()
	defer fullDoc.Close()
	header1 := ""
	header1Name := ""
	wikiBaseUrl := "https://github.com/datacharmer/dbdeployer/wiki"

	fmt.Fprintf(homeFile, "# dbdeployer\n")
	fmt.Fprintf(fullDoc, "# dbdeployer\n")
	reHeader1 := regexp.MustCompile(`^#\s+`)
	reHeader2 := regexp.MustCompile(`^##\s+`)
	reCleanHeader := regexp.MustCompile(`^#+\s+`)
	reCodeBoundary := regexp.MustCompile("^\\s*```")
	insideCode := false
	addHomeLink := false

	var header1File *os.File
	var filesList = []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if reCodeBoundary.MatchString(line) {
			if insideCode {
				insideCode = false
			} else {
				insideCode = true
			}
		}
		if !insideCode && reHeader1.MatchString(line) {
			header1Name = reCleanHeader.ReplaceAllString(line, "")
			header1 = header1Name
			header1Name = strings.ReplaceAll(header1Name, " ", "-")
			header1Name = strings.ToLower(header1Name)
			if header1File != nil {
				header1File.Close()
			}
			header1File, err = os.Create(header1Name + ".md")
			if err != nil {
				fmt.Printf("error opening %s : %s\n", header1Name, err)
				os.Exit(1)
			}
			filesList = append(filesList, header1Name + ".md")
			fmt.Fprintf(homeFile, "- [%s](%s/%s)\n", header1, wikiBaseUrl, header1Name)
			fmt.Fprintf(fullDoc, "- [%s](#%s)\n", header1, header1Name)
			addHomeLink = true
		}
		if !insideCode && reHeader2.MatchString(line) {
			header2Name := reCleanHeader.ReplaceAllString(line, "")
			header2 := header2Name
			header2Name = strings.ReplaceAll(header2Name, " ", "-")
			header2Name = strings.ToLower(header2Name)
			fmt.Fprintf(homeFile, "    - [%s](%s/%s#%s)\n", header2, wikiBaseUrl, header1Name, header2Name)
			fmt.Fprintf(fullDoc, "    - [%s](#%s)\n", header2,  header2Name)
		}

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

			fmt.Fprintf(header1File, "    $ %s\n", commandText)
			fmt.Fprintf(header1File, "%s\n", out)
		} else {
			if header1File == nil {
				fmt.Fprintf(homeFile, "%s\n", line)
				fmt.Fprintf(fullDoc, "%s\n", line)
			} else {
				fmt.Fprintf(header1File, "%s\n", line)
				if addHomeLink {
					fmt.Fprintf(header1File, "[[HOME](https://github.com/datacharmer/dbdeployer/wiki)]\n")
					addHomeLink = false
				}
			}
		}

	}
	if header1File != nil {
		header1File.Close()
	}

	if err = scanner.Err(); err != nil {
		common.Exitf(1, "# ERROR: %s", err)
	}
	fullDoc.Close()
	for _, fileName := range filesList {
		text, err := common.SlurpAsLines(fileName)
		if err != nil {
			fmt.Printf("error reading file %s: %s\n", fileName, err)
			os.Exit(1)
		}
	    var cleanText []string
		for _, line := range text {
			if !strings.Contains(line, "[HOME]")	{
				cleanText = append(cleanText, line)
			}
		}
		err = common.AppendStrings(cleanText, fullDocName, "\n")
		if err != nil {
			fmt.Printf("error writing to %s: %s\n", fullDocName, err)
			os.Exit(1)
		}
	}
}
