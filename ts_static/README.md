# Tests with testscript

Like the tests in package `ts`, also the ones here use `testscript`, with a notable difference: they don't use real
database servers, and in fact can run even in places without the possibility of downloading them.

A large portion of the tests use a mocking recipe. It is similar to what is done in the `sandbox` package, but much more
resilient here, as it uses the `txtar` capabilities of `testscript` to generate fake database binaries while also
isolating the testing environment. When the fake database binaries are created, dbdeployer will use them as if they were real ones.