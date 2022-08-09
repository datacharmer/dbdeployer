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

package ts

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/datacharmer/dbdeployer/common"
)

func customConditions(condition string) (bool, error) {
	elements := strings.Split(condition, ":")
	if len(elements) == 0 {
		return false, fmt.Errorf("no condition found")
	}
	name := elements[0]
	switch name {
	// minimum_version_for_gtid checks that the current version supports GTID operations
	// example:
	// [!minimum_version_for_gtid:$db_version:$db_flavor] skip 'minimum requirements for GTID not met'
	case "minimum_version_for_gtid":
		return minimumVersionForFeature(name, common.GTID, elements)

	// minimum_version_for_group checks that the current version supports group replication
	// example:
	// [!minimum_version_for_group:$db_version:$db_flavor] skip 'minimum requirements for group replication not met'
	case "minimum_version_for_group":
		return minimumVersionForFeature(name, common.GroupReplication, elements)

	// minimum_version_for_semisync checks that the current version supports semi-synchronous replication
	// example:
	// [!minimum_version_for_semisync:$db_version:$db_flavor] skip 'minimum requirements for semi-sync not met'
	case "minimum_version_for_semisync":
		return minimumVersionForFeature(name, common.SemiSynch, elements)

	// minimum_version_for_multi_source checks that the current version supports multi-source replication
	// example:
	// [!minimum_version_for_multi_source:$db_version:$db_flavor] skip 'minimum requirements for multi-source not met'
	case "minimum_version_for_multi_source":
		return minimumVersionForFeature(name, common.MultiSource, elements)

	// version_is_at_least checks that the current version is at least equal or greater than a given version
	// example:
	// [version_is_at_least:$db_version:8.0.21] conditional_command
	case "version_is_at_least":
		return twoVersionsCondition(name, elements)

	//  version_is checks that the given version is exactly what is required
	// example:
	// [version_is:$db_version:8.0.21] conditional_command
	case "version_is":
		return twoVersionsCondition(name, elements)

	// exists_within_seconds checks that the given file exists within the request number of seconds
	// example
	// [exists_within_seconds:file_name:3] conditional_command
	case "exists_within_seconds":
		return existsWithinSeconds(elements)
	default:
		return false, fmt.Errorf("unrecognized condition name '%s'", name)
	}
}

func minimumVersionForFeature(name, feature string, elements []string) (bool, error) {
	if len(elements) < 3 {
		return false, fmt.Errorf("condition '%s' requires version and flavor", name)
	}
	version := elements[1]
	flavor := elements[2]
	return common.HasCapability(flavor, feature, version)
}

func twoVersionsCondition(name string, elements []string) (bool, error) {
	if len(elements) < 3 {
		return false, fmt.Errorf("condition '%s' requires current version and version to compare to", name)
	}
	versionList, compareToList, err := twoVersionList(elements[1], elements[2])
	if err != nil {
		return false, fmt.Errorf("condition '%s' requires two valid versions: %s", name, err)
	}

	switch name {
	case "version_is_at_least":
		return common.GreaterOrEqualVersionList(versionList, compareToList)
	case "version_is":
		return elements[1] == elements[2], nil
	default:
		return false, fmt.Errorf("unrecofnized condition name %s", name)
	}
}

func existsWithinSeconds(elements []string) (bool, error) {
	if len(elements) < 3 {
		return false, fmt.Errorf("condition 'exists_within_seconds' requires a file name and the number of seconds")
	}
	fileName := elements[1]
	delay, err := strconv.Atoi(elements[2])
	if err != nil {
		return false, err
	}
	if delay == 0 {
		return common.FileExists(fileName), nil
	}
	elapsed := 0
	for elapsed < delay {
		time.Sleep(time.Second)
		if common.FileExists(fileName) {
			return true, nil
		}
		elapsed++
	}
	fmt.Printf("file %s not found within %d seconds\n", fileName, delay)
	return false, nil
}

func twoVersionList(s1, s2 string) ([]int, []int, error) {
	s1List, err := common.VersionToList(s1)
	if err != nil {
		return nil, nil, fmt.Errorf("string '%s' is not a valid version: %s", s1, err)
	}
	var s2List []int
	if len(s2) > 0 {
		s2List, err = common.VersionToList(s2)
		if err != nil {
			return nil, nil, fmt.Errorf("string '%s' is not a valid version: %s", s2, err)
		}
	}
	return s1List, s2List, nil
}
