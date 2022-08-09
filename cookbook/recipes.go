// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cookbook

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/alexeyco/simpletable"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

var (
	AuxiliaryRecipes        = []string{"prerequisites", "include"}
	PrerequisitesShown bool = false
)

type TemplateSort struct {
	name       string
	scriptName string
	flavor     string
}

type ByScriptName []TemplateSort
type ByName []TemplateSort
type ByFlavorAndName []TemplateSort

func (a ByScriptName) Len() int           { return len(a) }
func (a ByScriptName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScriptName) Less(i, j int) bool { return a[i].scriptName < a[j].scriptName }

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].name < a[j].name }

func (a ByFlavorAndName) Len() int      { return len(a) }
func (a ByFlavorAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByFlavorAndName) Less(i, j int) bool {
	return a[i].flavor < a[j].flavor || (a[i].flavor == a[j].flavor && a[i].name < a[j].name)
}

func ListRecipes(flavor, sortBy string) {
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "recipe"},
			{Align: simpletable.AlignCenter, Text: "script name"},
			{Align: simpletable.AlignCenter, Text: "description"},
			{Align: simpletable.AlignCenter, Text: "needed\n flavor"},
		},
	}

	var recipeSortList []TemplateSort

	for name, template := range RecipesList {
		recipeSortList = append(recipeSortList, TemplateSort{
			name:       name,
			scriptName: template.ScriptName,
			flavor:     template.RequiredFlavor,
		})
	}

	switch sortBy {
	case "name":
		sort.Sort(ByName(recipeSortList))
	case "script":
		sort.Sort(ByScriptName(recipeSortList))
	case "flavor":
		sort.Sort(ByFlavorAndName(recipeSortList))
	default:
		common.Exitf(1, "sort parameter '%s' incorrect. Accepted: [name, script, flavor]", sortBy)
	}
	for _, sortTemplate := range recipeSortList {
		template := RecipesList[sortTemplate.name]
		if template.IsExecutable {
			if flavor == "" || flavor == template.RequiredFlavor {
				var cells []*simpletable.Cell
				if template.RequiredFlavor == "" {
					template.RequiredFlavor = "-"
				}
				cells = append(cells, &simpletable.Cell{Text: sortTemplate.name})
				cells = append(cells, &simpletable.Cell{Text: template.ScriptName})
				cells = append(cells, &simpletable.Cell{Text: template.Description})
				cells = append(cells, &simpletable.Cell{Text: template.RequiredFlavor})
				table.Body.Cells = append(table.Body.Cells, cells)
			}
		}
	}
	table.SetStyle(simpletable.StyleRounded)
	table.Println()
}

func getCookbookDirectory() string {
	cookbookDir := defaults.Defaults().CookbookDirectory
	if !common.DirExists(cookbookDir) {
		err := os.Mkdir(cookbookDir, globals.PublicDirectoryAttr)
		if err != nil {
			common.Exitf(1, "error creating cookbook directory %s: %s", cookbookDir, err)
		}
	}
	return cookbookDir
}

func recipeExists(recipeName string) bool {
	_, ok := RecipesList[recipeName]
	return ok
}

func createPrerequisites() string {
	cookbookDir := getCookbookDirectory()
	preReqScript := path.Join(cookbookDir, CookbookPrerequisites)
	for _, recipeName := range AuxiliaryRecipes {
		CreateRecipe(recipeName, "")
	}
	return preReqScript
}

func showPrerequisites(flavor string) {
	if PrerequisitesShown {
		return
	}
	prerequisitesScript := createPrerequisites()
	fmt.Printf("No tarballs for flavor %s were found in your environment\n", flavor)
	fmt.Printf("Please read instructions in %s\n", prerequisitesScript)
	PrerequisitesShown = true
}

func ShowRecipe(recipeName string, flavor string, raw bool) {
	if !recipeExists(recipeName) {
		fmt.Printf("recipe %s not found\n", recipeName)
		os.Exit(1)
	}
	if raw {
		fmt.Printf("%s\n", RecipesList[recipeName].Contents)
		return
	}
	recipe := RecipesList[recipeName]
	if recipe.RequiredFlavor != "" && flavor == "" {
		flavor = recipe.RequiredFlavor
	}
	if flavor == "" {
		flavor = common.MySQLFlavor
	}
	recipeText, _, err := GetRecipe(recipeName, flavor)
	if err != nil {
		showPrerequisites(flavor)
	}
	fmt.Printf("%s\n", recipeText)
}

func CreateRecipe(recipeName, flavor string) {
	var isRecursive bool = false

	for _, auxRecipeName := range AuxiliaryRecipes {
		if auxRecipeName == recipeName {
			isRecursive = true
		}
	}
	if strings.ToLower(recipeName) == "all" {
		for name := range RecipesList {
			CreateRecipe(name, flavor)
		}
		return
	}
	recipe := RecipesList[recipeName]
	if recipe.RequiredFlavor != "" {
		flavor = recipe.RequiredFlavor
	}
	if flavor == "" {
		flavor = common.MySQLFlavor
	}
	if !recipeExists(recipeName) {
		fmt.Printf("recipe %s not found\n", recipeName)
		os.Exit(1)
	}
	recipeText, versionCode, err := GetRecipe(recipeName, flavor)
	if err != nil && !isRecursive {
		showPrerequisites(flavor)
		common.Exitf(1, "error getting recipe %s: %s", recipeName, err)
	}
	if versionCode == globals.ErrNoVersionFound && !isRecursive {
		showPrerequisites(flavor)
	}
	cookbookDir := getCookbookDirectory()
	if recipe.ScriptName != CookbookInclude {
		targetInclude := path.Join(cookbookDir, CookbookInclude)
		if !common.FileExists(targetInclude) && !isRecursive {
			CreateRecipe("include", flavor)
		}
	}
	targetScript := path.Join(cookbookDir, recipe.ScriptName)
	//if common.FileExists(targetScript) {
	//	fmt.Printf("Script %s already created\n", targetScript)
	//	return
	//}
	err = common.WriteString(recipeText, targetScript)
	if err != nil {
		common.Exitf(1, "error writing file %s: %s", targetScript, err)
	}
	if recipe.IsExecutable {
		err = os.Chmod(targetScript, globals.ExecutableFileAttr)
		if err != nil {
			common.Exitf(1, "error while making file %s executable: %s", targetScript, err)
		}
	}
	fmt.Printf("%s created\n", targetScript)
}

func GetRecipe(recipeName, flavor string) (string, int, error) {
	var text string

	recipe, ok := RecipesList[recipeName]
	if !ok {
		return text, globals.ErrNoRecipeFound, fmt.Errorf("recipe %s not found", recipeName)
	}
	latestVersions := make(map[string]string)
	for _, version := range globals.SupportedMySQLVersions {
		latest := common.GetLatestVersion(defaults.Defaults().SandboxBinary, version, common.MySQLFlavor)
		if latest != "" {
			latestVersions[version] = latest
		} else {
			latestVersions[version] = fmt.Sprintf("%s_%s", globals.VersionNotFound, version)
		}
	}
	latestVersion := common.GetLatestVersion(defaults.Defaults().SandboxBinary, "", flavor)
	versionCode := 0
	if latestVersion == globals.VersionNotFound {
		versionCode = globals.ErrNoVersionFound
	}
	var data = defaults.DefaultsToMap()
	data["Copyright"] = globals.ShellScriptCopyright
	data["TemplateName"] = recipeName
	data["LatestVersion"] = latestVersion
	for version, latest := range latestVersions {
		reDot := regexp.MustCompile(`\.`)
		versionName := reDot.ReplaceAllString(version, "_")
		fieldName := fmt.Sprintf("Latest%s", versionName)
		data[fieldName] = latest
	}
	text, err := common.SafeTemplateFill(recipeName, recipe.Contents, data)
	if err != nil {
		return globals.EmptyString, versionCode, err
	}
	return text, versionCode, nil
}
