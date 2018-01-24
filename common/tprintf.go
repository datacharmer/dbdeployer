package common

import (
	"bytes"
	"regexp"
	"text/template"
)

// Smap defines the map of variable types, for brevity
type Smap map[string]interface{}

// Given a multi-line string, this function removes leading
// spaces from every line.
// It also removes the first line, if it is empty
func TrimmedLines(s string) string {
	// matches the start of the text followed by an EOL
	re := regexp.MustCompile(`(?m)\A\s*$`)
	s = re.ReplaceAllString(s, "")

	re = regexp.MustCompile(`(?m)^\t\t`)
	s = re.ReplaceAllString(s, "")
	return s
	/*

	// matches the start of every line, followed by any spaces
	re = regexp.MustCompile(`(?m)^\s*`)
	return re.ReplaceAllString(s, "")
	*/
}

// Tprintf passed template string is formatted using its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
// Based on code from https://play.golang.org/p/COHKlB2RML
func Tprintf(tmpl string, data Smap) string {

	// Creates a template
	t := template.Must(template.New("tmp").Parse(tmpl))
	buf := &bytes.Buffer{}

	// If an error occurs, returns an empty string
	if err := t.Execute(buf, data); err != nil {
		return ""
	}

	// Returns the populated template
	return buf.String()
}
