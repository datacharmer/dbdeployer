// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
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

//go:build go1.16
// +build go1.16

package sandbox

import (
	_ "embed"

	"github.com/datacharmer/dbdeployer/globals"
)

// Templates for replication

var (
	//go:embed templates/replication/init_slaves.gotxt
	initSlavesTemplate string

	//go:embed templates/replication/semi_sync_start.gotxt
	semiSyncStartTemplate string

	//go:embed templates/replication/start_all.gotxt
	startAllTemplate string

	//go:embed templates/replication/restart_all.gotxt
	restartAllTemplate string

	//go:embed templates/replication/exec_all.gotxt
	execAllTemplate string

	//go:embed templates/replication/use_all.gotxt
	useAllTemplate string

	//go:embed templates/replication/metadata_all.gotxt
	metadataAllTemplate string

	//go:embed templates/replication/use_all_admin.gotxt
	useAllAdminTemplate string

	//go:embed templates/replication/use_all_slaves.gotxt
	useAllSlavesTemplate string

	//go:embed templates/replication/use_all_masters.gotxt
	useAllMastersTemplate string

	//go:embed templates/replication/exec_all_slaves.gotxt
	execAllSlavesTemplate string

	//go:embed templates/replication/exec_all_masters.gotxt
	execAllMastersTemplate string

	//go:embed templates/replication/wipe_and_restart_all.gotxt
	wipeAndRestartAllTemplate string

	//go:embed templates/replication/stop_all.gotxt
	stopAllTemplate string

	//go:embed templates/replication/send_kill_all.gotxt
	sendKillAllTemplate string

	//go:embed templates/replication/clear_all.gotxt
	clearAllTemplate string

	//go:embed templates/replication/status_all.gotxt
	statusAllTemplate string

	//go:embed templates/replication/test_sb_all.gotxt
	testSbAllTemplate string

	//go:embed templates/replication/check_slaves.gotxt
	checkSlavesTemplate string

	//go:embed templates/replication/master.gotxt
	masterTemplate string

	//go:embed templates/replication/master_admin.gotxt
	masterAdminTemplate string

	//go:embed templates/replication/slave.gotxt
	slaveTemplate string

	//go:embed templates/replication/slave_admin.gotxt
	slaveAdminTemplate string

	//go:embed templates/replication/test_replication.gotxt
	testReplicationTemplate string

	//go:embed templates/replication/multi_source.gotxt
	multiSourceTemplate string

	//go:embed templates/replication/multi_source_use_slaves.gotxt
	multiSourceUseSlavesTemplate string

	//go:embed templates/replication/multi_source_use_masters.gotxt
	multiSourceUseMastersTemplate string

	//go:embed templates/replication/multi_source_exec_slaves.gotxt
	multiSourceExecSlavesTemplate string

	//go:embed templates/replication/multi_source_exec_masters.gotxt
	multiSourceExecMastersTemplate string

	//go:embed templates/replication/check_multi_source.gotxt
	checkMultiSourceTemplate string

	//go:embed templates/replication/multi_source_test.gotxt
	multiSourceTestTemplate string

	//go:embed templates/replication/repl_replicate_from.gotxt
	replicateFromReplTemplate string

	//go:embed templates/replication/repl_sysbench.gotxt
	sysbenchReplTemplate string

	//go:embed templates/replication/repl_sysbench_ready.gotxt
	sysbenchReadyReplTemplate string

	ReplicationTemplates = TemplateCollection{
		globals.TmplInitSlaves: TemplateDesc{
			Description: "Initialize slaves after deployment",
			Notes:       "Can also be run after calling './clear_all'",
			Contents:    initSlavesTemplate,
		},
		globals.TmplSemiSyncStart: TemplateDesc{
			Description: "Starts semi synch replication ",
			Notes:       "",
			Contents:    semiSyncStartTemplate,
		},
		globals.TmplStartAll: TemplateDesc{
			Description: "Starts nodes in replication order (with optional mysqld arguments)",
			Notes:       "",
			Contents:    startAllTemplate,
		},
		globals.TmplRestartAll: TemplateDesc{
			Description: "stops all nodes and restarts them (with optional mysqld arguments)",
			Notes:       "",
			Contents:    restartAllTemplate,
		},
		globals.TmplUseAll: TemplateDesc{
			Description: "Execute a query for all nodes",
			Notes:       "",
			Contents:    useAllTemplate,
		},
		globals.TmplExecAll: TemplateDesc{
			Description: "Execute a command in all nodes",
			Notes:       "",
			Contents:    execAllTemplate,
		},
		globals.TmplMetadataAll: TemplateDesc{
			Description: "Execute a metadata query for all nodes",
			Notes:       "",
			Contents:    metadataAllTemplate,
		},
		globals.TmplUseAllAdmin: TemplateDesc{
			Description: "Execute a query (as admin user) for all nodes",
			Notes:       "",
			Contents:    useAllAdminTemplate,
		},
		globals.TmplUseAllSlaves: TemplateDesc{
			Description: "Execute a query for all slaves",
			Notes:       "master-slave topology",
			Contents:    useAllSlavesTemplate,
		},
		globals.TmplUseAllMasters: TemplateDesc{
			Description: "Execute a query for all masters",
			Notes:       "master-slave topology",
			Contents:    useAllMastersTemplate,
		},
		globals.TmplExecAllSlaves: TemplateDesc{
			Description: "Execute a command in all slave nodes",
			Notes:       "master-slave topology",
			Contents:    execAllSlavesTemplate,
		},
		globals.TmplExecAllMasters: TemplateDesc{
			Description: "Execute a command in all master nodes",
			Notes:       "master-slave topology",
			Contents:    execAllMastersTemplate,
		},
		globals.TmplStopAll: TemplateDesc{
			Description: "Stops all nodes in reverse replication order",
			Notes:       "",
			Contents:    stopAllTemplate,
		},
		globals.TmplSendKillAll: TemplateDesc{
			Description: "Send kill signal to all nodes",
			Notes:       "",
			Contents:    sendKillAllTemplate,
		},
		globals.TmplClearAll: TemplateDesc{
			Description: "Remove data from all nodes",
			Notes:       "",
			Contents:    clearAllTemplate,
		},
		globals.TmplStatusAll: TemplateDesc{
			Description: "Show status of all nodes",
			Notes:       "",
			Contents:    statusAllTemplate,
		},
		globals.TmplTestSbAll: TemplateDesc{
			Description: "Run sb test on all nodes",
			Notes:       "",
			Contents:    testSbAllTemplate,
		},
		globals.TmplTestReplication: TemplateDesc{
			Description: "Tests replication flow",
			Notes:       "",
			Contents:    testReplicationTemplate,
		},
		globals.TmplCheckSlaves: TemplateDesc{
			Description: "Checks replication status in master and slaves",
			Notes:       "",
			Contents:    checkSlavesTemplate,
		},
		globals.TmplMaster: TemplateDesc{
			Description: "Runs the MySQL client for the master",
			Notes:       "",
			Contents:    masterTemplate,
		},
		globals.TmplMasterAdmin: TemplateDesc{
			Description: "Runs the MySQL client for the master as admin user",
			Notes:       "",
			Contents:    masterAdminTemplate,
		},
		globals.TmplSlave: TemplateDesc{
			Description: "Runs the MySQL client for a slave",
			Notes:       "",
			Contents:    slaveTemplate,
		},
		globals.TmplSlaveAdmin: TemplateDesc{
			Description: "Runs the MySQL client for a slave as admin_user",
			Notes:       "",
			Contents:    slaveAdminTemplate,
		},
		globals.TmplMultiSource: TemplateDesc{
			Description: "Initializes nodes for multi-source replication",
			Notes:       "fan-in and all-masters",
			Contents:    multiSourceTemplate,
		},
		globals.TmplMultiSourceUseSlaves: TemplateDesc{
			Description: "Runs a query for all slave nodes",
			Notes:       "group replication and multi-source topologies",
			Contents:    multiSourceUseSlavesTemplate,
		},
		globals.TmplMultiSourceUseMasters: TemplateDesc{
			Description: "Runs a query for all master nodes",
			Notes:       "group replication and multi-source topologies",
			Contents:    multiSourceUseMastersTemplate,
		},
		globals.TmplMultiSourceExecSlaves: TemplateDesc{
			Description: "Runs a command in each slave node",
			Notes:       "group replication and multi-source topologies",
			Contents:    multiSourceExecSlavesTemplate,
		},
		globals.TmplMultiSourceExecMasters: TemplateDesc{
			Description: "Runs a command in each slave node",
			Notes:       "group replication and multi-source topologies",
			Contents:    multiSourceExecMastersTemplate,
		},
		globals.TmplWipeAndRestartAll: TemplateDesc{
			Description: "clears the databases and restarts them all",
			Notes:       "group replication and multi-source topologies",
			Contents:    wipeAndRestartAllTemplate,
		},
		globals.TmplMultiSourceTest: TemplateDesc{
			Description: "Test replication flow for multi-source replication",
			Notes:       "fan-in and all-masters",
			Contents:    multiSourceTestTemplate,
		},
		globals.TmplCheckMultiSource: TemplateDesc{
			Description: "checks replication status for multi-source replication",
			Notes:       "fan-in and all-masters",
			Contents:    checkMultiSourceTemplate,
		},
		globals.TmplReplReplicateFrom: TemplateDesc{
			Description: "use replicate_from script from the master",
			Notes:       "",
			Contents:    replicateFromReplTemplate,
		},
		globals.TmplReplSysbench: TemplateDesc{
			Description: "use sysbench script from the master",
			Notes:       "",
			Contents:    sysbenchReplTemplate,
		},
		globals.TmplReplSysbenchReady: TemplateDesc{
			Description: "use sysbench_ready script from the master",
			Notes:       "",
			Contents:    sysbenchReadyReplTemplate,
		},
	}
)
