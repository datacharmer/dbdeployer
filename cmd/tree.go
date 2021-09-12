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

//go:build docs
// +build docs

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
)

func writeApi(showHidden bool) {
	fmt.Println("This is the list of commands and modifiers available for")
	fmt.Println("dbdeployer {{.Version}} as of {{.Date}}")
	fmt.Println("")
	fmt.Println("# main")
	fmt.Println("{{dbdeployer -h }}")
	fmt.Println("")
	fmt.Println("{{dbdeployer-docs tree }}")
	traverse(rootCmd, "", 0, true, showHidden)
}

func writeBashCompletion() {
	completionFile := "dbdeployer_completion.sh"
	rootCmd.GenBashCompletionFile(completionFile)
	common.CondPrintf("Copy %s to the completion directory (/etc/bash_completion.d or /usr/local/etc/bash_completion.d)\n", completionFile)

}

func writeManPages() {
	manDir := "man_pages"
	if common.DirExists(manDir) {
		common.Exitf(1, globals.ErrNamedDirectoryAlreadyExists, "manual pages", manDir)
	}
	common.Mkdir(manDir)
	header := &doc.GenManHeader{
		Title:   "dbdeployer",
		Section: "1",
	}
	err := doc.GenManTree(rootCmd, header, manDir)
	common.ErrCheckExitf(err, 1, "%s", err)
	common.CondPrintf("Man pages generated in '%s'\n", manDir)
}

func writeMarkdownPages() {
	mdDir := "markdown_pages"
	if common.DirExists(mdDir) {
		common.Exitf(1, globals.ErrNamedDirectoryAlreadyExists, "Markdown pages", mdDir)
	}
	common.Mkdir(mdDir)
	err := doc.GenMarkdownTree(rootCmd, mdDir)
	common.ErrCheckExitf(err, 1, "%s", err)
	common.CondPrintf("Markdown pages generated in '%s'\n", mdDir)
}

func writeRstPages() {
	rstDir := "rst_pages"
	if common.DirExists(rstDir) {
		common.Exitf(1, globals.ErrNamedDirectoryAlreadyExists, "restructured Text pages", rstDir)
	}
	common.Mkdir(rstDir)
	err := doc.GenReSTTree(rootCmd, rstDir)
	common.ErrCheckExitf(err, 1, "%s", err)
	common.CondPrintf("Restructured Text pages generated in '%s'\n", rstDir)
}

func makeDocumentation(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	api, _ := flags.GetBool("api")
	showHidden, _ := flags.GetBool("show-hidden")
	bashCompletion, _ := flags.GetBool("bash-completion")
	manPages, _ := flags.GetBool("man-pages")
	mdPages, _ := flags.GetBool("markdown-pages")
	rstPages, _ := flags.GetBool("rst-pages")
	if (manPages && api) || (api && bashCompletion) || (api && mdPages) || (api && rstPages) {
		common.Exit(1, "choose one option only")
	}
	if rstPages {
		writeRstPages()
		return
	}
	if manPages {
		writeManPages()
		return
	}
	if mdPages {
		writeMarkdownPages()
		return
	}
	if bashCompletion {
		writeBashCompletion()
		return
	}
	if api {
		writeApi(showHidden)
		return
	}
	traverse(rootCmd, "", 0, api, showHidden)
}

func traverse(cmd *cobra.Command, parent string, level int, api, showHidden bool) {
	subcommands := cmd.Commands()
	indent := strings.Repeat(" ", level*4) + "-"
	for _, c := range subcommands {
		hidden_flag := ""
		if c.Hidden || c.Name() == "help" {
			if showHidden {
				hidden_flag = " (HIDDEN) "
			} else {
				continue
			}
		}
		size := len(c.Commands())
		if api {
			if size > 0 || level == 0 {
				fmt.Printf("\n##%s%s\n", parent+" "+c.Name(), hidden_flag)
			}
			fmt.Printf("{{dbdeployer%s %s -h}}\n", parent, c.Name())
		} else {
			fmt.Printf("%s %-20s%s\n", indent, c.Name(), hidden_flag)
		}
		if size > 0 {
			traverse(c, parent+" "+c.Name(), level+1, api, showHidden)
		}
	}
}

var treeCmd = &cobra.Command{
	Use:     "tree",
	Short:   "shows command tree and other docs",
	Aliases: []string{"docs"},
	Long: `This command is only used to create API documentation. 
You can, however, use it to show the command structure at a glance.`,
	Hidden: true,
	Run:    makeDocumentation,
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
