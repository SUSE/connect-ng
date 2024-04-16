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
Requires Go >= 1.21

```
make build
```
This will create a `out/suseconnect` binary.

### Build in container
```
cd connect-ng
podman run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.16 make build
```
This will create a `out/suseconnect` binary on the host.

### Testing

Run the unit tests: `make test`
Run selected unit tests, eg: `go test ./internal/collectors/`
