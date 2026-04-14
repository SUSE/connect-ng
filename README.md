# SUSEConnect-ng

[![build result](https://build.opensuse.org/projects/systemsmanagement:SCC/packages/suseconnect-ng/badge.svg?type=default)](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)

SUSEConnect is a Golang command line tool for connecting a client system to the SUSE Customer Center.
It will connect the system to your product subscriptions and enable the product repositories/services locally.

SUSEConnect-ng reduces the size of its runtime dependencies compared to the
replaced [Ruby SUSEConnect](https://github.com/SUSE/connect).

SUSEConnect-ng is distributed as RPM for all SUSE distributions and gets built in
the [openSUSE build service](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng).

Please visit https://scc.suse.com to see and manage your subscriptions.

SUSEConnect-ng communicates with SCC over this [REST API](https://github.com/SUSE/connect/blob/master/doc/SCC-API-%28Implemented%29.md).

### Requirements

Requirements:
  * Go >= 1.24
  * Docker

### Build

```
make build
```
This will create a `out/suseconnect` binary.

### Build in container
If you don't have a go compiler installed, you can run the build in a container: 
```
docker run --rm -v $(pwd):/connect registry.suse.com/bci/golang:1.24-openssl sh -c "git config --global --add safe.directory /connect; cd /connect; make build"
```
This will create a `out/suseconnect` binary on the host.

### Testing

There are three test suites available to run:
  * unit tests
  * feature tests
  * YaST2 registration tests

#### Unit Tests

You can run all unit tests by running `make test`. If you then want to run unit
tests for a specific package, you can simply run it as you would do for any Go
project, for example: `go test ./internal/collectors/`.

#### Feature Tests

For feature tests you first need to create an `.env` file in the root directory
of the project with the following contents:

``` sh
VALID_REGCODE="<regcode>"
EXPIRED_REGCODE="<regcode>"
NOT_ACTIVATED_REGCODE="<regcode>"
```

These values can be picked up from Glue's production environment. Once that is
done, you can then simply run `make feature-tests`. This will run a all feature
tests inside of a container by using the registration codes as provided by the
`.env` file.

#### YaST2 Registration Tests

You can run the YaST2 registration tests using `make test-yast`, which uses
the [yast/yast-registration repo](https://github.com/yast/yast-registration)
from within a customer container image, defined in [third_party/Dockerfile.yast](third_party/Dockerfile.yast),
to exercise the `libsuseconnect.so` library via the suseconnect Ruby bindings.

#### Running all test suites

You can run all the test suites sequentially using `make run-tests`. Please
ensure you have setup the `.env` file appropriately to support running the
feature tests first.

### Coverage Reporting (Experimental)

Experimental support for reporting unified coverage of all test suite runs
is available by passing `COVERAGE=true` on the command line when running the
`make` command, or by setting the value in the environment.

This will enable coverage collection by the individual tests under the `coverage`
directory.

When coverage collection is enabled each of the test suites will report their
respective coverage details after the test suite completes successfully.

To report the detailed coverage data you can use the `coverage` target in the
[Makefile](Makefile). Alternatively to get a summary of the percentage coverage
by subpackage using the `coverage-percent` target.

**Notes about coverage collection**:
  * While YaST2 registration test suite driving infrastructure has been updated
    to support enabling coverage collection for the `libsuseconnect.so` library
    the current implementation is missing the step to flush the collected data
    before the process exits. This will be addressed by future updates. Until
    this is resolved the YaST2 test suite will always report zeros for coverage.
  * Additional work will be needed to enable coverage collecting and reporting
    by the GitHub Action CI workflows.