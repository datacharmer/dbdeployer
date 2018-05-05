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

// +build docs

package cmd

import (
	"fmt"
	"strings"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/datacharmer/dbdeployer/common"
)

func WriteApi(show_hidden bool) {
	fmt.Println("This is the list of commands and modifiers available for")
	fmt.Println("dbdeployer {{.Version}} as of {{.Date}}")
	fmt.Println("")
	fmt.Println("# main")
	fmt.Println("{{dbdeployer -h }}")
	fmt.Println("")
	fmt.Println("{{dbdeployer-docs tree }}")
	traverse(rootCmd, "", 0, true, show_hidden)
}

func WriteBashCompletion() {
	completion_file := "dbdeployer_completion.sh"
	rootCmd.GenBashCompletionFile(completion_file)
	fmt.Printf("Copy %s to the completion directory (/etc/bash_completion.d or /usr/local/etc/bash_completion.d)\n",completion_file)

}

func WriteManPages() {
	man_dir := "man_pages"
	if common.DirExists(man_dir) {
		common.Exit(1, fmt.Sprintf("manual pages directory '%s' exists already.",man_dir))
	}
	common.Mkdir(man_dir)
	header := &doc.GenManHeader{
		Title: "dbdeployer",
		Section: "1",
	}
	err := doc.GenManTree(rootCmd, header, man_dir)
	if err != nil {
		common.Exit(1, fmt.Sprintf("%s", err))
	}
	fmt.Printf("Man pages generated in '%s'\n", man_dir)
}

func WriteMarkdownPages() {
	md_dir := "markdown_pages"
	if common.DirExists(md_dir) {
		common.Exit(1, fmt.Sprintf("Markdown pages directory '%s' exists already.",md_dir))
	}
	common.Mkdir(md_dir)
	err := doc.GenMarkdownTree(rootCmd, md_dir)
	if err != nil {
		common.Exit(1, fmt.Sprintf("%s", err))
	}
	err = doc.GenReSTTree(rootCmd, md_dir)
	fmt.Printf("Markdown pages generated in '%s'\n", md_dir)
}

func WriteRstPages() {
	rst_dir := "rst_pages"
	if common.DirExists(rst_dir) {
		common.Exit(1, fmt.Sprintf("Restructured Text pages directory '%s' exists already.",rst_dir))
	}
	common.Mkdir(rst_dir)
	err := doc.GenReSTTree(rootCmd, rst_dir)
	if err != nil {
		common.Exit(1, fmt.Sprintf("%s", err))
	}
	fmt.Printf("Restructured Text pages generated in '%s'\n", rst_dir)
}

func MakeDocumentation(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	api, _  := flags.GetBool("api")
	show_hidden, _  := flags.GetBool("show-hidden")
	bash_completion, _  := flags.GetBool("bash-completion")
	man_pages, _  := flags.GetBool("man-pages")
	md_pages, _  := flags.GetBool("markdown-pages")
	rst_pages, _  := flags.GetBool("rst-pages")
	if (man_pages && api) || (api && bash_completion) || (api && md_pages) || (api && rst_pages) {
		common.Exit(1, "Choose one option only")
	}
	if rst_pages {
		WriteRstPages()
		return
	}
	if man_pages {
		WriteManPages()
		return
	}
	if md_pages {
		WriteMarkdownPages()
		return
	}
	if bash_completion {
		WriteBashCompletion()
		return
	}
	if api {
		WriteApi(show_hidden)
		return
	}
	traverse(rootCmd, "", 0, api, show_hidden)
}

func traverse(cmd *cobra.Command, parent string, level int, api, show_hidden bool) {
	subcommands := cmd.Commands()
	indent := strings.Repeat(" ", level*4) + "-"
	for _, c := range subcommands {
		hidden_flag := ""
		if c.Hidden || c.Name() == "help" {
			if show_hidden {
				hidden_flag = " (HIDDEN) "
			} else {
				continue
			}
		}
		size := len(c.Commands())
		if api {
			if size > 0  || level == 0 {
				fmt.Printf("\n##%s%s\n", parent + " " + c.Name(), hidden_flag)
			}
			fmt.Printf("{{dbdeployer%s %s -h}}\n", parent, c.Name())
		} else {
			fmt.Printf("%s %-20s%s\n", indent, c.Name(), hidden_flag)
		}
		if size > 0 {
			traverse(c, parent + " " + c.Name(), level + 1, api, show_hidden)
		}
	}
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "shows command tree and other docs",
	Aliases: []string{"docs"},
	Long: `This command is only used to create API documentation. 
You can, however, use it to show the command structure at a glance.`,
	Hidden : true,
	Run: MakeDocumentation,
}

func init() {
	rootCmd.AddCommand(treeCmd)
	treeCmd.PersistentFlags().Bool("man-pages", false, "Writes man pages")
	treeCmd.PersistentFlags().Bool("markdown-pages", false, "Writes Markdown docs")
	treeCmd.PersistentFlags().Bool("rst-pages", false, "Writes Restructured Text docs")
	treeCmd.PersistentFlags().Bool("api", false, "Writes API template")
	treeCmd.PersistentFlags().Bool("bash-completion", false, "creates bash-completion file")
	treeCmd.PersistentFlags().Bool("show-hidden", false, "Shows also hidden commands")
}
