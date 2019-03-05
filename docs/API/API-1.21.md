This is the list of commands and modifiers available for
dbdeployer 1.21.0 as of 05-Mar-2019 20:29 UTC

# main
    $ dbdeployer -h 
    dbdeployer makes MySQL server installation an easy task.
    Runs single, multiple, and replicated sandboxes.
    
    Usage:
      dbdeployer [command]
    
    Available Commands:
      admin           sandbox management tasks
      defaults        tasks related to dbdeployer defaults
      delete          delete an installed sandbox
      delete-binaries delete an expanded tarball
      deploy          deploy sandboxes
      global          Runs a given command in every sandbox
      help            Help about any command
      remote          Manages remote tarballs
      sandboxes       List installed sandboxes
      unpack          unpack a tarball into the binary directory
      usage           Shows usage of installed sandboxes
      versions        List available versions
    
    Flags:
          --config string           configuration file (default "$HOME/.dbdeployer/config.json")
      -h, --help                    help for dbdeployer
          --sandbox-binary string   Binary repository (default "$HOME/opt/mysql")
          --sandbox-home string     Sandbox deployment directory (default "$HOME/sandboxes")
          --version                 version for dbdeployer
    
    Use "dbdeployer [command] --help" for more information about a command.
    

    $ dbdeployer-docs tree 
    - admin               
        - capabilities        
        - lock                
        - unlock              
        - upgrade             
    - defaults            
        - export              
        - load                
        - reset               
        - show                
        - store               
        - templates           
            - describe            
            - export              
            - import              
            - list                
            - reset               
            - show                
        - update              
    - delete              
    - delete-binaries     
    - deploy              
        - multiple            
        - replication         
        - single              
    - global              
        - restart             
        - start               
        - status              
        - stop                
        - test                
        - test-replication    
        - use                 
    - remote              
        - download            
        - list                
    - sandboxes           
    - unpack              
    - usage               
    - versions            
    

## admin
    $ dbdeployer admin -h
    Runs commands related to the administration of sandboxes.
    
    Usage:
      dbdeployer admin [command]
    
    Aliases:
      admin, manage
    
    Available Commands:
      capabilities Shows capabilities of a given flavor [and optionally version]
      lock         Locks a sandbox, preventing deletion
      unlock       Unlocks a sandbox
      upgrade      Upgrades a sandbox to a newer version
    
    Flags:
      -h, --help   help for admin
    
    
    $ dbdeployer admin capabilities -h
    Shows the capabilities of all flavors. 
    If a flavor is specified, only the capabilities of that flavor are shown.
    If also a version is specified, we show what that version supports
    
    Usage:
      dbdeployer admin capabilities [flavor [version]] [flags]
    
    Examples:
    dbdeployer admin capabilities
    dbdeployer admin capabilities mysql
    dbdeployer admin capabilities mysql 5.7.11
    dbdeployer admin capabilities mysql 5.7.13
    
    
    Flags:
      -h, --help   help for capabilities
    
    
    $ dbdeployer admin lock -h
    Prevents deletion for a given sandbox.
    Note that the deletion being prevented is only the one occurring through dbdeployer. 
    Users can still delete locked sandboxes manually.
    
    Usage:
      dbdeployer admin lock sandbox_name [flags]
    
    Aliases:
      lock, preserve
    
    Flags:
      -h, --help   help for lock
    
    
    $ dbdeployer admin unlock -h
    Removes lock, allowing deletion of a given sandbox
    
    Usage:
      dbdeployer admin unlock sandbox_name [flags]
    
    Aliases:
      unlock, unpreserve
    
    Flags:
      -h, --help   help for unlock
    
    
    $ dbdeployer admin upgrade -h
    Upgrades a sandbox to a newer version.
    The sandbox with the new version must exist already.
    The data directory of the old sandbox will be moved to the new one.
    
    Usage:
      dbdeployer admin upgrade sandbox_name newer_sandbox [flags]
    
    Examples:
    dbdeployer admin upgrade msb_8_0_11 msb_8_0_12
    
    Flags:
      -h, --help   help for upgrade
    
    

## defaults
    $ dbdeployer defaults -h
    Runs commands related to the administration of dbdeployer,
    such as showing the defaults and saving new ones.
    
    Usage:
      dbdeployer defaults [command]
    
    Aliases:
      defaults, config
    
    Available Commands:
      export      Export current defaults to a given file
      load        Load defaults from file
      reset       Remove current defaults file
      show        shows defaults
      store       Store current defaults
      templates   Templates management
      update      Load defaults from file
    
    Flags:
      -h, --help   help for defaults
    
    
    $ dbdeployer defaults export -h
    Saves current defaults to a user-defined file
    
    Usage:
      dbdeployer defaults export filename [flags]
    
    Flags:
      -h, --help   help for export
    
    
    $ dbdeployer defaults load -h
    Reads defaults from file and saves them to dbdeployer configuration file ($HOME/.dbdeployer/config.json)
    
    Usage:
      dbdeployer defaults load file_name [flags]
    
    Aliases:
      load, import
    
    Flags:
      -h, --help   help for load
    
    
    $ dbdeployer defaults reset -h
    Removes current dbdeployer configuration file ($HOME/.dbdeployer/config.json)
    Afterwards, dbdeployer will use the internally stored defaults.
    
    Usage:
      dbdeployer defaults reset [flags]
    
    Aliases:
      reset, remove
    
    Flags:
      -h, --help   help for reset
    
    
    $ dbdeployer defaults show -h
    Shows currently defined defaults
    
    Usage:
      dbdeployer defaults show [flags]
    
    Aliases:
      show, list
    
    Flags:
      -h, --help   help for show
    
    
    $ dbdeployer defaults store -h
    Saves current defaults to dbdeployer configuration file ($HOME/.dbdeployer/config.json)
    
    Usage:
      dbdeployer defaults store [flags]
    
    Flags:
      -h, --help   help for store
    
    

## defaults templates
    $ dbdeployer defaults templates -h
    The commands in this section show the templates used
    to create and manipulate sandboxes.
    
    Usage:
      dbdeployer defaults templates [command]
    
    Aliases:
      templates, template, tmpl, templ
    
    Available Commands:
      describe    Describe a given template
      export      Exports templates to a directory
      import      imports templates from a directory
      list        list available templates
      reset       Removes all template files
      show        Show a given template
    
    Flags:
      -h, --help   help for templates
    
    
    $ dbdeployer defaults templates describe -h
    Describe a given template
    
    Usage:
      dbdeployer defaults templates describe template_name [flags]
    
    Aliases:
      describe, descr, structure, struct
    
    Flags:
      -h, --help            help for describe
          --with-contents   Shows complete structure and contents
    
    
    $ dbdeployer defaults templates export -h
    Exports a group of templates (or "ALL") to a given directory
    
    Usage:
      dbdeployer defaults templates export group_name directory_name [template_name] [flags]
    
    Flags:
      -h, --help   help for export
    
    
    $ dbdeployer defaults templates import -h
    Imports a group of templates (or "ALL") from a given directory
    
    Usage:
      dbdeployer defaults templates import group_name directory_name [template_name] [flags]
    
    Flags:
      -h, --help   help for import
    
    
    $ dbdeployer defaults templates list -h
    list available templates
    
    Usage:
      dbdeployer defaults templates list [group] [flags]
    
    Flags:
      -h, --help     help for list
      -s, --simple   Shows only the template names, without description
    
    
    $ dbdeployer defaults templates reset -h
    Removes all template files that were imported and starts using internal values.
    
    Usage:
      dbdeployer defaults templates reset [flags]
    
    Aliases:
      reset, remove
    
    Flags:
      -h, --help   help for reset
    
    
    $ dbdeployer defaults templates show -h
    Show a given template
    
    Usage:
      dbdeployer defaults templates show template_name [flags]
    
    Flags:
      -h, --help   help for show
    
    
    $ dbdeployer defaults update -h
    Updates one field of the defaults. Stores the result in the dbdeployer configuration file.
    Use "dbdeployer defaults show" to see which values are available
    
    Usage:
      dbdeployer defaults update label value [flags]
    
    Examples:
    
    	$ dbdeployer defaults update master-slave-base-port 17500		
    
    
    Flags:
      -h, --help   help for update
    
    

## delete
    $ dbdeployer delete -h
    Stops the sandbox (and its depending sandboxes, if any), and removes it.
    Warning: this command is irreversible!
    
    Usage:
      dbdeployer delete sandbox_name (or "ALL") [flags]
    
    Aliases:
      delete, remove, destroy
    
    Examples:
    
    	$ dbdeployer delete msb_8_0_4
    	$ dbdeployer delete rsandbox_5_7_21
    
    Flags:
          --concurrent     Runs multiple deletion tasks concurrently.
          --confirm        Requires confirmation.
      -h, --help           help for delete
          --skip-confirm   Skips confirmation with multiple deletions.
    
    

## delete-binaries
    $ dbdeployer delete-binaries -h
    Removes the given directory and all its subdirectories.
    It will fail if the directory is still used by any sandbox.
    Warning: this command is irreversible!
    
    Usage:
      dbdeployer delete-binaries binaries_dir_name [flags]
    
    Examples:
    
    	$ dbdeployer delete-binaries 8.0.4
    	$ dbdeployer delete ps5.7.25
    
    Flags:
      -h, --help           help for delete-binaries
          --skip-confirm   Skips confirmation.
    
    

## deploy
    $ dbdeployer deploy -h
    Deploys single, multiple, or replicated sandboxes
    
    Usage:
      dbdeployer deploy [command]
    
    Available Commands:
      multiple    create multiple sandbox
      replication create replication sandbox
      single      deploys a single sandbox
    
    Flags:
          --base-port int                 Overrides default base-port (for multiple sandboxes)
          --binary-version string         Specifies the version when the basedir directory name does not contain it (i.e. it is not x.x.xx)
          --bind-address string           defines the database bind-address  (default "127.0.0.1")
          --client-from string            Where to get the client binaries from
          --concurrent                    Runs multiple sandbox deployments concurrently
          --custom-mysqld string          Uses an alternative mysqld (must be in the same directory as regular mysqld)
      -p, --db-password string            database password (default "msandbox")
      -u, --db-user string                database user (default "msandbox")
          --defaults strings              Change defaults on-the-fly (--defaults=label:value)
          --disable-mysqlx                Disable MySQLX plugin (8.0.11+)
          --enable-general-log            Enables general log for the sandbox (MySQL 5.1+)
          --enable-mysqlx                 Enables MySQLX plugin (5.7.12+)
          --expose-dd-tables              In MySQL 8.0+ shows data dictionary tables
          --flavor string                 Defines the tarball flavor (MySQL, NDB, Percona Server, etc)
          --flavor-in-prompt              Add flavor values to prompt
          --force                         If a destination sandbox already exists, it will be overwritten
          --gtid                          enables GTID
      -h, --help                          help for deploy
          --history-dir string            Where to store mysql client history (default: in sandbox directory)
          --init-general-log              uses general log during initialization (MySQL 5.1+)
      -i, --init-options strings          mysqld options to run during initialization
          --keep-server-uuid              Does not change the server UUID
          --log-directory string          Where to store dbdeployer logs (default "$HOME/sandboxes/logs")
          --log-sb-operations             Logs sandbox operations to a file
          --my-cnf-file string            Alternative source file for my.sandbox.cnf
      -c, --my-cnf-options strings        mysqld options to add to my.sandbox.cnf
          --native-auth-plugin            in 8.0.4+, uses the native password auth plugin
          --port int                      Overrides default port
          --post-grants-sql strings       SQL queries to run after loading grants
          --post-grants-sql-file string   SQL file to run after loading grants
          --pre-grants-sql strings        SQL queries to run before loading grants
          --pre-grants-sql-file string    SQL file to run before loading grants
          --remote-access string          defines the database access  (default "127.%")
          --repl-crash-safe               enables Replication crash safe
          --rpl-password string           replication password (default "rsandbox")
          --rpl-user string               replication user (default "rsandbox")
          --sandbox-directory string      Changes the default sandbox directory
          --skip-load-grants              Does not load the grants
          --skip-report-host              Does not include report host in my.sandbox.cnf
          --skip-report-port              Does not include report port in my.sandbox.cnf
          --skip-start                    Does not start the database server
          --socket-in-datadir             Create socket in datadir instead of $TMPDIR
          --use-template strings          [template_name:file_name] Replace existing template with one from file
    
    
    $ dbdeployer deploy multiple -h
    Creates several sandboxes of the same version,
    without any replication relationship.
    For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
    the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
    Use the "unpack" command to get the tarball into the right directory.
    
    Usage:
      dbdeployer deploy multiple MySQL-Version [flags]
    
    Examples:
    
    	$ dbdeployer deploy multiple 5.7.21
    	
    
    Flags:
      -h, --help        help for multiple
      -n, --nodes int   How many nodes will be installed (default 3)
    
    
    $ dbdeployer deploy replication -h
    The replication command allows you to deploy several nodes in replication.
    Allowed topologies are "master-slave" for all versions, and  "group", "all-masters", "fan-in"
    for  5.7.17+.
    Topology "pcx" is available for binaries of type Percona Xtradb Cluster.
    For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
    the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
    Use the "unpack" command to get the tarball into the right directory.
    
    Usage:
      dbdeployer deploy replication MySQL-Version [flags]
    
    Examples:
    
    		$ dbdeployer deploy replication 5.7    # deploys highest revision for 5.7
    		$ dbdeployer deploy replication 5.7.21 # deploys a specific revision
    		$ dbdeployer deploy replication /path/to/5.7.21 # deploys a specific revision in a given path
    		# (implies topology = master-slave)
    
    		$ dbdeployer deploy --topology=master-slave replication 5.7
    		# (explicitly setting topology)
    
    		$ dbdeployer deploy --topology=group replication 5.7
    		$ dbdeployer deploy --topology=group replication 8.0 --single-primary
    		$ dbdeployer deploy --topology=all-masters replication 5.7
    		$ dbdeployer deploy --topology=fan-in replication 5.7
    		$ dbdeployer deploy --topology=pxc replication pxc5.7.25
    	
    
    Flags:
      -h, --help                     help for replication
          --master-ip string         Which IP the slaves will connect to (default "127.0.0.1")
          --master-list string       Which nodes are masters in a multi-source deployment (default "1,2")
      -n, --nodes int                How many nodes will be installed (default 3)
          --read-only-slaves         Set read-only for slaves
          --repl-history-dir         uses the replication directory to store mysql client history
          --semi-sync                Use semi-synchronous plugin
          --single-primary           Using single primary for group replication
          --slave-list string        Which nodes are slaves in a multi-source deployment (default "3")
          --super-read-only-slaves   Set super-read-only for slaves
      -t, --topology string          Which topology will be installed (default "master-slave")
    
    
    $ dbdeployer deploy single -h
    single installs a sandbox and creates useful scripts for its use.
    MySQL-Version is in the format x.x.xx, and it refers to a directory named after the version
    containing an unpacked tarball. The place where these directories are found is defined by 
    --sandbox-binary (default: $HOME/opt/mysql.)
    For example:
    	dbdeployer deploy single 5.7     # deploys the latest release of 5.7.x
    	dbdeployer deploy single 5.7.21  # deploys a specific release
    	dbdeployer deploy single /path/to/5.7.21  # deploys a specific release in a given path
    
    For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
    the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
    Use the "unpack" command to get the tarball into the right directory.
    
    Usage:
      dbdeployer deploy single MySQL-Version [flags]
    
    Flags:
      -h, --help            help for single
          --master          Make the server replication ready
          --prompt string   Default prompt for the single client (default "mysql")
    
    

## global
    $ dbdeployer global -h
    This command can propagate the given action through all sandboxes.
    
    Usage:
      dbdeployer global [command]
    
    Examples:
    
    	$ dbdeployer global use "select version()"
    	$ dbdeployer global status
    	$ dbdeployer global stop
    	
    
    Available Commands:
      restart          Restarts all sandboxes
      start            Starts all sandboxes
      status           Shows the status in all sandboxes
      stop             Stops all sandboxes
      test             Tests all sandboxes
      test-replication Tests replication in all sandboxes
      use              Runs a query in all sandboxes
    
    Flags:
      -h, --help   help for global
    
    
    $ dbdeployer global restart -h
    Restarts all sandboxes
    
    Usage:
      dbdeployer global restart [options] [flags]
    
    Flags:
      -h, --help   help for restart
    
    
    $ dbdeployer global start -h
    Starts all sandboxes
    
    Usage:
      dbdeployer global start [options] [flags]
    
    Flags:
      -h, --help   help for start
    
    
    $ dbdeployer global status -h
    Shows the status in all sandboxes
    
    Usage:
      dbdeployer global status [flags]
    
    Flags:
      -h, --help   help for status
    
    
    $ dbdeployer global stop -h
    Stops all sandboxes
    
    Usage:
      dbdeployer global stop [flags]
    
    Flags:
      -h, --help   help for stop
    
    
    $ dbdeployer global test -h
    Tests all sandboxes
    
    Usage:
      dbdeployer global test [flags]
    
    Aliases:
      test, test_sb, test-sb
    
    Flags:
      -h, --help   help for test
    
    
    $ dbdeployer global test-replication -h
    Tests replication in all sandboxes
    
    Usage:
      dbdeployer global test-replication [flags]
    
    Aliases:
      test-replication, test_replication
    
    Flags:
      -h, --help   help for test-replication
    
    
    $ dbdeployer global use -h
    Runs a query in all sandboxes.
    It does not check if the query is compatible with every version deployed.
    For example, a query using @@port won't run in MySQL 5.0.x
    
    Usage:
      dbdeployer global use {query} [flags]
    
    Examples:
    
    	$ dbdeployer global use "select @@server_id, @@port"
    
    Flags:
      -h, --help   help for use
    
    

## remote
    $ dbdeployer remote -h
    Manages remote tarballs
    
    Usage:
      dbdeployer remote [command]
    
    Available Commands:
      download    download a remote tarball into a local file
      list        list remote tarballs
    
    Flags:
      -h, --help   help for remote
    
    
    $ dbdeployer remote download -h
    If no file name is given, the file name will be mysql-<version>.tar.xz
    
    Usage:
      dbdeployer remote download version [file-name] [flags]
    
    Aliases:
      download, get
    
    Flags:
      -h, --help   help for download
    
    
    $ dbdeployer remote list -h
    list remote tarballs
    
    Usage:
      dbdeployer remote list [version] [flags]
    
    Aliases:
      list, index
    
    Flags:
      -h, --help   help for list
    
    

## sandboxes
    $ dbdeployer sandboxes -h
    Lists all sandboxes installed in $SANDBOX_HOME.
    If sandboxes are installed in a different location, use --sandbox-home to 
    indicate where to look.
    Alternatively, using --catalog will list all sandboxes, regardless of where 
    they were deployed.
    
    Usage:
      dbdeployer sandboxes [flags]
    
    Aliases:
      sandboxes, installed, deployed
    
    Flags:
          --catalog     Use sandboxes catalog instead of scanning directory
          --flavor      Shows flavor in sandbox list
          --full-info   Shows all info in table format
          --header      Shows header with catalog output
      -h, --help        help for sandboxes
          --table       Shows sandbox list as a table
    
    

## unpack
    $ dbdeployer unpack -h
    If you want to create a sandbox from a tarball (.tar.gz or .tar.xz), you first need to unpack it
    into the sandbox-binary directory. This command carries out that task, so that afterwards 
    you can call 'deploy single', 'deploy multiple', and 'deploy replication' commands with only 
    the MySQL version for that tarball.
    If the version is not contained in the tarball name, it should be supplied using --unpack-version.
    If there is already an expanded tarball with the same version, a new one can be differentiated with --prefix.
    
    Usage:
      dbdeployer unpack MySQL-tarball [flags]
    
    Aliases:
      unpack, extract, untar, unzip, inflate, expand
    
    Examples:
    
        $ dbdeployer unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
        Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4
    
        $ dbdeployer unpack --prefix=ps Percona-Server-5.7.21-linux.tar.gz
        Unpacking tarball Percona-Server-5.7.21-linux.tar.gz to $HOME/opt/mysql/ps5.7.21
    
        $ dbdeployer unpack --unpack-version=8.0.18 --prefix=bld mysql-mybuild.tar.gz
        Unpacking tarball mysql-mybuild.tar.gz to $HOME/opt/mysql/bld8.0.18
    	
    
    Flags:
          --flavor string           Defines the tarball flavor (MySQL, NDB, Percona Server, etc)
      -h, --help                    help for unpack
          --overwrite               Overwrite the destination directory if already exists
          --prefix string           Prefix for the final expanded directory
          --shell                   Unpack a shell tarball into the corresponding server directory
          --target-server string    Uses a different server to unpack a shell tarball
          --unpack-version string   which version is contained in the tarball
          --verbosity int           Level of verbosity during unpack (0=none, 2=maximum) (default 1)
    
    

## usage
    $ dbdeployer usage -h
    Shows syntax and examples of tools installed in database sandboxes.
    
    Usage:
      dbdeployer usage [single|multiple] [flags]
    
    Flags:
      -h, --help   help for usage
    
    

## versions
    $ dbdeployer versions -h
    List available versions
    
    Usage:
      dbdeployer versions [flags]
    
    Aliases:
      versions, available
    
    Flags:
      -h, --help   help for versions
    
    
