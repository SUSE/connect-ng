# Similate client interactions with SCC or RMT server

The `public-api-demo` can be used for adhoc testing purposes to similate client interactions with an upstream SCC or RMT server.

# Usage

Assuming that the command has been built using `make build` at the top-level it will be available as `out/public-api-demo` under the top-level directory.

```bash
% out/public-api-demo
public-api-demo: A connect client library demo
./public-api-demo IDENTIFIER VERSION ARCH <system_information>
```

The command runs through a number of stages, exercising the corresponding client interaction with the server at each stage, pausing at the end of each stage waiting for user input before continuing.

This allows a developer to check by other means that client <=> server interactions have had the expected effects, e.g. that relevant fields have been populated in a database table.

## Command Line Arguments

The command takes 3 required and a 4th optional argument:

* `IDENTIFIER` - the product stream identifier, e.g. `SLES`.
* `VERSION` - the product codestream version, e.g. `15.7` for SLE 15 SP7 or `16.0` for SLE 16.0.
* `ARCH` - the architecture, e.g. `x86_64`.
* `<system_information>` - an optional path to a file containing a system information JSON blob.

NOTE: The `IDENTIFIER`, `VERSION`, and `ARCH` arguments, when joined using `/` should form a valid product triplet, e.g. `SLES/15.7/x86_64`.

## Environment Variables

The following environment variables can be used to control what the `public-api-demo` does:

* `REGCODE` - specifies the product registration code to use when registering a simulated client against an upstream SCC server; registration codes are optional when registering against an RMT.
* `SCC_HOST` - allows specification of an alternate upstream SCC target URL, e.g. a local SCC dev-env instance.
* `RMT_HOST` - allows specification of an upstream RMT target URL, e.g. a real RMT or local RMT dev-env instance. Note that this will also disable certain interactions that are not supported by RMTs.
* `API_CERT` - specifies the path to a cert file; required when SCC_HOST or RMT_HOST specified an HTTPS target using a self-signed certificate.
* `NON_INTERACTIVE` - disables to pause for user input after each demo stage.
* `PROFILES_DIR` - specifies a directory that should contain files named for simulated profile types, e.g. `pci_data`, `mod_list` or `rpm_packages`, that should contain the corresponding data field for that profile type as a JSON blob.
* `DISABLE_TOKEN_HANDLING` - disables token handling when interacting with the upstream server.
* `TRACE_CREDENTIAL_UPDATES` - enabled verbose tracing of updates to the client's SCC credentials.

NOTE: If both `SCC_HOST` and `RMT_HOST` are specified, the `RMT_HOST` url will be used.