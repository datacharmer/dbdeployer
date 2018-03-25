# dbdeployer

[DBdeployer](https://github.com/datacharmer/dbdeployer) is a tool that deploys MySQL database servers easily.
This is a port of [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox), originally written in Perl, and re-designed from the ground up in [Go](https://golang.org). See the [features comparison](https://github.com/datacharmer/dbdeployer/blob/master/docs/features.md) for more detail.

## Installation

The installation is simple, as the only thing you will need is a binary executable for your operating system.
Get the one for your O.S. from [dbdeployer releases](https://github.com/datacharmer/dbdeployer/releases) and place it in a directory in your $PATH.
(There are no binaries for Windows. See the [features list](https://github.com/datacharmer/dbdeployer/blob/master/docs/features.md) for more info.)

For example:

    $ VERSION=0.3.1
    $ origin=https://github.com/datacharmer/dbdeployer/releases/download/$VERSION
    $ wget $origin/dbdeployer-$VERSION.linux.tar.gz
    $ tar -xzf dbdeployer-$VERSION.linux.tar.gz
    $ chmod +x dbdeployer-$VERSION.linux
    $ sudo mv dbdeployer-$VERSION.linux /usr/local/bin/dbdeployer

Of course, there are **prerequisites**: your machine must be able to run the MySQL server. Be aware that version 5.6+ and higher require some libraries that are not installed by default in all flavors of Linux (libnuma, libaio.)

## Main operations

(See this ASCIIcast for a demo of its operations.)
[![asciicast](https://asciinema.org/a/165707.png)](https://asciinema.org/a/165707)

With dbdeployer, you can deploy a single sandbox, or many sandboxes  at once, with or without replication.

The main command is **deploy** with its subcommands **single**, **replication**, and **multiple**, which work with MySQL tarball that have been unpacked into the _sandbox-binary_ directory (by default, $HOME/opt/mysql.)

To use a tarball, you must first run the **unpack** command, which will unpack the tarball into the right directory.

For example:

    $ dbdeployer --unpack-version=8.0.4 unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4
    .........100.........200.........292

    $ dbdeployer deploy single 8.0.4
    Database installed in $HOME/sandboxes/msb_8_0_4
    . sandbox server started


The program doesn't have any dependencies. Everything is included in the binary. Calling *dbdeployer* without arguments or with '--help' will show the main help screen.

    $ dbdeployer --version
    dbdeployer version 0.3.8
    

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
          --sandbox-home string     Sandbox deployment direcory (default "$HOME/sandboxes")
          --version                 version for dbdeployer
    
    Use "dbdeployer [command] --help" for more information about a command.
    

The flags listed in the main screen can be used with any commands.
The flags _--my-cnf-options_ and _--init-options_ can be used several times.

If you don't have any tarballs installed in your system, you should first *unpack* it (see an example above).

    $ dbdeployer unpack -h
    If you want to create a sandbox from a tarball, you first need to unpack it
    into the sandbox-binary directory. This command carries out that task, so that afterwards 
    you can call 'single', 'multiple', and 'replication' commands with only the MySQL version
    for that tarball.
    
    Usage:
      dbdeployer unpack MySQL-tarball [flags]
    
    Aliases:
      unpack, extract, untar, unzip, inflate, expand
    
    Examples:
    
        $ dbdeployer --unpack-version=8.0.4 unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
        Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4
        .........100.........200.........292
    	
    
    Flags:
      -h, --help                    help for unpack
          --prefix string           Prefix for the final expanded directory
          --unpack-version string   which version is contained in the tarball (mandatory)
    
    

The easiest command is *deploy single*, which installs a single sandbox.

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
    
    

If you want more than one sandbox of the same version, without any replication relationship, use the *multiple* command with an optional "--node" flag (default: 3).

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
    
    

The *replication* command will install a master and two or more slaves, with replication started. You can change the topology to "group" and get three nodes in peer replication, or compose multi-source topologies with *all-masters* or *fan-in*.

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
          --master-list string   Which nodes are masters in a multi-source deployment (default "1 2")
      -n, --nodes int            How many nodes will be installed (default 3)
          --semi-sync            Use semi-synchronous plugin
          --single-primary       Using single primary for group replication
          --slave-list string    Which nodes are slaves in a multi-source deployment (default "3")
      -t, --topology string      Which topology will be installed (default "master-slave")
    
    

## Multiple sandboxes, same version and type

If you want to deploy several instances of the same version and the same type (for example two single sandboxes of 8.0.4, or two group replication instances with different single-primary setting) you can specify the data directory name and the ports manually.

    $ dbdeployer deploy single 8.0.4
    # will deploy in msb_8_0_4 using port 8004

    $ dbdeployer deploy single 8.0.4 --sandbox-directory=msb2_8_0_4 --port=8005
    # will deploy in msb2_8_0_4 using port 8005

    $ dbdeployer deploy replication 8.0.4 --sandbox-directory=rsandbox2_8_0_4 --base-port=18600
    # will deploy replication in rsandbox2_8_0_4 using ports 18601, 18602, 18603

## Concurrent deployment and deletion

Starting with version 0.3.0, dbdeployer can deploy groups of sandboxes (*replication*, *multiple*) with the flag ``--concurrent``. When this flag is used, dbdeployed will run operations concurrently.
The same flag can be used with the *delete* command. It is useful when there are several sandboxes to be deleted at once.
Concurrent operations run from 2 to 5 times faster than sequential ones, depending on the version of the server and the number of nodes.

## Replication topologies

Multiple sandboxes can be deployed using replication with several topologies (using ``dbdeployer deploy replication --topology=xxxxx``:

* **master-slave** is the default topology. It will install one master and two slaves. More slaves can be added with the option ``--nodes``.
* **group** will deploy three peer nodes in group replication. If you want to use a single primary deployment, add the option ``--single-primary``. Available for MySQL 5.7 and later.
* **fan-in** is the opposite of master-slave. Here we have one slave and several masters. This topology requires MySQL 5.7 or higher.
**all-masters** is a special case of fan-in, where all nodes are masters and are also slaves of all nodes.

It is possible to tune the flow of data in multi-source topologies. The default for fan-in is three nodes, where 1 and 2 are masters, and 2 are slaves. You can change the predefined settings by providing the list of components:

    $ dbdeployer deploy replication --topology=fan-in \
        --nodes=5 --master-list="1,2 3" \
        --slave-list="4,5" 8.0.4 \
        --concurrent
In the above example, we get 5 nodes instead of 3. The first three are master (``--master-list="1,2,3"``) and the last two are slaves (``--slave-list="4,5"``) which will receive data from all the masters. There is a test automatically generated to check the replication flow. In our case it shows the following:

    $ ~/sandboxes/fan_in_msb_8_0_4/test_replication
    # master 1
    # master 2
    # master 3
    # slave 4
    ok - '3' == '3' - Slaves received tables from all masters
    # slave 5
    ok - '3' == '3' - Slaves received tables from all masters
    # pass: 2
    # fail: 0

The first three lines show that each master has done something. In our case, each master has created a different table. Slaves in nodes 5 and 6 then count how many tables they found, and if they got the tables from all masters, the test succeeds.


## Sandbox customization

There are several ways of changing the default behavior of a sandbox.

1. You can add options to the sandbox being deployed using --my-cnf-options="some mysqld directive". This option can be used many times. The supplied options are added to my.sandbox.cnf
2. You can specify a my.cnf template (--my-cnf-file=filename) instead of defining options line by line. dbdeployer will skip all the options that are needed for the sandbox functioning.
3. You can run SQL statements or SQL files before or after the grants were loaded (--pre-grants-sql, --pre-grants-sql-file, etc). You can also use these options to peek into the state of the sandbox and see what is happening at every stage.
4. For more advanced needs, you can look at the templates being used for the deployment, and load your own instead of the original s(--use-template=TemplateName:FileName.)

For example:

    $ dbdeployer deploy single 5.6.33 --my-cnf-options="general_log=1" \
        --pre-grants-sql="select host, user, password from mysql.user" \
        --post-grants-sql="select @@general_log"

    $ dbdeployer defaults templates list
    $ dbdeployer defaults templates show templateName > mytemplate.txt
    # edit the template
    $ dbdeployer deploy single --use-template=templateName:mytemplate.txt 5.7.21

dbdeployer will use your template instead of the original.

5. You can also export the templates, edit them, and ask dbdeployer to edit your changes.

Example:

    $ dbdeployer defaults templates export single my_templates
    # Will export all the templates for the "single" group to the direcory my_templates/single
    $ dbdeployer defaults templates export ALL my_templates
    # exports all templates into my_templates, one directory for each group
    # Edit the templates that you want to change. You can also remove the ones that you want to leave untouched.
    $ dbdeployer defaults templates import single my_templates
    # Will import all templates from my_templates/single

Warning: modifying templates may block the regular work of the sandboxes. Use this feature with caution!

6. Finally, you can modify the defaults for the application, using the "defaults" command. You can export the defaults, import them from a modified JSON file, or update a single one on-the-fly.

Here's how:

    $ dbdeployer defaults show
    # Internal values:
    {
     	"version": "0.3.7",
     	"sandbox-home": "$HOME/sandboxes",
     	"sandbox-binary": "$HOME/opt/mysql",
     	"use-sandbox-catalog": true,
     	"master-slave-base-port": 11000,
     	"group-replication-base-port": 12000,
     	"group-replication-sp-base-port": 13000,
     	"fan-in-replication-base-port": 14000,
     	"all-masters-replication-base-port": 15000,
     	"multiple-base-port": 16000,
     	"group-port-delta": 125,
     	"master-name": "master",
     	"master-abbr": "m",
     	"node-prefix": "node",
     	"slave-prefix": "slave",
     	"slave-abbr": "s",
     	"sandbox-prefix": "msb_",
     	"master-slave-prefix": "rsandbox_",
     	"group-prefix": "group_msb_",
     	"group-sp-prefix": "group_sp_msb_",
     	"multiple-prefix": "multi_msb_",
     	"fan-in-prefix": "fan_in_msb_",
     	"all-masters-prefix": "all_masters_msb_"
     }
    

    $ dbdeployer defaults update master-slave-base-port 15000
    # Updated master-slave-base-port -> "15000"
    # Configuration file: $HOME/.dbdeployer/config.json
    {
     	"version": "0.3.7",
     	"sandbox-home": "$HOME/sandboxes",
     	"sandbox-binary": "$HOME/opt/mysql",
     	"use-sandbox-catalog": true,
     	"master-slave-base-port": 15000,
     	"group-replication-base-port": 12000,
     	"group-replication-sp-base-port": 13000,
     	"fan-in-replication-base-port": 14000,
     	"all-masters-replication-base-port": 15000,
     	"multiple-base-port": 16000,
     	"group-port-delta": 125,
     	"master-name": "master",
     	"master-abbr": "m",
     	"node-prefix": "node",
     	"slave-prefix": "slave",
     	"slave-abbr": "s",
     	"sandbox-prefix": "msb_",
     	"master-slave-prefix": "rsandbox_",
     	"group-prefix": "group_msb_",
     	"group-sp-prefix": "group_sp_msb_",
     	"multiple-prefix": "multi_msb_",
     	"fan-in-prefix": "fan_in_msb_",
     	"all-masters-prefix": "all_masters_msb_"
     }
    

Another way of modifying the defaults, which does not store the new values in dbdeployer's configuration file, is through the ``--defaults`` flag. The above change could be done like this:

    $ dbdeployer --defaults=master-slave-base-port:15000 \
        deploy replication 5.7.21

The difference is that using ``dbdeployer defaults update`` the value is changed permanently for the next commands, or until you run a ``dbdeployer defaults reset``. Using the ``--defaults`` flag, instead, will modify the defaults only for the active command.

## Sandbox management

You can list the available MySQL versions with

    $ dbdeployer versions

Also "available" is a recognized alias for this command.

And you can list which sandboxes were already installed

    $ dbdeployer installed  # Aliases: sandboxes, deployed

The command "usage" shows how to use the scripts that were installed with each sandbox.

    $ dbdeployer usage
    
    	USING A SANDBOX
    
    Change directory to the newly created one (default: $SANDBOX_HOME/msb_VERSION 
    for single sandboxes)
    [ $SANDBOX_HOME = $HOME/sandboxes unless modified with flag --sandbox-home ]
    
    The sandbox directory of the instance you just created contains some handy 
    scripts to manage your server easily and in isolation.
    
    "./start", "./status", "./restart", and "./stop" do what their name suggests. 
    start and restart accept parameters that are eventually passed to the server. 
    e.g.:
    
      ./start --server-id=1001
    
      ./restart --event-scheduler=disabled
    
    "./use" calls the command line client with the appropriate parameters,
    Example:
    
        ./use -BN -e "select @@server_id"
        ./use -u root
    
    "./clear" stops the server and removes everything from the data directory, 
    letting you ready to start from scratch. (Warning! It's irreversible!)
    
    "./my" is a prefix script to invoke any command named "my*" from the 
    MySQL /bin directory. It is important to use it rather than the 
    corresponding globally installed tool, because this guarantees 
    that you will be using the tool for the version you have deployed.
    Examples:
    
        ./my sqldump db_name
        ./my sqlbinlog somefile
    
     USING MULTIPLE SERVER SANDBOX
    On a replication sandbox, you have the same commands (run "dbdeployer usage single"), 
    with an "_all" suffix, meaning that you propagate the command to all the members. 
    Then you have "./m" as a shortcut to use the master, "./s1" and "./s2" to access 
    the slaves (and "s3", "s4" ... if you define more).
    
    In group sandboxes without a master slave relationship (group replication and 
    multiple sandboxes) the nodes can be accessed by ./n1, ./n2, ./n3, and so on.
    
    start_all
    status_all
    restart_all
    stop_all
    use_all
    clear_all
    m
    s1, s2, n1, n2
    
    The scripts "check_slaves" or "check_nodes" give the status of replication in the sandbox.
    
    

## Sandbox macro operations

You can run a command in several sandboxes at once, using the *global* command, which propagates your command to all the installed sandboxes.

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
    
    

The sandboxes can also be deleted, either one by one or all at once:

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
    
    

You can lock one or more sandboxes to prevent deletion. Use this command to make the sandbox non-deletable.

    $ dbdeployer admin lock sandbox_name

A locked sandbox will not be deleted, even when running "dbdeployer delete ALL."

The lock can also be reverted using

    $ dbdeployer admin unlock sandbox_name

