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
 $ docker run --rm --privileged --env-file .env -ti -v $(pwd):/connect registry.suse.com/bci/golang:1.21-openssl
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
