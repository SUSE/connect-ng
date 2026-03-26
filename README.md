# SUSEConnect-ng

[![build result](https://build.opensuse.org/projects/systemsmanagement:SCC/packages/suseconnect-ng/badge.svg?type=default)](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)

SUSEConnect is a Golang command line tool for connecting a client system to the SUSE Customer Center.
It will connect the system to product subscriptions and enable the product repositories/services locally.

SUSEConnect-ng reduces the size of its runtime dependencies compared to the
replaced [Ruby SUSEConnect](https://github.com/SUSE/connect).

SUSEConnect-ng is distributed as RPM for all SUSE distributions and gets built in
the [openSUSE build service](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng).

Please visit https://scc.suse.com to view and manage subscriptions.

SUSEConnect-ng communicates with SCC over this [REST API](https://github.com/SUSE/connect/blob/master/doc/SCC-API-%28Implemented%29.md).

### Build
Requires Go >= 1.21

```
make build
```
This will create a `out/suseconnect` binary.

### Build in container
If the local system does not have a go compiler installed, the build can be done in a container: 
```
docker run --rm -v $(pwd):/connect registry.suse.com/bci/golang:1.21-openssl sh -c "git config --global --add safe.directory /connect; cd /connect; make build"
```
This will create a `out/suseconnect` binary on the host.

### Testing
Pull requests must pass CI testing to be approved. These tests can be run locally before the PR is created. There are three sets of CI tests:
* Format, unit, and build tests
* Feature tests
* Yast tests

#### Unit Tests
Users can run all unit tests by running `make test`. Unit
tests for a specific package can be run it as for any Go
project, for example: `go test ./internal/collectors/`.

#### Format Test
Users can run the go format test by running `make check-format`

#### Feature Tests

Feature tests require an `.env` file in the root directory
of the project. There is an .env-example; copy .env-example to .env and update the following entries.

``` sh
REGCODE=
HA_REGCODE=
EXPIRED_REGCODE=

```

These values can be picked up from Glue's production environment. Once that is
done, run `make feature-tests`. This will run all feature
tests inside of a container by using the registration codes as provided by the
`.env` file.

**Running Featre tests in a container**

The Feture tests can be run using the official SUSE Golang container:

```
export IMAGE="registry.suse.com/bci/golang:1.21-openssl"
# Run the required container:
$ docker run --rm -it --env-file .env -w /usr/src/connect-ng -v $(pwd):/usr/src/connect-ng $IMAGE

# To build the connect-ng rpms within the container use:
$ docker run --rm -it -w /usr/src/connect-ng -v $(pwd):/usr/src/connect-ng $IMAGE 'build/ci/build-rpm'

# Run feature tests in the container
$ docker run --rm -it --env-file .env -w /usr/src/connect-ng -v $(pwd):/usr/src/connect-ng $IMAGE bash -c 'build/ci/build-rpm && build/ci/configure && build/ci/run-feature-tests'
```

### YaST integration tests / yast (pull_request)
The YaST integration tests are run via the Makefile
```
make test-yast
```

#### Workflow Definitions

##### Lint and unit tests
workflow definition: [.github/workflows/lint-unit.yml](https://github.com/SUSE/connect-ng/blob/main/.github/workflows/lint-unit.yml)

##### Feature Tests
workflow definition: [.github/workflows/features.yml](https://github.com/SUSE/connect-ng/blob/main/.github/workflows/features.yml)

##### Yast Test
workflow definition: [.github/workflows/yast.yml](https://github.com/SUSE/connect-ng/blob/main/.github/workflows/yast.yml)



