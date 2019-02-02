# Contributions to dbdeployer

dbdeployer is open source, and as such contributions are welcome.

The following guidelines want to help and simplify the process of contributing to dbdeployer development.

## Principles

1. Contributions should follow the classic GitHub workflow, i.e. forking, cloning, then submitting a Pull Request (PR) 
with the code that you propose.
2. Every new feature or bug fix should have an [issue](https://github.com/datacharmer/dbdeployer/issues), where the 
improvement gets discussed before any code is written.
3. If the contribution is a quick fix, such as a grammar or spelling mistake, go right ahead and submit a PR.
4. A new feature should either have tests associated to it or state a very good reason for why not.
5. Any addition to dbdeployer **must not break** the test suite (see *Running the test suite* below).
6. Any new major feature should be documented in `README.md` (See *Changes to documentation* below).

## Changes to documentation

A caveat about changes in the following files:

* `README.md`
* `docs/dbdeployer_completion.sh`
* `docs/API/*`

These files are **generated**, not written directly. Any change to them must happen through the templates and executables 
in `./mkreadme`. Specifically, changes to `README.md` happen through `./mkreadme/readme_template.md`, which is then
processed using `./mkreadme/build_readme.sh`. 


## Running the test suite


### 1. Availability of binaries

The test suite requires the availability of MySQL binaries in `$HOME/opt/mysql` from version 5.0 to 8.0. There should be 
at least one binary for each major version.

For example, this one is very good:

```bash
$ dbdeployer versions
Basedir: $HOME/opt/mysql
5.0.91        5.1.73        5.5.52        5.5.53        5.6.33        5.6.41
5.7.18        5.7.21        5.7.22        5.7.23        5.7.24        8.0.11
8.0.13        8.0.14
```

But also this is acceptable:

```bash
$ dbdeployer versions
Basedir: $HOME/opt/mysql
4.1.22  5.0.96  5.1.72  5.5.62  5.6.43  5.7.25  8.0.15
```

### 2. producing the binaries.

To run the test suite, you need to create the dbdeployer executables.

`.build/set_version NEW_VERSION`

(Please read the Semantic Versioning section in the README)

If the addition changes the API, you should also change the compatible version

`.build/set_version NEW_VERSION compatible`

Then build the  executables:

```bash
./scripts/build.sh all
MKDOCS=1 ./scripts/build.sh all
```

You will have, among other things, binaries named `dbdeployer-NEW_VERSION.linux` and `dbdeployer-NEW_VERSION-docs.linux`

You need to copy these executable to a place in your `$PATH` where they are picked up automatically. For example:

```bash
cp dbdeployer-NEW_VERSION.YOUR_OS $GOPATH/bin
cp dbdeployer-NEW_VERSION-docs.YOUR_OS $GOPATH/bin
```

Make sure that running `$ dbdeployer --version` you now get the newest one.


### 3. Running the tests with your own MySQL binaries

```bash
./test/go-unit-tests.sh
./test/functional-test.sh concurrent exitfail all
```

Depending on how many binaries you have and the speed of your machine, this test will take from 20 to 40 minutes.


### 4. Running the test in a Docker container

This is the same test that runs automatically on GitHub when a pull request is submitted. It uses a specific Docker
image (`datacharmer/mysql-sb-full`), which contains MySQL binaries from version 4.1.22 to 8.0.x. To run it on your
machine you will need a working Docker server and the ability of downloading the image from the Internet.

While the previous one may occasionally fail (if you don't have all the versions available), this one **must succeed**, 
or your PR will remain in a failed state and thus delay its review. I recommend running this test before pushing a commit.

```bash
export EXIT_ON_FAILURE=1
export RUN_CONCURRENTLY=1
./test/docker-test.sh NEW_VERSION
```

### 5. Running the mock tests

These tests are much faster, and will not use any real database server. Mock servers will be created for totally
made up MySQL versions (such as 5.7.99) and make sure that dbdeployer works as expected for those fake servers.

```bash
./test/mock/defaults-change.sh
./test/mock/short-versions.sh
./test/mock/direct-paths.sh
./test/mock/expected_ports.sh
./test/mock/read-only-replication.sh
```


### 6. Or run them all together

If your environment has all the requirements to run everything listed above, a single command will run all the tests:

```bash
./tests/all_tests.sh NEW_VERSION
```

This one will take approximately 90 minutes to complete.

