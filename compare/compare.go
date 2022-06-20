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

// Collection of comparison functions used in testing
package compare

import (
	"os"
	"regexp"
	"testing"
)

func SkipOnDemand(envVar string, t *testing.T) {
	if os.Getenv(envVar) != "" {
		t.Skipf("Skipped on user request: environment variable '%s' was set ", envVar)
	}
}

func OkIsNil(label string, val interface{}, t *testing.T) {
	if val == nil {
		t.Logf("ok - %s is nil\n", label)
	} else {
		t.Logf("not ok - %s is NOT nil\n", label)
		t.Fail()
	}
}

func OkIsNotNil(label string, val interface{}, t *testing.T) {
	if val != nil {
		t.Logf("ok - %s is not nil\n", label)
	} else {
		t.Logf("not ok - %s is nil\n", label)
		t.Fail()
	}
}

// Compares two empty interfaces
func OkEqualInterface(label string, a, b interface{}, t *testing.T) {
	if a == b {
		t.Logf("ok - %s: expected: '%v'\n", label, a)
	} else {
		t.Logf("not ok - %s: Numbers are not equal - expected %v, but got %v", label, b, a)
		t.Fail()
	}
}

// Compares two integers
func OkEqualInt(label string, a, b int, t *testing.T) {
	if a == b {
		t.Logf("ok - %s: expected: %d\n", label, a)
	} else {
		t.Logf("not ok - %s: Numbers are not equal - expected %d, but got %d", label, b, a)
		t.Fail()
	}
}

// Compares two strings
func OkEqualString(label, a, b string, t *testing.T) {
	if a == b {
		t.Logf("ok - %s: expected: '%s'\n", label, a)
	} else {
		t.Logf("not ok - %s: Strings are not equal - expected '%s', but got '%s'", label, b, a)
		t.Fail()
	}
}

// Checks whether a string is not empty
func OkNotEmptyString(label, a string, t *testing.T) {
	if a != "" {
		t.Logf("ok - %s: string is not empty as expected\n", label)
	} else {
		t.Logf("not ok - %s: String is empty", label)
		t.Fail()
	}
}

// Compares two booleans
func OkEqualBool(label string, a, b bool, t *testing.T) {
	if a == b {
		t.Logf("ok - %s: expected: %v\n", label, a)
	} else {
		t.Logf("not ok - %s: Values are not the same - expected %v, but got %v", label, b, a)
		t.Fail()
	}
}

// Checks that a string matches a given regular expression
func OkMatchesString(label, val, regex string, t *testing.T) {
	re := regexp.MustCompile(regex)
	if re.MatchString(val) {
		t.Logf("ok - %s: '%s' matches '%s'\n", label, val, regex)
	} else {
		t.Logf("not ok - %s: String '%s'  doesn't match '%s'", label, val, regex)
		t.Fail()
	}
}

// Compares two string slices
func OkEqualStringSlices(t *testing.T, found, expected []string) {
	if len(expected) != len(found) {
		t.Logf("not ok - slice Found has %d elements, while slice Expected has %d\n", len(found), len(expected))
		t.Logf("Found: %v", found)
		t.Logf("Expected: %v", expected)
		t.Fail()
		return
	}
	for N := 0; N < len(found); N++ {
		if found[N] == expected[N] {
			t.Logf("ok - element %d of Found and the same in Expected are equal [%v]\n", N, found[N])
		} else {
			t.Logf("not ok - element %d of Found differs from the corresponding one in Expected. "+
				"Expected '%s' - found: '%s'\n", N, expected[N], found[N])
			t.Fail()
		}
	}
}

// Compares two integer slices
func OkEqualIntSlices(t *testing.T, found, expected []int) {
	if len(expected) != len(found) {
		t.Logf("not ok - slice Found has %d elements, while slice Expected has %d\n", len(found), len(expected))
		t.Logf("Found: %v", found)
		t.Logf("Expected: %v", expected)
		t.Fail()
		return
	}
	for N := 0; N < len(found); N++ {
		if found[N] == expected[N] {
			t.Logf("ok - element %d of Found and the same in Expected are equal [%v]\n", N, found[N])
		} else {
			t.Logf("not ok - element %d of Found differs from the corresponding one in Expected. "+
				"Expected '%d' - found: '%d'\n", N, expected[N], found[N])
			t.Fail()
		}
	}
}

// Compares two byte slices
func OkEqualByteSlices(t *testing.T, found, expected []byte) {
	if len(expected) != len(found) {
		t.Logf("not ok - slice Found has %d elements, while slice Expected has %d\n", len(found), len(expected))
		t.Logf("Found: %v", found)
		t.Logf("Expected: %v", expected)
		t.Fail()
		return
	}
	for N := 0; N < len(found); N++ {
		if found[N] == expected[N] {
			t.Logf("ok - byte %d of Found and the same in Expected are equal [%x]\n", N, found[N])
		} else {
			t.Logf("not ok - byte %d of Found differs from the corresponding one in Expected. "+
				"Expected '%x' - found: '%x'\n", N, expected[N], found[N])
			t.Fail()
		}
	}
}
