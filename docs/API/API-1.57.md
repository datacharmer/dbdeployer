This is the list of commands and modifiers available for
dbdeployer 1.57.0 as of 09-Dec-2020 12:05 UTC

# main
    $ dbdeployer -h 
    dbdeployer makes MySQL server installation an easy task.
    Runs single, multiple, and replicated sandboxes.
    
    Usage:
      dbdeployer [command]
    
    Available Commands:
      admin           sandbox management tasks
      cookbook        Shows dbdeployer samples
      data-load       tasks related to dbdeployer data loading
      defaults        tasks related to dbdeployer defaults
      delete          delete an installed sandbox
      delete-binaries delete an expanded tarball
      deploy          deploy sandboxes
      downloads       Manages remote tarballs
      export          Exports the command structure in JSON format
      global          Runs a given command in every sandbox
      help            Help about any command
      import          imports one or more MySQL servers into a sandbox
      info            Shows information about dbdeployer environment samples
      init            initializes dbdeployer environment
      sandboxes       List installed sandboxes
      unpack          unpack a tarball into the binary directory
      update          Gets dbdeployer newest version
      usage           Shows usage of installed sandboxes
      use             uses a sandbox
      versions        List available versions
    
    Flags:
          --config string           configuration file (default "$HOME/.dbdeployer/config.json")
      -h, --help                    help for dbdeployer
          --sandbox-binary string   Binary repository (default "$HOME/opt/mysql")
          --sandbox-home string     Sandbox deployment directory (default "$HOME/sandboxes")
          --shell-path string       Which shell to use for generated scripts (default "/usr/local/bin/bash")
          --skip-library-check      Skip check for needed libraries (may cause nasty errors)
      -v, --version                 version for dbdeployer
    
    Use "dbdeployer [command] --help" for more information about a command.
    

    $ dbdeployer-docs tree 
    - admin               
        - capabilities        
        - lock                
        - remove-default      
        - set-default         
        - unlock              
        - upgrade             
    - cookbook            
        - create              
        - list                
        - show                
    - data-load           
        - export              
        - get                 
        - import              
        - list                
        - reset               
        - show                
    - defaults            
        - enable-bash-completion
        - export              
        - flag-aliases        
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
    - downloads           
        - add                 
        - export              
        - get                 
        - get-by-version      
        - get-unpack          
        - import              
        - list                
        - reset               
        - show                
    - export              
    - global              
        - metadata            
        - restart             
        - start               
        - status              
        - stop                
        - test                
        - test-replication    
        - use                 
    - import              
        - single              
    - info                
        - defaults            
        - releases            
        - version             
    - init                
    - remote              
        - download            
        - list                
    - sandboxes           
    - unpack              
    - update              
    - usage               
    - use                 
    - versions            
    

## admin
    $ dbdeployer admin -h
    Runs commands related to the administration of sandboxes.
    
    Usage:
      dbdeployer admin [command]
    
    Aliases:
      admin, manage
    
    Available Commands:
      capabilities   Shows capabilities of a given flavor [and optionally version]
      lock           Locks a sandbox, preventing deletion
      remove-default Removes default sandbox
      set-default    Sets a sandbox as default
      unlock         Unlocks a sandbox
      upgrade        Upgrades a sandbox to a newer version
    
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
    
    
    $ dbdeployer admin remove-default -h
    Removes the default sandbox
    
    Usage:
      dbdeployer admin remove-default [flags]
    
    Flags:
          --default-sandbox-executable string   Name of the executable to run commands in the default sandbox (default "default")
      -h, --help                                help for remove-default
    
    
    $ dbdeployer admin set-default -h
    Sets a given sandbox as default, so that it can be used with $SANDBOX_HOME/default
    
    Usage:
      dbdeployer admin set-default sandbox_name [flags]
    
    Flags:
          --default-sandbox-executable string   Name of the executable to run commands in the default sandbox (default "default")
      -h, --help                                help for set-default
    
    
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
          --dry-run   Shows upgrade operations, but don't execute them
      -h, --help      help for upgrade
          --verbose   Shows upgrade operations
    
    

## cookbook
    $ dbdeployer cookbook -h
    Shows practical examples of dbdeployer usages, by creating usage scripts.
    
    Usage:
      dbdeployer cookbook [command]
    
    Aliases:
      cookbook, recipes, samples
    
    Available Commands:
      create      creates a script for a given recipe
      list        Shows available dbdeployer samples
      show        Shows the contents of a given recipe
    
    Flags:
          --flavor string   For which flavor this recipe is
      -h, --help            help for cookbook
    
    
    $ dbdeployer cookbook create -h
    creates a script for a given recipe
    
    Usage:
      dbdeployer cookbook create recipe_name or ALL [flags]
    
    Aliases:
      create, make
    
    Flags:
      -h, --help   help for create
    
    
    $ dbdeployer cookbook list -h
    Shows list of available cookbook recipes
    
    Usage:
      dbdeployer cookbook list [flags]
    
    Flags:
      -h, --help             help for list
          --sort-by string   Sort order for the list (name, flavor, script) (default "name")
    
    
    $ dbdeployer cookbook show -h
    Shows the contents of a given recipe, without actually running it
    
    Usage:
      dbdeployer cookbook show recipe_name [flags]
    
    Flags:
      -h, --help   help for show
          --raw    Shows the recipe without variable substitution
    
    

## data-load
    $ dbdeployer data-load -h
    Runs commands related to the database loading
    
    Usage:
      dbdeployer data-load [command]
    
    Aliases:
      data-load, load-data
    
    Available Commands:
      export      Saves the archives details into a file
      get         Loads an archived database into a sandbox
      import      Imports the archives details from a file
      list        list databases available for loading
      reset       Resets the archives to their default values
      show        show details of a database
    
    Flags:
      -h, --help   help for data-load
    
    
    $ dbdeployer data-load export -h
    Saves the archives details into a file
    
    Usage:
      dbdeployer data-load export file-name [flags]
    
    Flags:
      -h, --help   help for export
    
    
    $ dbdeployer data-load get -h
    Loads an archived database into a sandbox
    
    Usage:
      dbdeployer data-load get archive-name sandbox-name [flags]
    
    Flags:
      -h, --help        help for get
          --overwrite   overwrite previously downloaded archive
    
    
    $ dbdeployer data-load import -h
    
    Imports modified archives from a JSON file.
    In the archive specification, the strings "$use" and "$my"
    will be expanded to the relative scripts in the target sandbox directory.
    
    Usage:
      dbdeployer data-load import file-name [flags]
    
    Flags:
      -h, --help   help for import
    
    
    $ dbdeployer data-load list -h
    List databases available for loading
    
    Usage:
      dbdeployer data-load list [flags]
    
    Flags:
          --full-info   Shows all archive details
      -h, --help        help for list
    
    
    $ dbdeployer data-load reset -h
    Resets the archives to their default values
    
    Usage:
      dbdeployer data-load reset [flags]
    
    Flags:
      -h, --help   help for reset
    
    
    $ dbdeployer data-load show -h
    show details of a database
    
    Usage:
      dbdeployer data-load show archive-name [flags]
    
    Flags:
          --full-info   Shows all archive details
      -h, --help        help for show
    
    

## defaults
    $ dbdeployer defaults -h
    Runs commands related to the administration of dbdeployer,
    such as showing the defaults and saving new ones.
    
    Usage:
      dbdeployer defaults [command]
    
    Aliases:
      defaults, config
    
    Available Commands:
      enable-bash-completion Enables bash-completion for dbdeployer
      export                 Export current defaults to a given file
      flag-aliases           Shows flag aliases
      load                   Load defaults from file
      reset                  Remove current defaults file
      show                   shows defaults
      store                  Store current defaults
      templates              Templates management
      update                 Change defaults value
    
    Flags:
      -h, --help   help for defaults
    
    
    $ dbdeployer defaults enable-bash-completion -h
    Enables bash completion using either a local copy of dbdeployer_completion.sh or a remote one
    
    Usage:
      dbdeployer defaults enable-bash-completion [flags]
    
    Flags:
          --completion-file string   Use this file as completion
      -h, --help                     help for enable-bash-completion
          --remote                   Download dbdeployer_completion.sh from GitHub
          --remote-url string        Where to downloads dbdeployer_completion.sh from (default "https://raw.githubusercontent.com/datacharmer/dbdeployer/master/docs/dbdeployer_completion.sh")
          --run-it                   Run the command instead of just showing it
    
    
    $ dbdeployer defaults export -h
    Saves current defaults to a user-defined file
    
    Usage:
      dbdeployer defaults export filename [flags]
    
    Flags:
      -h, --help   help for export
    
    
    $ dbdeployer defaults flag-aliases -h
    Shows the aliases available for some flags
    
    Usage:
      dbdeployer defaults flag-aliases [flags]
    
    Aliases:
      flag-aliases, option-aliases, aliases
    
    Flags:
      -h, --help   help for flag-aliases
    
    
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
          --camel-case   Show defaults in CamelCase format
      -h, --help         help for show
    
    
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
    Halts the sandbox (and its depending sandboxes, if any), and removes it.
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
          --use-stop       Use 'stop' instead of 'send_kill destroy' to halt the database servers
    
    

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
          --base-port int                   Overrides default base-port (for multiple sandboxes)
          --base-server-id int              Overrides default server_id (for multiple sandboxes)
          --binary-version string           Specifies the version when the basedir directory name does not contain it (i.e. it is not x.x.xx)
          --bind-address string             defines the database bind-address  (default "127.0.0.1")
          --client-from string              Where to get the client binaries from
          --concurrent                      Runs multiple sandbox deployments concurrently
          --custom-mysqld string            Uses an alternative mysqld (must be in the same directory as regular mysqld)
          --custom-role-extra string        Extra instructions for custom role (8.0+) (default "WITH GRANT OPTION")
          --custom-role-name string         Name for custom role (8.0+) (default "R_CUSTOM")
          --custom-role-privileges string   Privileges for custom role (8.0+) (default "ALL PRIVILEGES")
          --custom-role-target string       Target for custom role (8.0+) (default "*.*")
      -p, --db-password string              database password (default "msandbox")
      -u, --db-user string                  database user (default "msandbox")
          --default-role string             Which role to assign to default user (8.0+) (default "R_DO_IT_ALL")
          --defaults stringArray            Change defaults on-the-fly (--defaults=label:value)
          --disable-mysqlx                  Disable MySQLX plugin (8.0.11+)
          --enable-admin-address            Enables admin address (8.0.14+)
          --enable-general-log              Enables general log for the sandbox (MySQL 5.1+)
          --enable-mysqlx                   Enables MySQLX plugin (5.7.12+)
          --expose-dd-tables                In MySQL 8.0+ shows data dictionary tables
          --flavor string                   Defines the tarball flavor (MySQL, NDB, Percona Server, etc)
          --flavor-in-prompt                Add flavor values to prompt
          --force                           If a destination sandbox already exists, it will be overwritten
          --gtid                            enables GTID
      -h, --help                            help for deploy
          --history-dir string              Where to store mysql client history (default: in sandbox directory)
          --init-general-log                uses general log during initialization (MySQL 5.1+)
      -i, --init-options stringArray        mysqld options to run during initialization
          --keep-server-uuid                Does not change the server UUID
          --log-directory string            Where to store dbdeployer logs (default "$HOME/sandboxes/logs")
          --log-sb-operations               Logs sandbox operations to a file
          --my-cnf-file string              Alternative source file for my.sandbox.cnf
      -c, --my-cnf-options stringArray      mysqld options to add to my.sandbox.cnf
          --native-auth-plugin              in 8.0.4+, uses the native password auth plugin
          --port int                        Overrides default port
          --port-as-server-id               Use the port number as server ID
          --post-grants-sql stringArray     SQL queries to run after loading grants
          --post-grants-sql-file string     SQL file to run after loading grants
          --pre-grants-sql stringArray      SQL queries to run before loading grants
          --pre-grants-sql-file string      SQL file to run before loading grants
          --remote-access string            defines the database access  (default "127.%")
          --repl-crash-safe                 enables Replication crash safe
          --rpl-password string             replication password (default "rsandbox")
          --rpl-user string                 replication user (default "rsandbox")
          --sandbox-directory string        Changes the default name of the sandbox directory
          --skip-load-grants                Does not load the grants
          --skip-report-host                Does not include report host in my.sandbox.cnf
          --skip-report-port                Does not include report port in my.sandbox.cnf
          --skip-start                      Does not start the database server
          --socket-in-datadir               Create socket in datadir instead of $TMPDIR
          --task-user string                Task user to be created (8.0+)
          --task-user-role string           Role to be assigned to task user (8.0+)
          --use-template stringArray        [template_name:file_name] Replace existing template with one from file
    
    
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
    Topologies "pcx" and "ndb" are available for binaries of type Percona Xtradb Cluster and MySQL Cluster.
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
    		$ dbdeployer deploy --topology=ndb replication ndb8.0.14
    	
    
    Flags:
          --change-master-options stringArray   options to add to CHANGE MASTER TO
      -h, --help                                help for replication
          --master-ip string                    Which IP the slaves will connect to (default "127.0.0.1")
          --master-list string                  Which nodes are masters in a multi-source deployment
          --ndb-nodes int                       How many NDB nodes will be installed (default 3)
      -n, --nodes int                           How many nodes will be installed (default 3)
          --read-only-slaves                    Set read-only for slaves
          --repl-history-dir                    uses the replication directory to store mysql client history
          --semi-sync                           Use semi-synchronous plugin
          --single-primary                      Using single primary for group replication
          --slave-list string                   Which nodes are slaves in a multi-source deployment
          --super-read-only-slaves              Set super-read-only for slaves
      -t, --topology string                     Which topology will be installed (default "master-slave")
    
    
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
          --server-id int   Overwrite default server-id
    
    

## downloads
    $ dbdeployer downloads -h
    Manages remote tarballs
    
    Usage:
      dbdeployer downloads [command]
    
    Available Commands:
      add            Adds a tarball to the list
      export         Exports the list of tarballs to a file
      get            Downloads a remote tarball
      get-by-version Downloads a remote tarball
      get-unpack     Downloads and unpacks a remote tarball
      import         Imports the list of tarballs from a file or URL
      list           list remote tarballs
      reset          Reset the custom list of tarballs and resume the defaults
      show           Downloads a remote tarball
    
    Flags:
      -h, --help   help for downloads
    
    
    $ dbdeployer downloads add -h
    Adds a tarball to the list
    
    Usage:
      dbdeployer downloads add tarball_name [flags]
    
    Flags:
          --OS string              Define the tarball OS (default: current OS)
          --flavor string          Define the tarball flavor
      -h, --help                   help for add
          --minimal                Define whether the tarball is a minimal one
          --overwrite              Overwrite existing entry
          --short-version string   Define the tarball short version
          --url string             Define the tarball URL
          --version string         Define the tarball version
    
    
    $ dbdeployer downloads export -h
    Exports the list of tarballs to a file
    
    Usage:
      dbdeployer downloads export file-name [options] [flags]
    
    Flags:
          --add-empty-item   Add an empty item to the tarballs list
      -h, --help             help for export
    
    
    $ dbdeployer downloads get -h
    Downloads a remote tarball
    
    Usage:
      dbdeployer downloads get tarball_name [options] [flags]
    
    Flags:
          --dry-run             Show what would be downloaded, but don't run it
      -h, --help                help for get
          --progress-step int   Progress interval (default 10485760)
          --quiet               Do not show download progress
    
    
    $ dbdeployer downloads get-by-version -h
    
    Download a tarball identified by a combination of
    version, flavor, operating system, and optionally its minimal state.
    If you don't specify the Operating system, the current one will be assumed.
    If the flavor is not specified, 'mysql' is assumed.
    Use the option '--dry-run' to see what dbdeployer would download.
    
    Usage:
      dbdeployer downloads get-by-version version [options] [flags]
    
    Examples:
    
    $ dbdeployer downloads get-by-version 5.7 --newest --dry-run
    $ dbdeployer downloads get-by-version 5.7 --newest --minimal --dry-run --OS=linux
    $ dbdeployer downloads get-by-version 5.7 --newest
    $ dbdeployer downloads get-by-version 8.0 --flavor=ndb
    $ dbdeployer downloads get-by-version 5.7.26 --minimal
    $ dbdeployer downloads get-by-version 5.7 --minimal
    
    
    Flags:
          --OS string           Choose only the given OS
          --dry-run             Show what would be downloaded, but don't run it
          --flavor string       Choose only the given flavor
          --guess-latest        Guess the latest version (highest version w/ increased revision number)
      -h, --help                help for get-by-version
          --minimal             Choose only minimal tarballs
          --newest              Choose only the newest tarballs not yet downloaded
          --progress-step int   Progress interval (default 10485760)
          --quiet               Do not show download progress
    
    
    $ dbdeployer downloads get-unpack -h
    get-unpack downloads a tarball and then unpacks it, using the same
    options available for "dbdeployer unpack".
    
    Usage:
      dbdeployer downloads get-unpack tarball_name [options] [flags]
    
    Flags:
          --delete-after-unpack     Delete the tarball after successful unpack
          --flavor string           Defines the tarball flavor (MySQL, NDB, Percona Server, etc)
      -h, --help                    help for get-unpack
          --overwrite               Overwrite the destination directory if already exists
          --prefix string           Prefix for the final expanded directory
          --progress-step int       Progress interval (default 10485760)
          --shell                   Unpack a shell tarball into the corresponding server directory
          --target-server string    Uses a different server to unpack a shell tarball
          --unpack-version string   which version is contained in the tarball
          --verbosity int           Level of verbosity during unpack (0=none, 2=maximum) (default 1)
    
    
    $ dbdeployer downloads import -h
    
    Imports the list of tarballs from a file or a URL.
    If the argument is "remote-github" or "remote-tarballs", dbdeployer will get the file from
    its Github repository.
    (See: dbdeployer info defaults remote-tarball-url)
    
    Usage:
      dbdeployer downloads import {file-name | URL} [flags]
    
    Flags:
      -h, --help   help for import
    
    
    $ dbdeployer downloads list -h
    List remote tarballs.
    By default it includes tarballs for current operating system.
    Use '--OS=os_name' or '--OS=all' to change.
    
    Usage:
      dbdeployer downloads list [options] [flags]
    
    Aliases:
      list, index
    
    Flags:
          --OS string       Which OS will be listed
          --flavor string   Which flavor will be listed
      -h, --help            help for list
          --show-url        Show the URL
    
    
    $ dbdeployer downloads reset -h
    Reset the custom list of tarballs and resume the defaults
    
    Usage:
      dbdeployer downloads reset [flags]
    
    Flags:
      -h, --help   help for reset
    
    
    $ dbdeployer downloads show -h
    Downloads a remote tarball
    
    Usage:
      dbdeployer downloads show tarball_name [flags]
    
    Aliases:
      show, display
    
    Flags:
      -h, --help   help for show
    
    

## export
    $ dbdeployer export -h
    Exports the command line structure, with examples and flags, to a JSON structure.
    If a command is given, only the structure of that command and below will be exported.
    Given the length of the output, it is recommended to pipe it to a file or to another command.
    
    Usage:
      dbdeployer export [command [sub-command]] [ > filename ] [ | command ]  [flags]
    
    Aliases:
      export, dump
    
    Flags:
          --force-output-to-terminal   display output to terminal regardless of pipes being used
      -h, --help                       help for export
    
    

## global
    $ dbdeployer global -h
    This command can propagate the given action through all sandboxes.
    
    Usage:
      dbdeployer global [command]
    
    Examples:
    
    	$ dbdeployer global use "select version()"
    	$ dbdeployer global status
    	$ dbdeployer global stop --version=5.7.27
    	$ dbdeployer global stop --short-version=8.0
    	$ dbdeployer global stop --short-version='!8.0' # or --short-version=no-8.0
    	$ dbdeployer global status --port-range=5000-8099
    	$ dbdeployer global start --flavor=percona
    	$ dbdeployer global start --flavor='!percona' --type=single
    	$ dbdeployer global metadata version --flavor='!percona' --type=single
    	
    
    Available Commands:
      metadata         Runs a metadata query in all sandboxes
      restart          Restarts all sandboxes
      start            Starts all sandboxes
      status           Shows the status in all sandboxes
      stop             Stops all sandboxes
      test             Tests all sandboxes
      test-replication Tests replication in all sandboxes
      use              Runs a query in all sandboxes
    
    Flags:
          --dry-run                Show what would be executed, without doing it
          --flavor string          Runs command only in sandboxes of the given flavor
      -h, --help                   help for global
          --name string            Runs command only in sandboxes of the given name
          --port string            Runs commands only in sandboxes containing the given port
          --port-range string      Runs command only in sandboxes containing a port in the given range
          --short-version string   Runs command only in sandboxes of the given short version
          --type string            Runs command only in sandboxes of the given type
          --verbose                Show what is matched when filters are used
          --version string         Runs command only in sandboxes of the given version
    
    
    $ dbdeployer global metadata -h
    Runs a metadata query in all sandboxes
    
    Usage:
      dbdeployer global metadata {keyword} [flags]
    
    Examples:
    
    	$ dbdeployer global metadata 
    
    Flags:
      -h, --help   help for metadata
    
    
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
    
    

## import
    $ dbdeployer import -h
    imports one or more MySQL servers into a sandbox
    
    Usage:
      dbdeployer import [command]
    
    Available Commands:
      single      imports a MySQL server into a sandbox
    
    Flags:
          --client-from string         Where to get the client binaries from
      -h, --help                       help for import
          --sandbox-directory string   Changes the default sandbox directory
    
    
    $ dbdeployer import single -h
    Imports an existing (local or remote) server into a sandbox,
    so that it can be used with the usual sandbox scripts.
    Requires host, port, user, password.
    
    Usage:
      dbdeployer import single host port user password [flags]
    
    Flags:
      -h, --help   help for single
    
    

## info
    $ dbdeployer info -h
    Shows current information about defaults and environment.
    
    Usage:
      dbdeployer info [command]
    
    Available Commands:
      defaults    displays a defaults value
      releases    displays info on releases, or a given release
      version     displays the latest version available
    
    Flags:
          --earliest        Return the earliest version
          --flavor string   For which flavor this info is
      -h, --help            help for info
    
    
    $ dbdeployer info defaults -h
    Displays one field of the defaults.
    
    Usage:
      dbdeployer info defaults field-name [flags]
    
    Examples:
    
    	$ dbdeployer info defaults master-slave-base-port 
    
    
    Flags:
      -h, --help   help for defaults
    
    
    $ dbdeployer info releases -h
    Displays info on all the available releases, or a specific one
    
    Usage:
      dbdeployer info releases [tag] [flags]
    
    Examples:
    
    	$ dbdeployer info releases
    	$ dbdeployer info releases v1.35.0
    	$ dbdeployer info releases latest
    
    
    Flags:
      -h, --help        help for releases
          --limit int   Limit number of releases to show (0 = unlimited) (default 3)
          --raw         Show the original data
          --stats       Show downloads statistics
    
    
    $ dbdeployer info version -h
    Displays the latest version available for deployment.
    If a short version is indicated (such as 5.7, or 8.0), only the versions belonging to that short
    version are searched.
    If "all" is indicated after the short version, displays all versions belonging to that short version.
    
    Usage:
      dbdeployer info version [short-version|all] [all] [flags]
    
    Examples:
    
        # Shows the latest version available
        $ dbdeployer info version
        8.0.16
    
        # shows the latest version belonging to 5.7
        $ dbdeployer info version 5.7
        5.7.26
    
        # shows the latest version for every short version
        $ dbdeployer info version all
        5.0.96 5.1.73 5.5.53 5.6.41 5.7.26 8.0.16
    
        # shows all the versions for a given short version
        $ dbdeployer info version 8.0 all
        8.0.11 8.0.12 8.0.13 8.0.14 8.0.15 8.0.16
    
    
    Flags:
      -h, --help   help for version
    
    

## init
    $ dbdeployer init -h
    Initializes dbdeployer environment: 
    * creates $SANDBOX_HOME and $SANDBOX_BINARY directories
    * downloads and expands the latest MySQL tarball
    * installs shell completion file
    
    Usage:
      dbdeployer init [flags]
    
    Flags:
          --dry-run                 Show operations but don't run them
      -h, --help                    help for init
          --skip-all-downloads      Do not download any file (skip both MySQL tarball and shell completion file)
          --skip-shell-completion   Do not download shell completion file
          --skip-tarball-download   Do not download MySQL tarball
    
    

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
    If no file name is given, the file name will be <version>.tar.xz
    
    Usage:
      dbdeployer remote download version [file-name] [flags]
    
    Aliases:
      download, get
    
    Flags:
      -h, --help                help for download
          --progress            Show download progress
          --progress-step int   Progress interval (default 10485760)
    
    
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
          --by-date      Show sandboxes in order of creation
          --by-flavor    Show sandboxes sorted by flavor
          --by-version   Show sandboxes sorted by version
          --catalog      Use sandboxes catalog instead of scanning directory
          --flavor       Shows flavor in sandbox list
          --full-info    Shows all info in table format
          --header       Shows header with catalog output
      -h, --help         help for sandboxes
          --latest       Show only latest sandbox
          --oldest       Show only oldest sandbox
          --table        Shows sandbox list as a table
    
    

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
    
    

## update
    $ dbdeployer update -h
    Updates dbdeployer in place using the latest version (or one of your choice)
    
    Usage:
      dbdeployer update [version] [flags]
    
    Examples:
    
    $ dbdeployer update
    # gets the latest release, overwrites current dbdeployer binaries 
    
    $ dbdeployer update --dry-run
    # shows what it will do, but does not do it
    
    $ dbdeployer update --new-path=$PWD
    # downloads the latest executable into the current directory
    
    $ dbdeployer update v1.34.0 --force-old-version
    # downloads dbdeployer 1.34.0 and replace the current one
    # (WARNING: a version older than 1.36.0 won't support updating)
    
    
    Flags:
          --OS string           Gets the executable for this Operating system
          --docs                Gets the docs version of the executable
          --dry-run             Show what would happen, but don't execute it
          --force-old-version   Force download of older version
      -h, --help                help for update
          --new-path string     Download updated dbdeployer into a different path
          --verbose             Gives more info
    
    

## usage
    $ dbdeployer usage -h
    Shows syntax and examples of tools installed in database sandboxes.
    
    Usage:
      dbdeployer usage [single|multiple] [flags]
    
    Flags:
      -h, --help   help for usage
    
    

## use
    $ dbdeployer use -h
    Uses a given sandbox.
    If a sandbox is indicated, it will be used.
    Otherwise, it will use the latest deployed sandbox.
    Optionally, an executable can be set as second argument.
    
    Usage:
      dbdeployer use [sandbox_name [executable]] [flags]
    
    Examples:
    
    $ dbdeployer use                    # runs "use" on the latest deployed sandbox
    $ dbdeployer use rsandbox_8_0_22    # runs "m" on replication sandbox rsandbox_8_0_22
    $ dbdeployer use rsandbox_8_0_22 s1 # runs "s1" on replication sandbox rsandbox_8_0_22
    $ echo 'SELECT @@SERVER_ID' | dbdeployer use # pipes an SQL query to latest deployed sandbox
    
    
    Flags:
      -h, --help         help for use
          --ls           List files in sandbox
          --run string   Name of executable to run
    
    

## versions
    $ dbdeployer versions -h
    List available versions
    
    Usage:
      dbdeployer versions [flags]
    
    Aliases:
      versions, available
    
    Flags:
          --by-flavor       Shows versions list by flavor
          --flavor string   Get only versions of the given flavor
      -h, --help            help for versions
    
    
