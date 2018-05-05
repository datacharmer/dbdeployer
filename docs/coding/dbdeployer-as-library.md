# Using dbdeployer code from other applications

If you want to create a MySQL sandbox from your application, you need to fill in a structure
``sandbox.SandboxDef``, with at least the following fields:

```go    

import "github.com/datacharmer/dbdeployer/sandbox"

var sdef =	sandbox.SandboxDef{
	 Version: "5.7.22",
	 Basedir: os.Getenv("HOME") + "/opt/mysql/5.7.22",
	 SandboxDir: os.Getenv("HOME") + "/sandboxes", 
     DirName: "msb_5_7_22",
	 LoadGrants: true,
	 InstalledPorts: []int{1186, 3306, 33060},
	 Port: 5722,
	 DbUser: "msandbox",
	 RplUser: "rsandbox",
	 DbPassword: "msandbox",
	 RplPassword: "rsandbox",
	 RemoteAccess: "127.%",
	 BindAddress: "127.0.0.1",
}
```

This is the full structure of SandboxDef:

```go
type SandboxDef struct {
	DirName           string    // name of the directory cointaining the sandbox
	SBType            string    // Type of sandbox (single, multiple, replication-node, group-node)
	Multi             bool      // either single or part of a multiple sandbox
	NodeNum           int	    // in multiple sandboxes, which node is this
	Version           string    // MySQL version
	Basedir           string    // Where to get binaries from
	SandboxDir        string    // Target directory for sandboxes
	LoadGrants        bool      // Should we load grants?
	SkipReportHost    bool      // Do not add report-host to my.sandbox.cnf
	SkipReportPort    bool      // Do not add report-port to my.sandbox.cnf
	SkipStart         bool	    // Do not start the server after deployment
	InstalledPorts    []int     // Which ports should be skipped in port assignment for this SB
	Port              int	    // port assigned to this sandbox
	MysqlXPort        int	    // XPlugin port for thsi sandbox
	UserPort          int	    // 
	BasePort          int       // Base port for calculating more ports in multiple SB
	MorePorts         []int     // Additional ports that belong to thos sandbox
	Prompt            string    // Prompt to use in "mysql" client
	DbUser            string    // Database user name
	RplUser           string    // Replication user name
	DbPassword        string    // Database password
	RplPassword       string    // Replication password
	RemoteAccess      string    // What access have the users created for this SB (127.%)
	BindAddress       string    // Bind address for this sandbox (127.0.0.1)
	CustomMysqld      string    // Use an alternative mysqld executable
	ServerId          int       // Server ID (for replication)
	ReplOptions       string    // Replication options, as string to append to my.sandbox.cnf
	GtidOptions       string    // Options needed for GTID
	SemiSyncOptions   string    // Options for semi-synchronous replication
	InitOptions       []string  // Options to be added to the initialization command
	MyCnfOptions      []string	// Options to be added to my.sandbox.cnf
	PreGrantsSql      []string	// SQL statements to execute before grants assignment
	PreGrantsSqlFile  string    // SQL file to load before grants assignment
	PostGrantsSql     []string  // SQL statements to run after grants assignment
	PostGrantsSqlFile string    // SQL file to load after grants assignment
	MyCnfFile         string    // options file to merge with the SB my.sandbox.cnf
	InitGeneralLog    bool      // enable general log during server initialization
	EnableGeneralLog  bool		// enable general log after initialization
	NativeAuthPlugin  bool	    // Use the native password plugin for MySQL 8.0.4+
	DisableMysqlX     bool		// Disable Xplugin (MySQL 8.0.11+)
	EnableMysqlX      bool		// Enable Xplugin (MySQL 5.7.12+)
	KeepUuid          bool		// Do not change UUID
	SinglePrimary     bool		// Use single primary for group replication
	Force             bool		// Overwrite an existing sandbox with same target
	ExposeDdTables    bool		// Show hidden data dictionary tables (MySQL 8.0.0+)
	RunConcurrently   bool		// Run multiple sandbox creation concurrently
}
```

Then you can call the function ``sandbox.CreateSingleSandbox(sdef)``.

This will create a fully functional single sandbox that you can then use like any other created by dbdeployer.

To remove a sandbox, you need two steps:

``` go
	sandbox.RemoveSandbox(sandbox_home, "msb_5_7_22", false)
	defaults.DeleteFromCatalog(sandbox_home+"/msb_5_7_22")
```

See the sample source file ``minimal-sandbox.go`` for a working example.

If you want to create multiple sandboxes, things are a bit more complicated. ``dbdeployer`` uses a concurrent execution engine that needs to be used with care.

Look at the invocation of the replication command in ``cmd/replication.go`` for an example of how to prepare the SandboxDef structure before calling the relevant function.

