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

package main

import (
	"bufio"
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"os"
	"sort"
)

/*
	This utility reads a list of versions (format x.x.xx)
	and sorts them in numerical order, taking into account
	all three fields, making sure that 5.7.9 comes before 5.7.10.
*/
func main() {

	scanner := bufio.NewScanner(os.Stdin)

	type version_list struct {
		text string
		mmr  []int
	}
	var vlist []version_list
	for scanner.Scan() {
		line := scanner.Text()
		vl := common.VersionToList(line)
		rec := version_list{
			text: line,
			mmr:  vl,
		}
		if vl[0] > 0 {
			vlist = append(vlist, rec)
		}
	}

	if err := scanner.Err(); err != nil {
		common.Exitf(1, "error: %s", err)
	}
	sort.Slice(vlist, func(i, j int) bool {
		return vlist[i].mmr[0] < vlist[j].mmr[0] ||
			(vlist[i].mmr[0] == vlist[j].mmr[0] && vlist[i].mmr[1] < vlist[j].mmr[1]) ||
			(vlist[i].mmr[0] == vlist[j].mmr[0] && vlist[i].mmr[1] == vlist[j].mmr[1] && vlist[i].mmr[2] < vlist[j].mmr[2])
	})
	for _, v := range vlist {
		fmt.Printf("%s\n", v.text)
	}
}
