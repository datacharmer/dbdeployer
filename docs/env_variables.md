# Environmental variables

The following variables are used by dbdeployer.

## Abbreviations

* ``DEBUG_ABBR`` Enables debug information for abbreviations engine.
* ``SKIP_ABBR`` Disables the abbreviations engine.
* ``SILENT_ABBR`` Disables the verbosity with abbreviations.
* ``DBDEPLOYER_ABBR_FILE`` Changes the abbreviations file

## Concurrency

* ``RUN_CONCURRENTLY`` Run operations concurrently (multiple sandboxes and deletions)
* ``DEBUG_CONCURRENCY`` Enables debug information for concurrent operations
* ``VERBOSE_CONCURRENCY`` Gives more info during concurrency.

## Catalog

* ``SKIP_DBDEPLOYER_CATALOG`` Stops using the centralized JSON catalog.

## Ports management

* ``SHOW_CHANGED_PORTS`` will show which ports were changed to avoid clashes.

## Sandbox deployment

* ``HOME``   Used to initialize sandboxes components: ``SANDBOX_HOME`` and ``SANDBOX_BINARY`` depend on this one.
* ``TMPDIR`` Used to define where the socket file will be located.
* ``USER``   Used to define which user will be used to run the database server.
* ``SANDBOX_BINARY`` Where the MySQL unpacked binaries are stored.
* ``SANDBOX_HOME``   Where the sandboxes are deployed.
* ``INIT_OPTIONS``   Options to be added to the initialization script command line.
* ``MY_CNF_OPTIONS`` Options to be added to the sandbox configuration file.
* ``MY_CNF_FILE``    Alternate file to be used as source for the sandbox my.cnf.
