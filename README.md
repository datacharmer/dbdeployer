# dbdeployer

[DBdeployer](https://github.com/datacharmer/dbdeployer) is a tool that deploys MySQL database servers easily.


## Operations

With dbdeployer, you can deploy a single sandbox, or many sandboxes  at once, with or without replication.

The main commands are **single**, **replication**, and **multiple**, which work with MySQL tarball that have been unpacked into the _sandbox-binary_ directory (by default, $HOME/opt/mysql.)

To use a tarball, you must first run the **unpack** command, which will unpack the tarball into the right directory.

For example:

    $ dbdeployer --unpack-version=8.0.4 unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to /home/gmax/opt/mysql/8.0.4
    .........100.........200.........292

    $ dbdeployer single 8.0.4
    Database installed in /home/gmax/sandboxes/msb_8_0_4
    . sandbox server started


The program doesn't have any dependencies. Everything is included in the binary. Calling *dbdeployer* without arguments or with '--help' will show the main help screen.

    $ dbdeployer -h
    Makes MySQL server installation an easy task.
        Runs single, multiple, and replicated sandboxes.

    Usage:
      dbdeployer [command]

    Available Commands:
      help        Help about any command
      multiple    create multiple sandbox
      replication create replication sandbox
      sandboxes   List installed sandboxes
      single      deploys a single sandbox
      unpack      unpack a tarball into the binary directory
      versions    List available versions

    Flags:
          --bind-address string     defines the database bind-address  (default "127.0.0.1")
          --config string           config file (default "./sandbox.json")
          --db-password string      database password (default "msandbox")
          --db-user string          database user (default "msandbox")
          --gtid                    enables GTID
      -h, --help                    help for dbdeployer
          --init-options strings    mysqld options to run during initialization
          --my-cnf-options strings  mysqld options to add to my.sandbox.cnf
          --remote-access string    defines the database access  (default "127.%")
          --rpl-password string     replication password (default "rsandbox")
          --rpl-user string         replication user (default "rsandbox")
          --sandbox-binary string   Binary repository (default "/Users/gmax/opt/mysql")
          --sandbox-home string     Sandbox deployment direcory (default "/Users/gmax/sandboxes")
          --version                 version for dbdeployer

    Use "dbdeployer [command] --help" for more information about a command.

The flags listed in the main screen can be used with any commands.
The flags _--my-cnf-options_ and _--init-options_ can be used several times.

If you don't have any tarballs installed in your system, you should first *unpack* it (see an example above).

	$ dbdeployer unpack -h
	unpack a tarball into the binary directory

	Usage:
	  dbdeployer unpack MySQL-tarball [flags]

	Flags:
	  -h, --help                    help for unpack
		  --prefix string           Prefix for the final expanded directory
		  --unpack-version string   which version is contained in the tarball

The main command is *single*, which installs a single sandbox.

	$ dbdeployer single -h
	single installs a sandbox and creates useful scripts for its use.
	MySQL-Version is in the format x.x.xx, and it refers to a directory named after the version
	containing an unpacked tarball. The place where these directories are found is defined by
	--sandbox-binary (default: $HOME/opt/mysql.)
	For example:
		dbdeployer single 5.7.21

	For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
	the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz

	Usage:
	  dbdeployer single MySQL-Version [flags]

	Flags:
	  -h, --help   help for single

If you want more than one sandbox of the same version, without any replication relationship, use the *multiple* command with an optional "--node" flag (default: 3).

	$ dbdeployer multiple -h
	create multiple sandbox

	Usage:
	  dbdeployer multiple MySQL-Version [flags]

	Flags:
	  -h, --help        help for multiple
		  --nodes int   How many nodes will be installed (default 3)

The *replication* command will install a master and two or more slaves, with replication started.

	$ dbdeployer replication -h
	create replication sandbox

	Usage:
	  dbdeployer replication MySQL-Version [flags]

	Flags:
	  -h, --help              help for replication
		  --nodes int         How many nodes will be installed (default 3)
		  --topology string   Which topology will be installed (default "master-slave")

The only topology currently supported is "master-slave". Others, such as group-replication and multi-source. will follow.
