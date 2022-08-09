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
package common

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/compare"
	"testing"
)

type TestCapabilities struct {
	flavors  []string
	feature  string
	version  string
	expected bool
}

func TestHasCapability(t *testing.T) {
	var capabilitiesList = []TestCapabilities{
		{[]string{MySQLFlavor, MariaDbFlavor, PerconaServerFlavor}, InstallDb, "5.1.72", true},
		{[]string{MariaDbFlavor}, InstallDb, "5.5.0", true},
		{[]string{MariaDbFlavor}, InstallDb, "10.0.0", true},
		{[]string{MariaDbFlavor}, InstallDb, "10.1.0", true},
		{[]string{MariaDbFlavor}, InstallDb, "10.2.0", true},
		{[]string{MariaDbFlavor}, InstallDb, "10.3.0", true},
		{[]string{MariaDbFlavor}, InstallDb, "10.4.0", true},
		{[]string{MySQLFlavor}, InstallDb, "3.3.22", false},
		{[]string{MySQLFlavor}, InstallDb, "3.3.23", true},
		{[]string{MySQLFlavor}, InstallDb, "4.0.0", true},
		{[]string{MySQLFlavor}, InstallDb, "5.0.0", true},
		{[]string{MySQLFlavor}, InstallDb, "5.1.0", true},
		{[]string{MySQLFlavor}, InstallDb, "5.5.0", true},
		{[]string{MySQLFlavor}, InstallDb, "5.6.0", true},
		{[]string{MySQLFlavor}, InstallDb, "5.7.0", false},
		{[]string{PerconaServerFlavor}, InstallDb, "5.1.0", true},
		{[]string{PerconaServerFlavor}, InstallDb, "5.5.0", true},
		{[]string{PerconaServerFlavor}, InstallDb, "5.6.0", true},
		{[]string{PerconaServerFlavor}, InstallDb, "5.7.0", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, SemiSynch, "5.1.72", false},
		{[]string{MySQLFlavor, PerconaServerFlavor, MariaDbFlavor}, SemiSynch, "5.5.40", true},
		{[]string{MySQLFlavor}, MySQLX, "5.5.40", false},
		{[]string{MySQLFlavor}, MySQLX, "5.7.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, MySQLXDefault, "5.7.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, MySQLXDefault, "8.0.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor, MariaDbFlavor}, DynVariables, "5.1.72", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, CrashSafe, "5.1.72", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, CrashSafe, "5.6.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, GTID, "5.6.8", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, GTID, "5.6.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, EnhancedGTID, "5.6.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, EnhancedGTID, "5.7.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, Initialize, "5.6.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, Initialize, "5.7.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, CreateUser, "5.6.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, CreateUser, "5.7.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, SuperReadOnly, "5.6.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, SuperReadOnly, "5.7.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, MultiSource, "5.6.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, MultiSource, "5.7.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, GroupReplication, "5.6.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, GroupReplication, "5.7.40", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, SetPersist, "5.7.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, SetPersist, "8.0.12", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, Roles, "5.7.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, Roles, "8.0.12", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, NativeAuth, "5.7.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, NativeAuth, "8.0.12", true},
		{[]string{MySQLFlavor, PerconaServerFlavor}, DataDict, "5.7.40", false},
		{[]string{MySQLFlavor, PerconaServerFlavor}, DataDict, "8.0.12", true},
		{[]string{"no-such-flavor"}, "no-such-feature", "8.0.22", false},
	}
	testHasCapability(capabilitiesList, t)
}

func testHasCapability(capabilitiesList []TestCapabilities, t *testing.T) {

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

func TestCopyCapabilities(t *testing.T) {

	copiedFeatureList := copyCapabilities(MySQLFlavor, []string{InstallDb, SemiSynch})
	dummyCapabilities := Capabilities{
		Flavor:   "dummy",
		Features: copiedFeatureList,
	}
	AllCapabilities["dummy"] = dummyCapabilities
	var capabilitiesList = []TestCapabilities{
		{[]string{"dummy"}, InstallDb, "5.1.72", true},
		{[]string{"dummy"}, InstallDb, "5.7.72", false},
		{[]string{"dummy"}, InstallDb, "8.0.22", false},
		{[]string{"dummy"}, SemiSynch, "5.0.22", false},
		{[]string{"dummy"}, SemiSynch, "8.0.22", true},
	}
	testHasCapability(capabilitiesList, t)
	delete(AllCapabilities, "dummy")
	for N := 0; N < len(capabilitiesList); N++ {
		capabilitiesList[N].expected = false
	}
	testHasCapability(capabilitiesList, t)
}
