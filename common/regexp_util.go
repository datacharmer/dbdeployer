// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2022 Giuseppe Maxia
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
	"regexp"
)

// GetRegexNamedGroups returns the strings captured by a regular expression with named groups
// It returns an error when no named groups were in the regular expression
// It returns nil when the expression did not match
func GetRegexNamedGroups(text string, expression *regexp.Regexp) (map[string]string, error) {
	reNamed := regexp.MustCompile(`\(\?P<\w+>`)
	if !reNamed.MatchString(expression.String()) {
		return nil, fmt.Errorf("expression %s does not contain named groups", expression.String())
	}
	if !expression.MatchString(text) {
		return nil, nil
	}
	match := expression.FindStringSubmatch(text)
	result := make(map[string]string)
	for i, name := range expression.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result, nil
}

// GetRegexPositionalGroups returns the strings captured by a regular expression with positional groups
// It returns an error when named groups were in the regular expression
// It returns nil when the expression did not match
func GetRegexPositionalGroups(text string, expression *regexp.Regexp) ([]string, error) {
	reNamed := regexp.MustCompile(`\(\?P<\w+>`)
	if reNamed.MatchString(expression.String()) {
		return nil, fmt.Errorf("expression %s contains named groups", expression.String())
	}
	matchList := expression.FindAllStringSubmatch(text, -1)
	if len(matchList) == 0 || len(matchList[0]) < 2 {
		return nil, nil
	}
	return matchList[0][1:], nil
}

// Matches returns true when a given expression matches the input string
func Matches(s string, expr string) bool {
	re := regexp.MustCompile(expr)
	return re.MatchString(s)
}

// BeginsWith returns true when a given expression is found at the start of the input string
// Unlike `strings.HasPrefix`, the expression can be a regular expression rather than a fixed string
func BeginsWith(s, expr string) bool {
	return Matches(s, `^`+expr)
}

// EndsWith returns true when a given expression is found at the end of the input string
// Unlike `strings.HasSuffix`, the expression can be a regular expression rather than a fixed string
func EndsWith(s, expr string) bool {
	return Matches(s, expr+`$`)
}
