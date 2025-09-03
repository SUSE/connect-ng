## connect-ng CI setup

Our CI setup runs the following steps:

### Lint and unit tests

workflow definition: [.github/workflows/lint-unit.yml](https://github.com/SUSE/connect-ng/blob/main/.github/workflows/lint-unit.yml)

This workflow runs our unit tests + the usual `gofmt` to insure we have a broad coverage in functionality and good code style.

**Running unit tests locally**

There is no special mechanism needed to run these steps locally. Check the workflow for hints how to run unit tests

### CLI feature tests

workflow definition: [.github/workflows/features.yml](https://github.com/SUSE/connect-ng/blob/main/.github/workflows/features.yml)

This workflow runs our simple CLI feature tests and build the rpm beforehand and runs feature test we imported from the deprecated
ruby connect version. Check [features/](https://github.com/SUSE/connect-ng/tree/main/features) for more information.

**Requirements to run feature tests locally**

To run feature tests locally, you need:

- A checkout of connect-ng
- You need multiple secrets to run the actual feature tests, since they register and deregister with SCC in the feature tests

```
# Add this to your .env file
BETA_VALID_REGCODE=
VALID_REGCODE=
EXPIRED_REGCODE=
NOT_ACTIVATED_REGCODE=
BETA_NOT_ACTIVATED_REGCODE=
```

**Running tasks in the container**

We use the official SUSE Golang container, providing all we need to run the tests:

```
export IMAGE="registry.suse.com/bci/golang:1.24-openssl"
# Run the required container:
$ docker run --rm -it --env-file .env -v $(pwd):/usr/src/connect-ng $IMAGE

# To build the connect-ng rpms within the container use:
$ docker run --rm -it -v $(pwd):/usr/src/connect-ng $IMAGE 'build/ci/build-rpm'

# Run feature tests in the container
$ docker run --rm -it --env-file .env -v $(pwd):/usr/src/connect-ng $IMAGE bash -c 'build/ci/build-rpm && build/ci/configure && build/ci/run-feature-tests'
```
