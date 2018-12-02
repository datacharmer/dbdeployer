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
)

/*
	This utility reads a list of versions (format x.x.xx)
	and sorts them in numerical order, taking into account
	all three fields, making sure that 5.7.9 comes before 5.7.10.
*/
func main() {

	scanner := bufio.NewScanner(os.Stdin)

	var verList []string
	for scanner.Scan() {
		line := scanner.Text()
		verList = append(verList, line)
	}

	if err := scanner.Err(); err != nil {
		common.Exitf(1, "error: %s", err)
	}
	verList = common.SortVersions(verList)
	for _, v := range verList {
		fmt.Println(v)
	}
}
