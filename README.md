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
If you don't have a go compiler installed, you can run the build in a container: 
```
docker run --rm --privileged -ti -v $(pwd):/connect registry.suse.com/bci/golang:1.21-openssl cd connect; make build
```
This will create a `out/suseconnect` binary on the host.

### Testing

You can run all unit tests by running `make test`. If you then want to run unit
tests for a specific package, you can simply run it as you would do for any Go
project, for example: `go test ./internal/collectors/`.

For feature tests you first need to create an `.env` file in the root directory
of the project with the following contents:

``` sh
BETA_VALID_REGCODE="<regcode>"
BETA_NOT_ACTIVATED_REGCODE="<regcode>"
VALID_REGCODE="<regcode>"
EXPIRED_REGCODE="<regcode>"
NOT_ACTIVATED_REGCODE="<regcode>"
```

These values can be picked up from Glue's production environment. Once that is
done, you can then simply run `make feature-tests`. This will run a all feature
tests inside of a container by using the registration codes as provided by the
`.env` file.
