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
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var words = []string{"one", "two", "three", "double words"}

var namedExpression = regexp.MustCompile(`(?P<command>` + strings.Join(words, "|") + `)\s{1}`)

func TestGetRegexNamedGroups(t *testing.T) {
	type args struct {
		text string
		re   *regexp.Regexp
	}
	var reDouble = regexp.MustCompile(`(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)`)
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{"not_named", args{"aaa ", regexp.MustCompile(`(aaa)`)}, nil, true},
		{"one", args{"one ", namedExpression}, map[string]string{"command": "one"}, false},
		{"two", args{"two ", namedExpression}, map[string]string{"command": "two"}, false},
		{"three", args{"three ", namedExpression}, map[string]string{"command": "three"}, false},
		{"double words", args{"double words ", namedExpression}, map[string]string{"command": "double words"}, false},
		{"turing", args{"Alan Turing ", reDouble}, map[string]string{"first": "Alan", "last": "Turing"}, false},
		{"named_empty", args{"  ", reDouble}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRegexNamedGroups(tt.args.text, tt.args.re)
			if err != nil {
				if !tt.wantErr {
					t.Logf("error getting named groups %s", tt.name)
					t.Fail()
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Logf("%#v\n", got)
					t.Errorf("getRegexGroups() = %v, want %v, wantErr %v", got, tt.want, tt.wantErr)
				}
			}
		},
		)
	}
}

func TestGetRegexPositionalGroups(t *testing.T) {
	type args struct {
		text       string
		expression *regexp.Regexp
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{"named", args{"one ", namedExpression}, nil, true},
		{"one_word1", args{"a", regexp.MustCompile(`(\w+)`)}, []string{"a"}, false},
		{"one_word2", args{"aaa", regexp.MustCompile(`(\w+)`)}, []string{"aaa"}, false},
		{"two_words1", args{"aaa bbb ccc", regexp.MustCompile(`(\w+) (\w+)`)}, []string{"aaa", "bbb"}, false},
		{"two_words2", args{"-!/ bbb ccc", regexp.MustCompile(`(\w+) (\w+)`)}, []string{"bbb", "ccc"}, false},
		{"three_words1", args{"aaa bbb ccc", regexp.MustCompile(`(\w+) (\w+) (\w+)`)}, []string{"aaa", "bbb", "ccc"}, false},
		{"three_words2", args{"aaa bbb ccc some extra words", regexp.MustCompile(`(\w+) (\w+) (\w+)`)}, []string{"aaa", "bbb", "ccc"}, false},
		{"one_word_empty", args{".,/ ", regexp.MustCompile(`(\w+)`)}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRegexPositionalGroups(tt.args.text, tt.args.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegexPositionalGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRegexPositionalGroups() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	type args struct {
		s    string
		expr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"word", args{"abc", `\w+`}, true},
		{"two word fail", args{"abc", `\w+ \w+`}, false},
		{"two word", args{"abc def", `\w+ \w+`}, true},
		{"non-empty", args{"abc def", `\S+`}, true},
		{"non-empty-with-boundaries", args{"abc def", `^\S+\s*\S+$`}, true},
		{"non-empty-with-boundaries-fail", args{"abc def", `^\S+$`}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(tt.args.s, tt.args.expr); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeginsWith(t *testing.T) {
	type args struct {
		s    string
		expr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"batman", args{"batman", "bat"}, true},
		{"non-batman", args{"superman", "bat"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BeginsWith(tt.args.s, tt.args.expr); got != tt.want {
				t.Errorf("BeginsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEndsWith(t *testing.T) {
	type args struct {
		s    string
		expr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"batman", args{"batman", "man"}, true},
		{"superman", args{"superman", "man"}, true},
		{"the flash", args{"the flash", "man"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EndsWith(tt.args.s, tt.args.expr); got != tt.want {
				t.Errorf("EndsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
