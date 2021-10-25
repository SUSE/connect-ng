# Functional differences between Go and Ruby implementations of SUSEConnect:

- empty `namespace` config/argument is treated the same as no namespace
- null string values received from API are silently converted to empty strings
  when data is received. when the same data is sent back to API, fields with
  "omitempty" tag will not be included in JSON (example: Product.release_type).
- API calls which pass JSON body as "query" (e.g. upgradeProduct() or
  deactivateProduct()) can include unexpected attributes (mostly bools) which
  don't support "omitempty" tag. API seems to ignore these correctly.
- When proxy credentials are incorrect, go version returns different error
  message than the original ruby one. Both are misleading and don't indicate
  any proxy related problems.
- When doing API calls, ruby version tries to parse all non-empty responses
  as JSON before checking HTTP return codes. With incorrect configuration
  and/or proxy with invalid credentials this leads to different error messages
  when response is non-JSON and has non-success HTTP return code.
  Go version will handle this as "API error" while ruby version will fail to
  parse the response and handle this as "JSON error".
- The Go HTTP client tries to reuse connections with keep-alive.
- Docs for GET APIs call for URL-encoded query params. Ruby version sends a
  JSON query in body (like for other verbs). Go implementation follows docs.
  This difference is mostly visible in `connect.showProduct()` API call.
- The connect.syncProducts() returns deserialized slice of Product
  objects. Original code returns raw body.
- Original `SUSEConnect --rollback` ignores CLI arguments like `--debug`.
  This is a bug and it's already fixed in the Go version.
- With --debug the Go version sends all debug output to stderr. The Ruby
  version sends http debug to stderr, and other debug to stdout.
- In zypper-migration plugin, `--download <mode>` flag doesn't validate `<mode>`.
- Additional `--debug` flag was added to zypper-migration plugin to enable
  `SUSEConnect` debug info.
- Contradicting flags are not allowed in zypper-migration plugin to match new
  zypper behavior (see e.g.: https://github.com/openSUSE/zypper/pull/215 for
  more details).
- Zypper backup doesn't use parallel gzip but tar's built in gzip functionality
  which doesn't require shell pipes.
- Zypper backup stores both tarball and restore script under `--root` path.
- Self-update in zypper-migration plugin returns more detailed error information
  on failure.
- Package search only reports missing API for 404 responses if there's no error
  message returned (e.g. "base product not found")
