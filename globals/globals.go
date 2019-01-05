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
package globals

import "strings"

// This variable is changed to true when the "cmd" package is activated,
// meaning that we're using the command line interface of dbdeployer.
// It is used to make decisions whether to write messages to the screen
// when calling sandbox creation functions from other apps.
var UsingDbDeployer bool = false

const (
	// Instantiated in cmd/root.go
	ConfigLabel        = "config"
	SandboxBinaryLabel = "sandbox-binary"
	SandboxHomeLabel   = "sandbox-home"

	// Instantiated in cmd/deploy.go
	BasePortLabel          = "base-port"
	BinaryVersionLabel     = "binary-version"
	BindAddressLabel       = "bind-address"
	BindAddressValue       = "127.0.0.1"
	ConcurrentLabel        = "concurrent"
	CustomMysqldLabel      = "custom-mysqld"
	DbPasswordLabel        = "db-password"
	DbPasswordValue        = "msandbox"
	DbUserLabel            = "db-user"
	DbUserValue            = "msandbox"
	DefaultsLabel          = "defaults"
	DisableMysqlXLabel     = "disable-mysqlx"
	EnableGeneralLogLabel  = "enable-general-log"
	EnableMysqlXLabel      = "enable-mysqlx"
	ExposeDdTablesLabel    = "expose-dd-tables"
	ForceLabel             = "force"
	GtidLabel              = "gtid"
	HistoryDirLabel        = "history-dir"
	InitGeneralLogLabel    = "init-general-log"
	InitOptionsLabel       = "init-options"
	KeepServerUuidLabel    = "keep-server-uuid"
	LogLogDirectoryLabel   = "log-directory"
	LogSBOperationsLabel   = "log-sb-operations"
	MyCnfFileLabel         = "my-cnf-file"
	MyCnfOptionsLabel      = "my-cnf-options"
	NativeAuthPluginLabel  = "native-auth-plugin"
	PortLabel              = "port"
	PostGrantsSqlFileLabel = "post-grants-sql-file"
	PostGrantsSqlLabel     = "post-grants-sql"
	PreGrantsSqlFileLabel  = "pre-grants-sql-file"
	PreGrantsSqlLabel      = "pre-grants-sql"
	RemoteAccessLabel      = "remote-access"
	RemoteAccessValue      = "127.%"
	ReplCrashSafeLabel     = "repl-crash-safe"
	RplPasswordLabel       = "rpl-password"
	RplPasswordValue       = "rsandbox"
	RplUserLabel           = "rpl-user"
	RplUserValue           = "rsandbox"
	SandboxDirectoryLabel  = "sandbox-directory"
	SkipLoadGrantsLabel    = "skip-load-grants"
	SkipReportHostLabel    = "skip-report-host"
	SkipReportPortLabel    = "skip-report-port"
	SkipStartLabel         = "skip-start"
	UseTemplateLabel       = "use-template"

	// Instantiated in cmd/single.go
	MasterLabel = "master"

	// Instantiated in cmd/replication.go
	AllMastersLabel     = "all-masters"
	FanInLabel          = "fan-in"
	GroupLabel          = "group"
	MasterIpLabel       = "master-ip"
	MasterIpValue       = "127.0.0.1"
	MasterListLabel     = "master-list"
	MasterListValue     = "1,2"
	MasterSlaveLabel    = "master-slave"
	NodesLabel          = "nodes"
	NodesValue          = 3
	ReplHistoryDirLabel = "repl-history-dir"
	SemiSyncLabel       = "semi-sync"
	SinglePrimaryLabel  = "single-primary"
	SlaveListLabel      = "slave-list"
	SlaveListValue      = "3"
	TopologyLabel       = "topology"
	TopologyValue       = "master-slave"

	// Instantiated in cmd/unpack.go and unpack/unpack.go
	GzExt              = ".gz"
	PrefixLabel        = "prefix"
	ShellLabel         = "shell"
	TarExt             = ".tar"
	TarGzExt           = ".tar.gz"
	TarXzExt           = ".tar.xz"
	TargetServerLabel  = "target-server"
	TgzExt             = ".tgz"
	UnpackVersionLabel = "unpack-version"
	VerbosityLabel     = "verbosity"

	// Instantiated in cmd/delete.go
	SkipConfirmLabel = "skip-confirm"
	ConfirmLabel     = "confirm"

	// Instantiated in cmd/sandboxes.go
	CatalogLabel = "catalog"
	HeaderLabel  = "header"

	// Instantiated in cmd/templates.go
	SimpleLabel       = "simple"
	WithContentsLabel = "with-contents"

	// Instantiated in sandbox package
	AutoCnfName         = "auto.cnf"
	DataDirName         = "data"
	ScriptAddOption     = "add_option"
	ScriptClear         = "clear"
	ScriptGrantsMysql   = "grants.mysql"
	ScriptInitDb        = "init_db"
	ScriptLoadGrants    = "load_grants"
	ScriptMy            = "my"
	ScriptMySandboxCnf  = "my.sandbox.cnf"
	ScriptMysqlsh       = "mysqlsh"
	ScriptNoClear       = "no_clear"
	ScriptPostGrantsSql = "post_grants.sql"
	ScriptPreGrantsSql  = "pre_grants.sql"
	ScriptRestart       = "restart"
	ScriptSbInclude     = "sb_include"
	ScriptSendKill      = "send_kill"
	ScriptShowBinlog    = "show_binlog"
	ScriptShowLog       = "show_log"
	ScriptShowRelayLog  = "show_relaylog"
	ScriptStart         = "start"
	ScriptStatus        = "status"
	ScriptStop          = "stop"
	ScriptTestSb        = "test_sb"
	ScriptUse           = "use"

	ScriptCheckMsNodes      = "check_ms_nodes"
	ScriptCheckNodes        = "check_nodes"
	ScriptCheckSlaves       = "check_slaves"
	ScriptClearAll          = "clear_all"
	ScriptInitializeMsNodes = "initialize_ms_nodes"
	ScriptInitializeNodes   = "initialize_nodes"
	ScriptInitializeSlaves  = "initialize_slaves"
	ScriptNoClearAll        = "no_clear_all"
	ScriptRestartAll        = "restart_all"
	ScriptSendKillAll       = "send_kill_all"
	ScriptStartAll          = "start_all"
	ScriptStatusAll         = "status_all"
	ScriptStopAll           = "stop_all"
	ScriptTestReplication   = "test_replication"
	ScriptTestSbAll         = "test_sb_all"
	ScriptUseAll            = "use_all"
	ScriptUseAllMasters     = "use_all_masters"
	ScriptUseAllSlaves      = "use_all_slaves"
)

// Common error messages
const (
	ErrFileNotFound                = "file '%s' not found"
	ErrGroupNotFound               = "group '%s' not found"
	ErrTemplateNotFound            = "template '%s' not found"
	ErrBaseDirectoryNotFound       = "base directory '%s' not found"
	ErrDirectoryNotFound           = "directory '%s' not found"
	ErrNamedDirectoryNotFound      = "%s directory '%s' not found"
	ErrScriptNotFound              = "script '%s' not found"
	ErrScriptNotFoundInUpper       = "script '%s' not found in '%s'"
	ErrDirectoryNotFoundInUpper    = "directory '%s' not found in '%s'"
	ErrExecutableNotFound          = "executable '%s' not found"
	ErrDirectoryAlreadyExists      = "directory '%s' already exists"
	ErrFileAlreadyExists           = "file '%s' already exists"
	ErrNamedDirectoryAlreadyExists = "%s directory '%s' already exists"
	ErrWhileRemoving               = "error while removing %s\n%s"
	ErrWhileDeletingDir            = "error while deleting directory %s\n%s"
	ErrWhileRenamingScript         = "error while renaming script\n%s"
	ErrWhileStoppingSandbox        = "error while stopping sandbox %s"
	ErrWhileDeletingSandbox        = "error while deleting sandbox %s"
	ErrWhileStartingSandbox        = "error while starting sandbox %s"
	ErrOptionRequiresVersion       = "option '--%s' requires MySQL version '%s'+"
	ErrFeatureRequiresVersion      = "'%s' requires MySQL version '%s'+"
	ErrArgumentRequired            = "argument required: %s"
	ErrEncodingDefaults            = "error encoding defaults: '%s'"
	ErrCreatingSandbox             = "error creating sandbox: '%s'"
	ErrCreatingDirectory           = "error creating directory '%s': %s"
	ErrRemovingFromCatalog         = "error removing sandbox '%s' from catalog"
	ErrRetrievingSandboxList       = "error retrieving sandbox list: %s"
	ErrWhileComparingVersions      = "error while comparing versions"
)

const MaxAllowedPort int = 64000

// Go doesn't allow constants to be compound types. Thus we use variables here.
// Although they can be potentially changed (not that anyone would dare,) they
// are used here for the sake of code readability.
//
// This list of variables represents a mini-history of
// MySQL incompatible changes, from installation standpoint
//
// 5.1 introduced dynamic variables (set @@var_name = "something")
// Semi-sync replication started in MySQL 5.5.1
// Crash safe tables were introduced in 5.6.2
// GTID came in 5.6.9
// Better GTID (with fewer mandatory options) came in 5.7
// mysqld --initialize became the default method in 5.7
// CREATE USER became mandatory in 5.7.6 (before we could use GRANT directly)
// Multi source replication was introduced in 5.7.9
// MySQLX (a.k.a. document store) started in 5.7.12
// Group replication was embedded in the server as of 5.7.17
// Roles and data dictionary were introduced in 8.0
// Authentication plugin changed in 8.0.4
// MySQLX was enabled by default starting with 8.0.11
var (
	MinimumDynVariablesVersion      = []int{5, 1, 0}
	MinimumSemiSyncVersion          = []int{5, 5, 1}
	MinimumCrashSafeVersion         = []int{5, 6, 2}
	MinimumGtidVersion              = []int{5, 6, 9}
	MinimumEnhancedGtidVersion      = []int{5, 7, 0}
	MinimumDefaultInitializeVersion = []int{5, 7, 0}
	MinimumCreateUserVersion        = []int{5, 7, 6}
	MinimumMultiSourceReplVersion   = []int{5, 7, 9}
	MinimumMysqlxVersion            = []int{5, 7, 12}
	MinimumGroupReplVersion         = []int{5, 7, 17}
	MinimumRolesVersion             = []int{8, 0, 0}
	MinimumDataDictionaryVersion    = []int{8, 0, 0}
	MinimumNativeAuthPluginVersion  = []int{8, 0, 4}
	MinimumMysqlxDefaultVersion     = []int{8, 0, 11}
	EmptyString                     = ""
	EmptyStrings                    = []string{}
	EmptyBytes                      = []byte{}
)

const lineLength = 80

var (
	DashLine string = strings.Repeat("-", lineLength)
	StarLine string = strings.Repeat("*", lineLength)
	HashLine string = strings.Repeat("#", lineLength)
)

const PublicDirectoryAttr = 0755
const ExecutableFileAttr = 0744
const SandboxDescriptionName = "sbdescription.json"
