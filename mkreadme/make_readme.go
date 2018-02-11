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
	./docs/make_readme < docs/readme_template.md > README.md

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
