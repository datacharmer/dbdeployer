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

import "strings"

// UsingDbDeployer is changed to true when the "cmd" package is activated,
// meaning that we're using the command line interface of dbdeployer.
// It is used to make decisions whether to write messages to the screen
// when calling sandbox creation functions from other apps.
var UsingDbDeployer = false

const (
	// Sandbox types
	SbTypeSingle         = "single"
	SbTypeMultiple       = "multiple"
	SbTypeSingleImported = "single-imported"

	// Instantiated in cmd/root.go
	ConfigLabel        = "config"
	SandboxBinaryLabel = "sandbox-binary"
	SandboxHomeLabel   = "sandbox-home"
	SkipLibraryCheck   = "skip-library-check"

	// Instantiated in cmd/admin.go
	RemoteLabel              = "remote"
	RemoteUrlLabel           = "remote-url"
	CompletionFileLabel      = "completion-file"
	CompletionFileValue      = "dbdeployer_completion.sh"
	RunItLabel               = "run-it"
	CamelCase                = "camel-case"
	DefaultSandboxExecutable = "default-sandbox-executable"

	// Instantiated in cmd/init.go
	SkipAllDownloadsLabel    = "skip-all-downloads"
	SkipTarballDownloadLabel = "skip-tarball-download"
	SkipShellCompletionLabel = "skip-shell-completion"

	// Instantiated in cmd/deploy.go
	DefaultRoleLabel          = "default-role"
	CustomRoleNameLabel       = "custom-role-name"
	CustomRolePrivilegesLabel = "custom-role-privileges"
	CustomRoleTargetLabel     = "custom-role-target"
	CustomRoleExtraLabel      = "custom-role-extra"
	TaskUserLabel             = "task-user"
	TaskUserRoleLabel         = "task-user-role"
	BasePortLabel             = "base-port"
	BaseServerIdLabel         = "base-server-id"
	BinaryVersionLabel        = "binary-version"
	BindAddressLabel          = "bind-address"
	BindAddressValue          = LocalHostIP
	ConcurrentLabel           = "concurrent"
	CustomMysqldLabel         = "custom-mysqld"
	DbPasswordLabel           = "db-password"
	DbPasswordValue           = "msandbox"
	DbUserLabel               = "db-user"
	DbUserValue               = "msandbox"
	DefaultsLabel             = "defaults"
	DisableMysqlXLabel        = "disable-mysqlx"
	EnableGeneralLogLabel     = "enable-general-log"
	EnableMysqlXLabel         = "enable-mysqlx"
	EnableAdminAddressLabel   = "enable-admin-address"
	ExposeDdTablesLabel       = "expose-dd-tables"
	ForceLabel                = "force"
	GtidLabel                 = "gtid"
	HistoryDirLabel           = "history-dir"
	InitGeneralLogLabel       = "init-general-log"
	InitOptionsLabel          = "init-options"
	KeepServerUuidLabel       = "keep-server-uuid"
	LogLogDirectoryLabel      = "log-directory"
	LogSBOperationsLabel      = "log-sb-operations"
	MyCnfFileLabel            = "my-cnf-file"
	MyCnfOptionsLabel         = "my-cnf-options"
	NativeAuthPluginLabel     = "native-auth-plugin"
	OverwriteLabel            = "overwrite"
	PortLabel                 = "port"
	PostGrantsSqlFileLabel    = "post-grants-sql-file"
	PostGrantsSqlLabel        = "post-grants-sql"
	PreGrantsSqlFileLabel     = "pre-grants-sql-file"
	PreGrantsSqlLabel         = "pre-grants-sql"
	RawLabel                  = "raw"
	RemoteAccessLabel         = "remote-access"
	RemoteAccessValue         = "127.%"
	ReplCrashSafeLabel        = "repl-crash-safe"
	RplPasswordLabel          = "rpl-password"
	RplPasswordValue          = "rsandbox"
	RplUserLabel              = "rpl-user"
	RplUserValue              = "rsandbox"
	SandboxDirectoryLabel     = "sandbox-directory"
	SkipLoadGrantsLabel       = "skip-load-grants"
	SkipReportHostLabel       = "skip-report-host"
	SkipReportPortLabel       = "skip-report-port"
	SkipStartLabel            = "skip-start"
	UseTemplateLabel          = "use-template"
	ClientFromLabel           = "client-from"
	PromptLabel               = "prompt"
	FlavorInPromptLabel       = "flavor-in-prompt"
	PromptValue               = "mysql"
	SocketInDatadirLabel      = "socket-in-datadir"
	PortAsServerIdLabel       = "port-as-server-id"

	// Instantiated in cmd/single.go
	MasterLabel    = "master"
	ServerIdLabel  = "server-id"
	ShellPathLabel = "shell-path"
	ShellPathValue = "/bin/bash"

	// Instantiated in cmd/info.go
	EarliestLabel = "earliest"
	LimitLabel    = "limit"
	StatsLabel    = "stats"

	// Instantiated in cmd/remote.go
	MB                = 1024 * 1024
	TenMB             = MB * 10
	ProgressLabel     = "progress"
	ProgressStepLabel = "progress-step"
	ProgressStepValue = TenMB

	// Instantiated in cmd/downloads.go
	OSLabel                = "OS"
	ShowUrlLabel           = "show-url"
	UrlLabel               = "url"
	QuietLabel             = "quiet"
	GuessLatestLabel       = "guess-latest"
	MinimalLabel           = "minimal"
	NewestLabel            = "newest"
	AddEmptyItemLabel      = "add-empty-item"
	DeleteAfterUnpackLabel = "delete-after-unpack"
	MaxItemsLabel          = "max-items"

	// Instantiated in cmd/admin.go
	VerboseLabel = "verbose"
	DryRunLabel  = "dry-run"

	// Instantiated in cmd/replication.go
	AllMastersLabel     = "all-masters"
	FanInLabel          = "fan-in"
	GroupLabel          = "group"
	MasterIpLabel       = "master-ip"
	MasterIpValue       = LocalHostIP
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
	ChangeMasterOptions = "change-master-options"

	// Instantiated in cmd/unpack.go and unpack/unpack.go
	GzExt              = ".gz"
	PrefixLabel        = "prefix"
	ShellLabel         = "shell"
	TarExt             = ".tar"
	TarGzExt           = ".tar.gz"
	TarXzExt           = ".tar.xz"
	TargetServerLabel  = "target-server"
	TgzExt             = ".tgz"
	ZipExt             = ".zip"
	UnpackVersionLabel = "unpack-version"
	VerbosityLabel     = "verbosity"
	FlavorLabel        = "flavor"
	TypeLabel          = "type"
	NameLabel          = "name"
	PortRangeLabel     = "port-range"
	VersionLabel       = "version"
	ShortVersionLabel  = "short-version"
	FlavorFileName     = "FLAVOR"

	// Instantiated in cmd/use.go
	RunLabel = "run"
	LsLabel  = "ls"

	// Instantiated in cmd/delete.go
	SkipConfirmLabel = "skip-confirm"
	ConfirmLabel     = "confirm"
	UseStopLabel     = "use-stop"

	// Instantiated in cmd/sandboxes.go
	CatalogLabel   = "catalog"
	HeaderLabel    = "header"
	TableLabel     = "table"
	FullInfoLabel  = "full-info"
	ByDateLabel    = "by-date"
	ByVersionLabel = "by-version"
	LatestLabel    = "latest"
	OldestLabel    = "oldest"
	LocalHostIP    = "127.0.0.1"

	// Instantiated in cmd/templates.go
	SimpleLabel       = "simple"
	WithContentsLabel = "with-contents"

	// Instantiated in cmd/cookbook.go
	SortByLabel = "sort-by"

	// Instantiated in cmd/update.go
	NewPathLabel         = "new-path"
	DocsLabel            = "docs"
	ForceOldVersionLabel = "force-old-version"

	// Instantiated in cmd/versions.go
	ByFlavorLabel = "by-flavor"

	// Instantiated in cmd/export.go

	ForceOutputToTermLabel       = "force-output-to-terminal"
	ExportSandboxDir             = "sandbox-dir"
	ExportVersionDir             = "version-dir"
	ExportTemplateGroup          = "template-group"
	ExportTemplateName           = "template-name"
	ExportString                 = "string"
	ExportInteger                = "integer"
	ExportBoolean                = "boolean"
	ExportCookbookName           = "cookbook-name"
	ExportTopology               = "topology"
	ExportAllowedTopologies      = "allowed-topologies"
	ExportSupportedAllVersions   = "supported-all-versions"
	ExportSupportedMysqlVersions = "supported-mysql-versions"

	// Instantiated in sandbox package
	AutoCnfName              = "auto.cnf"
	DataDirName              = "data"
	ScriptAddOption          = "add_option"
	ScriptClear              = "clear"
	ScriptGrantsMysql        = "grants.mysql"
	ScriptInitDb             = "init_db"
	ScriptAfterStart         = "after_start"
	ScriptLoadGrants         = "load_grants"
	ScriptMy                 = "my"
	ScriptMySandboxCnf       = "my.sandbox.cnf"
	ScriptMysqlsh            = "mysqlsh"
	ScriptNoClear            = "no_clear"
	ScriptPostGrantsSql      = "post_grants.sql"
	ScriptPreGrantsSql       = "pre_grants.sql"
	ScriptRestart            = "restart"
	ScriptSbInclude          = "sb_include"
	ScriptSendKill           = "send_kill"
	ScriptShowBinlog         = "show_binlog"
	ScriptShowLog            = "show_log"
	ScriptShowRelayLog       = "show_relaylog"
	ScriptStart              = "start"
	ScriptStatus             = "status"
	ScriptStop               = "stop"
	ScriptTestSb             = "test_sb"
	ScriptUse                = "use"
	ScriptUseAdmin           = "use_admin"
	ScriptConnectionConf     = "connection.conf"
	ScriptConnectionSql      = "connection.sql"
	ScriptConnectionJson     = "connection.json"
	ScriptReplicateFrom      = "replicate_from"
	ScriptCloneConnectionSql = "clone_connection.sql"
	ScriptCloneFrom          = "clone_from"
	ScriptMetadata           = "metadata"
	ScriptSysbench           = "sysbench"
	ScriptSysbenchReady      = "sysbench_ready"
	ScriptWipeAndRestart     = "wipe_and_restart"

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
	ScriptUseAllAdmin       = "use_all_admin"
	ScriptExecAll           = "exec_all"
	ScriptWipeRestartAll    = "wipe_and_restart_all"
	ScriptMetadataAll       = "metadata_all"

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
	MinimumMySQLUpgradeTool                   = NumericVersion{5, 0, 0}
	MaximumMySQLUpgradeTool                   = NumericVersion{8, 0, 15}
	MinimumMySQLUpgradeServer                 = NumericVersion{8, 0, 16}
	MinimumCloneMySQLServer                   = NumericVersion{8, 0, 17}
	MinimumMySQLInstallDb                     = NumericVersion{3, 3, 23}
	MaximumMySQLInstallDb                     = NumericVersion{5, 6, 999}
	MinimumDynVariablesVersion                = NumericVersion{5, 1, 0}
	MinimumSemiSyncVersion                    = NumericVersion{5, 5, 1}
	MinimumCrashSafeVersion                   = NumericVersion{5, 6, 2}
	MinimumGtidVersion                        = NumericVersion{5, 6, 9}
	MinimumEnhancedGtidVersion                = NumericVersion{5, 7, 0}
	MinimumDefaultInitializeVersion           = NumericVersion{5, 7, 0}
	MinimumCreateUserVersion                  = NumericVersion{5, 7, 6}
	MinimumSuperReadOnly                      = NumericVersion{5, 7, 8}
	MinimumMultiSourceReplVersion             = NumericVersion{5, 7, 9}
	MinimumMysqlxVersion                      = NumericVersion{5, 7, 12}
	MinimumGroupReplVersion                   = NumericVersion{5, 7, 17}
	MinimumPersistVersion                     = NumericVersion{8, 0, 0}
	MinimumRolesVersion                       = NumericVersion{8, 0, 0}
	MinimumDataDictionaryVersion              = NumericVersion{8, 0, 0}
	MinimumNativeAuthPluginVersion            = NumericVersion{8, 0, 4}
	MinimumMysqlxDefaultVersion               = NumericVersion{8, 0, 11}
	MariaDbMinimumGtidVersion                 = NumericVersion{10, 0, 0}
	MariaDbMinimumMultiSourceVersion          = NumericVersion{10, 0, 0}
	MinimumXtradbClusterVersion               = NumericVersion{5, 6, 14}
	MinimumXtradbClusterNoSlaveUpdatesVersion = NumericVersion{5, 7, 14}
	MinimumXtradbClusterEncryptCluster        = NumericVersion{5, 7, 14}
	MinimumXtradbClusterRsync                 = NumericVersion{5, 7, 14}
	MaximumXtradbClusterRsync                 = NumericVersion{5, 7, 99}
	MinimumXtradbClusterXtraBackup            = NumericVersion{8, 0, 15}
	MinimumNdbClusterVersion                  = NumericVersion{7, 0, 0}
	MinimumNdbInstallDb                       = NumericVersion{7, 0, 0}
	MaximumNdbInstallDb                       = NumericVersion{7, 4, 99}
	MinimumNdbInitialize                      = NumericVersion{7, 5, 0}
	MinimumRootAuthVersion                    = NumericVersion{10, 4, 3}
	MinimumAdminAddressVersion                = NumericVersion{8, 0, 14}
	MinimumMySQLShellEmbed                    = NumericVersion{8, 0, 4}
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
	FnMysqlsh                     = "mysqlsh"
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

// This structure is not used directly by dbdeployer.
// It is here to be used by third party applications that
// use metadata exported using cmd.Export()
var ExportReferenceData = map[string]interface{}{
	ExportAllowedTopologies:      AllowedTopologies,
	ExportSupportedAllVersions:   SupportedAllVersions,
	ExportSupportedMysqlVersions: SupportedMySQLVersions,
}

var (
	ReservedPorts = []int{
		1186,  // MySQL Cluster
		3306,  // MySQL Server regular port
		5432,  // PostgreSQL default port
		33060, // MySQLX
		33062, // MySQL Server admin port
	}
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

	SupportedMySQLVersions = []string{
		"4.1", "5.0", "5.1", "5.5", "5.6", "5.7", "8.0",
	}
	SupportedAllVersions = []string{
		"4.1", "5.0", "5.1", "5.5", "5.6", "5.7", "8.0",
		"10.0", "10.1", "10.2", "10.3", "10.4", "10.5",
	}
	// Extra executables needed for PXC
	NeededPxcExecutables = []string{"rsync", "lsof", "socat"}
)

var ShellScriptCopyright string = `
#    DBDeployer - The MySQL Sandbox
#    Copyright (C) 2006-2020 Giuseppe Maxia
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
var MockTemplatesFilled = false

type FlagAlias struct {
	Command  string
	FlagName string
	Alias    string
}

var FlagAliases = []FlagAlias{
	{"ANY", ByFlavorLabel, "by-flavour"},
	{"ANY", FlavorLabel, "flavour"},
	{"cookbook.list", SortByLabel, "order-by"},
	{"cookbook.show", RawLabel, "original"},
	{"delete", ConcurrentLabel, "parallel"},
	{"deploy.multiple", ConcurrentLabel, "parallel"},
	{"deploy.multiple", DbPasswordLabel, "sandbox-password"},
	{"deploy.multiple", DbUserLabel, "sandbox-user"},
	{"deploy.replication", ConcurrentLabel, "parallel"},
	{"deploy.replication", DbPasswordLabel, "sandbox-password"},
	{"deploy.replication", DbUserLabel, "sandbox-user"},
	{"deploy.replication", MasterIpLabel, "primary-ip"},
	{"deploy.replication", MasterListLabel, "primary-list"},
	{"deploy.replication", ReadOnlyLabel, "read-only-replicas"},
	{"deploy.replication", SlaveListLabel, "replica-list"},
	{"deploy.replication", SuperReadOnlyLabel, "super-read-only-replicas"},
	{"deploy.single", DbPasswordLabel, "sandbox-password"},
	{"deploy.single", DbUserLabel, "sandbox-user"},
	{"deploy.single", MasterLabel, "primary"},
	{"deploy.single", MasterLabel, "replication-ready"},
	{"deploy.single", PortLabel, "sandbox-port"},
	{"downloads.get-by-version", NewestLabel, "latest"},
	{"info", EarliestLabel, "oldest"},
}

const (
	ErrNoVersionFound = 1
	ErrNoRecipeFound  = 2
	VersionNotFound   = "NOTFOUND"
)
