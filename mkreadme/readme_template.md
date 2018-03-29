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

    $ dbdeployer unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4
    .........100.........200.........292

    $ dbdeployer deploy single 8.0.4
    Database installed in $HOME/sandboxes/msb_8_0_4
    . sandbox server started


The program doesn't have any dependencies. Everything is included in the binary. Calling *dbdeployer* without arguments or with '--help' will show the main help screen.

    {{dbdeployer --version}}

    {{dbdeployer -h}}

The flags listed in the main screen can be used with any commands.
The flags _--my-cnf-options_ and _--init-options_ can be used several times.

If you don't have any tarballs installed in your system, you should first *unpack* it (see an example above).

	{{dbdeployer unpack -h}}

The easiest command is *deploy single*, which installs a single sandbox.

	{{dbdeployer deploy -h}}

	{{dbdeployer deploy single -h}}

If you want more than one sandbox of the same version, without any replication relationship, use the *multiple* command with an optional "--node" flag (default: 3).

	{{dbdeployer deploy multiple -h}}

The *replication* command will install a master and two or more slaves, with replication started. You can change the topology to "group" and get three nodes in peer replication, or compose multi-source topologies with *all-masters* or *fan-in*.

	{{dbdeployer deploy replication -h}}

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

	{{dbdeployer defaults show}}

	{{dbdeployer defaults update master-slave-base-port 15000}}

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

    {{dbdeployer usage}}

## Sandbox macro operations

You can run a command in several sandboxes at once, using the *global* command, which propagates your command to all the installed sandboxes.

    {{dbdeployer global -h }}

The sandboxes can also be deleted, either one by one or all at once:

    {{dbdeployer delete -h }}

You can lock one or more sandboxes to prevent deletion. Use this command to make the sandbox non-deletable.

    $ dbdeployer admin lock sandbox_name

A locked sandbox will not be deleted, even when running "dbdeployer delete ALL."

The lock can also be reverted using

    $ dbdeployer admin unlock sandbox_name

## Semantic versioning

As of version 1.0.0, dbdeployer adheres to the principles of [semantic versioning](https://semver.org/). A version number is made of Major, Minor, and Revision. When changes are applied, the following happens:

* Backward-compatible bug fixes increment the **Revision** number.
* Backward-compatible new features increment the **Minor** number.
* Backward incompatible changes (either features or bug fixes that break compatibility with the API) increment the **Major** number.

The starting API is defined in [API-1.0.md](https://github.com/datacharmer/dbdeployer/blob/master/docs/API-1.0.md).

