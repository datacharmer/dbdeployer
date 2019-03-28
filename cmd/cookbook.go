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

package cmd

import (
	"fmt"
	"os"

	"github.com/datacharmer/dbdeployer/cookbook"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/spf13/cobra"
)

func checkArgs(label, syntax string, args []string, howManyWanted int) {
	size := len(args)
	if len(args) < howManyWanted {
		fmt.Printf("%s: Required %d args but provided %d\n", label, howManyWanted, size)
		fmt.Printf("Syntax: %s %s\n", label, syntax)
		os.Exit(1)
	}
}

func listCookbook(cmd *cobra.Command, args []string) {
	cookbook.ListRecipes()
}

func showCookbook(cmd *cobra.Command, args []string) {
	checkArgs("show", "show recipe_name", args, 1)
	raw, _ := cmd.Flags().GetBool(globals.RawLabel)
	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	cookbook.ShowRecipe(args[0], flavor, raw)
}

func createCookbook(cmd *cobra.Command, args []string) {
	checkArgs("run", "run recipe_name", args, 1)
	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	cookbook.CreateRecipe(args[0], flavor)
}

var cookbookCmd = &cobra.Command{
	Use:     "cookbook",
	Aliases: []string{"recipes", "samples"},
	Short:   "Shows dbdeployer samples",
	Long:    `Shows practical examples of dbdeployer usages, by creating usage scripts.`,
}

var listCookbookCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows available dbdeployer samples",
	Long:  `Shows list of available cookbook recipes`,
	Run:   listCookbook,
}

var showCookbookCmd = &cobra.Command{
	Use:   "show recipe_name",
	Short: "Shows the contents of a given recipe",
	Long:  `Shows the contents of a given recipe, without actually running it`,
	Run:   showCookbook,
}

var createCookbookCmd = &cobra.Command{
	Use:     "create recipe_name or ALL",
	Aliases: []string{"make"},
	Short:   "creates a script for a given recipe",
	Long:    `creates a script for given recipe`,
	Run:     createCookbook,
}

func init() {
	rootCmd.AddCommand(cookbookCmd)
	cookbookCmd.AddCommand(listCookbookCmd)
	cookbookCmd.AddCommand(createCookbookCmd)
	cookbookCmd.AddCommand(showCookbookCmd)
	showCookbookCmd.Flags().BoolP(globals.RawLabel, "", false, "Shows the recipe without variable substitution")
	setPflag(cookbookCmd, globals.FlavorLabel, "", "", "", "For which flavor this recipe is", false)
}
