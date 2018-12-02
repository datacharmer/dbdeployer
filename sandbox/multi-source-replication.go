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

package sandbox

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func checkNodeLists(nodes int, mlist, slist []int) error {
	for _, N := range mlist {
		if N > nodes {
			return fmt.Errorf("master num '%d' greater than number of nodes (%d)", N, nodes)
		}
	}
	for _, N := range slist {
		if N > nodes {
			return fmt.Errorf("slave num '%d' greater than number of nodes (%d)", N, nodes)
		}
	}
	for _, M := range mlist {
		for _, S := range slist {
			if S == M {
				return fmt.Errorf("overlapping values: %d is in both master and slave list", M)
			}
		}
	}
	totalNodes := len(mlist) + len(slist)
	if totalNodes != nodes {
		return fmt.Errorf("mismatched values: masters (%d) + slaves (%d) = %d. Expected: %d", len(mlist), len(slist), totalNodes, nodes)
	}
	return nil
}

func nodesListToIntSlice(nodesList string, nodes int) (intList []int, err error) {
	separator := " "
	if common.Includes(nodesList, ",") {
		separator = ","
	} else if common.Includes(nodesList, ":") {
		separator = ":"
	} else if common.Includes(nodesList, ";") {
		separator = ";"
	} else if common.Includes(nodesList, `\.`) {
		separator = "."
	} else {
		separator = " "
	}
	list := strings.Split(nodesList, separator)
	// fmt.Printf("# separator: <%s> %#v\n",separator, list)
	if len(list) == 0 {
		return []int{}, fmt.Errorf("empty nodes list given (%s)", nodesList)
	}
	for _, s := range list {
		if s != "" {
			num, err := strconv.Atoi(s)
			if err != nil {
				return []int{}, fmt.Errorf("error converting node number '%s' to int: %s", s, err)
			}
			intList = append(intList, num)
		}
	}
	if len(intList) == 0 {
		return []int{}, fmt.Errorf("List '%s' is empty\n", nodesList)
	}
	if len(intList) > nodes {
		return []int{}, fmt.Errorf("List '%s' is greater than the expected number of nodes (%d)\n", nodesList, nodes)
	}
	return intList, nil
}

func makeNodesList(nodes int) (nodesList string) {
	for N := 1; N <= nodes; N++ {
		nodesList += fmt.Sprintf("%d ", N)
	}
	return nodesList
}

func CreateAllMastersReplication(sandboxDef SandboxDef, origin string, nodes int, masterIp string) error {
	sandboxDef.SBType = "all-masters"

	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		var fileName string
		var err error
		logger, fileName, err = defaults.NewLogger(common.LogDirName(), globals.AllMastersLabel)
		if err != nil {
			return err
		}
		sandboxDef.LogFileName = common.ReplaceLiteralHome(fileName)
	}

	sandboxDef.GtidOptions = SingleTemplates["gtid_options_57"].Contents
	sandboxDef.ReplCrashSafeOptions = SingleTemplates["repl_crash_safe_options"].Contents
	if sandboxDef.DirName == "" {
		sandboxDef.DirName += defaults.Defaults().AllMastersPrefix + common.VersionToName(origin)
	}
	sandboxDir := sandboxDef.SandboxDir
	sandboxDef.SandboxDir = common.DirName(sandboxDef.SandboxDir)
	if sandboxDef.BasePort == 0 {
		sandboxDef.BasePort = defaults.Defaults().AllMastersReplicationBasePort
	}
	masterAbbr := defaults.Defaults().MasterAbbr
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterLabel := defaults.Defaults().MasterName
	slaveLabel := defaults.Defaults().SlavePrefix

	data, err := CreateMultipleSandbox(sandboxDef, origin, nodes)
	if err != nil {
		return err
	}

	rawSandboxDir := data["SandboxDir"]
	if rawSandboxDir != nil {
		sandboxDef.SandboxDir = rawSandboxDir.(string)
	} else {
		return fmt.Errorf("Empty Sandbox directory received from multiple deployment")
	}

	masterList := makeNodesList(nodes)
	slaveList, err := nodesListToIntSlice(masterList, nodes)
	if err != nil {
		return err
	}
	data["MasterIp"] = masterIp
	data["MasterAbbr"] = masterAbbr
	data["MasterLabel"] = masterLabel
	data["MasterList"] = normalizeNodeList(masterList)
	data["SlaveAbbr"] = slaveAbbr
	data["SlaveLabel"] = slaveLabel
	data["SlaveList"] = normalizeNodeList(masterList)
	data["RplUser"] = sandboxDef.RplUser
	data["RplPassword"] = sandboxDef.RplPassword
	data["NodeLabel"] = defaults.Defaults().NodePrefix
	logger.Printf("Writing master and slave scripts in %s\n", sandboxDef.SandboxDir)
	for _, node := range slaveList {
		data["Node"] = node
		err = writeScript(logger, ReplicationTemplates, fmt.Sprintf("s%d", node), "slave_template", sandboxDir, data, true)
		if err != nil {
			return err
		}
		err = writeScript(logger, ReplicationTemplates, fmt.Sprintf("m%d", node), "slave_template", sandboxDir, data, true)
		if err != nil {
			return err
		}
	}
	logger.Printf("Writing all-masters replication scripts in %s\n", sandboxDef.SandboxDir)
	sbMulti := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptTestReplication, "multi_source_test_template", true},
			{globals.ScriptUseAllSlaves, "multi_source_use_slaves_template", true},
			{globals.ScriptUseAllMasters, "multi_source_use_masters_template", true},
			{globals.ScriptCheckMsNodes, "check_multi_source_template", true},
			{globals.ScriptInitializeMsNodes, "multi_source_template", true},
		},
	}
	err = writeScripts(sbMulti)
	if err != nil {
		return err
	}
	if !sandboxDef.SkipStart {
		logger.Printf("Initializing all-masters replication \n")
		fmt.Println(path.Join(common.ReplaceLiteralHome(sandboxDir), globals.ScriptInitializeMsNodes))
		_, err = common.RunCmd(path.Join(sandboxDir, globals.ScriptInitializeMsNodes))
		if err != nil {
			return err
		}
	}
	return nil
}

func normalizeNodeList(list string) string {
	re := regexp.MustCompile(`[,:.]`)
	return re.ReplaceAllString(list, " ")
}

func CreateFanInReplication(sandboxDef SandboxDef, origin string, nodes int, masterIp, masterList, slaveList string) error {
	sandboxDef.SBType = "fan-in"

	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		var fileName string
		var err error
		logger, fileName, err = defaults.NewLogger(common.LogDirName(), globals.FanInLabel)
		if err != nil {
			return err
		}
		sandboxDef.LogFileName = common.ReplaceLiteralHome(fileName)
	}

	sandboxDef.GtidOptions = SingleTemplates["gtid_options_57"].Contents
	sandboxDef.ReplCrashSafeOptions = SingleTemplates["repl_crash_safe_options"].Contents
	if sandboxDef.DirName == "" {
		sandboxDef.DirName = defaults.Defaults().FanInPrefix + common.VersionToName(origin)
	}
	if sandboxDef.BasePort == 0 {
		sandboxDef.BasePort = defaults.Defaults().FanInReplicationBasePort
	}
	sandboxDir := sandboxDef.SandboxDir
	sandboxDef.SandboxDir = common.DirName(sandboxDef.SandboxDir)
	mlist, err := nodesListToIntSlice(masterList, nodes)
	if err != nil {
		return err
	}
	slist, err := nodesListToIntSlice(slaveList, nodes)
	if err != nil {
		return err
	}
	err = checkNodeLists(nodes, mlist, slist)
	if err != nil {
		return err
	}
	data, err := CreateMultipleSandbox(sandboxDef, origin, nodes)
	if err != nil {
		return err
	}

	rawSandboxDir := data["SandboxDir"]
	if rawSandboxDir != nil {
		sandboxDef.SandboxDir = rawSandboxDir.(string)
	} else {
		return fmt.Errorf("Empty Sandbox directory received from multiple deployment")
	}
	masterAbbr := defaults.Defaults().MasterAbbr
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterLabel := defaults.Defaults().MasterName
	slaveLabel := defaults.Defaults().SlavePrefix
	data["MasterList"] = normalizeNodeList(masterList)
	data["SlaveList"] = normalizeNodeList(slaveList)
	data["MasterAbbr"] = masterAbbr
	data["MasterLabel"] = masterLabel
	data["SlaveAbbr"] = slaveAbbr
	data["SlaveLabel"] = slaveLabel
	data["RplUser"] = sandboxDef.RplUser
	data["RplPassword"] = sandboxDef.RplPassword
	data["NodeLabel"] = defaults.Defaults().NodePrefix
	data["MasterIp"] = masterIp
	logger.Printf("Writing master and slave scripts in %s\n", sandboxDef.SandboxDir)
	for _, slave := range slist {
		data["Node"] = slave
		err = writeScript(logger, ReplicationTemplates, fmt.Sprintf("s%d", slave), "slave_template", sandboxDir, data, true)
		if err != nil {
			return err
		}
	}
	sbMulti := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptTestReplication, "multi_source_test_template", true},
			{globals.ScriptCheckMsNodes, "check_multi_source_template", true},
			{globals.ScriptUseAllSlaves, "multi_source_use_slaves_template", true},
			{globals.ScriptUseAllMasters, "multi_source_use_masters_template", true},
			{globals.ScriptInitializeMsNodes, "multi_source_template", true},
		},
	}
	for _, master := range mlist {
		data["Node"] = master
		err = writeScript(logger, ReplicationTemplates, fmt.Sprintf("m%d", master), "slave_template", sandboxDir, data, true)
		if err != nil {
			return err
		}
	}
	logger.Printf("writing fan-in replication scripts in %s\n", sandboxDef.SandboxDir)
	err = writeScripts(sbMulti)
	if err != nil {
		return err
	}
	if !sandboxDef.SkipStart {
		logger.Printf("Initializing fan-in replication\n")
		fmt.Println(path.Join(common.ReplaceLiteralHome(sandboxDir), globals.ScriptInitializeMsNodes))
		_, err = common.RunCmd(path.Join(sandboxDir, globals.ScriptInitializeMsNodes))
		if err != nil {
			return fmt.Errorf("error initializing fan-in sandbox: %s", err)
		}
	}
	return nil
}
