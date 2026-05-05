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

### Build
Requires Go >= 1.24

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

**NOTE**: You may find that the `vendor` directory is owned by root after running
the `feature-tests` target; to delete it you may need to run `sudo rm -rf vendor`.

#### YaST2 Registration Tests

You can run the YaST2 registration tests using `make test-yast`, which uses
the [yast/yast-registration repo](https://github.com/yast/yast-registration)
from within a customer container image, defined in [third_party/Dockerfile.yast](third_party/Dockerfile.yast),
to exercise the `libsuseconnect.so` library via the suseconnect Ruby bindings.

#### Coverage Reporting for Unit Tests

Coverage reporting for unit tests has been added, and is enabled by default.
To disable it you can add `COVERAGE=false` to your `make test` command line,
or set it in your environment.

By default the `make test` will report the percentage of statements covered
on a per package basis as they are tested (similar to the `coverage-percent`
target described below), and will then generate the same detailed coverage
report as the `coverage-func` target described below.

**NOTE**: Support for generating coverage testing for the feature and YaST2
registratoon tests is under development and will be available at a later date.

##### Reviewing the most recent coverage data

Coverage data will be saved under the `coverage` directory, and you can review
the most recent unit test run's data using the the following coverage targets:
  * `coverage-func`
    This will report detail coverage stats for each function, with the overall
    summary coverage percentage for all functions in the code base at the end.
  * `coverage-percent`
    This will report the percentage of statements coverage on a per package
    basis as found in the tested codebase.
  * `coverage`
    This is currently an alias for `coverage-func`.
