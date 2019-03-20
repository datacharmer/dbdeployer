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
	"bytes"
	"fmt"
	"github.com/datacharmer/dbdeployer/globals"
	"reflect"
	"regexp"
	"text/template"
	"time"
)

// StringMap defines the map of variable types, for brevity
type StringMap map[string]interface{}

// Given a multi-line string, this function removes leading
// spaces from every line.
// It also removes the first line, if it is empty
func TrimmedLines(s string) string {
	// matches the start of the text followed by an EOL
	re := regexp.MustCompile(`(?m)\A\s*$`)
	s = re.ReplaceAllString(s, globals.EmptyString)

	re = regexp.MustCompile(`(?m)^\t\t`)
	s = re.ReplaceAllString(s, globals.EmptyString)
	return s
}

// TemplateFill passed template string is formatted using its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
// Based on code from https://play.golang.org/p/COHKlB2RML
// DEPRECATED: replaced by SafeTemplateFill
func TemplateFill(tmpl string, data StringMap) string {

	// Adds timestamp and version info
	timestamp := time.Now()
	_, timeStampExists := data["DateTime"]
	_, versionExists := data["AppVersion"]
	if !timeStampExists {
		data["DateTime"] = timestamp.Format(time.UnixDate)
	}
	if !versionExists {
		data["AppVersion"] = VersionDef
	}
	// Creates a template
	processTemplate := template.Must(template.New("tmp").Parse(tmpl))
	buf := &bytes.Buffer{}

	// If an error occurs, returns an empty string
	if err := processTemplate.Execute(buf, data); err != nil {
		return globals.EmptyString
	}

	// Returns the populated template
	return buf.String()
}

// Returns true if a StringMap (or an inner StringMap) contains a given key
func hasKey(sm StringMap, wantedKey string) bool {
	for key, value := range sm {
		if key == wantedKey {
			return true
		}
		valType := reflect.TypeOf(value)
		if valType == reflect.TypeOf(StringMap{}) {
			innerSm := value.(StringMap)
			return hasKey(innerSm, wantedKey)
		}
		if valType == reflect.TypeOf([]StringMap{}) {
			innerSm := value.([]StringMap)
			for _, ism := range innerSm {
				if hasKey(ism, wantedKey) {
					return true
				}
			}
		}
	}
	return false
}

// Gets a list of all variables mentioned in a template
func GetVarsFromTemplate(tmpl string) []string {
	var varList []string

	reTemplateVar := regexp.MustCompile(`\{\{\.([^{]+)\}\}`)
	captureList := reTemplateVar.FindAllStringSubmatch(tmpl, -1)
	if len(captureList) > 0 {
		for _, capture := range captureList {
			varList = append(varList, capture[1])
		}
	}
	return varList
}

// SafeTemplateFill passed template string is formatted using its operands and returns the resulting string.
// It checks that the data was safely initialized
func SafeTemplateFill(template_name, tmpl string, data StringMap) (string, error) {

	// Adds timestamp, version info, and empty engine clause if one was not provided
	timestamp := time.Now()
	_, engineClauseExists := data["EngineClause"]
	_, timeStampExists := data["DateTime"]
	_, versionExists := data["AppVersion"]
	if !timeStampExists {
		data["DateTime"] = timestamp.Format(time.UnixDate)
	}
	if !versionExists {
		data["AppVersion"] = VersionDef
	}
	if !engineClauseExists {
		data["EngineClause"] = ""
	}

	// Checks that all data was initialized
	// This check is especially useful when introducing new templates

	/**/
	// First, we get all variables in the pattern {{.VarName}}
	varList := GetVarsFromTemplate(tmpl)
	if len(varList) > 0 {
		for _, capture := range varList {
			// For each variable in the template text, we look whether it is
			// in the map
			if !hasKey(data, capture) {
				//fmt.Printf("### >>> %#v<<<\n", data)
				return globals.EmptyString,
					fmt.Errorf("data field '%s' (intended for template '%s') was not initialized ",
						capture, template_name)
			}
		}
	}
	/**/
	// Creates a template
	processTemplate := template.Must(template.New("tmp").Parse(tmpl))
	buf := &bytes.Buffer{}

	// If an error occurs, returns an empty string
	if err := processTemplate.Execute(buf, data); err != nil {
		return globals.EmptyString, err
	}

	// Returns the populated template
	return buf.String(), nil
}
