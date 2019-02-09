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
package common

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/compare"
	"testing"
)

func TestHasCapability(t *testing.T) {

	type TestCapabilities struct {
		flavors  []string
		feature  string
		version  string
		expected bool
	}

	var capabilitiesList = []TestCapabilities{
		{[]string{MySQLFlavor, MariaDbFlavor, PerconaServerFlavor}, InstallDb, "5.1.72", true},
		{[]string{MySQLFlavor}, InstallDb, "5.7.72", false},
		{[]string{MariaDbFlavor}, InstallDb, "5.7.72", true},
		{[]string{MariaDbFlavor}, InstallDb, "10.4.1", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, SemiSynch, "5.1.72", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, SemiSynch, "5.5.40", true},
		{[]string{MySQLFlavor}, MySQLX, "5.5.40", false},
		{[]string{MySQLFlavor}, MySQLX, "5.7.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, MySQLXDefault, "5.7.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, MySQLXDefault, "8.0.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor, MariaDbFlavor}, DynVariables, "5.1.72", true},
	}
	for _, cl := range capabilitiesList {

		fmt.Printf("%s %s %s [%v]\n", cl.flavors, cl.feature, cl.version, cl.expected)
		for _, flavor := range cl.flavors {
			matches, err := HasCapability(flavor, cl.feature, cl.version)
			if err != nil {
				t.Logf("error getting %s capability %s: %s", cl.flavors, cl.feature, err)
				t.Fail()
			}
			compare.OkEqualBool(fmt.Sprintf("%s for %s %s", cl.feature, flavor, cl.version), matches, cl.expected, t)
		}
	}
}
