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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

type TemplateInfo struct {
	Origin      int
	Group       string
	Name        string
	Description string
}

func FindTemplate(requested string) (group, template_name, contents string) {
	for name, tvar := range sandbox.AllTemplates {
		for k, v := range tvar {
			if k == requested || k == requested+"_template" {
				contents = v.Contents
				group = name
				template_name = k
				return
			}
		}
	}
	common.Exitf(1, "template '%s' not found", requested)
	return
}

func ShowTemplate(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1, "Argument required: template name")
	}
	requested := args[0]
	_, _, contents := FindTemplate(requested)
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
		common.Exitf(1, "group %s not found\n", wanted)
	}
	return
}

func ListTemplates(cmd *cobra.Command, args []string) {
	wanted := ""
	if len(args) > 0 {
		wanted = args[0]
	}
	flags := cmd.Flags()
	simple_list, _ := flags.GetBool(defaults.SimpleLabel)

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
		common.Exit(1, "Argument required: template name")
	}
	requested := args[0]
	flags := cmd.Flags()
	complete_listing, _ := flags.GetBool(defaults.WithContentsLabel)
	DescribeTemplate(requested, complete_listing)
}

func GetTemplatesDescription(requested string, complete_listing bool) string {
	group, template_name, contents := FindTemplate(requested)
	out := ""
	origin := "   "
	if sandbox.AllTemplates[group][requested].Origin == sandbox.TEMPLATE_FILE {
		origin = "{F}"
	}
	out += fmt.Sprintf("# Collection    : %s\n", group)
	out += fmt.Sprintf("# Name   %s    : %s\n", origin, template_name)
	out += fmt.Sprintf("# Description 	: %s\n", sandbox.AllTemplates[group][template_name].Description)
	out += fmt.Sprintf("# Notes     	: %s\n", sandbox.AllTemplates[group][template_name].Notes)
	out += fmt.Sprintf("# Length     	: %d\n", len(contents))
	if complete_listing {
		out += fmt.Sprintf("##START %s\n", template_name)
		out += fmt.Sprintf("%s\n", contents)
		out += fmt.Sprintf("##END %s\n\n", template_name)
	}
	return out
}

func DescribeTemplate(requested string, complete_listing bool) {
	fmt.Printf("%s", GetTemplatesDescription(requested, complete_listing))
}

func ExportTemplates(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		common.Exit(1,
			"The export command requires two arguments: group_name and directory_name",
			"If group_name is 'all', it will export all groups")
	}
	wanted := args[0]
	dir_name := args[1]
	template_name := ""
	if len(args) > 2 {
		template_name = args[2]
	}
	if wanted == "all" || wanted == "ALL" {
		wanted = ""
	}
	if common.DirExists(dir_name) {
		common.Exitf(1, "# Directory <%s> already exists", dir_name)
	}
	common.Mkdir(dir_name)
	common.WriteString(common.VersionDef, dir_name+"/version.txt")

	found_group := false
	found_template := false
	for group_name, group := range sandbox.AllTemplates {
		if group_name == wanted || wanted == "" {
			found_group = true
			group_dir := dir_name + "/" + group_name
			if !common.DirExists(group_dir) {
				common.Mkdir(group_dir)
			}
			for name, template := range group {
				if template_name == "" || common.Includes(name, template_name) {
					file_name := group_dir + "/" + name
					common.WriteString(common.TrimmedLines(template.Contents), file_name)
					fmt.Printf("%s/%s exported\n", group_name, name)
					found_template = true
				}
			}
		}
	}
	if !found_group {
		common.Exitf(1, "Group %s not found", wanted)
	}
	if !found_template {
		common.Exitf(1, "template %s not found", template_name)
	}
	fmt.Printf("Exported to %s\n", dir_name)
}

// Called by rootCmd when dbdeployer starts
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
		common.Exit(1,
			"The import command requires two arguments: group_name and dir_name",
			"If group_name is 'all', it will import all groups")
	}
	wanted := args[0]
	if wanted == "all" || wanted == "ALL" {
		wanted = ""
	}
	dir_name := args[1]
	if !common.DirExists(dir_name) {
		common.Exitf(1, "# Directory <%s> doesn't exist", dir_name)
	}
	template_name := ""
	if len(args) > 2 {
		template_name = args[2]
	}
	version_file := dir_name + "/version.txt"
	if !common.FileExists(version_file) {
		common.Exitf(1, "File %s not found. Unable to validate templates.", version_file)
	}
	template_version := strings.TrimSpace(common.SlurpAsString(version_file))
	version_list := common.VersionToList(template_version)
	// fmt.Printf("%v\n",version_list)
	compatible_version_list := common.VersionToList(common.CompatibleVersion)
	if version_list[0] < 0 {
		common.Exitf(1, "Invalid version (%s) found in %s", template_version, version_file)
	}
	if !common.GreaterOrEqualVersion(template_version, compatible_version_list) {
		common.Exitf(1, "Templates are for version %s. The minimum compatible version is %s", template_version, common.CompatibleVersion)
	}
	found_group := false
	found_template := false
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
				found_group = true
			} else {
				continue
			}
			if template_name == "" || common.Includes(name, template_name) {
				found_template = true
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
	if !found_group {
		common.Exitf(1, "Group %s not found", wanted)
	}
	if !found_template {
		common.Exitf(1, "template %s not found", template_name)
	}
}

func ResetTemplates(cmd *cobra.Command, args []string) {
	// TODO: loop through the templates directories and remove all the ones that have compatible versions.
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
		Short:   "Templates management",
		Hidden:  false,
		Long: `The commands in this section show the templates used
to create and manipulate sandboxes.
`,
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
		Use:   "export group_name directory_name [template_name]",
		Short: "Exports templates to a directory",
		Long:  `Exports a group of templates (or "ALL") to a given directory`,
		Run:   ExportTemplates,
	}
	templatesImportCmd = &cobra.Command{
		Use:   "import group_name directory_name [template_name]",
		Short: "imports templates from a directory",
		Long:  `Imports a group of templates (or "ALL") from a given directory`,
		Run:   ImportTemplates,
	}
	templatesResetCmd = &cobra.Command{
		Use:     "reset",
		Aliases: []string{"remove"},
		Short:   "Removes all template files",
		Long:    `Removes all template files that were imported and starts using internal values.`,
		Run:     ResetTemplates,
	}
)

func init() {
	defaultsCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(templatesListCmd)
	templatesCmd.AddCommand(templatesShowCmd)
	templatesCmd.AddCommand(templatesDescribeCmd)
	templatesCmd.AddCommand(templatesExportCmd)
	templatesCmd.AddCommand(templatesImportCmd)
	templatesCmd.AddCommand(templatesResetCmd)

	templatesListCmd.Flags().BoolP(defaults.SimpleLabel, "s", false, "Shows only the template names, without description")
	templatesDescribeCmd.Flags().BoolP(defaults.WithContentsLabel, "", false, "Shows complete structure and contents")
}
