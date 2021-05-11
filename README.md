# Go SuseConnect

PoC - evaluate rewriting SUSEConnect in Go.

Only the json status option (-s or --status) is implemented.

### Build
`go build cmd/suseconnect.go`

### Build in container
`cd go-connect`
`podman run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.16 go build -v cmd/suseconnect.go`
This will leave a `suseconnect` binary on the host.

### Shared library
`go build -buildmode=c-shared -o libsuseconnect.so ext/main.go`

See `ext/use-lib.py` for example use from python.
TODO ruby.
