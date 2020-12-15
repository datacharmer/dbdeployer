## 1.58.1	15-Dec-2020

### BUGS FIXED

* Fix Issue #124 (Can't reset defaults after an upgrade)

## 1.58.0	12-Dec-2020

### NEW FEATURES

* Add script `wipe_and_restart` to single sandboxes
* Add scripts `exec_all`, `exec_all_masters`, `exec_all_slaves`, `wipe_and_restart_all` to replication sandboxes
* Add subcommand `exec` to command `global`.

## 1.57.0	09-Dec-2020

### NEW FEATURES

* Add subcommands `import`, `export`, and `reset` to command `data-load`

## 1.56.0	02-Nov-2020

### NEW FEATURES

* Add options `--ls` and `--run` to command `use`
* Add scripts `sysbench` and `sysbench_ready` to all sandboxes
* Add command `data-load` with subcommands `list`, `show`, `get`

### BUGS FIXED

* Fix Issue #120 (Can't SELECT the mysql.user table in mariadb 10.4 or later)

## 1.55.0	10-Oct-2020

### NEW FEATURES

* Add second optional argument to `dbdeployer use` to indicate the executable to run within the sandbox.

## 1.54.1	27-Sep-2020

### BUGS FIXED

* Fix Issue #118 (dbdeployer fails to autodetect pxc flavor from binaries of 5.6.x versions)

## 1.54.0	13-Sep-2020

### NEW FEATURES

* The `send_kill` script accepts an argument `destroy` (aliases `-9` or `crash`) 
  to halt the server immediately.
* The `delete` command is much faster (uses `send_kill destroy`)

### BUGS FIXED

* The `use` command should re-start a sandbox that was stopped, but it was not doing it (Issue #116)
* Listing of small tarball for 8.0.21 lacked the 'minimal' attribute

### NOTES

* Added MySQL shell 8.0.21 to downloads list

### TESTING

* Improve `all_tests`: Now it doesn't require the version
* Improve unit test script to detect where to run


## 1.53.3	29-Aug-2020

### BUGS FIXED

* Fixed issue #115: failing to detect missing `socat` for PXC

## 1.53.2	07-Aug-2020

### BUGS FIXED

* Fixed bug during `init`. When the sandbox-binary directory existed, but empty,
  the download was skipped.
* `dbdeployer downloads export` would not export the field `date_added`. Such field was
   lost during build, due to a missed field in `code-generation.go`

### NOTES

* Added MySQL 5.7.31 to downloads list

### TESTING

* Added test for `dbdeployer init` under Docker containers for Ubuntu 18, Ubuntu 20, CentOS 7, CentOS 8


## 1.53.1	26-Jul-2020

### BUGS FIXED

* Fix bash completion for CentOS (missing completion script and unchecked `sudo` call)

## 1.53.0	26-Jul-2020

### NEW FEATURES

* Add command 'use' and sandboxes options
* Add options `--by-date`, `--by-flavor`, `--by-version`, `--newest`, `--oldest` to command `sandboxes`

### NOTES

* Add latest MySQL tarballs to downloads list
* Add Percona Server minimal to downloads list

### BUGS FIXED

* Fix issue #111 - allows using NDB 7.4/7.5


## 1.52.0	28-Jun-2020

### NEW FEATURES

* Add  command `init` that initializes dbdeployer environment
    * creates `$SANDBOX_HOME` and `$SANDBOX_BINARY` directories
    * downloads and expands the latest MySQL tarball
    * installs shell completion file

## 1.51.2	15-Jun-2020

### BUGS FIXED

* Fix error handling on default directories detection for bash completion.

## 1.51.1	14-Jun-2020

### BUGS FIXED

* Fix bug in defaults handling: "default-sandbox-executable" was not used in `info defaults`
* Prevent `ldd` from running on mocked binaries

### TESTING

* Moved mock tests under docker to improve consistency


## 1.51.0	12-Jun-2020

### NEW FEATURES

* Add default sandbox (commands `admin set-default` and `admin remove-default`)

## 1.50.2	07-Jun-2020

### BUGS FIXED

* Fix incorrect privileges when updating dbdeployer as root

## 1.50.1	06-Jun-2020

### BUGS FIXED

* Fix Issue #108: panic error when sandbox binary is not a directory

## 1.50.0	16-May-2020

### NEW FEATURES

* Add new options `--raw` and `--stats` to  command `info releases`
* Add checksum files (SHA256) to release
* Add signature check during dbdeployer update


## 1.49.0	26-Apr-2020

### NEW FEATURES

* Add `--change-master-options` (Issue #107)
* Add more downloads (8.0.20, 5.7.30)

### TESTING

* Add more tests
* Docker test now downloads latest 8.0 binaries instead of using built-in.

## 1.48.0	10-Apr-2020 (not released)

* Improve PXC handling of defaults across versions
* Fix defaults for PXC 5.6, which were broken after introducing support for PXC8.0 (Issue #106)


## 1.47.0	29-Mar-2020

### ADJUSTMENTS

* Add `pxc_encrypt_cluster_traffic` option to PXC options file
  to adjust to recent default changes and allow PXC 8.0.18 to work
  out of the box

### Code improvements

* Change all explicit `bash` paths in shell scripts to `env bash`

## 1.46.0	08-Mar-2020

### NEW FEATURES

* Add ability to deploy PXC 8.0 without using any options (dbdeployer
  uses `wsrep_sst_method=rsync` for 5.7 and `wsrep_sst_method=xtrabackup-v2` for 8.0)

## 1.45.0	22-Feb-2020

### NEW FEATURES

* Add option `--task-user` and `--task-user-role` to create additional user with custom role.
* Add cookbook recipe `single-custom-users` (include sample creation of orchestator user)

### Code improvements

* Add security checks with `gosec` scans
* Update copyright notice

### TESTING

* Add test for cookbook scripts

## 1.44.0	16-Feb-2020

### NEW FEATURES

* Add custom role to 8.0 sandboxes
* Add role related options: `--default-role`, `--custom-role-name`, `--custom-role-privileges`,
  `--custom-role-target`, `--custom-role-extra`
* Add filters to `global` command:
  `--type`, `--flavor`, `--version`, `--short-version`, `--name`, `--port`, `--port-range`.
  Each option can also be negated (`--type='!single'`, `--flavor=no-mysql`, `--port='!8019'`)
* Add verbose and dry-run flags to `global`.
* Add `metadata_all_template` script to composite sandboxes
* Add command `downloads add`
* Add tarballs for 5.7.29, 8.0.19 for MySQL and NDB (Linux and MacOS)

### TESTING

* Add test for tarball integrity and reachability

## 1.43.2	31-Jan-2020 (not released)

### BUGS FIXED

* Fix bug in NDB cluster: the cluster name was not reported correctly
  in the NDB configuration file

## 1.43.1	22-Dec-2019

### BUGS FIXED

* Fix determination of shell interpreter

## 1.43.0	10-Nov-2019 

### ADJUSTMENTS
* Add 'IF NOT EXISTS' to 'CREATE ROLE', to account to changed
  behavior of NDB server in 8.0.18

### BUGS FIXED

* Fix for hanging during library check (Issue #99)

## 1.42.0	01-Nov-2019 

###	NEW FEATURES
* Explicit template for PXC replication options (Issues #87 and #92)
  Allows installing PXC with version 5.6
* Explicit template for Group replication options

###	BUGS FIXED
* Fix command "dbdeployer info version --flavor VER"
  It would not display the requested short version but
  the latest one.

## 1.41.0	26-Oct-2019 

###	NEW FEATURES
* dbdeployer can guess the next downloadable version with the command
  `dbdeployer downloads get-by-version 8.0 --guess-latest`
  or
  `dbdeployer downloads get-by-version 5.7 --guess-latest`
	
###	BUGS FIXED
* Issue #95 mismatched values in fan-in when using nodes > 3
* Issue #93 Client detection for imported sandbox doesn't take flavor into account

###	TESTING
* Add test for `FindOrGuessTarballByVersionFlavorOS`
* Add mock test for fan-in nodes

## 1.40.0	19-Oct-2019 

###	NEW FEATURES
* dbdeployer now detects when essential libraries are missing (Linux only)
* New flag '--skip-library-check' will avoid the above automated detection
  but will expose to the risk of having an incomplete sandbox if the
  prerequisites were not satisfied.
* Add tarballs for 8.0.18 and 5.7.28 to the downloads list
* Add check to verify tarball correctness (they must contain a directory
	  that matches the tarball name)

## 1.39.1	02-Oct-2019 

###	BUGS FIXED:
* Show stderr text when external commands fail - Issue #94
* Enable start.log by default.
* Add missing exit directive on failure in `init_db` template
* Improve error messages on `init_db` failure
* Build Linux amd64 instead of 386

## 1.39.0	21-Sep-2019 
* Add script `clone_from` to clone from another 8.0.17+ server
* Add option "clone" to script `replicate_from`
* Add ability of detecting and using GTID to script `replicate_from`
* Add ability of finding a suitable client for imported sandboxes
* Add option --server-id to "deploy single" command
* Add option --base-server-id to deploy command
* Add option --sandbox-directory to import command

## 1.38.0	25-Aug-2019 

###	NEW FEATURES
* add command dbdeployer import single 


## 1.37.0	24-Aug-2019 (not released)

### NEW FEATURES
* Enable Go modules
* Remove vendor directory
* Change calls to StringSliceP with StringArrayP
* (No changes in behavior. Some flags help items will display "stringArray" 
	  instead of "strings")

## 1.36.1	18-Aug-2019

### BUGS FIXED
* Fix a bug in 'dbdeployer update': it only worked with --verbose enabled
* Fix output of 'dbdeployer update': it now confirms that the update has
	  happened

## 1.36.0	17-Aug-2019

### NEW FEATURES
* Add command "dbdeployer update"
* Add command "dbdeployer info releases"


### BUGS FIXED
* Fix 'dbdeployer info defaults': some values were not reported


### TESTING
* Add test for defaults info

## 1.35.0	11-Aug-2019

### NEW FEATURES
* Improve command "dbdeployer downloads import": it can also import from a URL
* Add downloads binaries for 8.0.17 and 5.7.27
* Add mysql-shell binaries in downloads


### TESTING
* More tests for `common/*`

## 1.34.0	30-Jun-2019

### NEW FEATURES
* Add command "defaults enable-bash-completion" to help install bash
	  completion
* Add command "defaults flag-aliases" to show aliases for some flags
* Add flag "--earliest" to command "info version"


### BUGS FIXED
* Downloads list should show files for current OS only
* Fix tarball list generation to be rebuilt only when
	  contents change
* Fix OS user detection to rely on operating system library calls
	  instead of environment variables.
	

### TESTING
* Replace `sort_versions` with calls to "dbdeployer info version"

## 1.33.0	16-Jun-2019

### NEW FEATURES
* Add command "downloads get-unpack" to combine the effects of
	  "dbdeployer downloads get" and "dbdeployer unpack" into a
	  single command (Issue #81)

### BUGS FIXED
* Issue #83 'unpack --overwrite' should not ask for confirmation
* Issue #84 '--sandbox-binary' does not expand '~' to $HOME

## 1.32.0	09-Jun-2019

### NEW FEATURE
* Add option `--shell-path` to choose alternative Bash interpreter
	  used in all scripts, including cookbook recipes and mock scripts

### TESTING
* Add test for alternative Bash interpreter

## 1.31.0	22-May-2019

### NEW FEATURES
* Add command 'downloads'
* add subcommand 'downloads list'
* add subcommand 'downloads get'
* add subcommand 'downloads get-by-version'
* add subcommand 'downloads import'
* add subcommand 'downloads export'
* add subcommand 'downloads reset'
	
### DEPRECATION
The command 'remote' is deprecated with its subcommands
Its functionality is replaced by 'downloads', albeit with
a different syntax.

## 1.30.1	22-May-2019

### BUGS FIXED
* Fix issue #80 NDB tarballs are wrongly labeled as regular "mysql" flavor

## 1.30.0	05-May-2019

### BUGS FIXED
* Fix Issue #72 Catalog operations can clash with multiple dbdeployer runs
* Fix wrong description for "defaults update"
* Fix wrong description for "cookbook create"
* Fix option `--sort-by` incorrectly assigned to command "cookbook" instead 
	  of subcommand "cookbook list"

### NEW FEATURES
* Add command "dbdeployer info" to improve dbdeployer usage in scripts

### TESTING
* Add test for parallel usage
* Reinstate parallel deployment in functional test
* Improve replication between sandboxes test, by checking
	  for existing versions better.

## 1.29.0	30-Apr-2019

### ADJUSTMENTS
* Change upgrade procedure for MySQL 8.0.16+
* Adjust cookbook upgrade recipe to use verbose output

### NEW FEATURES
* Add options `--dry-run` and --verbose to upgrade command

### TESTING
* Enhance export tests

## 1.28.2	22-Apr-2019

### BUGS FIXED
* Add missing error checks
* Clean code that was failing static checks

### TESTING
* Add export tests
* Add conditional static check to sanity test

## 1.28.1	14-Apr-2019

### BUGS FIXED
Add missing `$` in recipe skip-start-replication

## 1.28.0	14-Apr-2019

### NEW FEATURES
* Add option --sort-by=[name,flavor,script] to cookbook list
* Add option --flavor for versions command
* Add command export to show dbdeployer command structure as JSON 
* Add cookbook recipes 
	* Single sandbox deployed without starting the server
	* Replication sandbox deployed without starting the servers
	* Replication sandbox with custom names for directories and scripts
	* Single sandbox with admin address enabled

### BUGS FIXES
* Fixed plural of custom replication names


## 1.27.0	31-Mar-2019

### NEW FEATURES
* Add cookbook recipes:
    * replication across multiple versions
    * replication from group replication to single sandbox
    * replication from group replication to master/slave
    * replication from single sandbox to group replication
    * replication from master/slave to group replication
    * replication between single sandboxes
    * circular replication


## 1.26.0	26-Mar-2019

### NEW FEATURES
* Add custom `replication_from` script to each sandbox.
* Add metadata script to each sandbox
* Add option -port-as-server-id
* Add more info in `sb_include` script
* Add connection info scripts to each sandbox
* Add cookbook scripts for replication between groups, ndb, master/slave.

### BUGS FIXED
* Fix "remote get" bug: it should fail on Mac (Issue #76)
* Fix port detection bug for NDB. The cluster port was not used in the search for free ports.
* Sort cookbook listing (it was being displayed in random order)

### TESTING
* Add test for replication between sandboxes (single and master/slave)

## 1.25.0	24-Mar-2019 (not released)

### ADJUSTMENTS/NEW FEATURES:
* Add option `--enable-admin-address` for 8.0.14+
* When admin address is enabled, `./use_admin` is created for each sandbox
* multiple sandboxes have `./ma`, `./sa1` `./sa2`, `./na1`, `./na2` and `use_all_admin`

## 1.24.0	22-Mar-2019

### NEW FEATURES:
* Add command `cookbook` with subcommands list, show, create
* Replace cookbook static sample scripts with dynamically created ones

## 1.23.0	14-Mar-2019

### NEW FEATURES:
* Add support for MySQL Cluster (topology=ndb)
* Add option `--ndb-nodes` to replication

## 1.22.0	07-Mar-2019

### ADJUSTMENTS:
Change default authentication plugin for MariaDB 10.4.3 (Issue #67)

## 1.21.0	05-Mar-2019

### NEW FEATURES:
* Added support for Percona Xtradb Cluster (`--topology=pxc`) [Linux only]
* Added check for Unix tools before deploying.

## 1.20.0  24-Feb-2019 (not released)

### NEW FEATURES:
* Added command `delete-binaries`
* Added option `--overwrite` to `unpack` command
* Added flavors to `sandboxes`
* Added options `--flavor` and `--full-info` to `sandboxes`
* Added options `--flavor-in-prompt` to `deploy`
* Added option `--socket-in-datadir` to `deploy`
* Added option `--prompt` to `deploy single`

## 1.19.0	17-Feb-2019

### ADJUSTMENTS:
* Changed `remote get` docs to reflect changed names

### BUGS FIXED:
* Fixed Issue #48 error with multiple plugins
* Fixed Issue #62 TiDB fails on MacOS

### NEW FEATURES:
* Added script `after_start` to sandboxes (does nothing by default,
	  but it is used by TiDB to clean-up unwanted scripts)

## 1.18.0	09-Feb-2019

### NEW FEATURES:
* Added support for TiDB
* Added option `--client-from=X` (Issue #49)
* Added option `--flavor`
* Added flavor recording during unpack (Issue #52)
* Added command `dbdeployer admin capabilities [flavor [version]]`
* Added flavor detection based on files in expanded tarball (Issue #53)

	Code improvements:
* Version evaluation replaced by *capabilities*, i.e. flavor + version (Issue #50)

### BUGS FIXED:
* Fix Issue #51 unpack command fails when tarball name doesn't include a version
* Fix unpack panic when tarball does not include top directory
* Fix count of total tests in `./test/all_tests`.sh

### TESTING:
* Added tests for TiDB, flavor detection, capabilities evaluation

## 1.17.1	26-Jan-2019
* Fix Issue#46 - error creating catalog
* Some test improvements

## 1.17.0	11-Jan-2019

### NEW FEATURES:
* Added options `--read-only-slaves` and `--super-read-only`-slaves
	  to `deploy replication` command. It only works for slaves of
	  `master-slave` and `fan-in` topologies.

### BUGS FIXED:
* Fixed bug in multi-master topologies, where replication
	  ports were not calculated correctly

### Code improvements:
* Added check for IP address in replication creation

### TESTING:
* Added test for read-only replication options
* Added common failure tests for sandbox creation

## 1.16.0	05-Jan-2019

### NEW FEATURES:
* Add support for remote tarballs.
* New commands: `remote list` and `remote download`

## 1.15.1	09-Dec-2018

### Code improvements:

* Removed all unconditional output in reusable code (common, sandbox)
* Changed many public functions to private when unused outside package. 
* Added tests for most of functions in common  
	Minor bug fixes

## 1.15.0	02-Dec-2018

### Code improvements:
* Changed all code that can Exit instead of returning an error in 
	  packages common, sandbox, defaults, unpack.
* Packages common and defaults keep the functions that can exit
	  if they are only used with the CLI.
* Changed all functions returning errors to have the error as the last
	  returning item.
* Moved constants and many global variables from 'defaults' to new
	  'globals' package. This reduce the risk of circular import between
	  dbdeployer packages.

## 1.14.0	11-Nov-2018
### Code readability improvements:
* Added function `IsEnvSet`
* Added variables for features minimum versions
* Added constants for sandbox script names
* Added constants for most used error messages
* Added comparison functions for testing in a separate package
* Replaced manual path composition with path.Join

### NEW FEATURES:
* Added support for MySQL 4.1
* Added version info to sandbox scripts

## 1.13.0	27-Oct-2018

### ADJUSTMENTS
* Added ability of unpacking "tar.xz" files (tarballs for Linux are 
	  compressed with xz instead of gzip as of MySQL 8.0.12)

### BUGS FIXED
* Fixed minor issue in unpack-shell. The unpacker was trying to
	  move the shell extracted directory to the server directory.

## 1.12.4	14-Oct-2018 (not released)
* Code cleanup: fixed many code style violations
* Added unit test for abbreviations module.

## 1.12.3	07-Oct-2018
* Merged in PR #39 (add port to prompt)
* Added tests for common/checks and common/tprintf
* Refactored some code to improve testability

## 1.12.2	30-Sep-2018 (not released)
* Replaced stack implementation
* Added tests for stack and concurrency

## 1.12.1	23-Sep-2018

### NEW FEATURES:
* Added cookbook scripts

### BUGS FIXED:
* Fixed Issue #38 "--force does not check for locked sb"
* Added missing copyright on some files.
* Added check for missing copyright to `sanity_check.sh`

## 1.12.0	22-Sep-2018

### NEW FEATURES:
* Added sandbox operations logging (on demand)
* Added option `--log-sb-operations` to enable logging
* Added option `--log-directory` to change the default directory (`$SANDBOX_HOME`)
* Added log-sb-operations and log-directory fields to default values.

## 1.11.0	09-Sep-2018

### NEW FEATURES:
* Added option `--repl-crash-safe` (Issue #36) to get crash safe params without GTID.

### BUGS FIXED:
* Fixed Issue #35 (--gtid should include relay-log-recovery)
* Fixed Issue #37 (slave initialization with GTID should use
	  `MASTER_AUTO_POSITION`) and added test for GTID behavior.
* Some code refactoring: simplified exit procedure

## 1.10.2	08-Sep-2018 (not released)
* Code cleanup and formatting
* Moved build script to ./scripts 
* Added option to build compressed executables (but defaults is still
	  uncompressed)
* Added ./scripts/sanity_check.sh to verify formatting and code complaint.

## 1.10.1	01-Sep-2018
* Fixed Issue#34 "dbdeployer fails when .mylogin.cnf is found"

## 1.10.0	27-Aug-2018
* Fixed Issue#32 "Error log was not created in data directory"
* Added command 'dbdeployer admin upgrade sandbox1 sandbox2' (Issue#33)

## 1.9.0	12-Aug-2018
* Implemented Issue#31 "Add feature to extract and merge a mysql-shell
	  tarball into a regular one"

## 1.8.4	11-Aug-2018
* Fixed Issue #30 "--rpl-user and --db-user should not allow root"

## 1.8.3	05-Aug-2018
* Fixed Issue #26 "Port management glitch"

## 1.8.2	05-Aug-2018
* Added ./vendor folder (simplifies dependencies)

## 1.8.1	03-Aug-2018
* Fixed Issue #27 "undetected deps when building with MKDOCS"
	  No binary releases for this issue, as only build.sh is affected.

## 1.8.1	12-Jul-2018
* Improved naming of error log for each sandbox

## 1.8.0	08-Jul-2018

### NEW FEATURES
* Implemented Issue 21 "Add support for directly using path to binaries"


### BUGS FIXED
* unpack would not act on old tarball where files were not 
	  explicitly marked as regular.
* Fixed Issue 22 "dbdeployer should check whether the binaries are for the
	current OS"

### TESTING
* Added test for Issue 21

## 1.7.0	01-Jul-2018

### NEW FEATURES
* Added option for custom history file for each sandbox
* Added option for unified history file in replication sandboxes.

### BUGS FIXED
* Fixed bug in functional-test.sh, where some of the subtests were not
	  executed.

### TESTING
* Improved error checking in all test scripts

## 1.6.0	19-Jun-2018

### NEW FEATURES
	  Now it is possible to invoke dbdeployer with a short version number,
	  such as 5.7 or 8.0. In this case, dbdeployer will look for the latest
	  release of that MySQL version and use it.

### BUGS FIXED
* Command line was not saved to dbdeployer catalog

### TESTING
* Added test for short versions
* Added test for command line in catalog

## 1.5.3	10-Jun-2018
* Fixed Issue #16 "Rename during unpack command fails" by making
	  sure the path for sandbox-home and sandbox-binary are absolute ones.

## 1.5.2	09-Jun-2018

### BUGS FIXED
* Added a stack of cleanup operations for group sandboxes when a depending 
	  sandbox installation fails
* Fixed help message of "unpack" command.

### ADJUSTMENTS
* When using `mysqld-debug`, plugins are loaded from
	  $BASEDIR/lib/plugin/debug (https://bugs.mysql.com/bug.php?id=89688)

### TESTING
* Added test for data dictionary tables exposure

## 1.5.1	05-Jun-2018
* Improved documentation.
* Minor code refactoring

## 1.5.0	12-May-2018

### NEW FEATURES
* Added option `--binary-version`, which allows the basedir to be other than
	  {prefix}x.x.xx
* Added command line record to sbdescription.json, to the catalog,
	  and to the defaults listing.

### BUGS FIXED
* Fixed Issue #10 (again). The directory for symlinks was not created
	  timely, resulting in errors with some tarballs.
	VARIOUS
* Minor code refactoring
* Changes in templates for single sandbox scripts. Now almost all scripts
	  source a file containing common code (sb_include). The behavior
	  of the scripts is unchanged.

### TESTING
* Added test for `--binary-version`

## 1.4.2	06-May-2018
* Code reformatting and minor refactoring 
* Fixed syntax error in code for tree.go
* Added coding sample minimal-sandbox2.go

## 1.4.1	05-May-2018
* Merged pull request #11 from percona-csalguero/include_version_in_dir
* Fixed Issue #12 "deploying a sandbox with invalid version does not fail"
* Fixed minor bugs
* Removed unnecessary parameter in CreateSingleSandbox
* Added instructions to call CreateSingleSandbox from other apps
	  See ./docs/coding
* Minor code refactoring for Exit calls
* Added mock sandbox creation for unit tests

## 1.4.0	28-Apr-2018

### NEW FEATURES:
* Added option `--enable-mysqlx` (MySQL 5.7.12+)
* Added options `--enable-general-log` and `--init-general-log`
* Added list of "reserverd-ports" to defaults
* Increased documentation inside command "usage"
* Added dbdeployer version and timestamp to sandbox descriprtion files.
* Added "mysqlsh" script to sandboxes 5.7.12+ with Xplugin enabled
* Added `show_log` script to all sandboxes
* Improved interface of `show_binlog` and `show_relaylog`

### TESTING
* Added tests for reserved-ports
* Added test for mysqlsh and `show_*` creation
	VARIOUS
* Updated documentation

## 1.3.0	21-Apr-2018

### ADJUSTMENTS:
* Added support for mysqlx plugin being enabled by default (MySQL 8.0.11+)
* Added flag "--disable-mysqlx" to disable mysqlx plugin (8.0.11+)

### NEW FEATURES:
* Added scripts `use_all_masters` and `use_all_slaves` to all replication
	  sandboxes.
* Added option `--verbosity={0,1,2}` to *unpack* command.

### BUGS FIXED:
* Fixed Issue#10 "dbdeployer unpack does not handle symlinks"
* Fixed minor bug in documentation test builder.

### TESTING
* Added tests for number of ports, log errors, `use_all_masters`,
	  `use_all_slaves`, running processes.
* Added options to stop tests after a given set of operations.
* Removed restriction on running 5.6 tests in docker for Mac.

## 1.2.0	14-Apr-2018
* Added option `--skip-start`
* Added report-port and report-host automatically to my.sandbox.cnf
* Added options `--skip-report-host` and `--skip-report-port`
* Added documentation dbdeployer compiling.
* Added documentation for `--skip-start`
* Enhanced build.sh to handle dependencies.
* Added tests for `--skip-start` and report-host/report-port behavior.

## 1.1.1	02-Apr-2018
* Added more documentation
* Added bash-completion script
* Moved hidden command "tree" to conditional compiling.
	  Now there is a separated build for docs-enabled dbdeployer
* Added ability of producing more documentation using command "tree"

## 1.1.0	30-Mar-2018
* Added ability of handling environment variables
	  in configuration file. $HOME and $PWD are expanded
	  to actual values.
* Added hidden command "tree" that can generate the
	  full dbdeployer API. Using this feature, from now on
	  we can compare API changes automatically.
* Fixed visualization of sandboxes from catalog
* Fixed minor code issues.
* Added tests for environment variable replacement

## 1.0.1	28-Mar-2018
* Fixed Issue #5 "Single deployment doesn't show the location of the
	  sandbox"
* Added API definition (`./docs/API-1.0.md`)
* Added test for Issue #5
* Fixed typos and improved docs.

## 1.0.0	26-Mar-2018
* General Availability.
* Fixed bug with single deployment and --force. On the second deployment,
	  the port was changed.
* More tests added. The test suite now runs a total of 3,013 tests (MacOS)
	  and 3,143 (Linux). A total of 6,156 tests that ran at least twice (once 
	  with concurrency and once without)

## 0.3.9	25-Mar-2018
* Added version detection to `unpack` command. Now `--unpack-version`
	  becomes mandatory only if a version is not detected from the tarball
	  name.
* Added `--header` flag to `sandboxes` command.
* More tests and improved tests.

## 0.3.8	24-Mar-2018
* Fixed deployment bug in fan-in replication
* Added tests for fan-in replication, sandbox completeness, start,
	  restart, and `add_options`.

## 0.3.7	24-Mar-2018
* Added `--semi-sync` option to replication
* Added more tests

## 0.3.6	21-Mar-2018
* Minor change to templates
* Added test for export/import templates
* Added more tests for pre/post grants SQL

## 0.3.5	21-Mar-2018
* Added test for on-the-fly template replacement
* Trivial changes to "sandboxes" output

## 0.3.4	20-Mar-2018
* Changed test for group replication (now uses the
	  same defined for multi-source replication)
* Improved usability of tests.
* Made tests easier to extend.
* Added test for pre/post grants SQL.

## 0.3.3	16-Mar-2018
* Added (mock) tests for unpack command
* Improved test reporting
* Added list of environment variables

## 0.3.2	15-Mar-2018
* Minor bug fixes
* Added more tests

## 0.3.1	12-Mar-2018
* Added topologies "fan-in" and "all-masters"
* Feature complete: This is release candidate for 1.0
* Fixed bug on UUID generation.

## 0.3.0	11-Mar-2018
* Implemented parallel deployment of multiple sandboxes
* Flag --concurrent is available for *deploy* and *delete*
* Improved tests

## 0.2.5	10-Mar-2018
* Added --catalog to "sandboxes" command
* Improved tests

## 0.2.4	08-Mar-2018
* INCOMPATIBLE CHANGES:
    * MySQL 8.0.x now starts with `caching_sha2_password` by default.
    * flag "`--keep-auth-plugin`" was removed. Instead, we now have
	    "`--native-auth-plugin`", false by default, which will use the old
		plugin if enabled.
    * The sandbox catalog is now enabled by default. It can be disabled
	    using either the environment variable `SKIP_DBDEPLOYER_CATALOG`
		or using the configuration file.
* Added workaround for bug#89959: replication with 8.0 and
	  `caching_sha2_password` fails


## 0.2.3	07-Mar-2018 (not released)
* Improved mock test speed by parametrizing sleep intervals:
    * 200 mock sandboxes tested in 73 seconds (previously, 15 minutes).
    * 2144 mock sandboxes tested in 23 minutes (previously, 145 minutes)

## 0.2.2	07-Mar-2018
* Now dbdeployer finds unused ports automatically, to avoid conflicts.
* Added ability of running faster tests with mock MySQL packages.

## 0.2.1	04-Mar-2018
* Added `--defaults` flag
* Removed hardcoded names for multiple sandbox directories and shortcuts.
* Added directory names and shortcuts for multiple sandboxes to configuration data
* Added ability of exporting/importing a single template.
* Fixed visualization error with template export
* Added interactivity to main test script.

## 0.2.0	27-Feb-2018
* INCOMPATIBLE CHANGES:
    * "single", "multiple", and "replication" are now subcommands of "deploy".
    * Previous "admin" commands are now under "defaults"
    * "templates" is now a subcommand of "defaults"
    * New "admin" command only supports "lock" and "unlock"

* EXPERIMENTAL FEATURE:
		There is a sandbox catalog being created and updated in
		`$HOME/.dbdeployer/sandboxes.json`.
		The deployment and deletion commands handle the catalog
		transparently. Disabled by default. It can be enabled by
		setting the environment variable `DBDEPLOYER_CATALOG`


## 0.1.25	26-Feb-2018
* Added commands "admin lock" and "admin unlock" to prevent/allow deletion
	of a sandbox.
* Added placeholder fields for multi-source clustering in defaults

## 0.1.24	20-Feb-2018
* Fixed bug with "sandboxes" command. It would not check if the
	  `sandbox_home` directory existed.
* Fixed bug in "sandboxes" command. It would not report sandboxes
	  created by other applications (MySQL-Sandbox)
* Added check for template version during export/import
* Added tests for UUID generation
* Improved docker test

## 0.1.23	19-Feb-2018
* Added "test-replication" to "global" command
* Added several aliases to "unpack"
* Changed template `init_db`, to allow for easier customization
* Added test for docker. The full test suite can run in a container.
* Simplified test.sh by using "dbdeployer global" rather than hand made
	  loops.

## 0.1.22	18-Feb-2018
* All values used for sandbox deployment are now modifiable.
* Added command "admin" to deal with default values:
	  show, store, remove, load, update, export
* Refactored global variables to become modifiable through the "admin"
	  command
* Added commands for template export and import.

## 0.1.21	16-Feb-2018
* Added flag `--expose-dd-tables` to show data dictionary hidden tables
* Added flag `--custom-mysqld` to use different server executable

## 0.1.20	14-Feb-2018
* Added flags for pre and post grants SQL execution.
* `--pre-grants-sql-file`
* `--pre-grants-sql`
* `--post-grants-sql-file`
* `--post-grants-sql`
* Fixed bug (in cobra+pflag package) that splits multiple commands by comma.

## 0.1.19	14-Feb-2018
* MySQL 8.0+ sandboxes now use roles instead of direct privilege
	assignments.
* Added global flag --force to overwrite an existing deployment
* Added global flag `--my-cnf-file` to use a customized my.cnf
* Added flag `--master-ip` to replication deployments
* Fixed bug in abbreviations: flags were not considered correctly.

## 0.1.18	12-Feb-2018
* The "delete" command now supports "ALL" as argument. It will delete all installed sandboxes.
* Added flag "--skip-confirm" for the "delete" command, to delete without confirmation.
* Fixed mistake during initialization: the version search was happening before the check
	  for the sandbox home directory.
* Added the "global" command to propagate a command to all sandboxes

## 0.1.17	11-Feb-2018
* Added automated README generation
* minor code changes

## 0.1.16	10-Feb-2018
* Added automatic generation of human-readable server-UUID
* Added flag `--keep-server-uuid` to prevent the above change

## 0.1.15	08-Feb-2018
* Changed default port and sandbox directory for single-primary group
	  replication.
* Added custom abbreviations feature.

## 0.1.14	07-Feb-2018
* Added script `test_sb` to every single sandbox
* Added script `test_sb_all` to all multiple/group/replication sandbox
* Added script `test_replication` to replication sandboxes
* Added test/test.sh, which runs a comprehensive test of most dbdeployer features

## 0.1.13	06-Feb-2018
* Added command "templates export"
* Added flag `--single-primary` for group replication
* Added flags `--sandbox-directory`, --port, and base-port
	  to allow deploying several sandboxes of the same version.
* Added a check for clash on installed ports
* INCOMPATIBLE change: Changed format of sbdescription.json:
	  now can list several ports per sandbox.

## 0.1.12	04-Feb-2018
* Added a check for version before applying gtid.
* Added commands templates list/show/describe
* Added `--use-template=template_name:file_name` flag

## 0.1.11	31-Jan-2018
* Improved check for tarball as an argument to single, replication,
	multiple.
* Improved help for single, multiple, and replication
* Added customized prompt for configuration file

## 0.1.10	30-Jan-2018
* Changed initialization method to use tarball libraries
* Fixed glitch in "unpack" when original tarball has clashing name

## 0.1.09	30-Jan-2018
* Updated README.md
* Changed formatting for "usage" command
* Run detection of invalid group replication earlier.
* Added version to sandbox description file

## 0.1.08	29-Jan-2018
* Added sandbox description file
* 'sandboxes' command uses above file for sandbox listing
* Added 'delete' command

## 0.1.07	29-Jan-2018
* improved documentation
* Added "usage" command
* Added description to "sandboxes" output
* Added check for version format
* Changed message for missing argument
* Added check for sandbox-home existence

## 0.1.06	28-Jan-2018
* Added group replication topology.

## 0.1.05	27-Jan-2018
* Added option --master to 'single' command
* Added new commands to each sandbox: `add_option`, `show_binlog`,
	`show_relaylog`, my.

## 0.1.04	26-Jan-2018
* Added short names for some flags.
* Improved commands usage text

## 0.1.03	26-Jan-2018
* Modified `--my-cnf-options` and --init-options to be accepted multiple
	times

## 0.1.02	25-Jan-2018
* Fixed bug in unpack when basedir was not created.

## 0.1.01	25-Jan-2018
* Fixed inclusion of options in my.sandbox.cnf (`--my-cnf-options`)
* Added command 'multiple'
* Enhanced documentation

## 0.1.00	24-Jan-2018
* Initial commit with basic features migrated from MySQL-Sandbox
