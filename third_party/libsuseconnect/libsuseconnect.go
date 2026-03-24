package main

// #include <stdlib.h>
// typedef void (*logLineFunc)(int level, const char* message);
// void log_bridge_fun(logLineFunc f, int level, const char* message);
import "C"

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"unsafe"

	"github.com/SUSE/connect-ng/internal/connect"
	cred "github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/SUSE/connect-ng/pkg/search"
)

// trace entry & exit to exported calls
func trace(fmt string, args ...interface{}) {
	// use util.Info for now as util.Debug output doesn't show up in /var/log/YaST2/y2log
	util.Info.Printf("libsuseconnect - "+fmt, args...)
}

// log level
const (
	llDebug   = 1
	llInfo    = 2
	llWarning = 3
	llError   = 4
	llFatal   = 5
)

// simple Writer interface implementation which forwards messages
// to log callback
type callbackWriter struct {
	level int
}

func (w callbackWriter) Write(p []byte) (n int, err error) {
	message := C.CString(string(p))
	C.log_bridge_fun(logFun, C.int(w.level), message)
	C.free(unsafe.Pointer(message))
	return len(p), nil
}

// function pointer to log callback passed from ruby
var logFun C.logLineFunc

//export set_log_callback
func set_log_callback(logCallback C.logLineFunc) {
	trace("set_log_callback - call args - logCallback: %v", logCallback)
	logFun = logCallback
	// NOTE: Debug is not redirected here as it is disabled by default
	util.Info.SetOutput(callbackWriter{llInfo})
	// TODO: add other levels?
	trace("set_log_callback - exit")
}

//export free_string
func free_string(str *C.char) {
	trace("free_string - call args - str: %s", C.GoString(str))
	C.free(unsafe.Pointer(str))
	trace("free_string - exit")
}

//export announce_system
func announce_system(clientParams, distroTarget *C.char) *C.char {
	trace("announce_system - call args - clientParams: %s, distroTarget: %s",
		C.GoString(clientParams), C.GoString(distroTarget))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	if err := api.Register(opts); err != nil {
		trace("announce_system - api.Register() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}

	creds, err := cred.ReadCredentials(cred.SystemCredentialsPath(opts.FsRoot))
	if err != nil {
		trace("announce_system - cred.ReadCredentials() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}

	login := creds.Username
	password := creds.Password

	var res struct {
		Credentials []string `json:"credentials"`
	}
	res.Credentials = []string{login, password, ""}
	jsn, _ := json.Marshal(&res)
	trace("announce_system - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export update_system
func update_system(clientParams, distroTarget *C.char) *C.char {
	trace("update_system - call args - clientParams: %s, distroTarget: %s",
		C.GoString(clientParams), C.GoString(distroTarget))
	opts := loadConfig(C.GoString(clientParams))

	api := connect.NewWrappedAPI(opts)
	if err := api.KeepAlive(opts.EnableSystemUptimeTracking); err != nil {
		trace("update_system - api.KeepAlive() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("update_system - exit")
	return C.CString("{}")
}

//export deactivate_system
func deactivate_system(clientParams *C.char) *C.char {
	trace("update_system - call args - clientParams: %s", C.GoString(clientParams))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	err := connect.Deregister(api, opts)
	if err != nil {
		trace("deactivate_system - connect.Deregister() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}

	trace("deactivate_system - exit")
	return C.CString("{}")
}

//export credentials
func credentials(path *C.char) *C.char {
	trace("credentials - call args - path: %s", C.GoString(path))
	creds, err := cred.ReadCredentials(C.GoString(path))
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, _ := json.Marshal(&creds)
	trace("credentials - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export create_credentials_file
func create_credentials_file(login, password, token, path *C.char) *C.char {
	trace("create_credentials_file - call args - login: %s, password: %s, token: %s, path: %s",
		C.GoString(login), "[REDACTED]", "[REDACTED]", C.GoString(path))
	credPath := C.GoString(path)

	if !filepath.IsAbs(credPath) {
		credPath = filepath.Join(cred.DefaultCredentialsDir, credPath)
	}

	err := cred.CreateCredentials(
		C.GoString(login), C.GoString(password), C.GoString(token), credPath)
	if err != nil {
		trace("create_credentials_file - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("create_credentials_file - exit")
	return C.CString("") // TODO need more consistent return path
}

//export curlrc_credentials
func curlrc_credentials() *C.char {
	trace("curlrc_credentials - call args")
	// NOTE: errors are ignored to match original
	creds, _ := cred.ReadCurlrcCredentials()
	jsn, _ := json.Marshal(&creds)
	trace("curlrc_credentials - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export show_product
func show_product(clientParams, product *C.char) *C.char {
	trace("show_product - call args - clientParams: %s, product: %s",
		C.GoString(clientParams), C.GoString(product))
	opts := loadConfig(C.GoString(clientParams))

	var productQuery registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &productQuery)
	if err != nil {
		trace("show_product - json.Unmarshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}

	wrapper := connect.NewWrappedAPI(opts)
	productData, err := registration.FetchProductInfo(wrapper.GetConnection(), productQuery.Identifier, productQuery.Version, productQuery.Arch)
	if err != nil {
		trace("show_product - registration.FetchProductInfo() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(productData)
	if err != nil {
		trace("show_product - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("show_product - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export activate_product
func activate_product(clientParams, product, email *C.char) *C.char {
	trace("activate_product - call args - clientParams: %s, product: %s, email: %s",
		C.GoString(clientParams), C.GoString(product), C.GoString(email))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	var p registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &p)
	if err != nil {
		trace("activate_product - json.Unmarshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	trace("activate_product - json.Unmarshal() - p: %+v", p)
	service, err := connect.ActivateProduct(api.GetConnection(), opts.Token, p)
	if err != nil {
		trace("activate_product - connect.ActivateProduct() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("activate_product - connect.ActivateProduct() - service: %+v", service)
	jsn, err := json.Marshal(service)
	if err != nil {
		trace("activate_product - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("activate_product - exit - jsn: %q", jsn)
	return C.CString(string(jsn))
}

//export activated_products
func activated_products(clientParams *C.char) *C.char {
	trace("activated_products - call args - clientParams: %s", C.GoString(clientParams))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	products, err := connect.ActivatedProducts(api.GetConnection())
	if err != nil {
		trace("activated_products - connect.ActivatedProducts() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(products)
	if err != nil {
		trace("activated_products - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("activated_products - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export deactivate_product
func deactivate_product(clientParams, product *C.char) *C.char {
	trace("deactivate_product - call args - clientParams: %s, product: %s",
		C.GoString(clientParams), C.GoString(product))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	var p registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &p)
	if err != nil {
		trace("deactivate_product - json.Unmarshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}

	metadata, pr, err := registration.Deactivate(api.GetConnection(), p.Identifier, p.Version, p.Arch)
	if err != nil {
		trace("deactivate_product - registration.Deactivate() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	service := &registration.Service{
		ID:            metadata.ID,
		URL:           metadata.URL,
		Name:          metadata.Name,
		ObsoletedName: metadata.ObsoletedName,
		Product:       *pr,
	}
	jsn, err := json.Marshal(service)
	if err != nil {
		trace("deactivate_product - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("deactivate_product - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export get_config
func get_config(path *C.char) *C.char {
	trace("get_config - call args - path: %s", C.GoString(path))
	opts, _ := connect.ReadFromConfiguration(C.GoString(path))
	jsn, err := json.Marshal(opts)
	if err != nil {
		trace("get_config - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("get_config - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export write_config
func write_config(clientParams *C.char) *C.char {
	trace("write_config - call args - clientParams: %s", C.GoString(clientParams))
	opts := loadConfig(C.GoString(clientParams))

	err := opts.SaveAsConfiguration()
	if err != nil {
		trace("write_config - opts.SaveAsConfiguration() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("write_config - exit")
	return C.CString("{}")
}

func loadConfig(clientParams string) *connect.Options {
	// unmarshal extra config fields only for local use
	var extConfig struct {
		Debug string `json:"debug"`
	}
	json.Unmarshal([]byte(clientParams), &extConfig)
	// enable debug output if "debug" was set in json
	if v, _ := strconv.ParseBool(extConfig.Debug); v {
		trace("loadConfig - enable debug output")
		util.Debug.SetOutput(callbackWriter{llDebug})
	}

	// Read the options from the default configuration path and merge the
	// provided clientParams into as well.
	opts, _ := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	_ = json.Unmarshal([]byte(clientParams), opts)

	trace("loadConfig - exit - opts: %+v", opts)
	return opts
}

func certToPEM(cert *x509.Certificate) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
}

func certsToPEM(certs []*x509.Certificate) string {
	slices.Reverse(certs)
	var pemString string
	for _, cert := range certs {
		pemString += certToPEM(cert)
	}
	return pemString
}

func errorToJSON(err error) string {
	trace("errorToJSON - call args - err: (%T) %v", err, err)
	var s struct {
		ErrType string `json:"err_type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
		// [optional] auxiliary error data
		Data string `json:"data,omitempty"`
	}

	// map Go x509 errors to OpenSSL verify return values
	// see: https://www.openssl.org/docs/man1.0.2/man1/verify.html
	sslErrorMap := map[int]int{
		int(x509.Expired): 10, // X509_V_ERR_CERT_HAS_EXPIRED
		// TODO: add other values as needed
	}

	if ae, ok := err.(*connection.ApiError); ok {
		trace("errorToJSON - (connection.ApiError) - ae: %v", ae)
		s.ErrType = "APIError"
		s.Code = ae.Code
		s.Message = ae.Message
	} else if uerr, ok := err.(*url.Error); ok {
		trace("errorToJSON - (url.Error) - uerr: %v", uerr)
		ierr := errors.Unwrap(err)
		if uerr.Timeout() {
			trace("errorToJSON - Timeout - ierr: %v", ierr)
			s.ErrType = "Timeout"
			s.Message = ierr.Error()
		} else if ce, ok := ierr.(x509.CertificateInvalidError); ok {
			trace("errorToJSON - SSLError CertInvalid - ierr: %v", ierr)
			s.ErrType = "SSLError"
			s.Message = ierr.Error()
			s.Data = certToPEM(ce.Cert)
			s.Code = sslErrorMap[int(ce.Reason)]
		} else if ce, ok := ierr.(x509.UnknownAuthorityError); ok {
			trace("errorToJSON - SSLError UnknownAuthority - ierr: %v", ierr)
			s.ErrType = "SSLError"
			s.Message = ierr.Error()
			s.Data = certToPEM(ce.Cert)
			// this could be:
			// 18 (X509_V_ERR_DEPTH_ZERO_SELF_SIGNED_CERT),
			// 19 (X509_V_ERR_SELF_SIGNED_CERT_IN_CHAIN) or
			// 20 (X509_V_ERR_UNABLE_TO_GET_ISSUER_CERT_LOCALLY)
			s.Code = 19 // this seems to best match original behavior
		} else if ce, ok := ierr.(*tls.CertificateVerificationError); ok {
			trace("errorToJSON - SSLError CertVerification - ierr: %v", ierr)
			// starting with go1.20, we receive this error (https://go.dev/doc/go1.20#crypto/tls)
			s.ErrType = "SSLError"
			s.Message = ierr.Error()
			s.Data = certsToPEM(ce.UnverifiedCertificates)
			// this could be:
			// 18 (X509_V_ERR_DEPTH_ZERO_SELF_SIGNED_CERT),
			// 19 (X509_V_ERR_SELF_SIGNED_CERT_IN_CHAIN) or
			// 20 (X509_V_ERR_UNABLE_TO_GET_ISSUER_CERT_LOCALLY)
			s.Code = 19 // this seems to best match original behavior
		} else if ce, ok := ierr.(x509.HostnameError); ok {
			trace("errorToJSON - SSLError Hostname - ierr: %v", ierr)
			s.ErrType = "SSLError"
			s.Message = ierr.Error()
			// ruby version doesn't have this but it might be useful
			s.Data = certToPEM(ce.Certificate)
		} else if _, ok := ierr.(*net.OpError); ok {
			trace("errorToJSON - NetError - ierr: %v", ierr)
			s.ErrType = "NetError"
			s.Message = ierr.Error()
		} else {
			util.Debug.Printf("url.Error: %T: %v", ierr, err)
			trace("errorToJSON - unrecognised url.Error - ierr: (%T) %v", ierr, ierr)
			s.Message = err.Error()
		}
	} else if je, ok := err.(connect.JSONError); ok {
		trace("errorToJSON - JSONError - err: %v", je)
		s.ErrType = "JSONError"
		s.Message = errors.Unwrap(je).Error()
	} else {
		switch err {
		case cred.ErrMalformedSccCredFile:
			s.ErrType = "MalformedSccCredentialsFile"
		case cred.ErrMissingCredentialsFile:
			s.ErrType = "MissingCredentialsFile"
		}
		util.Debug.Printf("Error: %T: %v", err, err)
		trace("errorToJSON - %s url.Error - ierr: (%T) %v", s.ErrType, err, err)
		s.Message = err.Error()
	}

	trace("errorToJSON - struct - s: %+v", s)
	jsn, _ := json.Marshal(&s)
	trace("errorToJSON - exit - jsn: %s", string(jsn))
	return string(jsn)
}

//export getstatus
func getstatus(format *C.char) *C.char {
	trace("getstatus - call args - format: %s", C.GoString(format))
	opts, _ := connect.ReadFromConfiguration(connect.DefaultConfigPath)

	gFormat := C.GoString(format)
	var f connect.StatusFormat
	if gFormat == "text" {
		f = connect.StatusText
	} else {
		f = connect.StatusJSON
	}
	output, err := connect.GetProductStatuses(opts, f)
	if err != nil {
		trace("getstatus - err: (%T) %q", err, err)
		return C.CString(err.Error())
	}
	trace("getstatus - exit - output: %s", output)
	return C.CString(output)
}

//export update_certificates
func update_certificates() *C.char {
	trace("update_certificates - call args")
	// NOTE: this is no longer relevant, but we keep it for
	// backwards-compatibility.
	trace("update_certificates - exit - no longer relevant")
	return C.CString("{}")
}

//export reload_certificates
func reload_certificates() *C.char {
	trace("reload_certificates - call args")
	// NOTE: this is no longer relevant, but we keep it for
	// backwards-compatibility.
	trace("reload_certificates - exit - no longer relevant")
	return C.CString("{}")
}

//export list_installer_updates
func list_installer_updates(clientParams, product *C.char) *C.char {
	trace("list_installer_updates - call args - clientParams: %s, product: %s",
		C.GoString(clientParams), C.GoString(product))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	var productQuery registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &productQuery)
	if err != nil {
		trace("list_installer_updates - json.Unmarshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	repos, err := connect.InstallerUpdates(api.GetConnection(), productQuery)
	if err != nil {
		trace("list_installer_updates - connect.InstallerUpdates() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(repos)
	if err != nil {
		trace("list_installer_updates - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("list_installer_updates - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export system_migrations
func system_migrations(clientParams, products *C.char) *C.char {
	trace("system_migrations - call args - clientParams: %s, products: %s",
		C.GoString(clientParams), C.GoString(products))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	installed := make([]registration.Product, 0)
	err := json.Unmarshal([]byte(C.GoString(products)), &installed)
	if err != nil {
		trace("system_migrations - json.Unmarshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	migrations, err := connect.ProductMigrations(api.GetConnection(), installed)
	if err != nil {
		trace("system_migrations - connect.ProductMigrations() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(migrations)
	if err != nil {
		trace("system_migrations - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("system_migrations - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export offline_system_migrations
func offline_system_migrations(clientParams, products, targetBaseProduct *C.char) *C.char {
	trace("offline_system_migrations - call args - clientParams: %s, products: %s, targetBaseProduct: %s",
		C.GoString(clientParams), C.GoString(products), C.GoString(targetBaseProduct))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	installed := make([]registration.Product, 0)
	err := json.Unmarshal([]byte(C.GoString(products)), &installed)
	if err != nil {
		trace("offline_system_migrations - json.Unmarshal(products) - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	var target registration.Product
	if err := json.Unmarshal([]byte(C.GoString(targetBaseProduct)), &target); err != nil {
		trace("offline_system_migrations - json.Unmarshal(targetBaseProduct) - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	migrations, err := connect.OfflineProductMigrations(api.GetConnection(), installed, target)
	if err != nil {
		trace("offline_system_migrations - connect.OfflineProductMigrations() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(migrations)
	if err != nil {
		trace("offline_system_migrations - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("offline_system_migrations - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export upgrade_product
func upgrade_product(clientParams, product *C.char) *C.char {
	trace("upgrade_product - call args - clientParams: %s, product: %s",
		C.GoString(clientParams), C.GoString(product))
	opts := loadConfig(C.GoString(clientParams))

	var prod registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &prod)
	if err != nil {
		trace("upgrade_product - json.Unmarshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}

	conn := connect.NewWrappedAPI(opts)
	meta, pr, err := registration.Upgrade(conn.GetConnection(), prod.Identifier, prod.Version, prod.Arch)
	service := &registration.Service{
		ID:            meta.ID,
		URL:           meta.URL,
		Name:          meta.Name,
		ObsoletedName: meta.ObsoletedName,
		Product:       *pr,
	}
	if err != nil {
		trace("upgrade_product - registration.Upgrade() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(service)
	if err != nil {
		trace("upgrade_product - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("upgrade_product - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export synchronize
func synchronize(clientParams, products *C.char) *C.char {
	trace("synchronize - call args - clientParams: %s, products: %s",
		C.GoString(clientParams), C.GoString(products))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	prods := make([]registration.Product, 0)
	err := json.Unmarshal([]byte(C.GoString(products)), &prods)
	if err != nil {
		trace("synchronize - json.Unmarshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	activated, err := connect.SyncProducts(api.GetConnection(), prods)
	if err != nil {
		trace("synchronize - connect.SyncProducts() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(activated)
	if err != nil {
		trace("synchronize - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("synchronize - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export system_activations
func system_activations(clientParams *C.char) *C.char {
	trace("system_activations - call args - clientParams: %s", C.GoString(clientParams))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	activations, err := registration.FetchActivations(api.GetConnection())
	if err != nil {
		trace("system_activations - registration.FetchActivations() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(activations)
	if err != nil {
		trace("system_activations - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("system_activations - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

//export search_package
func search_package(clientParams, product, query *C.char) *C.char {
	trace("search_package - call args - clientParams: %s, product: %s, query: %s",
		C.GoString(clientParams), C.GoString(product), C.GoString(query))
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	var p registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &p)
	if err != nil {
		trace("search_package - json.Unmarshal(product) - err: (%T) %q", err, err)
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}

	results, err := search.Package(api.GetConnection(), C.GoString(query), p.ToTriplet())
	if err != nil {
		trace("search_package - search.Package() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(results)
	if err != nil {
		trace("search_package - json.Marshal() - err: (%T) %q", err, err)
		return C.CString(errorToJSON(err))
	}
	trace("search_package - exit - jsn: %s", string(jsn))
	return C.CString(string(jsn))
}

func main() {}
