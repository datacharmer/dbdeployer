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
	"path"
	"regexp"
	"strconv"
	"strings"
)

func checkNodeLists(nodes int, mlist, slist []int) {
	for _, N := range mlist {
		if N > nodes {
			common.Exitf(1, "master num '%d' greater than number of nodes (%d)", N, nodes)
		}
	}
	for _, N := range slist {
		if N > nodes {
			common.Exitf(1, "slave num '%d' greater than number of nodes (%d)", N, nodes)
		}
	}
	for _, M := range mlist {
		for _, S := range slist {
			if S == M {
				common.Exitf(1, "overlapping values: %d is in both master and slave list", M)
			}
		}
	}
	totalNodes := len(mlist) + len(slist)
	if totalNodes != nodes {
		common.Exitf(1, "mismatched values: masters (%d) + slaves (%d) = %d. Expected: %d", len(mlist), len(slist), totalNodes, nodes)
	}
}

func nodesListToIntSlice(nodesList string, nodes int) (intList []int) {
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
		common.Exitf(1, "empty nodes list given (%s)", nodesList)
	}
	for _, s := range list {
		if s != "" {
			num, err := strconv.Atoi(s)
			common.ErrCheckExitf(err, 1, "error converting node number '%s' to int", s)
			intList = append(intList, num)
		}
	}
	if len(intList) == 0 {
		fmt.Printf("List '%s' is empty\n", nodesList)
	}
	if len(intList) > nodes {
		fmt.Printf("List '%s' is greater than the expected number of nodes (%d)\n", nodesList, nodes)
	}
	return
}

func makeNodesList(nodes int) (nodesList string) {
	for N := 1; N <= nodes; N++ {
		nodesList += fmt.Sprintf("%d ", N)
	}
	return nodesList
}

func CreateAllMastersReplication(sandboxDef SandboxDef, origin string, nodes int, masterIp string) {
	sandboxDef.SBType = "all-masters"

	fileName, logger := defaults.NewLogger(common.LogDirName(), defaults.AllMastersLabel)
	sandboxDef.LogFileName = common.ReplaceLiteralHome(fileName)
	sandboxDef.Logger = logger
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
	data := CreateMultipleSandbox(sandboxDef, origin, nodes)

	sandboxDef.SandboxDir = data["SandboxDir"].(string)
	masterList := makeNodesList(nodes)
	slist := nodesListToIntSlice(masterList, nodes)
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
	for _, node := range slist {
		data["Node"] = node
		writeScript(logger, ReplicationTemplates, fmt.Sprintf("s%d", node), "slave_template", sandboxDir, data, true)
		writeScript(logger, ReplicationTemplates, fmt.Sprintf("m%d", node), "slave_template", sandboxDir, data, true)
	}
	logger.Printf("Writing all-masters replication scripts in %s\n", sandboxDef.SandboxDir)
	writeScript(logger, ReplicationTemplates, defaults.ScriptTestReplication, "multi_source_test_template", sandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptUseAllSlaves, "multi_source_use_slaves_template", sandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptUseAllMasters, "multi_source_use_masters_template", sandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptCheckMsNodes, "check_multi_source_template", sandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptInitializeMsNodes, "multi_source_template", sandboxDir, data, true)
	if !sandboxDef.SkipStart {
		logger.Printf("Initializing all-masters replication \n")
		fmt.Println(path.Join(common.ReplaceLiteralHome(sandboxDir), defaults.ScriptInitializeMsNodes))
		common.RunCmd(path.Join(sandboxDir, defaults.ScriptInitializeMsNodes))
	}
}

func normalizeNodeList(list string) string {
	re := regexp.MustCompile(`[,:.]`)
	return re.ReplaceAllString(list, " ")
}

func CreateFanInReplication(sandboxDef SandboxDef, origin string, nodes int, masterIp, masterList, slaveList string) {
	sandboxDef.SBType = "fan-in"

	fileName, logger := defaults.NewLogger(common.LogDirName(), defaults.FanInLabel)
	sandboxDef.LogFileName = fileName
	sandboxDef.Logger = logger
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
	mlist := nodesListToIntSlice(masterList, nodes)
	slist := nodesListToIntSlice(slaveList, nodes)
	checkNodeLists(nodes, mlist, slist)
	data := CreateMultipleSandbox(sandboxDef, origin, nodes)

	sandboxDef.SandboxDir = data["SandboxDir"].(string)
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
		writeScript(logger, ReplicationTemplates, fmt.Sprintf("s%d", slave), "slave_template", sandboxDir, data, true)
	}
	for _, master := range mlist {
		data["Node"] = master
		writeScript(logger, ReplicationTemplates, fmt.Sprintf("m%d", master), "slave_template", sandboxDir, data, true)
	}
	logger.Printf("writing fan-in replication scripts in %s\n", sandboxDef.SandboxDir)
	writeScript(logger, ReplicationTemplates, defaults.ScriptTestReplication, "multi_source_test_template", sandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptCheckMsNodes, "check_multi_source_template", sandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptUseAllSlaves, "multi_source_use_slaves_template", sandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptUseAllMasters, "multi_source_use_masters_template", sandboxDir, data, true)
	writeScript(logger, ReplicationTemplates, defaults.ScriptInitializeMsNodes, "multi_source_template", sandboxDir, data, true)
	if !sandboxDef.SkipStart {
		logger.Printf("Initializing fan-in replication\n")
		fmt.Println(path.Join(common.ReplaceLiteralHome(sandboxDir), defaults.ScriptInitializeMsNodes))
		common.RunCmd(path.Join(sandboxDir, defaults.ScriptInitializeMsNodes))
	}
}
