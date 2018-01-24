# dbdeployer

DBdeployer is a tool that deploys MySQL database servers easily.


## Operations

With dbdeployer, you can deploy a single sandbox, or many sandboxes with replication.

The main commands are **single** and **replication**, which work with MySQL tarball that have been unpacked into the _sandbox-binary_ directory (by default, $HOME/opt/mysql.)

To use a tarball, you must first run the **unpack** command, which will unpack the tarball into the right directory.

For example:

    $ dbdeployer --unpack-version=8.0.4 unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to /home/gmax/opt/mysql/8.0.4
    .........100.........200.........292

    $ dbdeployer single 8.0.4
    Database installed in /home/gmax/sandboxes/msb_8_0_4
    . sandbox server started


    $ dbdeployer -h
    Makes MySQL server installation an easy task.
        Runs single, multiple, and replicated sandboxes.

    Usage:
      dbdeployer [command]

    Available Commands:
      help        Help about any command
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
          --init-options string     mysqld options to run during initialization
          --my-cnf-options string   mysqld options to add to my.sandbox.cnf
          --remote-access string    defines the database access  (default "127.%")
          --rpl-password string     replication password (default "rsandbox")
          --rpl-user string         replication user (default "rsandbox")
          --sandbox-binary string   Binary repository (default "/Users/gmax/opt/mysql")
          --sandbox-home string     Sandbox deployment direcory (default "/Users/gmax/sandboxes")
          --version                 version for dbdeployer


    $ dbdeployer single -h
    Installs a sandbox and creates useful scripts for its use.

    Usage:
      dbdeployer single [flags] version


    $ dbdeployer replication -h
    create replication sandbox

    Usage:
      dbdeployer replication [flags] version


    $ dbdeployer versions -h
    List available versions

    Usage:
      dbdeployer versions [flags] 

    Aliases:
      versions, available

    $ dbdeployer sandboxes -h
    List installed sandboxes

    Usage:
      dbdeployer sandboxes

    Aliases:
      sandboxes, installed, deployed

    $ dbdeployer unpack -h
    unpack a tarball into the binary directory

    Usage:
      dbdeployer unpack [flags] tarball

    Flags:
      -h, --help                    help for unpack
          --prefix string           Prefix for the final expanded directory
          --unpack-version string   which version is contained in the tarball

    Global Flags:
          --sandbox-binary string   Binary repository (default "/Users/gmax/opt/mysql")

