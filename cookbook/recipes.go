// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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
package cookbook

import (
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"os"
	"path"
	"regexp"
	"strings"
)

type CookbookError struct {
	Error error
	Code int
}

const (
	ErrNoVersionFound = 1
)

func ListRecipes() {
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "recipe"},
			{Align: simpletable.AlignCenter, Text: "script name"},
			{Align: simpletable.AlignCenter, Text: "description"},
		},
	}

	for name, template := range RecipesList {
		if template.IsExecutable {
			var cells []*simpletable.Cell
			cells = append(cells, &simpletable.Cell{Text: name})
			cells = append(cells, &simpletable.Cell{Text: template.ScriptName})
			cells = append(cells, &simpletable.Cell{Text: template.Description})
			table.Body.Cells = append(table.Body.Cells, cells)
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
	if ok {
		return true
	}
	return false
}

func createPrerequisites() string {
	cookbookDir := getCookbookDirectory()
	preReqScript := path.Join(cookbookDir, CookbookPrerequisites)

	for _, recipeName := range []string{"prerequisites", "include"} {
		recipe := RecipesList[recipeName]
		recipeText := recipe.Contents
		targetScript := path.Join(cookbookDir, recipe.ScriptName)
		err := common.WriteString(recipeText, targetScript)
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
	return preReqScript
}

func showPrerequisites() {
	prerequisitesScript := createPrerequisites()
	fmt.Printf("No tarballs were found in your environment\n")
	fmt.Printf("Please read instructions in %s\n", prerequisitesScript)
	os.Exit(0)
}

func ShowRecipe(recipeName string, raw bool) {
	if !recipeExists(recipeName) {
		fmt.Printf("recipe %s not found\n", recipeName)
		os.Exit(1)
	}
	if raw {
		fmt.Printf("%s\n",RecipesList[recipeName].Contents)
		return
	}
	recipeText, cberr := GetRecipe(recipeName)
	if cberr.Error != nil {
		if cberr.Code == ErrNoVersionFound {
			showPrerequisites()
		}
		common.Exitf(1, "error getting recipe %s: %s", recipeName, cberr.Error)
	}
	fmt.Printf("%s\n", recipeText)
}

func CreateRecipe(recipeName string) {
	if strings.ToLower(recipeName) == "all" {
		for name, _ := range RecipesList {
			CreateRecipe(name)
		}
		return
	}
	recipe := RecipesList[recipeName]
	if !recipeExists(recipeName) {
		fmt.Printf("recipe %s not found\n", recipeName)
		os.Exit(1)
	}
	recipeText, cberr := GetRecipe(recipeName)
	if cberr.Error != nil {
		if cberr.Code == ErrNoVersionFound {
			showPrerequisites()
		}
		common.Exitf(1, "error getting recipe %s: %s", recipeName, cberr.Error)
	}
	cookbookDir := getCookbookDirectory()
	if recipe.ScriptName != CookbookInclude {
		targetInclude := path.Join(cookbookDir, CookbookInclude)
		if !common.FileExists(targetInclude) {
			CreateRecipe("include")
		}
	}
	targetScript := path.Join(cookbookDir, recipe.ScriptName)
	if common.FileExists(targetScript) {
		fmt.Printf("Script %s already created\n", targetScript)
		return
	}
	err := common.WriteString(recipeText, targetScript)
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

func GetLatestVersion(wantedVersion string) (string, CookbookError) {
	if wantedVersion == "" {
		wantedVersion = os.Getenv("WANTED_VERSION")
	}
	if wantedVersion == "" {
		wantedVersion = "5.7"
	}
	sandboxBinary := os.Getenv("SANDBOX_BINARY")
	if sandboxBinary == "" {
		sandboxBinary = defaults.Defaults().SandboxBinary
	}
	versions, err := common.GetVersionsFromDir(sandboxBinary)
	if err != nil {
		return "", CookbookError{Error:err, Code: ErrNoVersionFound}
	}
	if len(versions) == 0 {
		return "", CookbookError{Error: fmt.Errorf("no sorted version found for %s", wantedVersion), Code: ErrNoVersionFound}
	}

	sortedVersions := common.SortVersionsSubset(versions, wantedVersion)
	if len(sortedVersions) < 1 {
		return globals.EmptyString, CookbookError{fmt.Errorf("no sorted versions found"), ErrNoVersionFound}
	}
	latestVersion := sortedVersions[len(sortedVersions)-1]
	return latestVersion, CookbookError{nil, 0}
}

func GetRecipe(recipeName string) (string, CookbookError) {
	var text string

	recipe, ok := RecipesList[recipeName]
	if !ok {
		return text, CookbookError{fmt.Errorf("recipe %s not found", recipeName), 0}
	}
	latestVersions := make(map[string]string)
	for _, version := range []string{"5.0", "5.1", "5.5", "5.6", "5.7", "8.0"} {
		latest, _ := GetLatestVersion(version)
		if latest != "" {
			latestVersions[version] = latest
		} else {
			latestVersions[version] =  fmt.Sprintf("NOTFOUND_%s", version)
		}
	}
	latestVersion, cberr := GetLatestVersion("")
	if cberr.Error != nil {
		return globals.EmptyString, cberr
	}
	var data = common.StringMap{
		"Copyright":     globals.Copyright,
		"TemplateName":  recipeName,
		"LatestVersion": latestVersion,
	}
	for version, latest := range latestVersions {
		reDot := regexp.MustCompile(`\.`)
		versionName := reDot.ReplaceAllString(version, "_")
		fieldName := fmt.Sprintf("Latest%s", versionName)
		data[fieldName] = latest
	}
	text, err := common.SafeTemplateFill(recipeName, recipe.Contents, data)
	if err != nil {
		return globals.EmptyString, CookbookError{err, 0}
	}
	return text, CookbookError{nil, 0}
}
