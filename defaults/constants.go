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

package defaults

const (
	// Instantiated in cmd/root.go
	ConfigLabel        = "config"
	SandboxBinaryLabel = "sandbox-binary"
	SandboxHomeLabel   = "sandbox-home"

	// Instantiated in cmd/deploy.go
	LogSBOperationsLabel   = "log-sb-operations"
	LogLogDirectoryLabel   = "log-directory"
	DbUserLabel            = "db-user"
	DbUserValue            = "msandbox"
	DbPasswordLabel        = "db-password"
	DbPasswordValue        = "msandbox"
	RplUserLabel           = "rpl-user"
	RplUserValue           = "rsandbox"
	RplPasswordLabel       = "rpl-password"
	RplPasswordValue       = "rsandbox"
	PortLabel              = "port"
	BasePortLabel          = "base-port"
	GtidLabel              = "gtid"
	ReplCrashSafeLabel     = "repl-crash-safe"
	NativeAuthPluginLabel  = "native-auth-plugin"
	KeepServerUuidLabel    = "keep-server-uuid"
	ForceLabel             = "force"
	SkipStartLabel         = "skip-start"
	DisableMysqlXLabel     = "disable-mysqlx"
	EnableMysqlXLabel      = "enable-mysqlx"
	SkipLoadGrantsLabel    = "skip-load-grants"
	SkipReportHostLabel    = "skip-report-host"
	SkipReportPortLabel    = "skip-report-port"
	ExposeDdTablesLabel    = "expose-dd-tables"
	ConcurrentLabel        = "concurrent"
	EnableGeneralLogLabel  = "enable-general-log"
	InitGeneralLogLabel    = "init-general-log"
	RemoteAccessLabel      = "remote-access"
	RemoteAccessValue      = "127.%"
	BindAddressLabel       = "bind-address"
	BindAddressValue       = "127.0.0.1"
	CustomMysqldLabel      = "custom-mysqld"
	BinaryVersionLabel     = "binary-version"
	DefaultsLabel          = "defaults"
	InitOptionsLabel       = "init-options"
	MyCnfOptionsLabel      = "my-cnf-options"
	PreGrantsSqlFileLabel  = "pre-grants-sql-file"
	PreGrantsSqlLabel      = "pre-grants-sql"
	PostGrantsSqlFileLabel = "post-grants-sql-file"
	PostGrantsSqlLabel     = "post-grants-sql"
	MyCnfFileLabel         = "my-cnf-file"
	UseTemplateLabel       = "use-template"
	SandboxDirectoryLabel  = "sandbox-directory"
	HistoryDirLabel        = "history-dir"

	// Instantiated in cmd/single.go
	MasterLabel = "master"

	// Instantiated in cmd/replication.go
	MasterListLabel     = "master-list"
	MasterListValue     = "1,2"
	SlaveListLabel      = "slave-list"
	SlaveListValue      = "3"
	MasterIpLabel       = "master-ip"
	MasterIpValue       = "127.0.0.1"
	TopologyLabel       = "topology"
	TopologyValue       = "master-slave"
	NodesLabel          = "nodes"
	NodesValue          = 3
	SinglePrimaryLabel  = "single-primary"
	SemiSyncLabel       = "semi-sync"
	ReplHistoryDirLabel = "repl-history-dir"
	MasterSlaveLabel    = "master-slave"
	GroupLabel          = "group"
	FanInLabel          = "fan-in"
	AllMastersLabel     = "all-masters"

	// Instantiated in cmd/unpack.go
	VerbosityLabel     = "verbosity"
	UnpackVersionLabel = "unpack-version"
	PrefixLabel        = "prefix"
	ShellLabel         = "shell"
	TargetServerLabel  = "target-server"

	// Instantiated in cmd/delete.go
	SkipConfirmLabel = "skip-confirm"
	ConfirmLabel     = "confirm"

	// Instantiated in cmd/sandboxes.go
	CatalogLabel = "catalog"
	HeaderLabel  = "header"

	// Instantiated in cmd/templates.go
	SimpleLabel       = "simple"
	WithContentsLabel = "with-contents"
)
