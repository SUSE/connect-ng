# connect-ng CI setup

Our CI setup runs the following steps:

* [Lint and Unit Tests](#lint-and-unit-tests)
* [Build RPM](#build-rpm)
* [CLI Feature Tests](#cli-feature-tests)
* [YaST2 Registration Tests](#yast2-registration-tests)
* [Agama Rust Integration Tests](#agama-rust-integration-tests)

## Lint and Unit Tests

workflow definition: [.github/workflows/lint-unit.yml](../../.github/workflows/lint-unit.yml)

This workflow runs the usual `gofmt` check to ensure that our coding style
is consistent, followed by running our unit tests to ensure we have a broad
coverage in functionality, and then verifies that the code base builds.

**Running unit tests locally**

This is equivalent to running the following [Makefile](../../Makefile)
targets locally:

* check-format
* test
* vendor
* build

## Build RPM

workflow definition: [.github/workflows/build-rpm.yml](../../.github/workflows/build-rpm.yml)

This workflow verifies that we can successfully build and install the RPM
packages defined by the [suseconnect-ng.spec](../packaging/suseconnect-ng.spec) file.

**Running build rpm tests locally**

This is equivalent to running `make build-rpm` locally.

## CLI Feature Tests

workflow definition: [.github/workflows/features.yml](../../.github/workflows/features.yml)

This workflow builds the suseconnect-ng components, installs them similarly to
how the associated RPMs would install them and then runs our CLI feature test
suite to help catch errors and help avoid regressions with respect to existing
functionality and the legacy ruby connect version.

Check [features/](../../features) for more information.

**Run feature tests locally using containers**

To run these tests you will need to ensure your local [.env](../../.env-example) file is
setup appropriately and run `make feature-tests` to build and test the feature tests
within a container. For more details see [Feature Tests](../../README.md#feature-tests).

## YaST2 Registration Tests

workflow definition: [.github/workflows/yast.yml](../../.github/workflows/yast.yml)

This workflow builds the suseconnect-ng components and then uses them as part of a
container image build based upon the [third_party/Dockerfile.yast](../../third_party/Dockerfile.yast)
which is used to create a customised container image to run the
[YaST2 Registration Test Suite](https://github.com/yast/yast-registration)

**Running YaST2 Registrationbuild rpm tests locally**

This is equivalent to running `make test-yast` locally.

## Agama Rust Integration Tests

workflow definition: [.github/workflows/agama.yml](../../.github/workflows/agama.yml)

This workflow builds the suseconnect-ng components using a SUSE BCI Golang
container and then pulls those components into a SUSE BCI Rust container
to build the rust based [Agama suseconnect integration testing examples](https://github.com/agama-project/agama/tree/master/rust/suseconnect-agama/examples)
and run them to verify basic Agama registration.

**Running Agama Rust Integration Tests locally**

This is equivalent to running `make agama-tests` locally.

**NOTE**: You can optionally define a valid RMT url using the `RMT_HOST`
setting in the [.env](../../.env-example) file for local testing to
additionally test Agama registrations against an RMT, using either http
or https protocol. See [Agama Rust Integration Tests](../../README.md#agama-rust-integration-tests)
for more details.