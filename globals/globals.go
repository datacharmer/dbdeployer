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
var UsingDbDeployer = false

const (
	// Instantiated in cmd/root.go
	ConfigLabel        = "config"
	SandboxBinaryLabel = "sandbox-binary"
	SandboxHomeLabel   = "sandbox-home"

	// Instantiated in cmd/deploy.go
	BasePortLabel           = "base-port"
	BinaryVersionLabel      = "binary-version"
	BindAddressLabel        = "bind-address"
	BindAddressValue        = "127.0.0.1"
	ConcurrentLabel         = "concurrent"
	CustomMysqldLabel       = "custom-mysqld"
	DbPasswordLabel         = "db-password"
	DbPasswordValue         = "msandbox"
	DbUserLabel             = "db-user"
	DbUserValue             = "msandbox"
	DefaultsLabel           = "defaults"
	DisableMysqlXLabel      = "disable-mysqlx"
	EnableGeneralLogLabel   = "enable-general-log"
	EnableMysqlXLabel       = "enable-mysqlx"
	EnableAdminAddressLabel = "enable-admin-address"
	ExposeDdTablesLabel     = "expose-dd-tables"
	ForceLabel              = "force"
	GtidLabel               = "gtid"
	HistoryDirLabel         = "history-dir"
	InitGeneralLogLabel     = "init-general-log"
	InitOptionsLabel        = "init-options"
	KeepServerUuidLabel     = "keep-server-uuid"
	LogLogDirectoryLabel    = "log-directory"
	LogSBOperationsLabel    = "log-sb-operations"
	MyCnfFileLabel          = "my-cnf-file"
	MyCnfOptionsLabel       = "my-cnf-options"
	NativeAuthPluginLabel   = "native-auth-plugin"
	OverwriteLabel          = "overwrite"
	PortLabel               = "port"
	PostGrantsSqlFileLabel  = "post-grants-sql-file"
	PostGrantsSqlLabel      = "post-grants-sql"
	PreGrantsSqlFileLabel   = "pre-grants-sql-file"
	PreGrantsSqlLabel       = "pre-grants-sql"
	RawLabel                = "raw"
	RemoteAccessLabel       = "remote-access"
	RemoteAccessValue       = "127.%"
	ReplCrashSafeLabel      = "repl-crash-safe"
	RplPasswordLabel        = "rpl-password"
	RplPasswordValue        = "rsandbox"
	RplUserLabel            = "rpl-user"
	RplUserValue            = "rsandbox"
	SandboxDirectoryLabel   = "sandbox-directory"
	SkipLoadGrantsLabel     = "skip-load-grants"
	SkipReportHostLabel     = "skip-report-host"
	SkipReportPortLabel     = "skip-report-port"
	SkipStartLabel          = "skip-start"
	UseTemplateLabel        = "use-template"
	ClientFromLabel         = "client-from"
	PromptLabel             = "prompt"
	FlavorInPromptLabel     = "flavor-in-prompt"
	PromptValue             = "mysql"
	SocketInDatadirLabel    = "socket-in-datadir"
	PortAsServerIdLabel     = "port-as-server-id"

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
	NdbNodesLabel       = "ndb-nodes"
	NodesValue          = 3
	NdbNodesValue       = 3
	ReplHistoryDirLabel = "repl-history-dir"
	SemiSyncLabel       = "semi-sync"
	ReadOnlyLabel       = "read-only-slaves"
	SuperReadOnlyLabel  = "super-read-only-slaves"
	SinglePrimaryLabel  = "single-primary"
	SlaveListLabel      = "slave-list"
	SlaveListValue      = "3"
	TopologyLabel       = "topology"
	TopologyValue       = "master-slave"
	PxcLabel            = "pxc"
	NdbLabel            = "ndb"

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
	FlavorLabel        = "flavor"
	FlavorFileName     = "FLAVOR"

	// Instantiated in cmd/delete.go
	SkipConfirmLabel = "skip-confirm"
	ConfirmLabel     = "confirm"

	// Instantiated in cmd/sandboxes.go
	CatalogLabel  = "catalog"
	HeaderLabel   = "header"
	TableLabel    = "table"
	FullInfoLabel = "full-info"

	// Instantiated in cmd/templates.go
	SimpleLabel       = "simple"
	WithContentsLabel = "with-contents"

	// Instantiated in cmd/cookbook.go
	SortByLabel = "sort-by"

	// Instantiated in cmd/versions.go
	ByFlavorLabel = "by-flavor"

	// Instantiated in cmd/export.go

	ForceOutputToTermLabel  = "force-output-to-terminal"
	ExportSandboxDir        = "sandbox-dir"
	ExportVersionDir        = "version-dir"
	ExportTemplateGroup     = "template-group"
	ExportTemplateName      = "template-name"
	ExportString            = "string"
	ExportInteger           = "integer"
	ExportBoolean           = "boolean"
	ExportCookbookName      = "cookbook-name"
	ExportTopology          = "topology"
	ExportAllowedTopologies = "allowed-topologies"

	// Instantiated in sandbox package
	AutoCnfName          = "auto.cnf"
	DataDirName          = "data"
	ScriptAddOption      = "add_option"
	ScriptClear          = "clear"
	ScriptGrantsMysql    = "grants.mysql"
	ScriptInitDb         = "init_db"
	ScriptAfterStart     = "after_start"
	ScriptLoadGrants     = "load_grants"
	ScriptMy             = "my"
	ScriptMySandboxCnf   = "my.sandbox.cnf"
	ScriptMysqlsh        = "mysqlsh"
	ScriptNoClear        = "no_clear"
	ScriptPostGrantsSql  = "post_grants.sql"
	ScriptPreGrantsSql   = "pre_grants.sql"
	ScriptRestart        = "restart"
	ScriptSbInclude      = "sb_include"
	ScriptSendKill       = "send_kill"
	ScriptShowBinlog     = "show_binlog"
	ScriptShowLog        = "show_log"
	ScriptShowRelayLog   = "show_relaylog"
	ScriptStart          = "start"
	ScriptStatus         = "status"
	ScriptStop           = "stop"
	ScriptTestSb         = "test_sb"
	ScriptUse            = "use"
	ScriptUseAdmin       = "use_admin"
	ScriptConnectionConf = "connection.conf"
	ScriptConnectionSql  = "connection.sql"
	ScriptConnectionJson = "connection.json"
	ScriptReplicateFrom  = "replicate_from"
	ScriptMetadata       = "metadata"

	ScriptCheckMsNodes      = "check_ms_nodes"
	ScriptCheckNodes        = "check_nodes"
	ScriptClearAll          = "clear_all"
	ScriptInitializeMsNodes = "initialize_ms_nodes"
	ScriptInitializeNodes   = "initialize_nodes"
	ScriptNoClearAll        = "no_clear_all"
	ScriptRestartAll        = "restart_all"
	ScriptSendKillAll       = "send_kill_all"
	ScriptStartAll          = "start_all"
	ScriptStatusAll         = "status_all"
	ScriptStopAll           = "stop_all"
	ScriptTestReplication   = "test_replication"
	ScriptTestSbAll         = "test_sb_all"
	ScriptUseAll            = "use_all"

	// These constants are kept for reference
	// although they are not used directly in the code.
	// Their value is determined by the variable names
	// for masters and slaves defined in defaults/defaults.go
	// and possibly modified by user options
	ScriptInitializeSlaves = "initialize_slaves"
	ScriptCheckSlaves      = "check_slaves"
	ScriptUseAllMasters    = "use_all_masters"
	ScriptUseAllSlaves     = "use_all_slaves"
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
	ErrFeatureRequiresCapability   = "'%s' requires flavor '%s' version '%s'+"
	ErrArgumentRequired            = "argument required: %s"
	ErrEncodingDefaults            = "error encoding defaults: '%s'"
	ErrCreatingSandbox             = "error creating sandbox: '%s'"
	ErrCreatingDirectory           = "error creating directory '%s': %s"
	ErrRemovingFromCatalog         = "error removing sandbox '%s' from catalog"
	ErrRetrievingSandboxList       = "error retrieving sandbox list: %s"
	ErrWhileComparingVersions      = "error while comparing versions"
)

type NumericVersion []int

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
// The super_read_only flag was introduced in 5.7.8
// Multi source replication was introduced in 5.7.9
// MySQLX (a.k.a. document store) started in 5.7.12
// Group replication was embedded in the server as of 5.7.17
// Roles, persistent variables, and data dictionary were introduced in 8.0
// Authentication plugin changed in 8.0.4
// MySQLX was enabled by default starting with 8.0.11
var (
	MinimumMySQLInstallDb            = NumericVersion{3, 3, 23}
	MaximumMySQLInstallDb            = NumericVersion{5, 6, 999}
	MinimumDynVariablesVersion       = NumericVersion{5, 1, 0}
	MinimumSemiSyncVersion           = NumericVersion{5, 5, 1}
	MinimumCrashSafeVersion          = NumericVersion{5, 6, 2}
	MinimumGtidVersion               = NumericVersion{5, 6, 9}
	MinimumEnhancedGtidVersion       = NumericVersion{5, 7, 0}
	MinimumDefaultInitializeVersion  = NumericVersion{5, 7, 0}
	MinimumCreateUserVersion         = NumericVersion{5, 7, 6}
	MinimumSuperReadOnly             = NumericVersion{5, 7, 8}
	MinimumMultiSourceReplVersion    = NumericVersion{5, 7, 9}
	MinimumMysqlxVersion             = NumericVersion{5, 7, 12}
	MinimumGroupReplVersion          = NumericVersion{5, 7, 17}
	MinimumPersistVersion            = NumericVersion{8, 0, 0}
	MinimumRolesVersion              = NumericVersion{8, 0, 0}
	MinimumDataDictionaryVersion     = NumericVersion{8, 0, 0}
	MinimumNativeAuthPluginVersion   = NumericVersion{8, 0, 4}
	MinimumMysqlxDefaultVersion      = NumericVersion{8, 0, 11}
	MariaDbMinimumGtidVersion        = NumericVersion{10, 0, 0}
	MariaDbMinimumMultiSourceVersion = NumericVersion{10, 0, 0}
	MinimumXtradbClusterVersion      = NumericVersion{5, 7, 14}
	MinimumNdbClusterVersion         = NumericVersion{7, 0, 0}
	MinimumRootAuthVersion           = NumericVersion{10, 4, 3}
	MinimumAdminAddressVersion       = NumericVersion{8, 0, 14}
)

const (
	lineLength             = 80
	PublicDirectoryAttr    = 0755
	ExecutableFileAttr     = 0744
	SandboxDescriptionName = "sbdescription.json"
	ForbiddenDirName       = "lost+found"

	// File names found in tarballs
	FnAriaChk                     = "aria_chk"
	FnGarbd                       = "garbd"
	FnLibGaleraSmmA               = "libgalera_smm.a"
	FnLibGaleraSmmDylib           = "libgalera_smm.dylib"
	FnLibGaleraSmmSo              = "libgalera_smm.so"
	FnLibMariadbA                 = "libmariadb.a"
	FnLibMariadbClientA           = "libmariadbclient.a"
	FnLibMariadbClientDylib       = "libmariadbclient.dylib"
	FnLibMariadbClientSo          = "libmariadbclient.so"
	FnLibMariadbDylib             = "libmariadb.dylib"
	FnLibMySQLClientA             = "libmysqlclient.a"
	FnLibMySQLClientDylib         = "libmysqlclient.dylib"
	FnLibMySQLClientSo            = "libmysqlclient.so"
	FnLibPerconaServerClientA     = "libperconaserverclient.a"
	FnLibPerconaServerClientDylib = "libperconaserverclient.dylib"
	FnLibPerconaServerClientSo    = "libperconaserverclient.so"
	FnMysql                       = "mysql"
	FnMysqlInstallDb              = "mysql_install_db"
	FnMysqlProvisionZip           = "mysqlprovision.zip"
	FnMysqld                      = "mysqld"
	FnMysqldDebug                 = "mysqld-debug"
	FnMysqldSafe                  = "mysqld_safe"
	FnNdbd                        = "ndbd"
	FnNdbdEngineSo                = "ndb_engine.so"
	FnNdbdMgm                     = "ndb_mgm"
	FnNdbdMgmd                    = "ndb_mgmd"
	FnNdbdMtd                     = "ndbmtd"
	FnTableH                      = "table.h"
	FnTiDbServer                  = "tidb-server"
)

var AllowedTopologies = []string{
	MasterSlaveLabel,
	GroupLabel,
	PxcLabel,
	FanInLabel,
	AllMastersLabel,
	NdbLabel,
}

var ExportReferenceData = map[string]interface{}{
	ExportAllowedTopologies: AllowedTopologies,
}

var (
	DashLine     = strings.Repeat("-", lineLength)
	StarLine     = strings.Repeat("*", lineLength)
	HashLine     = strings.Repeat("#", lineLength)
	EmptyString  = ""
	EmptyStrings []string
	EmptyBytes   []byte

	// Executables needed for dbdeployer generated scripts
	NeededExecutables = []string{
		"awk", "bash", "cat", "date", "echo", "grep", "hostname",
		"kill", "ls", "mkdir", "mv", "printf", "rm", "seq", "sh",
		"sleep", "stat", "tail", "test", "[", "touch", "tr", "wc"}

	// Extra executables needed for PXC
	NeededPxcExecutables = []string{"rsync", "lsof"}
)

var Copyright string = `
#    DBDeployer - The MySQL Sandbox
#    Copyright (C) 2006-2019 Giuseppe Maxia
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.
`
