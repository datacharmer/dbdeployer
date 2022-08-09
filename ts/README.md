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


## Environmant variables

There are a few environment variables that add information to the test runs:

* `DRY_RUN` will not tun the tests, but will create the `testdata` directory and populate it with the database versions needed for the test.
* `TEST_DEBUG` will add verbosity to the test, showing the initialization part and some details on what the test is doing.
* `TEST_SHORT_VERSIONS` allows the user to choose which versions to test: e.g. `export TEST_SHORT_VERSIONS=5.6,5.7`
* `GITHUB_ACTIONS` If this variable is set, the test will only run on 5.7 and 8.0
* `ONLY_LATEST` the test will only run on the latest 8.0
