Shim for YaST
=============

Yast-Registration uses the original SUSEConnect library and both are written in Ruby. However connect-ng is written in Go, so if it is to replace the original Ruby SUSEConnect, then it will have to provide a shim written in Ruby.

The shim will provide the same interface to Yast as the Ruby SUSEConnect, but underneath it will call the Go connect-ng via a shared library to do the work.

Layers
======

```
    |---------------------------|
    | YaST Registration (Ruby)  |
    |---------------------------|
    | Shim (Ruby)               |
    |---------------------------|
    | libsuseconnect (Cgo)      |
    |---------------------------|
    | connect-ng (Go)           |
    |---------------------------|

```

YaST Registration
-----------------
No change here. It will import the library as before like `require "suse/connect"`.

The packaging for connect-ng will install the new shim gem instead of the suse-connect gem.


Shim
----
Northbound the shim provides the same interface to Yast as the Ruby SUSEConnect. Classes currently implemented in the Ruby SUSEConnect and used by Yast will need to be implemented here.

YaST passes strings and hashes as parameters. The hashes will be converted to JSON and passed on to libsuseconnect as strings.

Southbound the shim uses the Ruby [FFI](https://github.com/ffi/ffi) (Foreign Function Interface) gem to interface with the C library provided by libsuseconnect. Libsuseconnect will return JSON with the results or error details if there was a problem. The shim will raise an exception if there was an error. Otherwise the results will be returned in whatever form YaST is expecting.


libsuseconnect
--------------
[Cgo](https://golang.org/cmd/cgo/) allows C shared libraries to be built from Go packages. See the `build-so` target in [Makefile](../Makefile) which builds libsuseconnect.so.

This layer provides a set of C functions that the shim will call. For each function it converts the C strings received in the parameters to Go strings and decodes any JSON. Then it calls the appropriate function in the `internal/connect` package. It checks the return for errors, and builds the appropriate JSON response to include the results and any error information.


connect-ng
----------
No change to this layer.


Example
=======
Currently `yast-registration` [here](https://github.com/yast/yast-registration/blob/134f553e0a0ea75e94b095cfd1fa1fb0fa9bca75/src/lib/registration/registration.rb#L58) calls `SUSE::Connect::YaST.announce_system` [here](https://github.com/SUSE/connect/blob/1b377c95cd08e4e536cbf3d6707eaa1ef21c412e/lib/suse/connect/yast.rb#L21-L28)

The shim will implement the `announce_system` method with the same parameters, return and exceptions. It will convert the `client_params` hash parameter to JSON so it can be passed as a string parameter to the library using FFI.

The libsuseconnect layer will provide a `announce_system` function that takes a `client_params` JSON string and a `distro_target` string. It will convert these to Go strings. Then load the default config, unmarshal the JSON and merge those settings.

Then it will call [AnnounceSystem](https://github.com/SUSE/connect-ng/blob/c534203603e3c7d5c3064c1de340b2195996992a/internal/connect/client.go#L174) which returns a login, password and error. C does not allow multiple return values, so these will need to be JSON encoded so they can be passed back together as one string. The details from any error returned from `AnnounceSystem()` will need to be in the JSON passed back.

The shim layer will decode the JSON returned from libsuseconnect. It will check for errors and raise appropriate exceptions so they can be handled by YaST [here](https://github.com/yast/yast-registration/blob/134f553e0a0ea75e94b095cfd1fa1fb0fa9bca75/src/lib/registration/connect_helpers.rb#L63). Otherwise it will return the login and password up to YaST.
