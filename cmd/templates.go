// Copyright © 2018 Giuseppe Maxia
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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
	"os"
)

type TemplateInfo struct {
	Origin      int
	Group       string
	Name        string
	Description string
}

func FindTemplate(requested string) (group, contents string) {
	for name, tvar := range sandbox.AllTemplates {
		for k, v := range tvar {
			if k == requested {
				contents = v.Contents
				group = name
				return
			}
		}
	}
	fmt.Printf("template '%s' not found\n", requested)
	os.Exit(1)
	return
}

func ShowTemplate(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Argument required: template name")
		os.Exit(1)
	}
	requested := args[0]
	_, contents := FindTemplate(requested)
	fmt.Println(contents)
}

func GetTemplatesList(wanted string) (tlist []TemplateInfo) {
	found := false
	for group_name, tvar := range sandbox.AllTemplates {
		will_include := true
		//fmt.Printf("[%s]\n", group_name)
		if wanted != "" {
			if wanted != group_name {
				will_include = false
			}
		}
		var td TemplateInfo
		if will_include {
			for k, v := range tvar {
				td.Description = v.Description
				td.Group = group_name
				td.Origin = v.Origin
				td.Name = k
				tlist = append(tlist, td)
				found = true
			}
		}
	}
	if !found {
		fmt.Printf("group %s not found\n", wanted)
		os.Exit(1)
	}
	return
}

func ListTemplates(cmd *cobra.Command, args []string) {
	wanted := ""
	if len(args) > 0 {
		wanted = args[0]
	}
	flags := cmd.Flags()
	simple_list, _ := flags.GetBool("simple")

	templates := GetTemplatesList(wanted)
	for _, template := range templates {
		origin := "   "
		if template.Origin == sandbox.TEMPLATE_FILE {
			origin = "{F}"
		}
		if simple_list {
			fmt.Printf("%s %-13s %-25s\n", origin, "["+template.Group+"]", template.Name)
		} else {
			fmt.Printf("%s %-13s %-25s : %s\n", origin, "["+template.Group+"]", template.Name, template.Description)
		}
	}
}

func RunDescribeTemplate(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Argument required: template name")
		os.Exit(1)
	}
	requested := args[0]
	flags := cmd.Flags()
	complete_listing, _ := flags.GetBool("with-contents")
	DescribeTemplate(requested, complete_listing)
}

func GetTemplatesDescription(requested string, complete_listing bool) string {
	group, contents := FindTemplate(requested)
	out := ""
	origin := "   "
	if sandbox.AllTemplates[group][requested].Origin == sandbox.TEMPLATE_FILE {
		origin = "{F}"
	}
	out += fmt.Sprintf("# Collection    : %s\n", group)
	out += fmt.Sprintf("# Name   %s    : %s\n", origin, requested)
	out += fmt.Sprintf("# Description 	: %s\n", sandbox.AllTemplates[group][requested].Description)
	out += fmt.Sprintf("# Notes     	: %s\n", sandbox.AllTemplates[group][requested].Notes)
	out += fmt.Sprintf("# Length     	: %d\n", len(contents))
	if complete_listing {
		out += fmt.Sprintf("##START %s\n", requested)
		out += fmt.Sprintf("%s\n", contents)
		out += fmt.Sprintf("##END %s\n\n", requested)
	}
	return out
}

func DescribeTemplate(requested string, complete_listing bool) {
	fmt.Printf("%s", GetTemplatesDescription(requested, complete_listing))
}

func ExportTemplates(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println("The export command requires two arguments: group_name and directory_name")
		fmt.Println("If group_name is 'all', it will export all groups")
		os.Exit(1)
	}
	wanted := args[0]
	dir_name := args[1]
	if wanted == "all" || wanted == "ALL" {
		wanted = ""
	}
	if common.DirExists(dir_name) {
		fmt.Printf("# Directory <%s> already exists\n", dir_name)
		os.Exit(1)
	}
	common.Mkdir(dir_name)

	found := false
	for group_name, group := range sandbox.AllTemplates {
		if group_name == wanted || wanted == "" {
			found = true
			group_dir := dir_name + "/" + group_name
			if !common.DirExists(group_dir) {
				common.Mkdir(group_dir)
			}
			for name, template := range group {
				file_name := group_dir + "/" + name
				common.WriteString(common.TrimmedLines(template.Contents), file_name)
			}
		}
	}
	if !found {
		fmt.Printf("Group %s not found\n", wanted)
		os.Exit(1)
	}
	fmt.Printf("Exported to %s\n", dir_name)
}

func LoadTemplates() {
	load_dir := defaults.ConfigurationDir + "/templates" + common.CompatibleVersion
	if !common.DirExists(load_dir) {
		return
	}
	for group_name, group := range sandbox.AllTemplates {
		group_dir := load_dir + "/" + group_name
		if !common.DirExists(group_dir) {
			continue
		}
		for name, template := range group {
			file_name := group_dir + "/" + name
			if !common.FileExists(file_name) {
				continue
			}
			new_contents := common.SlurpAsString(file_name)
			new_template := template
			new_template.Origin = sandbox.TEMPLATE_FILE
			new_template.Contents = new_contents
			sandbox.AllTemplates[group_name][name] = new_template
			// fmt.Printf("# Template %s loaded from %s\n",name, file_name)
		}
	}
}

func ImportTemplates(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println("The import command requires two arguments: group_name and dir_name")
		fmt.Println("If group_name is 'all', it will import all groups")
		os.Exit(1)
	}
	wanted := args[0]
	if wanted == "all" || wanted == "ALL" {
		wanted = ""
	}
	dir_name := args[1]
	if !common.DirExists(dir_name) {
		fmt.Printf("# Directory <%s> doesn't exist\n", dir_name)
		os.Exit(1)
	}
	found := false
	for group_name, group := range sandbox.AllTemplates {
		group_dir := dir_name + "/" + group_name
		if !common.DirExists(group_dir) {
			continue
		}
		for name, _ := range group {
			file_name := group_dir + "/" + name
			if !common.FileExists(file_name) {
				continue
			}
			if group_name == wanted || wanted == "" {
				found = true
			} else {
				continue
			}
			new_contents := common.SlurpAsString(file_name)
			// fmt.Printf("Group: %s - File: %s\n", group_name, name)
			// fmt.Printf("sizes: %d %d\n",len(template.Contents), len(new_contents))
			if !common.DirExists(defaults.ConfigurationDir) {
				common.Mkdir(defaults.ConfigurationDir)
			}
			destination_dir := defaults.ConfigurationDir + "/templates" + common.CompatibleVersion
			if !common.DirExists(destination_dir) {
				common.Mkdir(destination_dir)
			}
			dest_group_dir := destination_dir + "/" + group_name
			if !common.DirExists(dest_group_dir) {
				common.Mkdir(dest_group_dir)
			}
			dest_file := dest_group_dir + "/" + name
			common.WriteString(new_contents, dest_file)
			fmt.Printf("# Template %s written to %s\n", name, dest_file)
		}
	}
	if !found {
		fmt.Printf("Group %s not found\n", wanted)
		os.Exit(1)
	}
}

func ResetTemplates(cmd *cobra.Command, args []string) {

	templates_dir := defaults.ConfigurationDir + "/templates" + common.CompatibleVersion
	if !common.DirExists(templates_dir) {
		return
	}
	err := os.RemoveAll(templates_dir)
	if err != nil {
		fmt.Printf("Error removing %s\n%s\n", templates_dir, err)
	}
	fmt.Printf("Templates directory %s removed\n", templates_dir)
}

var (
	templatesCmd = &cobra.Command{
		Use:     "templates",
		Aliases: []string{"template", "tmpl", "templ"},
		Short:   "Admin operations on templates",
		Hidden:  false,
		Long: `The commands in this section show the templates used
to create and manipulate sandboxes.
More commands (and flags) will follow to allow changing templates
either temporarily or permanently.`,
	}

	templatesListCmd = &cobra.Command{
		Use:   "list [group]",
		Short: "list available templates",
		Long:  ``,
		Run:   ListTemplates,
	}

	templatesShowCmd = &cobra.Command{
		Use:   "show template_name",
		Short: "Show a given template",
		Long:  ``,
		Run:   ShowTemplate,
	}
	templatesDescribeCmd = &cobra.Command{
		Use:     "describe template_name",
		Aliases: []string{"descr", "structure", "struct"},
		Short:   "Describe a given template",
		Long:    ``,
		Run:     RunDescribeTemplate,
	}
	templatesExportCmd = &cobra.Command{
		Use:   "export group_name dir_name",
		Short: "Exports all templates to a directory",
		Long:  `Exports a group of templates to a given directory`,
		Run:   ExportTemplates,
	}
	templatesImportCmd = &cobra.Command{
		Use:   "import group_name directory_name",
		Short: "imports all templates from a directory",
		Long:  `Imports a group of templates to a given directory`,
		Run:   ImportTemplates,
	}
	templatesResetCmd = &cobra.Command{
		Use:     "reset",
		Aliases: []string{"remove"},
		Short:   "Removes all template files",
		Long:    `Removes all template files that were imported and uses internal values.`,
		Run:     ResetTemplates,
	}
)

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(templatesListCmd)
	templatesCmd.AddCommand(templatesShowCmd)
	templatesCmd.AddCommand(templatesDescribeCmd)
	templatesCmd.AddCommand(templatesExportCmd)
	templatesCmd.AddCommand(templatesImportCmd)
	templatesCmd.AddCommand(templatesResetCmd)

	templatesListCmd.Flags().BoolP("simple", "s", false, "Shows only the template names, without description")
	templatesDescribeCmd.Flags().BoolP("with-contents", "", false, "Shows complete structure and contents")
}
