// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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

func Matches(s string, expr string) bool {
	re := regexp.MustCompile(expr)
	return re.MatchString(s)
}

func BeginsWith(s, expr string) bool {
	return Matches(s, `^`+expr)
}

func EndsWith(s, expr string) bool {
	return Matches(s, expr+`$`)
}
