This is the list of commands and modifiers available for dbdeployer

    $ dbdeployer --version
    dbdeployer version 1.1.0
    
# main
    $ dbdeployer -h 
    dbdeployer makes MySQL server installation an easy task.
    Runs single, multiple, and replicated sandboxes.
    
    Usage:
      dbdeployer [command]
    
    Available Commands:
      admin       sandbox management tasks
      defaults    tasks related to dbdeployer defaults
      delete      delete an installed sandbox
      deploy      deploy sandboxes
      global      Runs a given command in every sandbox
      help        Help about any command
      sandboxes   List installed sandboxes
      unpack      unpack a tarball into the binary directory
      usage       Shows usage of installed sandboxes
      versions    List available versions
    
    Flags:
          --config string           configuration file (default "$HOME/.dbdeployer/config.json")
      -h, --help                    help for dbdeployer
          --sandbox-binary string   Binary repository (default "$HOME/opt/mysql")
          --sandbox-home string     Sandbox deployment directory (default "$HOME/sandboxes")
          --version                 version for dbdeployer
    
    Use "dbdeployer [command] --help" for more information about a command.
    

    $ dbdeployer tree 
    - admin                
        - lock                 
        - unlock               
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
      lock        Locks a sandbox, preventing deletion
      unlock      Unlocks a sandbox
    
    Flags:
      -h, --help   help for admin
    
    
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
          --bind-address string           defines the database bind-address  (default "127.0.0.1")
          --concurrent                    Runs multiple sandbox deployments concurrently
          --custom-mysqld string          Uses an alternative mysqld (must be in the same directory as regular mysqld)
      -p, --db-password string            database password (default "msandbox")
      -u, --db-user string                database user (default "msandbox")
          --defaults strings              Change defaults on-the-fly (--defaults=label:value)
          --expose-dd-tables              In MySQL 8.0+ shows data dictionary tables
          --force                         If a destination sandbox already exists, it will be overwritten
          --gtid                          enables GTID
      -h, --help                          help for deploy
      -i, --init-options strings          mysqld options to run during initialization
          --keep-server-uuid              Does not change the server UUID
          --my-cnf-file string            Alternative source file for my.sandbox.cnf
      -c, --my-cnf-options strings        mysqld options to add to my.sandbox.cnf
          --native-auth-plugin            in 8.0.4+, uses the native password auth plugin
          --port int                      Overrides default port
          --post-grants-sql strings       SQL queries to run after loading grants
          --post-grants-sql-file string   SQL file to run after loading grants
          --pre-grants-sql strings        SQL queries to run before loading grants
          --pre-grants-sql-file string    SQL file to run before loading grants
          --remote-access string          defines the database access  (default "127.%")
          --rpl-password string           replication password (default "rsandbox")
          --rpl-user string               replication user (default "rsandbox")
          --sandbox-directory string      Changes the default sandbox directory
          --skip-load-grants              Does not load the grants
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
    For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
    the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
    Use the "unpack" command to get the tarball into the right directory.
    
    Usage:
      dbdeployer deploy replication MySQL-Version [flags]
    
    Examples:
    
    		$ dbdeployer deploy replication 5.7.21
    		# (implies topology = master-slave)
    
    		$ dbdeployer deploy --topology=master-slave replication 5.7.21
    		# (explicitly setting topology)
    
    		$ dbdeployer deploy --topology=group replication 5.7.21
    		$ dbdeployer deploy --topology=group replication 8.0.4 --single-primary
    		$ dbdeployer deploy --topology=all-masters replication 5.7.21
    		$ dbdeployer deploy --topology=fan-in replication 5.7.21
    	
    
    Flags:
      -h, --help                 help for replication
          --master-ip string     Which IP the slaves will connect to (default "127.0.0.1")
          --master-list string   Which nodes are masters in a multi-source deployment (default "1,2")
      -n, --nodes int            How many nodes will be installed (default 3)
          --semi-sync            Use semi-synchronous plugin
          --single-primary       Using single primary for group replication
          --slave-list string    Which nodes are slaves in a multi-source deployment (default "3")
      -t, --topology string      Which topology will be installed (default "master-slave")
    
    
    $ dbdeployer deploy single -h
    single installs a sandbox and creates useful scripts for its use.
    MySQL-Version is in the format x.x.xx, and it refers to a directory named after the version
    containing an unpacked tarball. The place where these directories are found is defined by 
    --sandbox-binary (default: $HOME/opt/mysql.)
    For example:
    	dbdeployer deploy single 5.7.21
    
    For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
    the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
    Use the "unpack" command to get the tarball into the right directory.
    
    Usage:
      dbdeployer deploy single MySQL-Version [flags]
    
    Flags:
      -h, --help     help for single
          --master   Make the server replication ready
    
    

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
          --catalog   Use sandboxes catalog instead of scanning directory
          --header    Shows header with catalog output
      -h, --help      help for sandboxes
    
    

## unpack
    $ dbdeployer unpack -h
    If you want to create a sandbox from a tarball, you first need to unpack it
    into the sandbox-binary directory. This command carries out that task, so that afterwards 
    you can call 'single', 'multiple', and 'replication' commands with only the MySQL version
    for that tarball.
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
      -h, --help                    help for unpack
          --prefix string           Prefix for the final expanded directory
          --unpack-version string   which version is contained in the tarball
    
    

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
    
    
