## Running feature tests

To run feature tests locally the easiest way is:

```
 # Make sure your .env is populated!
 $ cp .env-example .env
 # You need a working docker setup for this
 $ make feature-tests

```

If you do not want to run a full rpm build process per feature test run:

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

With the above it is possible to rebuild the binary and try the feature tests on the new binary quickly.
**Note:** Be aware that installing `suseonnect` via `rpm` will shadow the existing executable!
