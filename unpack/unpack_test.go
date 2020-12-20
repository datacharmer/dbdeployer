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

package unpack

import (
	"path"
	"strings"
	"testing"
)

func Test_pathDepth(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"empty", args{""}, 0},
		{"~", args{"~"}, 0},
		{"lorem ipsum", args{"lorem ipsum"}, 0},
		{"/", args{"/"}, 1},
		{"./", args{"./"}, 1},
		{"repeat", args{strings.Repeat("/", 10)}, 10},
		{"path_join", args{path.Join("one", "two", "three")}, 2},
		{"strings_join", args{strings.Join([]string{"one", "two", "three"}, "/")}, 2},
		{"////", args{"////"}, 4},
		{"/etc/", args{"/etc/"}, 2},
		{"/etc/something", args{"/etc/something"}, 2},
		{"../etc", args{"../etc"}, 1},
		{"../../../etc", args{"../../../etc"}, 3},
		{"../../../../../etc", args{"../../../../../etc"}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathDepth(tt.args.s); got != tt.want {
				t.Errorf("pathDepth(%s) = %v, want %v", tt.args.s, got, tt.want)
			}
		})
	}
}
