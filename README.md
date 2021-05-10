# Go SuseConnect

PoC - evaluate rewriting SUSEConnect in Go.

Only the json status option (-s or --status) is implemented.

### Build
`go build cmd/suseconnect.go`

### Shared library
`go build -buildmode=c-shared -o libsuseconnect.so ext/main.go`

See `ext/use-lib.py` for example use from python.
TODO ruby.
