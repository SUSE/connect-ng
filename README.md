# SUSEConnect-ng

SUSEConnect-ng is a work-in-progress project to rewrite [SUSEConnect](https://github.com/SUSE/connect) in Golang.

SUSEConnect is a command line tool for connecting a client system to the SUSE Customer Center.
It will connect the system to your product subscriptions and enable the product repositories/services locally.

SUSEConnect is distributed as RPM for all SUSE distributions and gets built in
the [openSUSE build service](https://build.opensuse.org/package/show/systemsmanagement:SCC/SUSEConnect).

Please visit https://scc.suse.com to see and manage your subscriptions.

SUSEConnect communicates with SCC over this [REST API](https://github.com/SUSE/connect/blob/master/doc/SCC-API-%28Implemented%29.md).

### Build
Requires Go 1.16 for [embed](https://pkg.go.dev/embed).
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
