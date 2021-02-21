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

package globals

const (
	// import
	TmplImportMyCnf  = "import_my_cnf"
	TmplImportInitDb = "import_init_db"
	//TmplImportNoOp         = "import_no_op"
	TmplImportUse          = "import_use"
	TmplImportStop         = "import_stop"
	TmplImportLoadGrants   = "import_load_grants"
	TmplImportClear        = "import_clear"
	TmplImportShowRelaylog = "import_show_relaylog"
	TmplImportTestSb       = "import_test_sb"
	TmplImportMetadata     = "import_metadata"
	TmplImportRestart      = "import_restart"
	TmplImportStatus       = "import_status"
	TmplImportSendKill     = "import_send_kill"
	TmplImportShowBinlog   = "import_show_binlog"
	TmplImportShowLog      = "import_show_log"
	TmplImportSbInclude    = "import_sb_include"
	TmplImportStart        = "import_start"
	TmplImportAddOption    = "import_add_option"
	TmplImportMysqlsh      = "import_mysqlsh"

	// multiple
	TmplRestartMulti       = "restart_multi"
	TmplSendKillMulti      = "send_kill_multi"
	TmplReplicateFromMulti = "replicate_from_multi"
	TmplStartMulti         = "start_multi"
	TmplNodeAdmin          = "node_admin"
	TmplStatusMulti        = "status_multi"
	TmplNode               = "node"
	TmplSysbenchMulti      = "sysbench_multi"
	TmplMetadataMulti      = "metadata_multi"
	TmplExecMulti          = "exec_multi"
	TmplUseMultiAdmin      = "use_multi_admin"
	TmplStopMulti          = "stop_multi"
	TmplClearMulti         = "clear_multi"
	TmplTestSbMulti        = "test_sb_multi"
	TmplSysbenchReadyMulti = "sysbench_ready_multi"
	TmplUseMulti           = "use_multi"

	// replication
	TmplExecAllMasters         = "exec_all_masters"
	TmplMultiSourceTest        = "multi_source_test"
	TmplReplReplicateFrom      = "repl_replicate_from"
	TmplStatusAll              = "status_all"
	TmplTestSbAll              = "test_sb_all"
	TmplTestReplication        = "test_replication"
	TmplMultiSource            = "multi_source"
	TmplInitSlaves             = "init_slaves"
	TmplClearAll               = "clear_all"
	TmplSlaveAdmin             = "slave_admin"
	TmplCheckMultiSource       = "check_multi_source"
	TmplReplSysbenchReady      = "repl_sysbench_ready"
	TmplSemiSyncStart          = "semi_sync_start"
	TmplExecAll                = "exec_all"
	TmplUseAllMasters          = "use_all_masters"
	TmplMasterAdmin            = "master_admin"
	TmplSlave                  = "slave"
	TmplReplSysbench           = "repl_sysbench"
	TmplRestartAll             = "restart_all"
	TmplUseAllAdmin            = "use_all_admin"
	TmplUseAllSlaves           = "use_all_slaves"
	TmplSendKillAll            = "send_kill_all"
	TmplMultiSourceUseSlaves   = "multi_source_use_slaves"
	TmplWipeAndRestartAll      = "wipe_and_restart_all"
	TmplUseAll                 = "use_all"
	TmplStopAll                = "stop_all"
	TmplCheckSlaves            = "check_slaves"
	TmplMultiSourceExecSlaves  = "multi_source_exec_slaves"
	TmplStartAll               = "start_all"
	TmplMultiSourceExecMasters = "multi_source_exec_masters"
	TmplMetadataAll            = "metadata_all"
	TmplExecAllSlaves          = "exec_all_slaves"
	TmplMaster                 = "master"
	TmplMultiSourceUseMasters  = "multi_source_use_masters"

	// mock
	TmplNoOpMock       = "no_op_mock"
	TmplMysqldSafeMock = "mysqld_safe_mock"
	TmplTidbMock       = "tidb_mock"

	// single
	TmplShowBinlog            = "show_binlog"
	TmplShowRelaylog          = "show_relaylog"
	TmplCloneFrom             = "clone_from"
	TmplSemisyncSlaveOptions  = "semisync_slave_options"
	TmplGrants8x              = "grants8x"
	TmplExposeDdTables        = "expose_dd_tables"
	TmplUseAdmin              = "use_admin"
	TmplShowLog               = "show_log"
	TmplWipeAndRestart        = "wipe_and_restart"
	TmplCopyright             = "copyright"
	TmplReplCrashSafeOptions  = "repl_crash_safe_options"
	TmplSbLocked              = "sb_locked"
	TmplSbInclude             = "sb_include"
	TmplConnectionInfoSql     = "connection_info_sql"
	TmplMetadata              = "metadata"
	TmplSemisyncMasterOptions = "semisync_master_options"
	TmplGrants57              = "grants57"
	TmplLoadGrants            = "load_grants"
	TmplMy                    = "my"
	TmplAfterStart            = "after_start"
	TmplConnectionInfoJson    = "connection_info_json"
	TmplClear                 = "clear"
	TmplStatus                = "status"
	TmplTaskUserGrants        = "task_user_grants"
	TmplTestSb                = "test_sb"
	TmplMyCnf                 = "my_cnf"
	TmplRestart               = "restart"
	TmplSysbench              = "sysbench"
	TmplSysbenchReady         = "sysbench_ready"
	TmplStop                  = "stop"
	TmplGrants5x              = "grants5x"
	TmplAddOption             = "add_option"
	TmplReplicateFrom         = "replicate_from"
	TmplGtidOptions56         = "gtid_options_56"
	TmplGtidOptions57         = "gtid_options_57"
	TmplCloneConnectionSql    = "clone_connection_sql"
	TmplInitDb                = "init_db"
	TmplUse                   = "use"
	TmplMysqlsh               = "mysqlsh"
	TmplSendKill              = "send_kill"
	TmplConnectionInfoConf    = "connection_info_conf"
	TmplReplicationOptions    = "replication_options"
	TmplStart                 = "start"

	// pxc
	TmplPxcReplication = "pxc_replication"
	TmplPxcStart       = "pxc_start"
	TmplPxcCheckNodes  = "check_pxc_nodes"

	//ndb
	TmplNdbStartCluster = "ndb_start_cluster"
	TmplNdbStopCluster  = "ndb_stop_cluster"
	TmplNdbConfig       = "ndb_config"
	TmplNdbMgm          = "ndb_mgm"
	TmplNdbCheckStatus  = "ndb_check_status"

	// tidb
	TmplTidbStart      = "tidb_start"
	TmplTidbStop       = "tidb_stop"
	TmplTidbSendKill   = "tidb_send_kill"
	TmplTidbGrants5x   = "tidb_grants5x"
	TmplTidbAfterStart = "tidb_after_start"
	TmplTidbInitDb     = "tidb_init_db"
	TmplTidbMyCnf      = "tidb_my_cnf"

	// group
	TmplInitNodes        = "init_nodes"
	TmplCheckNodes       = "check_nodes"
	TmplGroupReplOptions = "group_repl_options"
)
