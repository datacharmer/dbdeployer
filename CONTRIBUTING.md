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

`common/set_version.sh NEW_VERSION`

(Please read the Semantic Versioning section in the README)

If the addition changes the API, you should also change the compatible version

`common/set_version NEW_VERSION compatible`

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
./test/mock/replication-setup.sh
```


### 6. Or run them all together

If your environment has all the requirements to run everything listed above, a single command will run all the tests:

```bash
./tests/all_tests.sh NEW_VERSION
```

This one will take approximately 90 minutes to complete.

Here is a sample of the results at the end:

```
# ----------------------------------------------------------------
# cookbook-test             - tests:   35 - pass:   35 - fail:    0
# cookbook                  - tests:  136 - pass:  136 - fail:    0
# defaults-change           - tests:   85 - pass:   85 - fail:    0
# direct-paths              - tests: 3798 - pass: 3798 - fail:    0
# docker-test               - tests: 5350 - pass: 5350 - fail:    0
# expected_ports            - tests:   18 - pass:   18 - fail:    0
# fanin_test                - tests:  216 - pass:  216 - fail:    0
# functional-test           - tests: 5460 - pass: 5460 - fail:    0
# go-unit-tests             - tests: 1366 - pass: 1366 - fail:    0
# ndb_test                  - tests:  522 - pass:  522 - fail:    0
# parallel                  - tests:    6 - pass:    6 - fail:    0
# port-clash                - tests:  180 - pass:  180 - fail:    0
# pxc_test                  - tests:  164 - pass:  164 - fail:    0
# read-only-replication     - tests: 3348 - pass: 3348 - fail:    0
# sanity_check              - tests:    0 - pass:    0 - fail:    0
# short-versions            - tests:   42 - pass:   42 - fail:    0
# ----------------------------------------------------------------
sanity_check                   []        - time:   19s (       19s) - exit code: 0
go-unit-tests                  []        - time:   88s (    1m:28s) - exit code: 0
functional-test                []        - time: 2230s (   37m:10s) - exit code: 0
docker-test                    [1.45.0]  - time: 2294s (   38m:14s) - exit code: 0
cookbook-test                  []        - time:  990s (   16m:30s) - exit code: 0
defaults-change                []        - time:    7s (        7s) - exit code: 0
short-versions                 []        - time:    6s (        6s) - exit code: 0
direct-paths                   []        - time:   18s (       18s) - exit code: 0
expected_ports                 []        - time:    2s (        2s) - exit code: 0
read-only-replication          []        - time:   21s (       21s) - exit code: 0
ndb_test                       []        - time:    4s (        4s) - exit code: 0
pxc_test                       []        - time:    1s (        1s) - exit code: 0
fanin_test                     []        - time:   18s (       18s) - exit code: 0
cookbook                       []        - time:    2s (        2s) - exit code: 0
parallel                       []        - time:    9s (        9s) - exit code: 0
port-clash                     [sparse]  - time:  159s (    2m:39s) - exit code: 0
# Deployed: 111 sandboxes (1143 total ports) - Changed: 0
# ----------------------------------------------------------------
# cookbook-test             - tests:   35 - pass:   35 - fail:    0
# cookbook                  - tests:  136 - pass:  136 - fail:    0
# defaults-change           - tests:   85 - pass:   85 - fail:    0
# direct-paths              - tests: 3798 - pass: 3798 - fail:    0
# docker-test               - tests: 5350 - pass: 5350 - fail:    0
# expected_ports            - tests:   18 - pass:   18 - fail:    0
# fanin_test                - tests:  216 - pass:  216 - fail:    0
# functional-test           - tests: 5460 - pass: 5460 - fail:    0
# go-unit-tests             - tests: 1366 - pass: 1366 - fail:    0
# ndb_test                  - tests:  522 - pass:  522 - fail:    0
# parallel                  - tests:    6 - pass:    6 - fail:    0
# port-clash                - tests:  180 - pass:  180 - fail:    0
# pxc_test                  - tests:  164 - pass:  164 - fail:    0
# read-only-replication     - tests: 3348 - pass: 3348 - fail:    0
# sanity_check              - tests:    0 - pass:    0 - fail:    0
# short-versions            - tests:   42 - pass:   42 - fail:    0
# ----------------------------------------------------------------
# ----------------------------------------------------------------
# Total tests: 20726
#       pass : 20726
#       fail : 0
# ----------------------------------------------------------------
OS:  Darwin
Started: Sat Feb 22 21:00:23 CET 2020
Ended  : Sat Feb 22 22:38:13 CET 2020
Elapsed: 5870 seconds (1h:37m:50s)
# Exit code: 0
Runs concurrently: yes
```
