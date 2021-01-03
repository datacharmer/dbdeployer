// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/dustin/go-humanize/english"
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
	// common.CondPrintf("# separator: <%s> %#v\n",separator, list)
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
		return []int{}, fmt.Errorf("list '%s' is empty", nodesList)
	}
	if len(intList) > nodes {
		return []int{}, fmt.Errorf("list '%s' is greater than the expected number of nodes (%d)", nodesList, nodes)
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

	sandboxDef.GtidOptions = SingleTemplates[globals.TmplGtidOptions57].Contents
	sandboxDef.ReplCrashSafeOptions = SingleTemplates[globals.TmplReplCrashSafeOptions].Contents
	if sandboxDef.DirName == "" {
		sandboxDef.DirName += defaults.Defaults().AllMastersPrefix + common.VersionToName(origin)
	}
	sandboxDir := sandboxDef.SandboxDir
	sandboxDef.SandboxDir = common.DirName(sandboxDef.SandboxDir)

	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
	rev := vList[0]
	if sandboxDef.BasePort == 0 {
		sandboxDef.BasePort = sandboxDef.Port + defaults.Defaults().AllMastersReplicationBasePort + (rev * 100)
	}
	masterAbbr := defaults.Defaults().MasterAbbr
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterLabel := defaults.Defaults().MasterName
	slaveLabel := defaults.Defaults().SlavePrefix

	readOnlyOptions, err := checkReadOnlyFlags(sandboxDef)
	if err != nil {
		return err
	}
	if readOnlyOptions != "" {
		return fmt.Errorf("options --read-only and --super-read-only can't be used for all-masters topology\n" +
			"as every slave node is also a master")
	}
	data, err := CreateMultipleSandbox(sandboxDef, origin, nodes)
	if err != nil {
		return err
	}

	rawSandboxDir := data["SandboxDir"]
	if rawSandboxDir != nil {
		sandboxDef.SandboxDir = rawSandboxDir.(string)
	} else {
		return fmt.Errorf("empty Sandbox directory received from multiple deployment")
	}

	masterList := makeNodesList(nodes)
	slaveList, err := nodesListToIntSlice(masterList, nodes)
	if err != nil {
		return err
	}

	setGlobal := "GLOBAL"
	// persistent, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumRolesVersion)
	persistent, err := common.HasCapability(sandboxDef.Flavor, common.SetPersist, sandboxDef.Version)
	if err != nil {
		return err
	}
	if persistent {
		setGlobal = "PERSIST"
	}

	data["SlavesReadOnly"] = readOnlyOptions
	data["SetGlobal"] = setGlobal
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
	data["ChangeMasterExtra"] = setChangeMasterProperties("", sandboxDef.ChangeMasterOptions, logger)
	logger.Printf("Writing master and slave scripts in %s\n", sandboxDef.SandboxDir)
	for _, node := range slaveList {
		data["Node"] = node
		err = writeScript(logger, ReplicationTemplates, fmt.Sprintf("s%d", node), globals.TmplSlave, sandboxDir, data, true)
		if err != nil {
			return err
		}
		err = writeScript(logger, ReplicationTemplates, fmt.Sprintf("m%d", node), globals.TmplSlave, sandboxDir, data, true)
		if err != nil {
			return err
		}
	}
	logger.Printf("Writing all-masters replication scripts in %s\n", sandboxDef.SandboxDir)

	slavePlural := english.PluralWord(2, slaveLabel, "")
	masterPlural := english.PluralWord(2, masterLabel, "")
	useAllMasters := "use_all_" + masterPlural
	useAllSlaves := "use_all_" + slavePlural
	execAllSlaves := "exec_all_" + slavePlural
	execAllMasters := "exec_all_" + masterPlural

	sbMulti := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptTestReplication, globals.TmplMultiSourceTest, true},
			{useAllSlaves, globals.TmplMultiSourceUseSlaves, true},
			{useAllMasters, globals.TmplMultiSourceUseMasters, true},
			{execAllMasters, globals.TmplMultiSourceExecMasters, true},
			{execAllSlaves, globals.TmplMultiSourceExecSlaves, true},
			{globals.ScriptCheckMsNodes, globals.TmplCheckMultiSource, true},
			{globals.ScriptInitializeMsNodes, globals.TmplMultiSource, true},
			{globals.ScriptWipeRestartAll, globals.TmplWipeAndRestartAll, true},
		},
	}
	err = writeScripts(sbMulti)
	if err != nil {
		return err
	}
	if !sandboxDef.SkipStart {
		logger.Printf("Initializing all-masters replication \n")
		common.CondPrintln(path.Join(common.ReplaceLiteralHome(sandboxDir), globals.ScriptInitializeMsNodes))
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

func intInSlice(i int, slice []int) bool {
	for _, N := range slice {
		if N == i {
			return true
		}
	}
	return false
}

func setLists(mList, sList []int, mListSet, sListSet bool, nodes int) ([]int, []int, error) {

	// If the count match, there is nothing to do
	if len(mList)+len(sList) == nodes {
		return mList, sList, nil
	}

	// If there are too many roles defined, we return an error
	if len(mList)+len(sList) > nodes {
		return []int{}, []int{}, fmt.Errorf("too many nodes defined in master list + slave list")
	}

	// When the total of defined nodes (possibly the default) is less than the nodes, we try to figure out
	// how to set it up

	// master list was defined. We try to set the slaves with the remaining nodes
	if mListSet {
		if len(mList) >= nodes {
			return []int{}, []int{}, fmt.Errorf("too many masters (%d) defined for %d nodes", len(mList), nodes)
		}
		sList = []int{}
		for N := 1; N <= nodes; N++ {
			if !intInSlice(N, mList) {
				sList = append(sList, N)
			}
		}
	}
	// slave list was defined. We try to set the masters with the remaining nodes
	if sListSet {
		if len(sList) >= nodes {
			return []int{}, []int{}, fmt.Errorf("too many slaves (%d) defined for %d nodes", len(sList), nodes)
		}
		mList = []int{}
		for N := 1; N <= nodes; N++ {
			if !intInSlice(N, sList) {
				mList = append(mList, N)
			}
		}
	}

	return mList, sList, nil
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

	var mlist []int
	var slist []int
	var err error

	if masterList == "" {
		masterList = globals.MasterListValue
	}
	if slaveList == "" {
		slaveList = globals.SlaveListValue
	}
	sandboxDef.GtidOptions = SingleTemplates[globals.TmplGtidOptions57].Contents
	sandboxDef.ReplCrashSafeOptions = SingleTemplates[globals.TmplReplCrashSafeOptions].Contents
	if sandboxDef.DirName == "" {
		sandboxDef.DirName = defaults.Defaults().FanInPrefix + common.VersionToName(origin)
	}
	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
	rev := vList[0]
	if sandboxDef.BasePort == 0 {
		sandboxDef.BasePort = sandboxDef.Port + defaults.Defaults().FanInReplicationBasePort + (rev * 100)
	}
	sandboxDir := sandboxDef.SandboxDir
	sandboxDef.SandboxDir = common.DirName(sandboxDef.SandboxDir)
	mlist, err = nodesListToIntSlice(masterList, nodes)
	if err != nil {
		return err
	}
	slist, err = nodesListToIntSlice(slaveList, nodes)
	if err != nil {
		return err
	}
	if nodes > 3 {
		// No defaults were changed. We can recalculate them, making the slave as the highest
		// node number, and masters! all the remaining ones
		if masterList == globals.MasterListValue && slaveList == globals.SlaveListValue {
			mlist = []int{}
			for N := 1; N <= nodes-1; N++ {
				mlist = append(mlist, N)
			}
			slist = []int{nodes}
		} else {
			mlist, slist, err = setLists(mlist, slist,
				masterList != globals.MasterListValue, slaveList != globals.SlaveListValue,
				nodes)
			if err != nil {
				return err
			}
		}
		masterList = common.IntSliceToSeparatedString(mlist, ",")
		slaveList = common.IntSliceToSeparatedString(slist, ",")
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
		return fmt.Errorf("empty Sandbox directory received from multiple deployment")
	}
	masterAbbr := defaults.Defaults().MasterAbbr
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterLabel := defaults.Defaults().MasterName
	slaveLabel := defaults.Defaults().SlavePrefix

	setGlobal := "GLOBAL"
	// persistent, err := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumRolesVersion)
	persistent, err := common.HasCapability(sandboxDef.Flavor, common.SetPersist, sandboxDef.Version)
	if err != nil {
		return err
	}
	if persistent {
		setGlobal = "PERSIST"
	}

	readOnlyOptions, err := checkReadOnlyFlags(sandboxDef)
	if err != nil {
		return err
	}
	data["SetGlobal"] = setGlobal
	data["SlavesReadOnly"] = readOnlyOptions
	data["MasterList"] = normalizeNodeList(masterList)
	data["SlaveList"] = normalizeNodeList(slaveList)
	data["MasterAbbr"] = masterAbbr
	data["MasterLabel"] = masterLabel
	data["SlaveAbbr"] = slaveAbbr
	data["SlaveLabel"] = slaveLabel
	data["RplUser"] = sandboxDef.RplUser
	data["RplPassword"] = sandboxDef.RplPassword
	data["NodeLabel"] = defaults.Defaults().NodePrefix
	data["ChangeMasterExtra"] = setChangeMasterProperties("", sandboxDef.ChangeMasterOptions, logger)
	data["MasterIp"] = masterIp
	logger.Printf("Writing master and slave scripts in %s\n", sandboxDef.SandboxDir)
	for _, slave := range slist {
		data["Node"] = slave
		err = writeScript(logger, ReplicationTemplates, fmt.Sprintf("s%d", slave), globals.TmplSlave, sandboxDir, data, true)
		if err != nil {
			return err
		}
	}

	slavePlural := english.PluralWord(2, slaveLabel, "")
	masterPlural := english.PluralWord(2, masterLabel, "")
	useAllMasters := "use_all_" + masterPlural
	useAllSlaves := "use_all_" + slavePlural
	execAllSlaves := "exec_all_" + slavePlural
	execAllMasters := "exec_all_" + masterPlural
	sbMulti := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptTestReplication, globals.TmplMultiSourceTest, true},
			{globals.ScriptCheckMsNodes, globals.TmplCheckMultiSource, true},
			{useAllSlaves, globals.TmplMultiSourceUseSlaves, true},
			{useAllMasters, globals.TmplMultiSourceUseMasters, true},
			{execAllMasters, globals.TmplMultiSourceExecMasters, true},
			{execAllSlaves, globals.TmplMultiSourceExecSlaves, true},
			{globals.ScriptInitializeMsNodes, globals.TmplMultiSource, true},
			{globals.ScriptWipeRestartAll, globals.TmplWipeAndRestartAll, true},
		},
	}
	for _, master := range mlist {
		data["Node"] = master
		err = writeScript(logger, ReplicationTemplates, fmt.Sprintf("m%d", master), globals.TmplSlave, sandboxDir, data, true)
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
		common.CondPrintln(path.Join(common.ReplaceLiteralHome(sandboxDir), globals.ScriptInitializeMsNodes))
		_, err = common.RunCmd(path.Join(sandboxDir, globals.ScriptInitializeMsNodes))
		if err != nil {
			return fmt.Errorf("error initializing fan-in sandbox: %s", err)
		}
	}
	return nil
}
