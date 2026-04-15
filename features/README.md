## Running feature tests

To run feature tests locally the easiest way is:

```
 # Make sure your .env is populated!
 $ cp .env-example .env
 # You need a working docker setup for this
 $ make feature-tests

```

This will create a new container, build an RPM inside of it and run feature
tests with the newly installed RPM. This is what the CI ends up doing.

If you do not want to run a full RPM build process you can do the following:

```
 # Make sure your .env is populated!
 $ cp .env-example .env
 $ docker run --rm --privileged --env-file .env -ti -v $(pwd):/connect registry.suse.com/bci/golang:1.24-openssl
 > git config --global --add safe.directory /connect
 > cd /connect
 > make build
 > ln -s /connect/out/suseconnect /usr/local/bin/suseconnect

 # Run the full test suite for suseconnect for example:
 > go test -v features/suseconnect/*
```

With the above it is possible to rebuild the binary and try the feature tests on
the new binary quickly.

### Words of warning

Be aware that installing `suseconnect` via `rpm` will shadow the existing
executable.

Also, feature tests modify the existing filesystem by adding/removing certain
configuration files. Hence, make sure to run this into a containerized scenario
if you don't want unexpected surprises. That's why `make feature-tests` runs in
a container, and why the above example does it too.

### Special Notes for Kubernetes Provider Info Feature Tests

The kubernetes provider info feature tests require a valid functional systemd
environment to work and customise that environment to add sumulated kubernetes
provider services.

As such they will not work inside the normal containerised feature testing
environment, and require either a VM or customised container image with
systemd installed and started. See below for details on how to do that.

#### Environment variables used to manage kubernetes provider info tests

As such these feasture tests will be skipped by default unless the required
environment variable `KUBERNETES_PROVIDER_TESTS_ENABLED` is set in the
environment.

Additionally setting the `KUBERNETES_PROVIDER_TESTS_NO_CLEANUP` environment
variable will disable the cleanup of the system modifications which can be
helpful for adhoc testings and debugging.

#### Starting a container with systemd enabled

The recommended approach for this is to use `podman` rather than `docker` as
`podman` is more friendly to systemd.

You will need a customised container image which has systemd installed which
can be accomplished using a Dockerfile like the following:

```dockerfile
FROM registry.suse.com/bci/golang:1.24-openssl

# Install systemd and make
RUN zypper -n install \
    systemd \
    dmidecode \
    make \
    jq \
    && \
    zypper clean --all

# Set up systemd as the default entrypoint
# Note: Requires running with privileged mode or specific systemd volumes in podman
CMD ["/usr/lib/systemd/systemd"]
```

The kubernetes provider info tests can now be run as follows:

* build a testing image using the above Dockerfile
* start a named background container using that image with systemd enabled and repo dir mounted as /connect
* exec into the container to run a bash shell
* build the suseconnect and symlink it into the accessible path
* run the kubernetes provider tests with K8S_PROVIDER_TESTS_ENABLED=true set in the environment

```bash
$ podman build -t k8s-provider-tester -F /tmp/Dockerfile.k8s-tester
$ podman run -d --name k8s-testing --systemd=true --env-file .env -v $(pwd):/connect k8s-provider-tester
$ podman exec -it k8s-testing /bin/bash
> git config --global --add safe.directory /connect
> cd /connect
> make build
> ln -s /connect/out/suseconnect /usr/local/bin/suseconnect
> env K8S_PROVIDER_TESTS_ENABLED=true go test -v features/suseconnect/kubernetes_provider_info_test.go
```
