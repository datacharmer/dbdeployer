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
	"testing"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/compare"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
)

func TestGetLatestVersion(t *testing.T) {
	err := sandbox.SetMockEnvironment(sandbox.DefaultMockDir)
	if err != nil {
		t.Fatal("mock dir creation failed")
	}
	compare.OkIsNil("creating mock environment", err, t)
	var versions = []string{
		"5.0.89",
		"5.1.67",
		"5.5.48",
		"5.6.78",
		"5.7.22",
		"8.0.11",
	}

	latest := common.GetLatestVersion(defaults.Defaults().SandboxBinary, "", common.MySQLFlavor)
	compare.OkMatchesString("latest version", latest, `NOTFOUND`, t)
	recipe, _, _ := GetRecipe("single", common.MySQLFlavor)
	compare.OkMatchesString("recipe single", recipe, "NOTFOUND", t)
	for _, version := range versions {
		fmt.Printf("%s\n", version)
		err = sandbox.CreateMockVersion(version)
		compare.OkIsNil(fmt.Sprintf("creating mock version %s", version), err, t)
		latest = common.GetLatestVersion(defaults.Defaults().SandboxBinary, "", common.MySQLFlavor)
		compare.OkEqualString("latest version", latest, version, t)
		recipe, _, _ = GetRecipe("single", common.MySQLFlavor)
		compare.OkMatchesString("recipe single", recipe, version, t)
	}
	dummyRecipe := RecipeTemplate{
		Contents: `sandbox prefix ="{{.SandboxPrefix}}" - replication prefix = "{{.MasterSlavePrefix}}" <{{.PxcBasePort}}>`,
	}
	RecipesList["dummy"] = dummyRecipe

	recipeText, _, _ := GetRecipe("dummy", common.MySQLFlavor)
	compare.OkMatchesString("dummy recipe", recipeText, defaults.Defaults().SandboxPrefix, t)
	compare.OkMatchesString("dummy recipe", recipeText, defaults.Defaults().MasterSlavePrefix, t)

	err = sandbox.RemoveMockEnvironment(sandbox.DefaultMockDir)
	compare.OkIsNil("removing mock environment", err, t)
}
