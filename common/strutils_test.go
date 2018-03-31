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

package common

import (
	"os"
	"testing"
)

type path_info struct {
	value string
	env_var string
	expected string
}

func TestReplaceLiteralHome(t *testing.T) {
	os.Setenv("HOME", "/home/Groucho")
	os.Setenv("PWD", "/var/lib/MarxBrothers")
	var  paths = []path_info{
		{"/home/Groucho/", "HOME", "$HOME/"},
		{"/home/Groucho/path1/path2", "HOME", "$HOME/path1/path2"},
		{"/home/Harpo/path1/path2", "HOME", "/home/Harpo/path1/path2"},
		{"/var/lib/MarxBrothers/path1/path2", "PWD", "$PWD/path1/path2"},
		{"/var/lib/MarxCousins/path1/path2", "PWD", "/var/lib/MarxCousins/path1/path2"},
	}
	for _, p := range paths {
		value := p.value
		env_var := p.env_var
		expected := p.expected
		canary := ReplaceLiteralEnvVar(value, env_var)
		if expected == canary {
			t.Logf("ok    %-35s %-10s =--> %-25s\n", value, "(" + env_var + ")", expected)
		} else {
			t.Logf("NOT OK %-35s %-10s =--> %-25s\n", value, "(" + env_var + ")", expected)
			t.Fail()
		}
	}
	for _, p := range paths {
		value := p.expected
		env_var := p.env_var
		expected := p.value
		canary := ReplaceEnvVar(value, env_var)
		if expected == canary {
			t.Logf("ok    %-35s %-10s --=> %-25s\n", value, "(" + env_var + ")", expected)
		} else {
			t.Logf("NOT OK %-35s %-10s --=> %-25s\n", value, "(" + env_var + ")", expected)
			t.Fail()
		}
	}
}
