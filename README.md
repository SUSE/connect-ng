# Go SuseConnect

PoC - evaluate rewriting SUSEConnect in Go.

Only the status options are implemented (--status and --status-text).

### Build
Requires Go 1.16 for [embed](https://pkg.go.dev/embed).
```
go build cmd/suseconnect.go
```

### Build in container
```
cd go-connect
podman run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.16 go build -v cmd/suseconnect.go
```
This will leave a `suseconnect` binary on the host.

### Shared library
`go build -buildmode=c-shared -o libsuseconnect.so ext/main.go`

See `ext/use-lib.rb` for example use from ruby.
See `ext/use-lib.py` for example use from python.

## Examples
```
# ./suseconnect --status
[{"identifier":"SUSE-MicroOS","version":"5.0","arch":"x86_64","status":"Registered","regcode":"INTERNAL-USE-ONLY-116f-4b58","starts_at":"2021-04-21 15:08:32 UTC","expires_at":"2026-04-21 15:08:32 UTC","subscription_status":"ACTIVE","type":"internal"}]
```
#### HTTP proxy
```
# podman run --name squid -d -p 3128:3128 datadog/squid
# HTTPS_PROXY=127.0.0.1:3128 ./suseconnect --status-text
Installed Products:
------------------------------------------

  SUSE Linux Enterprise Micro 5.0
  (SUSE-MicroOS/5.0/x86_64)

  Registered
  
    Subscription:

    Regcode: INTERNAL-USE-ONLY-116f-4b58
    Starts at: 2021-04-21 15:08:32 UTC
    Expires at: 2026-04-21 15:08:32 UTC
    Status: ACTIVE
    Type: internal
  

------------------------------------------
```
