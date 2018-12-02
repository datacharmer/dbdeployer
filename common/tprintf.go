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
	"github.com/datacharmer/dbdeployer/globals"
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
