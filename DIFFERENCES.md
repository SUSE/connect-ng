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
