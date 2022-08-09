# Testing with testscript

The tests in this directory test dbdeployer using [`testscript`](https://github.com/rogpeppe/go-internal/testscript), a library that allows writing tests for tools quickly and effectively.

The tests run like any other test:

```
go test
```

Running with `-v` will be **very** verbose. 


## Test initialization

The test will perform several initializations before running. If the environment is ready (`$SANDBOX_BINARY` and `$SANDBOX_HOME` created) it will skip the initialization. Otherwise, it will run [`dbdeployer init`](https://github.com/datacharmer/dbdeployer/wiki/initializing-the-environment).

Furthermore, the test will download the necessary database binaries for all the required versions.
By default, the versions being used are: 4.1, 5.0, 5.1, 5.5, 5.6, 5.7, 8.0
If any of these versions are not available, the test will skip the download. See [Environmant variables](#environment-variables) for how to change this behavior

## Template-based script generation

The test does not use static scripts. Instead, it uses several templates (in the `templates` folder), which will be used
to create the final scripts. For each template, the initialization procedure creates one script for each version being
recognized by dbdeployer (see previous section).

Some templates refer to specific capabilities that are not available for all versions. For example, group replication is
only available for 5.7.17+ and 8.0.x. The test initialization recognizes such capabilities and only creates scripts for
compatible versions.

## Environmant variables

There are a few environment variables that add information to the test runs:

* `DRY_RUN` will not tun the tests, but will create the `testdata` directory and populate it with the database versions needed for the test.
* `TEST_DEBUG` will add verbosity to the test, showing the initialization part and some details on what the test is doing.
* `TEST_SHORT_VERSIONS` allows the user to choose which versions to test: e.g. `export TEST_SHORT_VERSIONS=5.6,5.7`
* `GITHUB_ACTIONS` If this variable is set, the test will only run on 5.7 and 8.0
* `ONLY_LATEST` the test will only run on the latest 8.0
