# dbdeployer

[DBdeployer](https://github.com/datacharmer/dbdeployer) is a tool that deploys MySQL database servers easily.
This is a port of [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox), originally written in Perl, and re-designed from the ground up in [Go](https://golang.org). See the [features comparison](https://github.com/datacharmer/dbdeployer/blob/master/docs/features.md) for more detail.

Documentation updated for version {{.Version}} ({{.Date}})

[![Build Status](https://travis-ci.org/datacharmer/dbdeployer.svg "Travis CI status")](https://travis-ci.org/datacharmer/dbdeployer)

## Installation

The installation is simple, as the only thing you will need is a binary executable for your operating system.
Get the one for your O.S. from [dbdeployer releases](https://github.com/datacharmer/dbdeployer/releases) and place it in a directory in your $PATH.
(There are no binaries for Windows. See the [features list](https://github.com/datacharmer/dbdeployer/blob/master/docs/features.md) for more info.)

For example:

    $ VERSION={{.Version}}
    $ OS=linux
    $ origin=https://github.com/datacharmer/dbdeployer/releases/download/$VERSION
    $ wget $origin/dbdeployer-$VERSION.$OS.tar.gz
    $ tar -xzf dbdeployer-$VERSION.$OS.tar.gz
    $ chmod +x dbdeployer-$VERSION.$OS
    $ sudo mv dbdeployer-$VERSION.$OS /usr/local/bin/dbdeployer

Of course, there are **prerequisites**: your machine must be able to run the MySQL server. Be aware that version 5.6+ and higher require some libraries that are not installed by default in all flavors of Linux (libnuma, libaio.)

## Main operations

(See this ASCIIcast for a demo of its operations.)
[![asciicast](https://asciinema.org/a/165707.png)](https://asciinema.org/a/165707)

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

    {{dbdeployer --version}}

    {{dbdeployer -h}}

The flags listed in the main screen can be used with any commands.
The flags ``--my-cnf-options`` and ``--init-options`` can be used several times.

If you don't have any tarballs installed in your system, you should first ``unpack`` it (see an example above).

	{{dbdeployer unpack -h}}

The easiest command is ``deploy single``, which installs a single sandbox.

	{{dbdeployer deploy -h}}

	{{dbdeployer deploy single -h}}

If you want more than one sandbox of the same version, without any replication relationship, use the ``deploy multiple`` command with an optional ``--nodes`` flag (default: 3).

	{{dbdeployer deploy multiple -h}}

The ``deploy replication`` command will install a master and two or more slaves, with replication started. You can change the topology to *group* and get three nodes in peer replication, or compose multi-source topologies with *all-masters* or *fan-in*.

	{{dbdeployer deploy replication -h}}

## Standard and non-standard basedir names

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

## Using short version numbers

You can use, instead of a full version number (e.g. ``8.0.11``,) a short one, such as ``8.0``. This shortcut works starting with version 1.6.0.
When you invoke dbdeployer with a short number, it will look for the highest revision number within that version, and use it for deployment.

For example, if your sandbox binary directory contains the following:

    5.7.19    5.7.20    5.7.22    8.0.1    8.0.11    8.0.4

You can issue the command ``dbdeployer deploy single 8.0``, and it will use 8.0.11 for a single deployment. Or ``dbdeployer deploy replication 5.7`` and it will result in a replication system using 5.7.22 (the latest one.)


## Multiple sandboxes, same version and type

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

## Using the direct path to the expanded tarball

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

Methods #1 and #2 are equivalent. They set the sandbox binary directory temporarily to a new one, and use it for the current deployement

Methods #3 and #4  will set the sandbox binary directory permanently, with the difference that #3 is set for any invocation of dbdeployer system-wide (in a different terminal window, it will use the new value,) while #4 is set only for the current session (in a different terminal window, it will still use the default.)

Be aware that, using this kind of organization may see conflicts during deployment. For example, after installing Percona Server 5.7.22, if you want to install MySQL 5.7.22 you will need to specify a ``--sandbox-directory`` explicitly.
Instead, if you use the prefix approach defined in the "standard and non-standard basedir names," conflicts should be avoided.

## Ports management

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

## Concurrent deployment and deletion

Starting with version 0.3.0, dbdeployer can deploy groups of sandboxes (``deploy replication``, ``deploy multiple``) with the flag ``--concurrent``. When this flag is used, dbdeployed will run operations concurrently.
The same flag can be used with the ``delete`` command. It is useful when there are several sandboxes to be deleted at once.
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

## Skip server start

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

## MySQL Document store, mysqlsh, and defaults.

MySQL 5.7.12+ introduces the XPlugin (a.k.a. _mysqlx_) which enables operations using a separate port (33060 by default) on special tables that can be treated as NoSQL collections.
In MySQL 8.0.11+ the XPlugin is enabled by default, giving dbdeployer the task of defining an additional port and socket for this service. When you deploy MySQL 8.0.11 or later, dbdeployer sets the ``mysqlx-port`` to the value of the regular port + ``mysqlx-delta-port`` (= 10000).

If you want to avoid having the XPlugin enabled, you can deploy the sandbox with the option ``--disable-mysqlx``.

For MySQL between 5.7.12 and 8.0.4, the approach is the opposite. By default, the XPlugin is disabled, and if you want to use it you will run the deployment using ``--enable-mysqlx``. In both cases the port and socket will be computed by dbdeployer.

When the XPlugin is enabled, it makes sense to use [the MySQL shell](https://dev.mysql.com/doc/refman/8.0/en/mysql-shell.html) and dbdeployer will create a ``mysqlsh`` script for the sandboxes that use the plugin. Unfortunately, as of today (late April 2018) the MySQL shell is not released with the server tarball, and therefore we have to fix things manually (see next section.) dbdeployer will look for ``mysqlsh`` in the same directory where the other clients are, so if you manually merge the mysql shell and the server tarballs, you will get the appropriate version of MySQL shell. If not, you will use the version of the shell that is available in ``$PATH``. If there is no MySQL shell available, you will get an error.

## Installing MySQL shell

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

## Logs management.

Sometimes, when using sandboxes for testing, it makes sense to enable the general log, either during initialization or for regular operation. While you can do that with ``--my-cnf-options=general-log=1`` or ``--my-init-options=--general-log=1``, as of version 1.4.0 you have two simple boolean shortcuts: ``--init-general-log`` and ``--enable-general-log`` that will start the general log when requested.

Additionally, each sandbox has a convenience script named ``show_log`` that can easily display either the error log or the general log. Run `./show_log -h` for usage info.

For replication, you also have ``show_binlog`` and ``show_relaylog`` in every sandbox as a shortcut to display replication logs easily.

## Sandbox customization

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

## Sandbox management

You can list the available MySQL versions with

    $ dbdeployer versions

Also "available" is a recognized alias for this command.

And you can list which sandboxes were already installed

    $ dbdeployer sandboxes  # Aliases: installed, deployed

The command "usage" shows how to use the scripts that were installed with each sandbox.

    {{dbdeployer usage}}

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

## Sandbox macro operations

You can run a command in several sandboxes at once, using the ``global`` command, which propagates your command to all the installed sandboxes.

    {{dbdeployer global -h }}

The sandboxes can also be deleted, either one by one or all at once:

    {{dbdeployer delete -h }}

You can lock one or more sandboxes to prevent deletion. Use this command to make the sandbox non-deletable.

    $ dbdeployer admin lock sandbox_name

A locked sandbox will not be deleted, even when running ``dbdeployer delete ALL``.

The lock can also be reverted using

    $ dbdeployer admin unlock sandbox_name

## Sandbox upgrade

dbdeployer 1.10.0 introduces upgrades:

    {{dbdeployer admin upgrade -h}}

To perform an upgrade, the following conditions myst be met:

* Both sandboxes must be **single** deployments.
* The older version must be one major version behind (5.6.x to 5.7.x, or 5.7.x to 8.0.x, but not 5.6.x to 8.0.x) or same major version but different revision (e.g. 5.7.22 to 5.7.23)
* The newer version must have been already deployed.
* The newer version must have mysql_upgrade in its base directory (e.g $SANDBOX_BINARY/5.7.23/bin)

dbdeployer checks all the conditions, then

1. stops both databases;
2. renames the data directory of the newer version;
3. moves the data directory of the older version under the newer sandbox;
4. restarts the newer version;
5. runs ``mysql_upgrade``.

The older version is, at this point, not operational anymore, and can be deleted.

## Compiling dbdeployer

Should you need to compile your own binaries for dbdeployer, follow these steps:

1. Make sure you have go installed in your system, and that the ``$GOPATH`` variable is set.
2. Run ``go get -u github.com/datacharmer/dbdeployer``.  This will import all the code that is needed to build dbdeployer.
3. Change directory to ``$GOPATH/src/github.com/datacharmer/dbdeployer``.
4. Run ``./scripts/build.sh {linux|OSX} {{.Version}}``
5. If you need the docs enabled binaries (see the section "Generating additional documentation") run ``MKDOCS=1 ./scripts/build.sh {linux|OSX} {{.Version}}``

## Generating additional documentation

Between this file and [the API API list](https://github.com/datacharmer/dbdeployer/blob/master/docs/API/API-1.1.md), you have all the existing documentation for dbdeployer.
Should you need additional formats, though, dbdeployer is able to generate them on-the-fly. Tou will need the docs-enabled binaries: in the distribution list, you will find:

* dbdeployer-{{.Version}}-docs.linux.tar.gz
* dbdeployer-{{.Version}}-docs.osx.tar.gz
* dbdeployer-{{.Version}}.linux.tar.gz
* dbdeployer-{{.Version}}.osx.tar.gz

The executables containing ``-docs`` in their name have the same capabilities of the regular ones, but in addition they can run the *hidden* command ``tree``, with alias ``docs``.

This is the command used to help generating the API documentation. 

{{dbdeployer-docs tree -h}}

In addition to the API template, the ``tree`` command can produce:

* man pages;
* Markdown documentation;
* Restructured Text pages;
* Command line completion script (see next section).

## Command line completion

There is a file ``./docs/dbdeployer_completion.sh``, which is automatically generated with dbdeployer API documentation. If you want to use bash completion on the command line, copy the file to the bash completion directory. For example:

    # Linux
    $ sudo cp ./docs/dbdeployer_completion.sh /etc/bash_completion.d
    $ source /etc/bash_completion

    # OSX
    $ sudo cp ./docs/dbdeployer_completion.sh /usr/local/etc/bash_completion.d
    $ source /usr/local/etc/bash_completion

Then, you can use completion as follows:

    $ dbdeployer [tab]
        admin  defaults  delete  deploy  global  sandboxes  unpack  usage  versions
    $ dbdeployer dep[tab]
    $ dbdeployer deploy [tab][tab]
        multiple     replication  single
    $ dbdeployer deploy s[tab]
    $ dbdeployer deploy single --b[tab][tab]
        --base-port=     --bind-address=

## Using dbdeployer source for other projects

If you need to create sandboxes from other Go apps, see  [dbdeployer-as-library.md](https://github.com/datacharmer/dbdeployer/blob/master/docs/coding/dbdeployer-as-library.md).

## Semantic versioning

As of version 1.0.0, dbdeployer adheres to the principles of [semantic versioning](https://semver.org/). A version number is made of Major, Minor, and Revision. When changes are applied, the following happens:

* Backward-compatible bug fixes increment the **Revision** number.
* Backward-compatible new features increment the **Minor** number.
* Backward incompatible changes (either features or bug fixes that break compatibility with the API) increment the **Major** number.

The starting API is defined in [API-1.0.md](https://github.com/datacharmer/dbdeployer/blob/master/docs/API/API-1.0.md) (generated manually.)
The file [API-1.1.md](https://github.com/datacharmer/dbdeployer/blob/master/docs/API/API-1.1.md) contains the same API definition, but was generated automatically and can be used to better compare the initial API with further version.


## Do not edit

``README.md`` is **generated** by processing ``./mkreadme/readme_template.md``. Do not edit it directly, as its contents will be overwritten.

