# dbdeployer
[DBdeployer](https://github.com/datacharmer/dbdeployer) is a tool that deploys MySQL database servers easily.
This is a port of [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox), originally written in Perl, and re-designed from the ground up in [Go](https://golang.org). See the [features comparison](https://github.com/datacharmer/dbdeployer/blob/master/docs/features.md) for more detail.

Documentation updated for version 1.58.2 (20-Dec-2020 14:50 UTC)

![Build Status](https://github.com/datacharmer/dbdeployer/workflows/.github/workflows/all_tests.yml/badge.svg)

- [Installation](#installation)
    - [Manual installation](#manual-installation)
    - [Installation via script](#installation-via-script)
- [Prerequisites](#prerequisites)
- [Initializing the environment](#initializing-the-environment)
- [Updating dbdeployer](#updating-dbdeployer)
- [Main operations](#main-operations)
    - [Overview](#overview)
    - [Unpack](#unpack)
    - [Deploy single](#deploy-single)
    - [Deploy multiple](#deploy-multiple)
    - [Deploy replication](#deploy-replication)
    - [Re-deploy a sandbox](#re-deploy-a-sandbox)
- [Database users](#database-users)
- [Database server flavors](#database-server-flavors)
- [Getting remote tarballs](#getting-remote-tarballs)
    - [Looking at the available tarballs](#looking-at-the-available-tarballs)
    - [Getting a tarball](#getting-a-tarball)
    - [Customizing the tarball list](#customizing-the-tarball-list)
    - [Changing the tarball list permanently](#changing-the-tarball-list-permanently)
    - [From remote tarball to ready to use in one step](#from-remote-tarball-to-ready-to-use-in-one-step)
    - [Guessing the latest MySQL version](#guessing-the-latest-mysql-version)
- [Practical examples](#practical-examples)
- [Standard and non-standard basedir names](#standard-and-non-standard-basedir-names)
- [Using short version numbers](#using-short-version-numbers)
- [Multiple sandboxes, same version and type](#multiple-sandboxes,-same-version-and-type)
- [Using the direct path to the expanded tarball](#using-the-direct-path-to-the-expanded-tarball)
- [Ports management](#ports-management)
- [Concurrent deployment and deletion](#concurrent-deployment-and-deletion)
- [Replication topologies](#replication-topologies)
- [Skip server start](#skip-server-start)
- [MySQL Document store, mysqlsh, and defaults.](#mysql-document-store,-mysqlsh,-and-defaults.)
- [Installing MySQL shell](#installing-mysql-shell)
- [Database logs management.](#database-logs-management.)
- [dbdeployer operations logging](#dbdeployer-operations-logging)
- [Sandbox customization](#sandbox-customization)
- [Sandbox management](#sandbox-management)
- [Sandbox macro operations](#sandbox-macro-operations)
    - [dbdeployer global exec](#dbdeployer-global-exec)
    - [dbdeployer global use](#dbdeployer-global-use)
- [Sandbox deletion](#sandbox-deletion)
- [Default sandbox](#default-sandbox)
- [Using the latest sandbox](#using-the-latest-sandbox)
- [Sandbox upgrade](#sandbox-upgrade)
- [Dedicated admin address](#dedicated-admin-address)
- [Loading sample data into sandboxes](#loading-sample-data-into-sandboxes)
- [Running sysbench](#running-sysbench)
- [Obtaining sandbox metadata](#obtaining-sandbox-metadata)
- [Replication between sandboxes](#replication-between-sandboxes)
    - [a. NDB to NDB](#a.-ndb-to-ndb)
    - [b. Group replication to group replication](#b.-group-replication-to-group-replication)
    - [c. Master/slave to master/slave.](#c.-master/slave-to-master/slave.)
    - [d. Hybrid replication](#d.-hybrid-replication)
    - [e. Cloning](#e.-cloning)
- [Using dbdeployer in scripts](#using-dbdeployer-in-scripts)
- [Importing databases into sandboxes](#importing-databases-into-sandboxes)
- [Cloning databases](#cloning-databases)
- [Compiling dbdeployer](#compiling-dbdeployer)
- [Generating additional documentation](#generating-additional-documentation)
- [Command line completion](#command-line-completion)
- [Using dbdeployer source for other projects](#using-dbdeployer-source-for-other-projects)
- [Exporting dbdeployer structure](#exporting-dbdeployer-structure)
- [Semantic versioning](#semantic-versioning)
- [Do not edit](#do-not-edit)
# Installation

## Manual installation

The installation is simple, as the only thing you will need is a binary executable for your operating system.
Get the one for your O.S. from [dbdeployer releases](https://github.com/datacharmer/dbdeployer/releases) and place it in a directory in your $PATH.
(There are no binaries for Windows. See the [features list](https://github.com/datacharmer/dbdeployer/blob/master/docs/features.md) for more info.)

For example:

    $ VERSION=1.58.2
    $ OS=linux
    $ origin=https://github.com/datacharmer/dbdeployer/releases/download/v$VERSION
    $ wget $origin/dbdeployer-$VERSION.$OS.tar.gz
    $ tar -xzf dbdeployer-$VERSION.$OS.tar.gz
    $ chmod +x dbdeployer-$VERSION.$OS
    $ sudo mv dbdeployer-$VERSION.$OS /usr/local/bin/dbdeployer

## Installation via script

You can download the [installation script](https://raw.githubusercontent.com/datacharmer/dbdeployer/master/scripts/dbdeployer-install.sh), and run it in your computer.
The script will find the latest version, download the corresponding binaries, check the SHA256 checksum, and - if given privileges - copy the executable to a directory within `$PATH`.

```
$ curl -s https://raw.githubusercontent.com/datacharmer/dbdeployer/master/scripts/dbdeployer-install.sh | bash
```

# Prerequisites

Of course, there are **prerequisites**: your machine must be able to run the MySQL server. Be aware that version 5.5 and higher require some libraries that are not installed by default in all flavors of Linux (libnuma, libaio.)

As of version 1.40.0, dbdeployer tries to detect whether the host has the necessary libraries installed. When missing libraries are detected, the deployment fails with an error showing the missing pieces.
For example:

```
# dbdeployer deploy single 5.7
# 5.7 => 5.7.27
error while filling the sandbox definition: missing libraries will prevent MySQL from deploying correctly
client (/root/opt/mysql/5.7.27/bin/mysql): [	libncurses.so.5 => not found 	libtinfo.so.5 => not found]

server (/root/opt/mysql/5.7.27/bin/mysqld): [	libaio.so.1 => not found 	libnuma.so.1 => not found]
global: [libaio libnuma]

Use --skip-library-check to skip this check
```

If you use `--skip-library-check`, the above check won't be performed, and the deployment may fail and leave you with an incomplete sandbox.
Skipping the check may be justified when deploying a very old version of MySQL (4.1, 5.0, 5.1)


# Initializing the environment

Immediately after installing dbdeployer, you can get the environment ready for operations using the command

```
$ dbdeployer init
```

This command creates the necessary directories, then downloads the latest MySQL binaries, and expands them in the right place. It also enables [command line completion](#command-line-completion).

Running the command without options is what most users need. Advanced ones may look at the documentation to fine tune the initialization.

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
    
    


# Updating dbdeployer

Starting with version 1.36.0, dbdeployer is able to update itself by getting the newest release from GitHub.

The quickest way of doing it is by running 
```
$ dbdeployer update
```

This command will download the latest release of dbdeployer from GitHub, and, if the version of the release is higher than the local one, will overwrite the dbdeployer executable.

You can get more information during the operation by using the `--verbose` option. Other options are available for advanced users.

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
    
    


You can also see the details of a release using `dbdeployer info releases latest`.


# Main operations

(See this ASCIIcast for a demo of its operations.)
[![asciicast](https://asciinema.org/a/165707.png)](https://asciinema.org/a/165707)

## Overview

With dbdeployer, you can deploy a single sandbox, or many sandboxes  at once, with or without replication.

The main command is ``deploy`` with its subcommands ``single``, ``replication``, and ``multiple``, which work with MySQL tarball that have been unpacked into the _sandbox-binary_ directory (by default, $HOME/opt/mysql.)

To use a tarball, you must first run the ``unpack`` command, which will unpack the tarball into the right directory.

For example:

    $ dbdeployer unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4
    .........100.........200.........292

    $ dbdeployer deploy single 8.0.4
    Database installed in $HOME/sandboxes/msb_8_0_4
    . sandbox server started


The program doesn't have any dependencies. Everything is included in the binary. Calling *dbdeployer* without arguments or with ``--help`` will show the main help screen.

    $ dbdeployer --version
    dbdeployer version 1.58.2
    

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
          --shell-path string       Path to Bash, used for generated scripts (default "/usr/local/bin/bash")
          --skip-library-check      Skip check for needed libraries (may cause nasty errors)
      -v, --version                 version for dbdeployer

    Use "dbdeployer [command] --help" for more information about a command.


The flags listed in the main screen can be used with any commands.
The flags ``--my-cnf-options`` and ``--init-options`` can be used several times.

## Unpack

If you don't have any tarballs installed in your system, you should first ``unpack`` it (see an example above).

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
    
    

## Deploy single

The easiest command is ``deploy single``, which installs a single sandbox.

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
    
    

## Deploy multiple

If you want more than one sandbox of the same version, without any replication relationship, use the ``deploy multiple`` command with an optional ``--nodes`` flag (default: 3).

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
    
    

## Deploy replication

The ``deploy replication`` command will install a master and two or more slaves, with replication started. You can change the topology to *group* and get three nodes in peer replication, or compose multi-source topologies with *all-masters* or *fan-in*.

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
    
    

As of version 1.21.0, you can use Percona Xtradb Cluster tarballs to deploy replication of type *pxc*. This deployment only works on Linux.

## Re-deploy a sandbox

If you run a deploy statement a second time, the command will fail, because the sandbox exists already. To overcome the restriction, you can repeat the operation with the option `--force`.

```
$ dbdeployer deploy  single 8.0
# 8.0 => 8.0.22
Database installed in $HOME/sandboxes/msb_8_0_22
run 'dbdeployer usage single' for basic instructions'
. sandbox server started

$ dbdeployer deploy  single 8.0
# 8.0 => 8.0.22
error creating sandbox: 'check directory directory $HOME/sandboxes/msb_8_0_22 already exists. Use --force to override'

$ dbdeployer deploy  single 8.0 --force
# 8.0 => 8.0.22
Overwriting directory $HOME/sandboxes/msb_8_0_22
stop $HOME/sandboxes/msb_8_0_22
Database installed in $HOME/sandboxes/msb_8_0_22
run 'dbdeployer usage single' for basic instructions'
. sandbox server started
```

For a quicker reset, all single sandboxes have a command `wipe_and_restart`. The replication sandboxes (except NDB and PXC) have a command `wipe_and_restart_all`.
Running such command will delete the data directory in all nodes, and re-create them.

```
$ ~/sandboxes/msb_8_0_22/wipe_and_restart
Terminating the server immediately --- kill -9 72889
[...]
Database installed in /home/gmax/sandboxes/msb_8_0_22
. sandbox server started

$ ~/sandboxes/rsandbox_8_0_22/wipe_and_restart_all
# executing 'send_kill' on /home/gmax/sandboxes/rsandbox_8_0_22
[...]
Database installed in /home/gmax/sandboxes/rsandbox_8_0_22/master
[...]
Database installed in /home/gmax/sandboxes/rsandbox_8_0_22/node1
[...]
Database installed in /home/gmax/sandboxes/rsandbox_8_0_22/node2
[...]
# executing 'start' on /home/gmax/sandboxes/rsandbox_8_0_22
executing 'start' on master
. sandbox server started
executing 'start' on slave 1
. sandbox server started
executing 'start' on slave 2
. sandbox server started
initializing slave 1
initializing slave 2
```

# Database users

The default users for each server deployed by dbdeployer are:

* `root`, with the default grants as given by the server version being installed. 
* `msandbox`, with all privileges except GRANT option.
* `msandbox_rw`, with minimum read/write privileges.
* `msandbox_ro`, with read-only privileges.
* `rsandbox`, with only replication related privileges (password: `rsandbox`)

The main user name (`msandbox`) and password (`msandbox`) can be changed using options `--db-user` and `db-password` respectively.

Every user is assigned by default to a limited scope (`127.%`) so that they can only communicate with the local host.
The scope can be changed using options `--bind-address` and `--remote-access`.

In MySQL 8.0 the above users are instantiated using roles. You can also define a custom role, and assign it to the main user.

You can create a different role and assign it to the default user with options like the following:

```
dbdeployer deploy single 8.0.19 \
    --custom-role-name=R_POWERFUL \
    --custom-role-privileges='ALL PRIVILEGES' \
    --custom-role-target='*.*' \
    --custom-role-extra='WITH GRANT OPTION' \
    --default-role=R_POWERFUL \
    --bind-address=0.0.0.0 \
    --remote-access='%' \
    --db-user=differentuser \
    --db-password=somethingdifferent
```

The result of this operation will be:

```
$ ~/sandboxes/msb_8_0_19/use
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 9
Server version: 8.0.19 MySQL Community Server - GPL

Copyright (c) 2000, 2020, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql [localhost:8019] {differentuser} ((none)) > show grants\G
*************************** 1. row ***************************
Grants for differentuser@localhost: GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, 
DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, 
SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, 
CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, 
CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.* TO `differentuser`@`localhost` WITH GRANT OPTION
*************************** 2. row ***************************
Grants for differentuser@localhost: GRANT APPLICATION_PASSWORD_ADMIN,AUDIT_ADMIN,BACKUP_ADMIN,BINLOG_ADMIN,
BINLOG_ENCRYPTION_ADMIN,CLONE_ADMIN,CONNECTION_ADMIN,ENCRYPTION_KEY_ADMIN,GROUP_REPLICATION_ADMIN,
INNODB_REDO_LOG_ARCHIVE,PERSIST_RO_VARIABLES_ADMIN,REPLICATION_APPLIER,REPLICATION_SLAVE_ADMIN,
RESOURCE_GROUP_ADMIN,RESOURCE_GROUP_USER,ROLE_ADMIN,SERVICE_CONNECTION_ADMIN,SESSION_VARIABLES_ADMIN,
SET_USER_ID,SYSTEM_USER,SYSTEM_VARIABLES_ADMIN,TABLE_ENCRYPTION_ADMIN,XA_RECOVER_ADMIN 
ON *.* TO `differentuser`@`localhost` WITH GRANT OPTION
*************************** 3. row ***************************
Grants for differentuser@localhost: GRANT `R_POWERFUL`@`%` TO `differentuser`@`localhost`
3 rows in set (0.01 sec)
```

Instead of assigning the custom role to the default user, you can also create a task user.

```
$ dbdeployer deploy single 8.0 \
  --task-user=task_user \
  --custom-role-name=R_ADMIN \
  --task-user-role=R_ADMIN 
```

The options shown in this section only apply to MySQL 8.0.

There is a method of creating users during deployment in any versions:

1. create a SQL file containing the `CREATE USER` and `GRANT` statements you want to run
2. use the option `--post-grants-sql-file` to load the instructions.

```
cat << EOF > orchestrator.sql

CREATE DATABASE IF NOT EXISTS orchestrator;
CREATE USER orchestrator IDENTIFIED BY 'msandbox';
GRANT ALL PRIVILEGES ON orchestrator.* TO orchestrator;
GRANT SELECT ON mysql.slave_master_info TO orchestrator;

EOF

$ dbdeployer deploy single 5.7 \
  --post-grants-sql-file=$PWD/orchestrator.sql
```

# Database server flavors

Before version 1.19.0, dbdeployer assumed that it was dealing to some version of MySQL, using the version to decide which features it would support. In version 1.19.0 dbdeployer started using the concept of **capabilities**, which is a combination of server **flavor** + a version. Some flavors currently supported are

* `mysql` : the classic MySQL server
* `percona` : Percona Server, any version. For the purposes of deployment, it has the same capabilities as MySQL
* `mariadb`: MariaDB server. Mostly the same as MySQL, but with differences in deployment methods.
* `pxc`: Percona Xtradb Cluster
* `ndb`: MySQL Cluster (NDB)
* `tidb`: A stand-alone TiDB server.

To see what every flavor can do, you can use the command `dbdeployer admin capabilities`.

To see the features of a given flavor: `dbdeployer admin capabilities FLAVOR`.

And to see what a given version of a flavor can do, you can use `dbdeployer admin capabilities FLAVOR VERSION`.

For example

```shell
$ dbdeployer admin capabilities

$ dbdeployer admin capabilities percona

$ dbdeployer admin capabilities mysql 5.7.11
$ dbdeployer admin capabilities mysql 5.7.13
```

# Getting remote tarballs

**NOTE:** As of version 1.31.0, `dbdeployer remote` is **DEPRECATED** and its functionality is replaced by `dbdeployer downloads`.

As of version 1.31.0, dbdeployer can download remote tarballs of various flavors from several locations. Tarballs are listed for Linux and MacOS.

## Looking at the available tarballs

```
$ dbdeployer downloads list
Available tarballs
                          name                              OS     version   flavor     size   minimal
-------------------------------------------------------- -------- --------- -------- -------- ---------
 tidb-master-darwin-amd64.tar.gz                          Darwin     3.0.0   tidb      26 MB
 tidb-master-linux-amd64.tar.gz                           Linux      3.0.0   tidb      26 MB
 mysql-5.7.26-macos10.14-x86_64.tar.gz                    Darwin    5.7.26   mysql    337 MB
 mysql-8.0.16-macos10.14-x86_64.tar.gz                    Darwin    8.0.16   mysql    153 MB
 mysql-8.0.15-macos10.14-x86_64.tar.gz                    Darwin    8.0.15   mysql    139 MB
 mysql-5.7.25-macos10.14-x86_64.tar.gz                    Darwin    5.7.25   mysql    337 MB
 mysql-5.6.41-macos10.13-x86_64.tar.gz                    Darwin    5.6.41   mysql    176 MB
 mysql-5.5.53-osx10.9-x86_64.tar.gz                       Darwin    5.5.53   mysql    114 MB
 mysql-5.1.73-osx10.6-x86_64.tar.gz                       Darwin    5.1.73   mysql     82 MB
 mysql-5.0.96-osx10.5-x86_64.tar.gz                       Darwin    5.0.96   mysql     61 MB
 mysql-8.0.16-linux-glibc2.12-x86_64.tar.xz               Linux     8.0.16   mysql    461 MB
 mysql-8.0.16-linux-x86_64-minimal.tar.xz                 Linux     8.0.16   mysql     44 MB   Y
[...]
```
The list is kept internally by dbdeployer, but it can be exported, edited, and reloaded (more on that later).


## Getting a tarball

We can download one of the listed tarballs in two ways:

* using `dbdeployer downloads get file_name`, where we copy and paste the file name from the list above. For example: `dbdeployer downloads get mysql-8.0.16-linux-glibc2.12-x86_64.tar.xz`.
* using `dbdeployer downloads get-by-version VERSION [options]` where we use several criteria to identify the file we want.

For example:

```
$ dbdeployer downloads get-by-version 5.7 --newest --dry-run
Would download:

Name:          mysql-5.7.26-macos10.14-x86_64.tar.gz
Short version: 5.7
Version:       5.7.26
Flavor:        mysql
OS:            Darwin
URL:           https://dev.mysql.com/get/Downloads/MySQL-5.7/mysql-5.7.26-macos10.14-x86_64.tar.gz
Checksum:      SHA512:ae84b0cfe3cf274fc79adb3db03b764d47033aea970cc26edcdd4adbe5b2e3d28bf4f98f2ee321f16e788d69cbe3a08bf39fa5329d8d7a67bee928d964891ed8
Size:          337 MB

$ dbdeployer downloads get-by-version 8.0 --newest --dry-run
Would download:

Name:          mysql-8.0.16-macos10.14-x86_64.tar.gz
Short version: 8.0
Version:       8.0.16
Flavor:        mysql
OS:            Darwin
URL:           https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.16-macos10.14-x86_64.tar.gz
Checksum:      SHA512:30fb86c929ad1f384622277dbc3d686f5530953a8f7e2c7adeb183768db69464e93a46b4a0ec212d006e069f1b93db0bd0a51918eaa7e3697ea227d86082d892
Size:          153 MB
```
The above commands, executed on MacOS, look for tarballs for the current operating system, and gets the one with the highest version. Notice the option `--dry-run`, which shows what would be downloaded, but without actually doing it.

If there are multiple files that match the search criteria, dbdeployer returns an error.
```
$ dbdeployer downloads get-by-version 8.0 --newest --OS=linux --dry-run
tarballs mysql-8.0.16-linux-x86_64-minimal.tar.xz and mysql-8.0.16-linux-glibc2.12-x86_64.tar.xz have the same version - Get the one you want by name
```

In this case, we can fix the error by adding another parameter:

```
$ dbdeployer downloads get-by-version 8.0 --newest --OS=linux  --minimal --dry-run
Would download:

Name:          mysql-8.0.16-linux-x86_64-minimal.tar.xz
Short version: 8.0
Version:       8.0.16
Flavor:        mysql
OS:            Linux
URL:           https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.16-linux-x86_64-minimal.tar.xz
Checksum:      MD5: 7bac88f47e648bf9a38e7886e12d1ec5
Size:          44 MB

# On Linux

$ dbdeployer downloads get-by-version 8.0 --newest --OS=linux  --minimal
....  44 MB
File /home/gmax/tmp/mysql-8.0.16-linux-x86_64-minimal.tar.xz downloaded
Checksum matches
```

If we download a tarball that is not intended for the current operating system, we will get a warning:

```
# On MacOS
$ dbdeployer downloads get-by-version 8.0 --newest --OS=linux  --minimal
....  44 MB
File /Users/gmax/go/src/github.com/datacharmer/dbd-ui/mysql-8.0.16-linux-x86_64-minimal.tar.xz downloaded
Checksum matches
################################################################################
WARNING: Current OS is darwin, but the tarball's OS is linux
################################################################################
```

We can also add the tarball flavor to get yet a different result from the above criteria:

```
$ dbdeployer downloads get-by-version 8.0 --newest   --flavor=ndb --dry-run
Would download:

Name:          mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz
Short version: 8.0
Version:       8.0.16
Flavor:        ndb
OS:            Linux
URL:           https://dev.mysql.com/get/Downloads/MySQL-Cluster-8.0/mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz
Checksum:      SHA512:a587a774cc7a8f6cbe295272f0e67869c5077b8fb56917e0dc2fa0ea1c91548c44bd406fcf900cc0e498f31bb7188197a3392aa0d7df8a08fa5e43901683e98a
Size:          1.1 GB
```

    $ dbdeployer downloads get --help
    Downloads a remote tarball
    
    Usage:
      dbdeployer downloads get tarball_name [options] [flags]
    
    Flags:
          --dry-run             Show what would be downloaded, but don't run it
      -h, --help                help for get
          --progress-step int   Progress interval (default 10485760)
          --quiet               Do not show download progress
    
    
    $ dbdeployer downloads get-by-flavor --help
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
    
    


## Customizing the tarball list

The tarball list is embedded in dbdeployer, but it can be modified with a few steps:

1. Run `dbdeployer downloads export mylist.json --add-empty-item`
2. Edit `mylist.json`, by filling the fields left empty:

```
        {
            "name": "mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz",
            "checksum": "SHA512:a587a774cc7a8f6cbe295272f0e67869c5077b8fb56917e0dc2fa0ea1c91548c44bd406fcf900cc0e498f31bb7188197a3392aa0d7df8a08fa5e43901683e98a",
            "OS": "Linux",
            "url": "https://dev.mysql.com/get/Downloads/MySQL-Cluster-8.0/mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz",
            "flavor": "ndb",
            "minimal": false,
            "size": 1100516061,
            "short_version": "8.0",
            "version": "8.0.16"
        },
        {
            "name": "FillIt",
            "OS": "",
            "url": "",
            "flavor": "",
            "minimal": false,
            "size": 0,
            "short_version": "",
            "version": "",
            "updated_by": "Fill it",
            "notes": "Fill it"
        }
```

3. Run `dbdeployer downloads import mylist.json`

The file will be saved into dbdeployer custom directory (`$HOME/.dbdeployer`), but only if the file validates , but only if the file validates 
If not, an error message will show what changes are needed.

If you don't need the customized list any longer, you can remove it using `dbdeployer downloads reset`: the custom file will be removed from dbdeployer directory and the embedded one will be used again.

## Changing the tarball list permanently

Adding tarballs to a personal list could be time consuming, if you need to do it often. A better way is to clone this repository, then modify the [original list](https://github.com/datacharmer/dbdeployer/blob/master/downloads/tarball_list.json), and then open a pull request with the changes. The list is used when building dbdeployer, as the contents of the JSON file are converted into an internal list.

When entering a new tarball, it is important to fill all the details needed to identify the download. The checksum field is very important. as it is what makes sure that the file downloaded is really the original one.

dbdeployer can calculate checksums for `MD5` (currently used in MySQL downloads pages), `SHA512` (used in most of the downloads listed in version 1.31.0), as well as `SHA1` and `SHA256`. To communicate which checksum is being used, the checksum string must be prefixed by the algorithm, such as `MD5:7bac88f47e648bf9a38e7886e12d1ec5`. An optional space before and after the colon (`:`) is accepted.

## From remote tarball to ready to use in one step

dbdeployer 1.33.0 adds a command `dbdeployer downloads get-unpack tarball_name` which combines the effects of `dbdeployer get tarball_name` followed by `dbdeployer unpack tarball_name`. This command accepts all options defined for `unpack`, so that you can optionally indicate the tarball flavor and version, whether to overwrite it, and if you want to delete the tarball after the operation.

```
$ dbdeployer downloads get-unpack \
   mysql-8.0.16-linux-x86_64-minimal.tar.xz \
   --overwrite \
   --delete-after-unpack
Downloading mysql-8.0.16-linux-x86_64-minimal.tar.xz
....  44 MB
File mysql-8.0.16-linux-x86_64-minimal.tar.xz downloaded
Checksum matches
Unpacking tarball mysql-8.0.16-linux-x86_64-minimal.tar.xz to $HOME/opt/mysql/8.0.16
.........100.........200.219
Renaming directory $HOME/opt/mysql/mysql-8.0.16-linux-x86_64-minimal to $HOME/opt/mysql/8.0.16

$ dbdeployer downloads get-unpack \
  mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz \
  --flavor=ndb \
  --prefix=ndb \
  --overwrite \
  --delete-after-unpack
Downloading mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz
.........105 MB.........210 MB.........315 MB.........419 MB.........524 MB
.........629 MB.........734 MB.........839 MB.........944 MB.........1.0 GB....  1.1 GB
File mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz downloaded
Checksum matches
Unpacking tarball mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/ndb8.0.16
[...]
Renaming directory $HOME/opt/mysql/mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64 to $HOME/opt/mysql/ndb8.0.16
```

## Guessing the latest MySQL version

If you know that a new version of MySQL is available, but you don't have such version in the downloads list, you can try a shortcut with the command `dbdeployer downloads get-by-version 8.0 --guess-latest`
(Available in version 1.41.0)

When you use `--guess-latest`, dbdeployer looks for the latest download available in the list, increases the version by 1, and tries to get the tarball from MySQL downloads page.

For example, if the latest version in the tarballs list is `8.0.21`, and you know that 8.0.22 has just been released, you can run the command

```
$ dbdeployer downloads get-by-version --guess-latest 8.0 --dry-run
Would download:

Name:          mysql-8.0.22-macos10.14-x86_64.tar.gz
Short version: 8.0
Version:       8.0.22
Flavor:        mysql
OS:            darwin
URL:           https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.22-macos10.14-x86_64.tar.gz
Checksum:
Size:          0 B
Notes:         guessed
```

Without `--dry-run`, it would attempt downloading MySQL 8.0.22. If the download is not available, you will get an error:

```
$ dbdeployer downloads get-by-version --guess-latest 8.0
Guessed mysql-8.0.22-macos10.14-x86_64.tar.gz file not ready for download
```

Beware: when the download happens, there is no checksum to perform. Use this feature with caution.

```
$ dbdeployer downloads get-by-version --guess-latest 8.0
Downloading mysql-8.0.22-macos10.14-x86_64.tar.gz
.........105 MB.....  166 MB
File $PWD/mysql-8.0.22-macos10.14-x86_64.tar.gz downloaded
No checksum to compare
```

# Practical examples

Several examples of dbdeployer usages are avaibale with the command ``dbdeployer cookbook``


    $ dbdeployer cookbook list
    .----------------------------------.-------------------------------------.---------------------------------------------------------------------------------------.--------.
    |              recipe              |             script name             |                                      description                                      | needed |
    |                                  |                                     |                                                                                       | flavor |
    +----------------------------------+-------------------------------------+---------------------------------------------------------------------------------------+--------+
    | admin                            | admin-single.sh                     | Single sandbox with admin address enabled                                             | mysql  |
    | all-masters                      | all-masters-deployment.sh           | Creation of an all-masters replication sandbox                                        | mysql  |
    | circular_replication             | circular-replication.sh             | Shows how to run replication between nodes of a multiple deployment                   | -      |
    | custom-named-replication         | custom-named-replication.sh         | Replication sandbox with custom names for directories and scripts                     | -      |
    | custom-users                     | single-custom-users.sh              | Single sandbox with custom users                                                      | mysql  |
    | delete                           | delete-sandboxes.sh                 | Delete all deployed sandboxes                                                         | -      |
    | fan-in                           | fan-in-deployment.sh                | Creation of a fan-in (many masters, one slave) replication sandbox                    | mysql  |
    | group-multi                      | group-multi-primary-deployment.sh   | Creation of a multi-primary group replication sandbox                                 | mysql  |
    | group-single                     | group-single-primary-deployment.sh  | Creation of a single-primary group replication sandbox                                | mysql  |
    | master-slave                     | master-slave-deployment.sh          | Creation of a master/slave replication sandbox                                        | -      |
    | ndb                              | ndb-deployment.sh                   | Shows deployment with ndb                                                             | ndb    |
    | prerequisites                    | prerequisites.sh                    | Shows dbdeployer prerequisites and how to make them                                   | -      |
    | pxc                              | pxc-deployment.sh                   | Shows deployment with pxc                                                             | pxc    |
    | remote                           | remote.sh                           | Shows how to get a remote MySQL tarball                                               | -      |
    | replication-operations           | repl-operations.sh                  | Show how to run operations in a replication sandbox                                   | -      |
    | replication-restart              | repl-operations-restart.sh          | Show how to restart sandboxes with custom options                                     | -      |
    | replication_between_groups       | replication-between-groups.sh       | Shows how to run replication between two group replications                           | mysql  |
    | replication_between_master_slave | replication-between-master-slave.sh | Shows how to run replication between two master/slave replications                    | -      |
    | replication_between_ndb          | replication-between-ndb.sh          | Shows how to run replication between two NDB clusters                                 | ndb    |
    | replication_between_single       | replication-between-single.sh       | Shows how to run replication between two single sandboxes                             | -      |
    | replication_group_master_slave   | replication-group-master-slave.sh   | Shows how to run replication between a group replication and master/slave replication | mysql  |
    | replication_group_single         | replication-group-single.sh         | Shows how to run replication between a group replication and a single sandbox         | mysql  |
    | replication_master_slave_group   | replication-master-slave-group.sh   | Shows how to run replication between master/slave replication and group replication   | mysql  |
    | replication_multi_versions       | replication-multi-versions.sh       | Shows how to run replication between different MySQL versions                         | -      |
    | replication_single_group         | replication-single-group.sh         | Shows how to run replication between a single sandbox an group replication            | mysql  |
    | show                             | show-sandboxes.sh                   | Show deployed sandboxes                                                               | -      |
    | single                           | single-deployment.sh                | Creation of a single sandbox                                                          | -      |
    | single-reinstall                 | single-reinstall.sh                 | Re-installs a single sandbox                                                          | -      |
    | skip-start-replication           | skip-start-replication.sh           | Replication sandbox deployed without starting the servers                             | -      |
    | skip-start-single                | skip-start-single.sh                | Single sandbox deployed without starting the server                                   | -      |
    | tidb                             | tidb-deployment.sh                  | Shows deployment and some operations with TiDB                                        | tidb   |
    | upgrade                          | upgrade.sh                          | Shows a complete upgrade example from 5.5 to 8.0                                      | mysql  |
    '----------------------------------'-------------------------------------'---------------------------------------------------------------------------------------'--------'
    

Using this command, dbdeployer can produce sample scripts for common operations.

For example `dbdeployer cookbook create single` will create the directory `./recipes` containing the script `single-deployment.sh`, using the versions available in your machine. If no versions are found, the script `prerequisites.sh` will show which steps to take.

`dbdeployer cookbook create ALL` will create all the recipe scripts .

The scripts in the `./recipes` directory show some of the most interesting ways of using dbdeployer.

Each `*deployment*` or `*operations*` script runs with this syntax:

```bash
./recipes/script_name.sh [version]
```

where `version` is `5.7.23`, or `8.0.12`, or `ndb7.6.9`, or any other recent version of MySQL. For this to work, you need to have unpacked the tarball binaries for the corresponding version. 
See `./recipes/prerequisites.sh` for practical steps.

You can run the same command several times, provided that you use a different version at every call.

```bash
./recipes/single-deployment.sh 5.7.24
./recipes/single-deployment.sh 8.0.13
```

`./recipes/upgrade.sh` is a complete example of upgrade operations. It runs an upgrade from 5.5 to 5.6, then the upgraded database is upgraded to 5.7, and finally to 8.0. Along the way, each database writes to the same table, so that you can see the effects of the upgrade.
Here's an example.
```
+----+-----------+------------+----------+---------------------+
| id | server_id | vers       | urole    | ts                  |
+----+-----------+------------+----------+---------------------+
|  1 |      5553 | 5.5.53-log | original | 2019-03-22 07:48:46 |
|  2 |      5641 | 5.6.41-log | upgraded | 2019-03-22 07:48:54 |
|  3 |      5641 | 5.6.41-log | original | 2019-03-22 07:48:59 |
|  4 |      5725 | 5.7.25-log | upgraded | 2019-03-22 07:49:09 |
|  5 |      5725 | 5.7.25-log | original | 2019-03-22 07:49:14 |
|  6 |      8015 | 8.0.15     | upgraded | 2019-03-22 07:49:25 |
+----+-----------+------------+----------+---------------------+
```
dbdeployer will detect the latest versions available in you system. If you don't have all the versions mentioned here, you should edit the script and use only the ones you want (such as 5.7.25 and 8.0.15).

    $ dbdeployer cookbook
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
    
    

# Standard and non-standard basedir names

dbdeployer expects to get the binaries from ``$HOME/opt/mysql/x.x.xx``. For example, when you run the command ``dbdeployer deploy single 8.0.11``, you must have the binaries for MySQL 8.0.11 expanded into a directory named ``$HOME/opt/mysql/8.0.11``.

If you want to keep several directories with the same version, you can differentiate them using a **prefix**:

    $HOME/opt/mysql/
                8.0.11
                lab_8.0.11
                ps_8.0.11
                myown_8.0.11

In the above cases, running ``dbdeployer deploy single lab_8.0.11`` will do what you expect, i.e. dbdeployer will use the binaries in ``lab_8.0.11`` and recognize ``8.0.11`` as the version for the database.

When the extracted tarball directory name that you want to use doesn't contain the full version number (such as ``/home/dbuser/build/path/5.7-extra``) you need to provide the version using the option ``--binary-version``. For example:

    dbdeployer deploy single 5.7-extra \
        --sandbox-binary=/home/dbuser/build/path \
        --binary-version=5.7.22

In the above command, ``--sandbox-binary`` indicates where to search for the binaries, ``5.7-extra`` is where the binaries are, and ``--binary-version`` indicates which version should be used.

Just to be clear, dbdeployer will recognize the directory as containing a version if it is only "x.x.x" or if it **ends** with "x.x.x" (as in ``lab_8.0.11``.)

# Using short version numbers

You can use, instead of a full version number (e.g. ``8.0.11``,) a short one, such as ``8.0``. This shortcut works starting with version 1.6.0.
When you invoke dbdeployer with a short number, it will look for the highest revision number within that version, and use it for deployment.

For example, if your sandbox binary directory contains the following:

    5.7.19    5.7.20    5.7.22    8.0.1    8.0.11    8.0.4

You can issue the command ``dbdeployer deploy single 8.0``, and it will use 8.0.11 for a single deployment. Or ``dbdeployer deploy replication 5.7`` and it will result in a replication system using 5.7.22 (the latest one.)


# Multiple sandboxes, same version and type

If you want to deploy several instances of the same version and the same type (for example two single sandboxes of 8.0.4, or two replication instances with different settings) you can specify the data directory name and the ports manually.

    $ dbdeployer deploy single 8.0.4
    # will deploy in msb_8_0_4 using port 8004

    $ dbdeployer deploy single 8.0.4 --sandbox-directory=msb2_8_0_4
    # will deploy in msb2_8_0_4 using port 8005 (which dbdeployer detects and uses)

    $ dbdeployer deploy replication 8.0.4 --concurrent
    # will deploy replication in rsandbox_8_0_4 using default calculated ports 19009, 19010, 19011

    $ dbdeployer deploy replication 8.0.4 \
        --gtid \
        --sandbox-directory=rsandbox2_8_0_4 \
        --base-port=18600 --concurrent
    # will deploy replication in rsandbox2_8_0_4 using ports 18601, 18602, 18603

# Using the direct path to the expanded tarball

If you have a custom organization of expanded tarballs, you may want to use the direct path to the binaries, instead of a combination of ``--sandbox-binary`` and the version name.

For example, let's assume your binaries are organized as follows:

    $HOME/opt/
             /percona/
                     /5.7.21
                     /5.7.22
                     /8.0.11
            /mysql/
                  /5.7.21
                  /5.7.22
                  /8.0.11

You can deploy a single sandbox for a Percona server version 5.7.22 using any of the following approaches:

    #1
    dbdeployer deploy single --sandbox-binary=$HOME/opt/percona 5.7.22

    #2
    dbdeployer deploy single $HOME/opt/percona/5.7.22

    #3
    dbdeployer defaults update sandbox-binary $HOME/opt/percona
    dbdeployer deploy single 5.7.22

    #4
    export SANDBOX_BINARY=$HOME/opt/percona
    dbdeployer deploy single 5.7.22

Methods #1 and #2 are equivalent. They set the sandbox binary directory temporarily to a new one, and use it for the current deployment

Methods #3 and #4  will set the sandbox binary directory permanently, with the difference that #3 is set for any invocation of dbdeployer system-wide (in a different terminal window, it will use the new value,) while #4 is set only for the current session (in a different terminal window, it will still use the default.)

Be aware that, using this kind of organization may see conflicts during deployment. For example, after installing Percona Server 5.7.22, if you want to install MySQL 5.7.22 you will need to specify a ``--sandbox-directory`` explicitly.
Instead, if you use the prefix approach defined in the "standard and non-standard basedir names," conflicts should be avoided.

# Ports management

dbdeployer will try using the default port for each sandbox whenever possible. For single sandboxes, the port will be the version number without dots: 5.7.22 will deploy on port 5722. For multiple sandboxes, the port number is defined by using a prefix number (visible in the defaults: ``dbdeployer defaults list``) + the port number + the revision number (for some topologies multiplied by 100.)
For example, single-primary group replication with MySQL 8.0.11 will compute the ports like this:

    base port = 8011 (version number) + 13000 (prefix) + 11 (revision) * 100  = 22111
    node1 port = base port + 1 = 22112
    node2 port = base port + 2 = 22113
    node3 port = base port + 2 = 22114

For group replication we need to calculate the group port, and we use the ``group-port-delta`` (= 125) to obtain it from the regular port:

    node1 group port = 22112 + 125 = 22237
    node2 group port = 22113 + 125 = 22238
    node3 group port = 22114 + 125 = 22239

For MySQL 8.0.11+, we also need to assign a port for the XPlugin, and we compute that using the regular port + the ``mysqlx-port-delta`` (=10000).

Thus, for MySQL 8.0.11 group replication deployments, you would see this listing:

    $ dbdeployer sandboxes --header
    name                   type                  version  ports
    ----------------       -------               -------  -----
    group_msb_8_0_11     : group-multi-primary    8.0.11 [20023 20148 30023 20024 20149 30024 20025 20150 30025]
    group_sp_msb_8_0_11  : group-single-primary   8.0.11 [22112 22237 32112 22113 22238 32113 22114 22239 32114]

This method makes port clashes unlikely when using the same version in different deployments, but there is a risk of port clashes when deploying many multiple sandboxes of close-by versions.
Furthermore, dbdeployer doesn't let the clash happen. Thanks to its central catalog of sandboxes, it knows which ports were already used, and will search for free ones whenever a potential clash is detected.
Bear in mind that the concept of "used" is only related to sandboxes. dbdeployer does not know if ports may be used by other applications.
You can minimize risks by telling dbdeployer which ports may be occupied. The defaults have a field ``reserved-ports``, containing the ports that should not be used. You can add to that list by modifying the defaults. For example, if you want to exclude port 7001, 10000, and 15000 from being used, you can run

    dbdeployer defaults update reserved-ports '7001,10000,15000'

or, if you want to preserve the ones that are reserved by default:

    dbdeployer defaults update reserved-ports '1186,3306,33060,7001,10000,15000'

# Concurrent deployment and deletion

Starting with version 0.3.0, dbdeployer can deploy groups of sandboxes (``deploy replication``, ``deploy multiple``) with the flag ``--concurrent``. When this flag is used, dbdeployer will run operations concurrently.
The same flag can be used with the ``delete`` command. It is useful when there are several sandboxes to be deleted at once.
Concurrent operations run from 2 to 5 times faster than sequential ones, depending on the version of the server and the number of nodes.

# Replication topologies

Multiple sandboxes can be deployed using replication with several topologies (using ``dbdeployer deploy replication --topology=xxxxx``:

* **master-slave** is the default topology. It will install one master and two slaves. More slaves can be added with the option ``--nodes``.
* **group** will deploy three peer nodes in group replication. If you want to use a single primary deployment, add the option ``--single-primary``. Available for MySQL 5.7 and later.
* **fan-in** is the opposite of master-slave. Here we have one slave and several masters. This topology requires MySQL 5.7 or higher.
* **all-masters** is a special case of fan-in, where all nodes are masters and are also slaves of all nodes.

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

Two more topologies, **ndb** and **pxc** require binaries of dedicated flavors, respectively _MySQL Cluster_ and _Percona Xtradb Cluster_. dbdeployer detects whether an expanded tarball satisfies the flavor requirements, and deploys only when the criteria are met.

# Skip server start

By default, when sandboxes are deployed, the servers start and additional operations to complete the topology are executed automatically. It is possible to skip the server start, using the ``--skip-start`` option. When this option is used, the server is initialized, but not started. Consequently, the default users are not created, and the database, when started manually, is only accessible with user ``root`` without password.

If you deploy with ``--skip-start``, you can run the rest of the operations manually:

    $ dbdeployer deploy single --skip-start 5.7.21
    $ $HOME/sandboxes/msb_5_7_21/start
    $ $HOME/sandboxes/msb_5_7_21/load_grants

The same can be done for replication, but you need also to run the additional step of initializing the slaves:

    $ dbdeployer deploy replication --skip-start 5.7.21 --concurrent
    $ $HOME/sandboxes/rsandbox_5_7_21/start_all
    $ $HOME/sandboxes/rsandbox_5_7_21/master/load_grants
    # NOTE: only the master needs to load grants. The slaves receive the grants through replication
    $ $HOME/sandboxes/rsandbox_5_7_21/initialize_slaves

Similarly, for group replication

    $ dbdeployer deploy replication --skip-start 5.7.21 --topology=group --concurrent
    $ $HOME/sandboxes/group_msb_5_7_21/start_all
    $ $HOME/sandboxes/group_msb_5_7_21/node1/load_grants
    $ $HOME/sandboxes/group_msb_5_7_21/node2/load_grants
    $ $HOME/sandboxes/group_msb_5_7_21/node3/load_grants
    $ $HOME/sandboxes/rsandbox_5_7_21/initialize_nodes

WARNING: running sandboxes with ``--skip-start`` is provided for advanced users and is not recommended.
If the purpose of skipping the start is to inspect the server before the sandbox granting operations, you may consider using ``--pre-grants-sql`` and ``--pre-grants-sql-file`` to run the necessary SQL commands (see _Sandbox customization_ below.)

# MySQL Document store, mysqlsh, and defaults.

MySQL 5.7.12+ introduces the XPlugin (a.k.a. _mysqlx_) which enables operations using a separate port (33060 by default) on special tables that can be treated as NoSQL collections.
In MySQL 8.0.11+ the XPlugin is enabled by default, giving dbdeployer the task of defining an additional port and socket for this service. When you deploy MySQL 8.0.11 or later, dbdeployer sets the ``mysqlx-port`` to the value of the regular port + ``mysqlx-delta-port`` (= 10000).

If you want to avoid having the XPlugin enabled, you can deploy the sandbox with the option ``--disable-mysqlx``.

For MySQL between 5.7.12 and 8.0.4, the approach is the opposite. By default, the XPlugin is disabled, and if you want to use it you will run the deployment using ``--enable-mysqlx``. In both cases the port and socket will be computed by dbdeployer.

When the XPlugin is enabled, it makes sense to use [the MySQL shell](https://dev.mysql.com/doc/refman/8.0/en/mysql-shell.html) and dbdeployer will create a ``mysqlsh`` script for the sandboxes that use the plugin. Unfortunately, as of today (late April 2018) the MySQL shell is not released with the server tarball, and therefore we have to fix things manually (see next section.) dbdeployer will look for ``mysqlsh`` in the same directory where the other clients are, so if you manually merge the mysql shell and the server tarballs, you will get the appropriate version of MySQL shell. If not, you will use the version of the shell that is available in ``$PATH``. If there is no MySQL shell available, you will get an error.

# Installing MySQL shell

The MySQL shell is distributed as a tarball. You can install it within the server binaries directory, using dbdeployer (as of version 1.9.0.)

The simplest operation is:

    $ dbdeployer unpack --shell \
        mysql-shell-8.0.12-$YOUR_OS.tar.gz

This command will work if MySQL 8.0.12 was already unpacked in ``$SANDBOX_BINARY/8.0.12``. dbdeployer recognizes the version (from the tarball name) and looks for the corresponding server. If it is found, the shell package will be temporarily expanded, and the necessary files moved into the server directory tree.

If the corresponding server directory does not exist, you can specify the wanted target:

    $ dbdeployer unpack --shell \
        mysql-shell-8.0.12-$YOUR_OS.tar.gz \
        --target-server=5.7.23

Since the MySQL team recommends using the latest shell even for older versions of MySQL, we can insert the shell from 8.0.12 into the 5.7 server, and it will work as expected, as the shell brings with it all needed files.
Using the option ``--verbosity=2`` we can see which files were extracted and where each component was copied or moved. Notice that the unpacked MySQL shell directory is deleted after the operation is completed.

# Database logs management.

Sometimes, when using sandboxes for testing, it makes sense to enable the general log, either during initialization or for regular operation. While you can do that with ``--my-cnf-options=general-log=1`` or ``--my-init-options=--general-log=1``, as of version 1.4.0 you have two simple boolean shortcuts: ``--init-general-log`` and ``--enable-general-log`` that will start the general log when requested.

Additionally, each sandbox has a convenience script named ``show_log`` that can easily display either the error log or the general log. Run `./show_log -h` for usage info.

For replication, you also have ``show_binlog`` and ``show_relaylog`` in every sandbox as a shortcut to display replication logs easily.

# dbdeployer operations logging

In addition to enabling database logs, you can also have logs of the operations performed by dbdeployer when building and activating sandboxes.
The logs are disabled by default. You can enable them for a given operation using ``--log-sb-operations``. When the logs are enabled, dbdeployer will create one or more log files in a directory under ``$HOME/sandboxes/logs``.
For a single sandbox, the log directory will be named ``single_v_v_vv-xxxx``, where ``v_v_vv`` is the version number and ``xxxx`` is dbdeployer Process ID. Inside the directory, there will be a file names ``single.log``.

For a replication sandbox, the directory will be named ``replication_v_v_vv-xxxx`` and it will contain at least 3 files: ``master-slave-replication.log`` with replication operations, and two single sandbox (one for master and one for a slave) logs named ``replication-node-x.log``. If there is more than one slave, each one will have its own log.

dbdeployer logs will record which function ran which operation, with the data used for single and compound sandboxes.

The name of the log is available inside the file ``sbdescription.json`` in each sandbox. If logging is disabled, the log field is not listed.

The logs are preserved until the corresponding sandbox is deleted.

Logging can be enabled permanently using the defaults: ``dbdeployer defaults update log-sb-operations true``. Similarly, you can change the log-directory either for a single operation (``--log-directory=...``) or permanently (``dbdeployer defaults update log-directory /my/path/to/logs``)

What kind of information is in the logs? The most important things found in there is the data used to fill the templates. If something goes wrong, the data should give us a lead in the right direction. The logs also record the result of several choices that dbdeployer makes, such as enebling a given port or adding such and such option to the configuration file. Even if nothing is wrong, the logs can give the inquisitive user some insight on what happens when we deploy a less than usual configuration, and which templates and options can be used to alter the result.

# Sandbox customization

There are several ways of changing the default behavior of a sandbox.

1. You can add options to the sandbox being deployed using ``--my-cnf-options="some mysqld directive"``. This option can be used many times. The supplied options are added to my.sandbox.cnf
2. You can specify a my.cnf template (``--my-cnf-file=filename``) instead of defining options line by line. dbdeployer will skip all the options that are needed for the sandbox functioning.
3. You can run SQL statements or SQL files before or after the grants were loaded (``--pre-grants-sql``, ``--pre-grants-sql-file``, etc). You can also use these options to peek into the state of the sandbox and see what is happening at every stage.
4. For more advanced needs, you can look at the templates being used for the deployment, and load your own instead of the original s(``--use-template=TemplateName:FileName``.)

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
    # Will export all the templates for the "single" group to the directory my_templates/single
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
```json
{
    "version": "1.5.0",
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
    "mysqlx-port-delta": 10000,
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
    "all-masters-prefix": "all_masters_msb_",
    "reserved-ports": [
        1186,
        3306,
        33060
    ],
    "timestamp": "Sat May 12 14:37:53 CEST 2018"
 }
```

    $ dbdeployer defaults update master-slave-base-port 15000
    # Updated master-slave-base-port -> "15000"
    # Configuration file: $HOME/.dbdeployer/config.json
```json
{
    "version": "1.5.0",
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
    "mysqlx-port-delta": 10000,
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
    "all-masters-prefix": "all_masters_msb_",
    "reserved-ports": [
        1186,
        3306,
        33060
    ],
    "timestamp": "Sat May 12 14:37:53 CEST 2018"
}
```

Another way of modifying the defaults, which does not store the new values in dbdeployer's configuration file, is through the ``--defaults`` flag. The above change could be done like this:

    $ dbdeployer --defaults=master-slave-base-port:15000 \
        deploy replication 5.7.21

The difference is that using ``dbdeployer defaults update`` the value is changed permanently for the next commands, or until you run a ``dbdeployer defaults reset``. Using the ``--defaults`` flag, instead, will modify the defaults only for the active command.

# Sandbox management

You can list the available MySQL versions with

    $ dbdeployer versions # Alias: available

Optionally, you can ask for only the versions of a given flavor (`dndeployer versions --flavor=ndb`)  or to show all the versions distinct by flavor (`dbdeployer versions --by-flavor`)

And you can list which sandboxes were already installed

    $ dbdeployer sandboxes # Aliases: installed, deployed

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
    
    "./send_kill [destroy]" does almost the same as "./stop", as it sends a SIGTERM (-15) kill
    to shut down the server. Additionally, when the regular kill fails, it will
    send an unfriendly SIGKILL (-9) to the unresponsive server.
    The argument "destroy" will immediately kill the server with SIGKILL (-9).
    
    "./add_option" will add one or more options to my.sandbox.cnf, and restarts the
    server to apply the changes.
    
    "init_db" and "load_grants" are used during the server initialization, and should not be used
    in normal operations. They are nonetheless useful to see which operations were performed
    to set up the server.
    
    "./show_binlog" and "./show_relaylog" will show the latest binary log or relay-log.
    
    "./my" is a prefix script to invoke any command named "my*" from the 
    MySQL /bin directory. It is important to use it rather than the 
    corresponding globally installed tool, because this guarantees 
    that you will be using the tool for the version you have deployed.
    Examples:
    
        ./my sqldump db_name
        ./my sqlbinlog somefile
    
    "./mysqlsh" invokes the mysql shell. Unlike other commands, this one only works
    if mysqlsh was installed, with preference to the binaries found in "basedir".
    This script is created only if the X plugin was enabled (5.7.12+ with --enable-mysqlx
    or 8.0.11+ without --disable-mysqlx)
    
    "./use_admin" is created when the sandbox is deployed with --enable-admin-address (8.0.14+)
    and allows using the database as administrator, with a dedicated port.
    
     USING MULTIPLE SERVER SANDBOX
    On a replication sandbox, you have the same commands (run "dbdeployer usage single"), 
    with an "_all" suffix, meaning that you propagate the command to all the members. 
    Then you have "./m" as a shortcut to use the master, "./s1" and "./s2" to access 
    the slaves (and "s3", "s4" ... if you define more).
    
    In group sandboxes without a master slave relationship (group replication and 
    multiple sandboxes) the nodes can be accessed by ./n1, ./n2, ./n3, and so on.
    
    start_all    [options] > starts all nodes
    status_all             > get the status of all nodes
    restart_all  [options] > restarts all nodes
    stop_all               > stops all nodes
    use_all         "SQL"  > runs a SQL statement in all nodes
    use_all_masters "SQL"  > runs a SQL statement in all masters
    use_all_slaves "SQL"   > runs a SQL statement in all slaves
    clear_all              > stops all nodes and removes all data
    m                      > invokes MySQL client in the master
    s1, s2, n1, n2         > invokes MySQL client in slave 1, 2, node 1, 2
    
    The scripts "check_slaves" or "check_nodes" give the status of replication in the sandbox.
    
    When the sandbox is deployed with --enable-admin-address (8.0.14+) the following scripts
    are also created:
    
    ma                    > invokes the MySQL client in the master as admin
    sa1, sa2, na1, na2    > invokes MySQL client as admin in slave 1, 2, node 1, 2
    use_all_admin "SQL"   > runs a SQL statement in all nodes as admin
    
    

Every sandbox has a file named ``sbdescription.json``, containing important information on the sandbox. It is useful to determine where the binaries come from and on which conditions it was installed.

For example, a description file for a single sandbox would show:

```json
{
    "basedir": "/home/dbuser/opt/mysql/5.7.22",
    "type": "single",
    "version": "5.7.22",
    "port": [
        5722
    ],
    "nodes": 0,
    "node_num": 0,
    "dbdeployer-version": "1.5.0",
    "timestamp": "Sat May 12 14:26:41 CEST 2018",
    "command-line": "dbdeployer deploy single 5.7.22"
}
```

And for replication:

```json
{
    "basedir": "/home/dbuser/opt/mysql/5.7.22",
    "type": "master-slave",
    "version": "5.7.22",
    "port": [
        16745,
        16746,
        16747
    ],
    "nodes": 2,
    "node_num": 0,
    "dbdeployer-version": "1.5.0",
    "timestamp": "Sat May 12 14:27:04 CEST 2018",
    "command-line": "dbdeployer deploy replication 5.7.22 --gtid --concurrent"
}
```

# Sandbox macro operations

You can run a command in several sandboxes at once, using the ``global`` command, which propagates your command to all the installed sandboxes.

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
      exec             Runs a command in all sandboxes
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
    
    

Using `global`, you can see the status, start, stop, restart, test all sandboxes, or run SQL and metadata queries.

The `global` command accepts filters (as of version 1.44.0) to limit which sandboxes are affected.

The following sub-commands are accepted:

* **exec**             Runs a command in all sandboxes.
* **metadata**         Runs a metadata query in all sandboxes
* **restart**          Restarts all sandboxes
* **start**            Starts all sandboxes
* **status**           Shows the status in all sandboxes
* **stop**             Stops all sandboxes
* **test**             Tests all sandboxes
* **test-replication** Tests replication in all sandboxes
* **use**              Runs a query in all sandboxes

## dbdeployer global exec

Runs a command in all sandboxes.
The command will be executed inside the sandbox. Thus, you can reference a file that you know should be there.
There is no check to ensure that a give command is doable.
For example: the command `cat filename` will result in an error if filename is not present in the sandbox directory.
You can run complex shell commands by prepending them with either `bash -- -c` or `sh -- -c`. Such command must be quoted.
Only one command can be passed, although you can use a shell command as described above to overcome this limitation.
IMPORTANT: if your command argument contains flags, you must use a double dash (`--`) before any of the flags.

You may combine the command passed to `exec` with other drill-down commands in the sandbox directory,
such as `./exec_all`. In this case, you need to make sure that all your sandbox contain the command, or use `--type`
to only run `exec` in the specific topologies.

## dbdeployer global use

Runs a query in all sandboxes.
It does not check if the query is compatible with every version deployed.
For example, a query using `@@port` won't run in MySQL 5.0.x

Usage: `dbdeployer global use {query} [flags]`

Example: `$ dbdeployer global use "select @@server_id, @@port"`

# Sandbox deletion

The sandboxes can also be deleted, either one by one or all at once:

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
    
    

You can lock one or more sandboxes to prevent deletion. Use this command to make the sandbox non-deletable.

    $ dbdeployer admin lock sandbox_name

A locked sandbox will not be deleted, even when running ``dbdeployer delete ALL``.

The lock can also be reverted using

    $ dbdeployer admin unlock sandbox_name

# Default sandbox

You can set a default sandbox using the command `dbdeployer admin set-default sandbox_name`

    $ dbdeployer admin set-default -h
    Sets a given sandbox as default, so that it can be used with $SANDBOX_HOME/default
    
    Usage:
      dbdeployer admin set-default sandbox_name [flags]
    
    Flags:
          --default-sandbox-executable string   Name of the executable to run commands in the default sandbox (default "default")
      -h, --help                                help for set-default
    
    


For example:

    $ dbdeployer admin set-default msb_8_0_20

This command creates a script `$HOME/sandboxes/default` that will point to the sandbox you have chosen.
After that, you can use the sandbox using `~/sandboxes/default command`, such as

    $ ~/sandboxes/default status
    $ ~/sandboxes/default use   # will get the `mysql` prompt
    $ ~/sandboxes/default use -e 'select version()'


If the sandbox chosen as default is a multiple or replication sandbox, you can use the commands that are available there

    $ ~/sandboxes/default status_all
    $ ~/sandboxes/default use_all 'select @@version, @@server_id, @@port'


You can have more than one default sandbox, using the option `--default-sandbox-executable=name`.
For example:


    $ dbdeployer admin set-default msb_8_0_20 --default-sandbox-executable=single
    $ dbdeployer admin set-default repl_8_0_20 --default-sandbox-executable=repl
    $ dbdeployer admin set-default group_msb_8_0_20 --default-sandbox-executable=group

With the above commands, you will have three executables in ~/sandboxes, named `single`, `repl`, and `group`.
You can use them just like the `default` executable:

    $ ~/sandboxes/single status
    $ ~/sandboxes/repl check_slaves
    $ ~/sandboxes/group check_nodes


# Using the latest sandbox

With the command `dbdeployer use`, you will use the latest sandbox that was deployed. If it is a single sandbox, dbdeployer will invoke the `./use` command. If it is a compound sandbox, it will run the `./n1` command.
If you don't want the latest sandbox, you can indicate a specific one:

```
$ dbdeployer use msb_5_7_31
``` 

If that sandbox was stopped, this command will restart it.


# Sandbox upgrade

dbdeployer 1.10.0 introduces upgrades:

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
    
    

To perform an upgrade, the following conditions must be met:

* Both sandboxes must be **single** deployments.
* The older version must be one major version behind (5.6.x to 5.7.x, or 5.7.x to 8.0.x, but not 5.6.x to 8.0.x) or same major version but different revision (e.g. 5.7.22 to 5.7.23)
* The newer version must have been already deployed.
* The newer version must have `mysql_upgrade` in its base directory (e.g `$SANDBOX_BINARY/5.7.23/bin`), but see below about this requirement being lifted for 8.0.16+. 

dbdeployer checks all the conditions, then

1. stops both databases;
2. renames the data directory of the newer version;
3. moves the data directory of the older version under the newer sandbox;
4. restarts the newer version;
5. runs ``mysql_upgrade`` (except with MySQL 8.0.16+, where [the server does the upgrade on its own](https://mysqlserverteam.com/mysql-8-0-16-mysql_upgrade-is-going-away/)).

The older version is, at this point, not operational anymore, and can be deleted.

# Dedicated admin address

MySQL 8.0.14+ introduces the options [`--admin-address` and `--admin-port`](https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_admin_address) to allow a dedicated connection for admin users using a different port. In regular server deployments, the port is 33062, but sandboxes need a different port for each one. Starting with dbdeployer 1.25.0, the option `--enable-admin-address` will create an admin port for each sandbox. In addition to the `./use` script, each single sandbox has a `./use_admin` script that makes administrative access easier.

```
$ dbdeployer deploy single 8.0.15 --enable-admin-address
Database installed in $HOME/sandboxes/msb_8_0_15
run 'dbdeployer usage single' for basic instructions'
.. sandbox server started

$ ~/sandboxes/msb_8_0_15/use_admin
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 9
Server version: 8.0.15 MySQL Community Server - GPL

Copyright (c) 2000, 2019, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

## ADMIN ##mysql [127.0.0.1:19015] {root} ((none)) > select user(), @@port, @@admin_port;
+----------------+--------+--------------+
| user()         | @@port | @@admin_port |
+----------------+--------+--------------+
| root@localhost |   8015 |        19015 |
+----------------+--------+--------------+
1 row in set (0.00 sec)

## ADMIN ##mysql [127.0.0.1:19015] {root} ((none)) >
```
Multiple sandboxes have other shortcuts for the same purpose: `./ma` gives access to the master with admin user, as do the `./sa1` and `./sa2` scripts for slaves. There are similar `./na1` `./na2` scripts for all nodes, and a `./use_all_admin` script sends a query to all nodes through an admin user.

# Loading sample data into sandboxes

The command `data-load` manages the loading of sample databases into a sandbox. (Available since 1.56.0)
It has the following sub-commands:

* `list` shows the available databases (with the option `--full-info` that displays all the details on the archives)
* `show archive-name` displays the contents of one archive
* `get archive-name sandbox-name` downloads the database, unpacks it, and loads its contents into the given sandbox. If the chosen sandbox is not single, the data is loaded into the primary node (`master` or `node1`, depending on the topology)
* `import file-name` loads the archives specifications from a JSON file
* `reset` Restores the archives specifications to their default values

Example:
```
$ dbdeployer deploy replication 8.0 --concurrent
# 8.0 => 8.0.22
$HOME/sandboxes/rsandbox_8_0_22/initialize_slaves
initializing slave 1
initializing slave 2
Replication directory installed in $HOME/sandboxes/rsandbox_8_0_22
run 'dbdeployer usage multiple' for basic instructions'

$ dbdeployer data-load get employees rsandbox_8_0_22
downloading https://github.com/datacharmer/test_db/releases/download/v1.0.7/test_db-1.0.7.tar.gz
.........10 MB.........21 MB.........32 MB...  36 MB
Unpacking /Users/gmax/sandboxes/rsandbox_8_0_22/test_db-1.0.7.tar.gz
..26
Running /Users/gmax/sandboxes/rsandbox_8_0_22/load_db.sh
INFO
CREATING DATABASE STRUCTURE
INFO
storage engine: InnoDB
INFO
LOADING departments
INFO
LOADING employees
INFO
LOADING dept_emp
INFO
LOADING dept_manager
INFO
LOADING titles
INFO
LOADING salaries
data_load_time_diff
00:00:57
```

# Running sysbench

Sandboxes created with version 1.56.0+ include two scripts:

* `sysbench` invokes the sysbench utility with the necessary connection options already filled. Users can specify all remaining options to complete the task.
* `sysbench_ready` can perform two pre-defined actions: `prepare` or `run`.

In both cases, the sysbench utility must already be installed. The scripts look at the supporting files in standard locations. If sysbench was installed manually, errors may occur.

# Obtaining sandbox metadata

As of version 1.26.0, dbdeployer creates a `metadata` script in every single sandbox. Using this script, we can get quick information about the sandbox, even if the database server is not running.

For example:

```
$ ~/sandboxes/msb_8_0_15/metadata help
Syntax: ~/sandboxes/msb_8_0_15/metadata request
Available requests:
  version
  major
  minor
  rev
  short (= major.minor)

  basedir
  cbasedir (Client Basedir)
  datadir
  port
  xport (MySQLX port)
  aport (Admin port)
  socket
  serverid (server id)
  pid (Process ID)
  pidfile (PID file)
  flavor
  sbhome (SANDBOX_HOME)
  sbbin (SANDBOX_BINARY)
  sbtype (Sandbox Type)


$ ~/sandboxes/msb_8_0_15/metadata version
8.0.15

$ ~/sandboxes/msb_8_0_15/metadata short
8.0

$ ~/sandboxes/msb_8_0_15/metadata pid
27361
```

# Replication between sandboxes

Every sandbox (created by dbdeployer 1.26.0+) includes a script called `replicate_from`, which allows replication from another sandbox, provided that both sandboxes are well configured to start replication.

Here's an example:

```
# deploying a sandbox with binary log and server ID
$ dbdeployer deploy single 8.0 --master
# 8.0 => 8.0.15
Database installed in $HOME/sandboxes/msb_8_0_15
run 'dbdeployer usage single' for basic instructions'
. sandbox server started

# Same, for version 5.7
$ dbdeployer deploy single 5.7 --master
# 5.7 => 5.7.25
Database installed in $HOME/sandboxes/msb_5_7_25
run 'dbdeployer usage single' for basic instructions'
. sandbox server started

# deploying a sandbox without binary log and server ID
$ dbdeployer deploy single 5.6
# 5.6 => 5.6.41
Database installed in $HOME/sandboxes/msb_5_6_41
run 'dbdeployer usage single' for basic instructions'
. sandbox server started

# situation:
$ dbdeployer sandboxes --full-info
.------------.--------.---------.---------------.--------.-------.--------.
|    name    |  type  | version |     ports     | flavor | nodes | locked |
+------------+--------+---------+---------------+--------+-------+--------+
| msb_5_6_41 | single | 5.6.41  | [5641 ]       | mysql  |     0 |        |
| msb_5_7_25 | single | 5.7.25  | [5725 ]       | mysql  |     0 |        |
| msb_8_0_15 | single | 8.0.15  | [8015 18015 ] | mysql  |     0 |        |
'------------'--------'---------'---------------'--------'-------'--------'

# Try replicating from the sandbox without binlogs and server ID. It fails
$ ~/sandboxes/msb_5_7_25/replicate_from  msb_5_6_41
No binlog information found in /Users/gmax/sandboxes/msb_5_6_41

# Try replicating from a master of a bigger version than the slave. It fails
$ ~/sandboxes/msb_5_7_25/replicate_from  msb_8_0_15
Master major version should be lower than slave version (or equal)

# Try replicating from 5.7 to 8.0. It succeeds

$ ~/sandboxes/msb_8_0_15/replicate_from  msb_5_7_25
Connecting to /Users/gmax/sandboxes/msb_5_7_25
--------------
CHANGE MASTER TO master_host="127.0.0.1",
master_port=5725,
master_user="rsandbox",
master_password="rsandbox"
, master_log_file="mysql-bin.000001", master_log_pos=4089
--------------

--------------
start slave
--------------

              Master_Log_File: mysql-bin.000001
          Read_Master_Log_Pos: 4089
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
          Exec_Master_Log_Pos: 4089
           Retrieved_Gtid_Set:
            Executed_Gtid_Set:
                Auto_Position: 0

```

The same method can be used to replicate between composite sandboxes. However, some extra steps may be necessary when replicating between clusters, as conflicts and pipeline blocks may happen.
There are at least three things to keep in mind:

1. As seen above, the version of the slave must be either the same as the master or higher.
2. Some topologies need the activation of `log-slave-updates` for this kind of replication to work correctly. For example, `PXC` and `master-slave` need this option to get replication from another cluster to all their nodes.
3. **dbdeployer composite sandboxes have all the same server_id**. When replicating to another entity, we get a conflict, and replication does not start. To avoid this problem, we need to use  the option `--port-as-server-id` when deploying the cluster.

Here are examples of a few complex replication scenarios:

## a. NDB to NDB

Here we need to make sure that the server IDs are different.

```
$ dbdeployer deploy replication ndb8.0.14 --topology=ndb \
    --port-as-server-id \
    --sandbox-directory=ndb_ndb8_0_14_1 --concurrent
[...]
$ dbdeployer deploy replication ndb8.0.14 --topology=ndb \
    --port-as-server-id \
    --sandbox-directory=ndb_ndb8_0_14_2 --concurrent
[...]

$ dbdeployer sandboxes --full-info
.-----------------.--------.-----------.----------------------------------------------.--------.-------.--------.
|      name       |  type  |  version  |                    ports                     | flavor | nodes | locked |
+-----------------+--------+-----------+----------------------------------------------+--------+-------+--------+
| ndb_ndb8_0_14_1 | ndb    | ndb8.0.14 | [21400 28415 38415 28416 38416 28417 38417 ] | ndb    |     3 |        |
| ndb_ndb8_0_14_2 | ndb    | ndb8.0.14 | [21401 28418 38418 28419 38419 28420 38420 ] | ndb    |     3 |        |
'-----------------'--------'-----------'----------------------------------------------'--------'-------'--------'

$ ~/sandboxes/ndb_ndb8_0_14_1/replicate_from ndb_ndb8_0_14_2
[...]
```

## b. Group replication to group replication

Also here, the only caveat is to ensure uniqueness of server IDs.
```
$ dbdeployer deploy replication 8.0.15 --topology=group \
    --concurrent --port-as-server-id \
    --sandbox-directory=group_8_0_15_1
[...]

$ dbdeployer deploy replication 8.0.15 --topology=group \
    --concurrent --port-as-server-id \
    --sandbox-directory=group_8_0_15_2
[...]

$ ~/sandboxes/group_8_0_15_1/replicate_from group_8_0_15_2
[...]
```

## c. Master/slave to master/slave.

In addition to caring about the server ID, we also need to make sure that the replication spreads to the slaves.

```
$ dbdeployer deploy replication 8.0.15 --topology=master-slave \
    --concurrent --port-as-server-id \
    --sandbox-directory=ms_8_0_15_1 \
    -c log-slave-updates
[...]

$ dbdeployer deploy replication 8.0.15 --topology=master-slave \
    --concurrent --port-as-server-id \
    --sandbox-directory=ms_8_0_15_2 \
    -c log-slave-updates
[...]

$  ~/sandboxes/ms_8_0_15_1/replicate_from ms_8_0_15_2
[...]
```

## d. Hibrid replication

Using the same methods, we can replicate from a cluster to a single sandbox (e,g. group replication to single 8.0 sandbox) or the other way around (single 8.0 sandbox to group replication).
We only need to make sure there are no conflicts as mentioned above. The script `replicate_from` can catch some issues, but I am sure there is still room for mistakes. For example, replicating from a NDB cluster to a single sandbox won't work, as the single one can't process the `ndbengine` tables.

Examples:

```
# group replication to single
~/sandboxes/msb_8_0_15/replicate_from group_8_0_15_2

# single to master/slave
~/sandboxes/ms_8_0_15_1/replicate_from msb_8_0_15

# master/slave to group
~/sandboxes/group_8_0_15_2/replicate_from ms_8_0_15_1
```

## e. Cloning

When both master and slave run version 8.0.17+, the script `replicate_from` allows an extra option `clone`. When this
option is given, and both sandboxes meet the [cloning pre-requisites](https://dev.mysql.com/doc/refman/8.0/en/clone-plugin-remote.html),
the script will try to clone the donor before starting replication. If successful, it will use the clone coordinates to
initialize the slave.

# Using dbdeployer in scripts

dbdeployer has been designed to simplify automated operations. Using it in scripts is easy, as shown in the [cookbook examples](#Practical-examples).
In addition to run operations on sandboxes, dbdeployer can also provide information about the environment in a way that is suitable for scripting.

For example, if you want to deploy a sandbox using the most recent 5.7 binaries, you may run `dbdeployer versions`, look which versions are available, and pick the most recent one. But dbdeployer 1.30.0 can aytomate this procedure using `dbdeployer info version 5.7`. This command will print the latest 5.7 binaries to the standard output, allowing us to create dynamic scripts such as:

```bash
# the absolute latest version
latest=$(dbdeployer info version)
latest57=$(dbdeployer info version 5.7)
latest80=$(dbdeployer info version 8.0)

if [ -z "$latest" ]
then
    echo "No versions found"
    exit 1
fi

echo "The latest version is $latest"

if [ -n "$latest57" ]
then
    echo "# latest for 5.7 : $latest57"
    dbdeployer deploy single $latest57
fi

if [ -n "$latest80" ]
then
    echo "# latest for 8.0 : $latest80"
    dbdeployer deploy single $latest80
fi
```

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
    
    

Similarly to `versions`, the `defaults` subcommand allows us to get dbdeployer metadata in a way that can be used in scripts

    $ dbdeployer info defaults -h
    Displays one field of the defaults.
    
    Usage:
      dbdeployer info defaults field-name [flags]
    
    Examples:
    
    	$ dbdeployer info defaults master-slave-base-port 
    
    
    Flags:
      -h, --help   help for defaults
    
    

For example

```
$ dbdeployer info defaults sandbox-prefix
msb_

$ dbdeployer info defaults master-slave-ptrefix
rsandbox_
```
You can ask for any fields from the defaults (see `dbdeployer defaults list` for the field names).

# Importing databases into sandboxes

With dbdeployer 1.39.0, you have the ability of importing an existing database into a sandbox.
The *importing* doesn't involve any re-installation or data transfer: the resulting sandbox will access the existing
database server using the standard sandbox scripts.

Syntax: `dbdeployer import single hostIP/name port username password` 

For example, 

```
dbdeployer import single 192.168.0.164 5000 public nOtMyPassW0rd
 detected: 5.7.22
 # Using client version 5.7.22
 Database installed in $HOME/sandboxes/imp_msb_5_7_22
 run 'dbdeployer usage single' for basic instructions'`
```

We connect to a server running at IP address 192.168.0.164, listening to port 5000. We pass user name and password on
the command line, and dbdeployer, detecting that the database runs version 5.7.22, uses the client of the closest
version to connect to it, and builds a sandbox, which we can access by the usual scripts:

```
~/sandboxes/imp_msb_5_7_22/use
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 19
Server version: 5.7.22 MySQL Community Server (GPL)

Copyright (c) 2000, 2018, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql [192.168.0.164:5000] {public} ((none)) > select host, user, authentication_string from mysql.user;
+-----------+---------------+-------------------------------------------+
| host      | user          | authentication_string                     |
+-----------+---------------+-------------------------------------------+
| localhost | root          | *14E65567ABDB5135D0CFD9A70B3032C179A49EE7 |
| localhost | mysql.session | *THISISNOTAVALIDPASSWORDTHATCANBEUSEDHERE |
| localhost | mysql.sys     | *THISISNOTAVALIDPASSWORDTHATCANBEUSEDHERE |
| localhost | healthchecker | *36C82179AFA394C4B9655005DD2E482D30A4BDF7 |
| %         | public        | *129FD0B9224690392BCF7523AC6E6420109E5F70 |
+-----------+---------------+-------------------------------------------+
5 rows in set (0.00 sec)
```

You have to keep in mind that several assumptions that are taken for granted in regular sandboxes may not hold for an
imported one. This sandbox refers to an out-of-the-box MySQL deployment that lacks some settings that are expected in
a regular sandbox:

```
$ ~/sandboxes/imp_msb_5_7_22/test_sb
ok - version '5.7.22'
ok - version is 5.7.22 as expected
ok - query was successful for user public: 'select 1'
ok - query was successful for user public: 'select 1'
ok - query was successful for user public: 'use mysql; select count(*) from information_schema.tables where table_schema=schema()'
ok - query was successful for user public: 'use mysql; select count(*) from information_schema.tables where table_schema=schema()'
not ok - query failed for user public: 'create table if not exists test.txyz(i int)'
ok - query was successful for user public: 'drop table if exists test.txyz'
# Tests :     8
# pass  :     7
# FAIL  :     1
```

In the above example, the `test` database, which exists in every sandbox, was not found, and the test failed.

There could be bigger limitations. Here's an attempt with a [db4free.net](https://db4free.net) account that works fine
but has bigger problems than the previous one:

```
$ dbdeployer import single db4free.net 3306 dbdeployer $(cat ~/.db4free.pwd)
detected: 8.0.17
# Using client version 8.0.17
Database installed in $HOME/sandboxes/imp_msb_8_0_17
run 'dbdeployer usage single' for basic instructions'
```

A db4free account can only access the user database, and nothing else. Specifically, it can't create databases, access
databases `information_schema` or `mysql`, or start replication.

Speaking of replication, we can use imported sandboxes to start replication between a remote server and a sandbox, or
between a sandbox and a remote server, or even, if both sandboxes are imported, start replication between two remote
servers (provided that the credentials used for importing have the necessary privileges.)

```
$ ~/sandboxes/msb_8_0_17/replicate_from imp_msb_5_7_22
Connecting to /Users/gmax/sandboxes/imp_msb_5_7_22
--------------
CHANGE MASTER TO master_host="192.168.0.164",
master_port=5000,
master_user="public",
master_password="nOtMyPassW0rd"
, master_log_file="d6db0cd349b8-bin.000001", master_log_pos=154
--------------

--------------
start slave
--------------

              Master_Log_File: d6db0cd349b8-bin.000001
          Read_Master_Log_Pos: 154
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
          Exec_Master_Log_Pos: 154
           Retrieved_Gtid_Set:
            Executed_Gtid_Set:
                Auto_Position: 0
```

    $ dbdeployer import single --help
    Imports an existing (local or remote) server into a sandbox,
    so that it can be used with the usual sandbox scripts.
    Requires host, port, user, password.
    
    Usage:
      dbdeployer import single host port user password [flags]
    
    Flags:
      -h, --help   help for single
    
    


# Cloning databases

In addition to [replicating between sandboxes](#replication-between-sandboxes), we can also clone a database, if it is
of version 8.0.17+ and [meets the prerequisites](https://dev.mysql.com/doc/refman/8.0/en/clone-plugin-remote.html).

Every sandbox using version 8.0.17 or later will also have a script named `clone_from`, which works like `replicate_from`.

For example, this command will clone from a master-slave sandbox into a single sandbox:

```
$ ~/sandboxes/msb_8_0_17/clone_from rsandbox_8_0_17
 Installing clone plugin in recipient sandbox
 Installing clone plugin in donor sandbox
 Cloning from rsandbox_8_0_17/master
 Giving time to cloned server to restart
 .
```

# Compiling dbdeployer

Should you need to compile your own binaries for dbdeployer, follow these steps:

1. Make sure you have go 1.11+ installed in your system.
2. Run `git clone https://github.com/datacharmer/dbdeployer.git`.  This will import all the code that is needed to build dbdeployer.
3. Change directory to `./dbdeployer`.
4. Run ./scripts/build.sh {linux|OSX}`
5. If you need the docs enabled binaries (see the section "Generating additional documentation") run `MKDOCS=1 ./scripts/build.sh {linux|OSX}`

# Generating additional documentation

Between this file and [the API API list](https://github.com/datacharmer/dbdeployer/blob/master/docs/API/API-1.1.md), you have all the existing documentation for dbdeployer.
Should you need additional formats, though, dbdeployer is able to generate them on-the-fly. Tou will need the docs-enabled binaries: in the distribution list, you will find:

* dbdeployer-1.58.2-docs.linux.tar.gz
* dbdeployer-1.58.2-docs.osx.tar.gz
* dbdeployer-1.58.2.linux.tar.gz
* dbdeployer-1.58.2.osx.tar.gz

The executables containing ``-docs`` in their name have the same capabilities of the regular ones, but in addition they can run the *hidden* command ``tree``, with alias ``docs``.

This is the command used to help generating the API documentation.

    $ dbdeployer-docs tree -h
    This command is only used to create API documentation. 
    You can, however, use it to show the command structure at a glance.
    
    Usage:
      dbdeployer tree [flags]
    
    Aliases:
      tree, docs
    
    Flags:
          --api               Writes API template
          --bash-completion   creates bash-completion file
      -h, --help              help for tree
          --man-pages         Writes man pages
          --markdown-pages    Writes Markdown docs
          --rst-pages         Writes Restructured Text docs
          --show-hidden       Shows also hidden commands
    
    

In addition to the API template, the ``tree`` command can produce:

* man pages;
* Markdown documentation;
* Restructured Text pages;
* Command line completion script (see next section).

# Command line completion

There is a file ``./docs/dbdeployer_completion.sh``, which is automatically generated with dbdeployer API documentation. If you want to use bash completion on the command line, copy the file to the bash completion directory. For example:

    # Linux
    $ sudo cp ./docs/dbdeployer_completion.sh /etc/bash_completion.d
    $ source /etc/bash_completion

    # OSX
    $ sudo cp ./docs/dbdeployer_completion.sh /usr/local/etc/bash_completion.d
    $ source /usr/local/etc/bash_completion

There is a dbdeployer command that does all the above for you:

```
dbdeployer defaults enable-bash-completion --remote --run-it
```

When completion is enabled, you can use it as follows:

    $ dbdeployer [tab]
        admin  defaults  delete  deploy  global  sandboxes  unpack  usage  versions
    $ dbdeployer dep[tab]
    $ dbdeployer deploy [tab][tab]
        multiple     replication  single
    $ dbdeployer deploy s[tab]
    $ dbdeployer deploy single --b[tab][tab]
        --base-port=     --bind-address=

# Using dbdeployer source for other projects

If you need to create sandboxes from other Go apps, see  [dbdeployer-as-library.md](https://github.com/datacharmer/dbdeployer/blob/master/docs/coding/dbdeployer-as-library.md).

# Exporting dbdeployer structure

If you want to use dbdeployer from other applications, it may be useful to have the command structure in a format that can be used from several programming languages. 
There is a command for that (since dbdeployer 1.28.0) that produces the commands and options information structure as a JSON structure.

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
    
    

# Semantic versioning

As of version 1.0.0, dbdeployer adheres to the principles of [semantic versioning](https://semver.org/). A version number is made of Major, Minor, and Revision. When changes are applied, the following happens:

* Backward-compatible bug fixes increment the **Revision** number.
* Backward-compatible new features increment the **Minor** number.
* Backward incompatible changes (either features or bug fixes that break compatibility with the API) increment the **Major** number.

The starting API is defined in [API-1.0.md](https://github.com/datacharmer/dbdeployer/blob/master/docs/API/API-1.0.md) (generated manually.)
The file [API-1.1.md](https://github.com/datacharmer/dbdeployer/blob/master/docs/API/API-1.1.md) contains the same API definition, but was generated automatically and can be used to better compare the initial API with further version.


# Do not edit

This wiki is **generated** by processing `./mkwiki/wiki_template.markdown`. Do not edit it directly, as its contents will be overwritten.

