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

### Prerequisites
The following tools should be available and verified as working:

  * A Golang v1.24 or later environment
    * needed to run certain go commands locally for testing and
      validation.
  * `git`
  * `make`
  * `docker`

### Quick Start

  * Clone the repo
      ```
      git clone https://github.com/SUSE/connect-ng
      cd connect-ng
      ```

  * When building for the first time, execute
      ```
      make vendor
      ```

  * Run build which will create `out/suseconnect` binary.
      ```
      make build
      ```

  * Setup `.env` file to run tests

      Copy the `.env-example` file to `.env` file and fill in the following
      values appropriately:
      ```
      REGCODE="<regcode>"
      EXPIRED_REGCODE="<regcode>"
      HA_REGCODE="<ha regcode>"

      # Optional RMT_HOST for Agama integration testing
      # RMT_HOST=https://rmt.example.com
      ```

      These regcodes can be your personal regcodes attached to your organization.
      Please refer to the
      [Activating-and-Managing-Subscriptions](https://scc.suse.com/docs/userguide#UG-Activating-and-Managing-Subscriptions)
      documentation if not familiar with the process.

      The `RMT_HOST`, if specified, must be a valid url for an RMT that
      supports registration of clients running the distro that backs
      the `RUSTCONTAINER` image specified in the [Makefile](Makefile).

  * To run the unit tests
      ```
      make test
      ```

      See [Unit Tests](#unit-tests) for details.

  * To run the feature tests

      Run the feature tests within a container by using the registration codes as provided by the .env file.
      ```
      make feature-tests
      ```

      See [Feature Tests](#feature-tests) for details.

  * To run the YAST2 registration tests
      ```
      make test-yast
      ```

      See [YaST2 Registration Tests](#yast2-registration-tests) for details.

  * To run the Agama rust integration tests
      ```
      make agama-tests
      ```

      See [Agama Rust Integration Tests](#agama-rust-integration-tests) for details.

### Build
Requires Go >= 1.24

```
make build
```
This will create a `out/suseconnect` binary.

### Build in a container
If you don't have a go compiler installed, you can run the build in a container:
```
docker run --rm -v $(pwd):/connect registry.suse.com/bci/golang:1.24-openssl sh -c "git config --global --add safe.directory /connect; cd /connect; make vendor build"
```
Or you can use the `bci-build` Makefile target:
```
make bci-build
```

Either of these actions will create an `out/` directory on the host containing the built binaries, such as `suseconnect`.

### Testing

There are three test suites available to run:
  * unit tests
  * feature tests
  * YaST2 registration tests
  * Agama rust integration tests

See also the [Running GitHub Actions Locally](#running-github-actions-locally) section below.

#### Unit Tests

You can run all unit tests by running `make test`. If you then want to run unit
tests for a specific package, you can simply run it as you would do for any Go
project, for example: `go test ./internal/collectors/`.

#### Feature Tests

For feature tests you first need to create an `.env` file in the root directory
of the project with the following contents:
For that copy .env-example file to .env file and fill following values

``` sh
REGCODE="<regcode>"
EXPIRED_REGCODE="<regcode>"
HA_REGCODE="<regcode>"
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
from within a custom container image, defined in [third_party/Dockerfile.yast](third_party/Dockerfile.yast),
to exercise the `libsuseconnect.so` library via the suseconnect Ruby bindings.

#### Agama Rust Integration Tests

You can run the Agama Rust integration tests using `make agama-tests`, which
uses the rust examples from the [agama-project/agama repo](https://github.com/agama-project/agama)
to exercise the Agama Rust integration via the `libsuseconnect.so` library.

These tests require that a valid `REGCODE` be specified via the `.env` file,
and optionally support an `RMT_HOST` url being specified in the `.env` file,
either `http` or `https` protocol.

Specifically these tests exercise that:
  * basic activation registrations works with the SCC, via the
    [activation tool](https://github.com/agama-project/agama/blob/master/rust/suseconnect-agama/examples/activation.rs)
  * optionally tests registration against an RMT using the
    [rmt tool](https://github.com/agama-project/agama/blob/master/rust/suseconnect-agama/examples/rmt.rs)
    if an `RMT_HOST` is specified via the .env file, or potentially via the GitHub Actions
    secrets for CI runs. If the `RMT_HOST` is specified with a `https` protocol URL, it
    will download the RMT's self-sigtned cert and provide it to the `rmt` example via
    the optional second argument.

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

### Running GitHub Actions Locally

With the [nektos/act](github.com/nektos/act) tool installed, either directly
or as a [GitHub CLI](https://cli.github.com/) extension, it can be used to run
the [SUSE/connect-ng GitHub Action workflows](.github/workflows/) locally.

`act` can also assist in the development and testing of new and existing
workflows.

#### Install and Setup `act`

Follow the installation instruction for installing `act` in the preferred way,
either as a local command `act` or via the `gh act` extension.

##### SUSE/connect-ng specific act settings in `.actrc`

The standard platform images used by `act` don't always support all of the features
that may be needed by a GitHub Actions workflow. `act` provides a mechanism to
specify alternate platform images via the `--platform` option.

Additionally, enabling the emulation of the v4 artifact upload/download support
also requires appropriate command line options to be set.

Similarly, to ensure that `act` sets up the appropriate environment settings for
jobs that run, the `--env-file .env` option should be set.

The [.actrc](.actrc) file in the repo specifies appropriate values for these
options.

#### Running the workflows locally

The available workflows, and their associated trigger events can be listed by
running the `act --list` or `gh act --list` command, as follows:

```bash
% act --list
INFO[0000] Using docker host 'unix:///var/run/docker.sock', and daemon socket 'unix:///var/run/docker.sock'
Stage  Job ID                Job name                                        Workflow name           Workflow file  Events
0      build-suseconnect     Build SUSE/connect-ng inputs required by Agama  agama tests             agama.yml      pull_request
0      feature-tests         feature-tests                                   feature tests           features.yml   pull_request
0      unit-tests            unit-tests                                      lint + unit tests       lint-unit.yml  pull_request
0      yast                  yast                                            YaST integration tests  yast.yml       pull_request
1      run-agama-rust-tests  Run agama rust tests                            agama tests             agama.yml      pull_request
```

The `act` command takes as argument the trigger event to simulate,
defaulting to the `push` event if none is specified, or, if only one event
is supported by the targeted workflows, it will default to that event.

Since all of the SUSE/connect-ng workflows support just the `pull_request`
event, running `act` without any arguments will attempt to run all of the
workflows locally in parallel.

To run a specific a specific workflow locally the full path to the workflow
file, e.g. [.github/workflows/agama.yml](.github/workflows/agama.yml) can be
specified with the `--workflows` (`-W`) option, e.g.

```bash
$ act -W .github/workflows/agama.yml pull_request
INFO[0000] Using docker host 'unix:///var/run/docker.sock', and daemon socket 'unix:///var/run/docker.sock'
INFO[0000] Start server on http://192.168.0.100:34567
[agama tests/Build SUSE/connect-ng inputs required by Agama] ⭐ Run Set up job
[agama tests/Build SUSE/connect-ng inputs required by Agama] 🚀  Start image=registry.suse.com/bci/golang:1.24-openssl
...
[agama tests/Run agama rust tests                          ] Cleaning up container for job Run agama rust tests
[agama tests/Run agama rust tests                          ]   ✅  Success - Complete job
[agama tests/Run agama rust tests                          ] 🏁  Job succeeded
```

Alternatively a specific job can be run by specifing its job id using the
`--job` (`-j`) option.  If the specified job depends on other jobsi, via
a `needs` attribute, they will also be run.  For example, the job id of
the final stage of the `agama.yml` workflow is `run-agama-rust-tests`,
and specifying it will also run the `build-suseconnect` job that it
depends upon, as follows:

```bash
$ gh act -j run-agama-rust-tests
INFO[0000] Using docker host 'unix:///var/run/docker.sock', and daemon socket 'unix:///var/run/docker.sock'
INFO[0000] Start server on http://192.168.0.100:34567
[agama tests/Build SUSE/connect-ng inputs required by Agama] ⭐ Run Set up job
[agama tests/Build SUSE/connect-ng inputs required by Agama] 🚀  Start image=registry.suse.com/bci/golang:1.24-openssl
...
[agama tests/Run agama rust tests                          ] Cleaning up container for job Run agama rust tests
[agama tests/Run agama rust tests                          ]   ✅  Success - Complete job
[agama tests/Run agama rust tests                          ] 🏁  Job succeeded
```