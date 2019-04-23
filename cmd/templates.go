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
	"path"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

type TemplateInfo struct {
	TemplateInFile bool
	Group          string
	Name           string
	Description    string
}

func findTemplate(requested string) (group, templateName, contents string) {
	for name, tvar := range sandbox.AllTemplates {
		for k, v := range tvar {
			if k == requested || k == requested+"_template" {
				contents = v.Contents
				group = name
				templateName = k
				return
			}
		}
	}
	common.Exitf(1, globals.ErrTemplateNotFound, requested)
	return
}

func showTemplate(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exitf(1, globals.ErrArgumentRequired, "template name")
	}
	requested := args[0]
	_, _, contents := findTemplate(requested)
	fmt.Println(contents)
}

func getTemplatesList(wanted string) (tlist []TemplateInfo) {
	found := false
	for groupName, tvar := range sandbox.AllTemplates {
		willInclude := true
		//common.CondPrintf("[%s]\n", group_name)
		if wanted != "" {
			if wanted != groupName {
				willInclude = false
			}
		}
		var td TemplateInfo
		if willInclude {
			for k, v := range tvar {
				td.Description = v.Description
				td.Group = groupName
				td.TemplateInFile = v.TemplateInFile
				td.Name = k
				tlist = append(tlist, td)
				found = true
			}
		}
	}
	if !found {
		common.Exitf(1, globals.ErrGroupNotFound, wanted)
	}
	return
}

func listTemplates(cmd *cobra.Command, args []string) {
	wanted := ""
	if len(args) > 0 {
		wanted = args[0]
	}
	flags := cmd.Flags()
	simpleList, _ := flags.GetBool(globals.SimpleLabel)

	templates := getTemplatesList(wanted)
	for _, template := range templates {
		origin := "   "
		if template.TemplateInFile {
			origin = "{F}"
		}
		if simpleList {
			fmt.Printf("%s %-13s %-25s\n", origin, "["+template.Group+"]", template.Name)
		} else {
			fmt.Printf("%s %-13s %-25s : %s\n", origin, "["+template.Group+"]", template.Name, template.Description)
		}
	}
}

func runDescribeTemplate(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exitf(1, globals.ErrArgumentRequired, "template name")
	}
	requested := args[0]
	flags := cmd.Flags()
	completeListing, _ := flags.GetBool(globals.WithContentsLabel)
	describeTemplate(requested, completeListing)
}

func getTemplatesDescription(requested string, completeListing bool) string {
	group, templateName, contents := findTemplate(requested)
	out := ""
	origin := "   "
	if sandbox.AllTemplates[group][requested].TemplateInFile {
		origin = "{F}"
	}
	out += fmt.Sprintf("# Collection    : %s\n", group)
	out += fmt.Sprintf("# Name   %s    : %s\n", origin, templateName)
	out += fmt.Sprintf("# Description 	: %s\n", sandbox.AllTemplates[group][templateName].Description)
	out += fmt.Sprintf("# Notes     	: %s\n", sandbox.AllTemplates[group][templateName].Notes)
	out += fmt.Sprintf("# Length     	: %d\n", len(contents))
	if completeListing {
		out += fmt.Sprintf("##START %s\n", templateName)
		out += fmt.Sprintf("%s\n", contents)
		out += fmt.Sprintf("##END %s\n\n", templateName)
	}
	return out
}

func describeTemplate(requested string, completeListing bool) {
	fmt.Printf("%s", getTemplatesDescription(requested, completeListing))
}

func exportTemplates(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		common.Exit(1,
			"the export command requires two arguments: group_name and directory_name",
			"If group_name is 'all', it will export all groups")
	}
	wanted := args[0]
	dirName := args[1]
	templateName := ""
	if len(args) > 2 {
		templateName = args[2]
	}
	if wanted == "all" || wanted == "ALL" {
		wanted = ""
	}
	if common.DirExists(dirName) {
		common.Exitf(1, globals.ErrDirectoryAlreadyExists, dirName)
	}
	common.Mkdir(dirName)
	err := common.WriteString(common.VersionDef, path.Join(dirName, "version.txt"))
	if err != nil {
		common.Exitf(1, "error writing template version file")
	}

	foundGroup := false
	foundTemplate := false
	for groupName, group := range sandbox.AllTemplates {
		if groupName == wanted || wanted == "" {
			foundGroup = true
			groupDir := path.Join(dirName, groupName)
			if !common.DirExists(groupDir) {
				common.Mkdir(groupDir)
			}
			for name, template := range group {
				if templateName == "" || common.Includes(name, templateName) {
					fileName := path.Join(groupDir, name)
					err = common.WriteString(common.TrimmedLines(template.Contents), fileName)
					if err != nil {
						common.Exitf(1, "error writing template %s", fileName)
					}
					fmt.Printf("%s/%s exported\n", groupName, name)
					foundTemplate = true
				}
			}
		}
	}
	if !foundGroup {
		common.Exitf(1, globals.ErrGroupNotFound, wanted)
	}
	if !foundTemplate {
		common.Exitf(1, globals.ErrTemplateNotFound, templateName)
	}
	fmt.Printf("Exported to %s\n", dirName)
}

// Called by rootCmd when dbdeployer starts
func loadTemplates() {
	loadDir := path.Join(defaults.ConfigurationDir, "templates"+common.CompatibleVersion)
	if !common.DirExists(loadDir) {
		return
	}
	for groupName, group := range sandbox.AllTemplates {
		groupDir := path.Join(loadDir, groupName)
		if !common.DirExists(groupDir) {
			continue
		}
		for name, template := range group {
			fileName := path.Join(groupDir, name)
			if !common.FileExists(fileName) {
				continue
			}
			newContents, err := common.SlurpAsString(fileName)
			if err != nil {
				common.Exitf(1, "error reading %s\n", fileName)
			}
			newTemplate := template
			newTemplate.TemplateInFile = true
			newTemplate.Contents = newContents
			sandbox.AllTemplates[groupName][name] = newTemplate
			// fmt.Printf("# Template %s loaded from %s\n",name, file_name)
		}
	}
}

func importTemplates(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		common.Exit(1,
			"the import command requires two arguments: group_name and dir_name",
			"If group_name is 'all', it will import all groups")
	}
	wanted := args[0]
	if wanted == "all" || wanted == "ALL" {
		wanted = ""
	}
	dirName := args[1]
	if !common.DirExists(dirName) {
		common.Exitf(1, globals.ErrDirectoryNotFound, dirName)
	}
	templateName := ""
	if len(args) > 2 {
		templateName = args[2]
	}
	versionFile := path.Join(dirName, "version.txt")
	if !common.FileExists(versionFile) {
		common.Exitf(1, "file %s not found. Unable to validate templates.", versionFile)
	}
	versionFileContents, err := common.SlurpAsString(versionFile)
	if err != nil {
		common.Exitf(1, "error reading version file")
	}
	templateVersion := strings.TrimSpace(versionFileContents)
	_, err = common.VersionToList(templateVersion)
	if err != nil {
		common.Exitf(1, "error converting version %s\n", templateVersion)
	}
	// fmt.Printf("%v\n",version_list)
	compatibleVersionList, err := common.VersionToList(common.CompatibleVersion)
	if err != nil {
		common.Exitf(1, "error converting compatible version %s from file %s\n", common.CompatibleVersion, versionFile)
	}
	templateVersionList, err := common.VersionToList(templateVersion)
	if err != nil {
		common.Exitf(1, "error converting version %s from file %s\n", templateVersion, versionFile)
	}
	compatibleTemplate, err := common.GreaterOrEqualVersionList(templateVersionList, compatibleVersionList)
	common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
	if !compatibleTemplate {
		common.Exitf(1, "templates are for version %s. The minimum compatible version is %s", templateVersion, common.CompatibleVersion)
	}
	foundGroup := false
	foundTemplate := false
	for groupName, group := range sandbox.AllTemplates {
		groupDir := path.Join(dirName, groupName)
		if !common.DirExists(groupDir) {
			continue
		}
		for name := range group {
			fileName := path.Join(groupDir, name)
			if !common.FileExists(fileName) {
				continue
			}
			if groupName == wanted || wanted == "" {
				foundGroup = true
			} else {
				continue
			}
			if templateName == "" || common.Includes(name, templateName) {
				foundTemplate = true
			} else {
				continue
			}
			newContents, err := common.SlurpAsString(fileName)
			if err != nil {
				common.Exitf(1, "error reading template %s\n", fileName)
			}
			// fmt.Printf("Group: %s - File: %s\n", group_name, name)
			// fmt.Printf("sizes: %d %d\n",len(template.Contents), len(new_contents))
			if !common.DirExists(defaults.ConfigurationDir) {
				common.Mkdir(defaults.ConfigurationDir)
			}
			destinationDir := path.Join(defaults.ConfigurationDir, "templates"+common.CompatibleVersion)
			if !common.DirExists(destinationDir) {
				common.Mkdir(destinationDir)
			}
			destGroupDir := path.Join(destinationDir, groupName)
			if !common.DirExists(destGroupDir) {
				common.Mkdir(destGroupDir)
			}
			destFile := path.Join(destGroupDir, name)
			err = common.WriteString(newContents, destFile)
			if err != nil {
				common.Exitf(1, "error writing %s\n", destFile)
			}
			fmt.Printf("# Template %s written to %s\n", name, destFile)
		}
	}
	if !foundGroup {
		common.Exitf(1, globals.ErrGroupNotFound, wanted)
	}
	if !foundTemplate {
		common.Exitf(1, globals.ErrTemplateNotFound, templateName)
	}
}

func resetTemplates(cmd *cobra.Command, args []string) {
	// TODO: loop through the templates directories and remove all the ones that have compatible versions.
	templatesDir := path.Join(defaults.ConfigurationDir, "templates"+common.CompatibleVersion)
	if !common.DirExists(templatesDir) {
		return
	}
	err := os.RemoveAll(templatesDir)
	if err != nil {
		fmt.Printf("Error removing %s\n%s\n", templatesDir, err)
	}
	fmt.Printf("Templates directory %s removed\n", templatesDir)
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
		Run:   listTemplates,
	}

	templatesShowCmd = &cobra.Command{
		Use:         "show template_name",
		Short:       "Show a given template",
		Long:        ``,
		Run:         showTemplate,
		Annotations: map[string]string{"export": ExportAnnotationToJson(TemplateNameExport)},
	}
	templatesDescribeCmd = &cobra.Command{
		Use:         "describe template_name",
		Aliases:     []string{"descr", "structure", "struct"},
		Short:       "Describe a given template",
		Long:        ``,
		Run:         runDescribeTemplate,
		Annotations: map[string]string{"export": ExportAnnotationToJson(TemplateNameExport)},
	}
	templatesExportCmd = &cobra.Command{
		Use:         "export group_name directory_name [template_name]",
		Short:       "Exports templates to a directory",
		Long:        `Exports a group of templates (or "ALL") to a given directory`,
		Run:         exportTemplates,
		Annotations: map[string]string{"export": ExportAnnotationToJson(TemplateGroupExport)},
	}
	templatesImportCmd = &cobra.Command{
		Use:         "import group_name directory_name [template_name]",
		Short:       "imports templates from a directory",
		Long:        `Imports a group of templates (or "ALL") from a given directory`,
		Run:         importTemplates,
		Annotations: map[string]string{"export": ExportAnnotationToJson(TemplateGroupExport)},
	}
	templatesResetCmd = &cobra.Command{
		Use:     "reset",
		Aliases: []string{"remove"},
		Short:   "Removes all template files",
		Long:    `Removes all template files that were imported and starts using internal values.`,
		Run:     resetTemplates,
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

	templatesListCmd.Flags().BoolP(globals.SimpleLabel, "s", false, "Shows only the template names, without description")
	templatesDescribeCmd.Flags().BoolP(globals.WithContentsLabel, "", false, "Shows complete structure and contents")
}
