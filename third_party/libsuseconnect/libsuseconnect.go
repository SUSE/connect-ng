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
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/SUSE/connect-ng/pkg/search"
)

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
	logFun = logCallback
	// NOTE: Debug is not redirected here as it is disabled by default
	util.Info.SetOutput(callbackWriter{llInfo})
	// TODO: add other levels?
}

//export free_string
func free_string(str *C.char) {
	C.free(unsafe.Pointer(str))
}

//export announce_system
func announce_system(clientParams, distroTarget *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	if err := connect.Register(api, opts); err != nil {
		return C.CString(errorToJSON(err))
	}

	creds, err := cred.ReadCredentials(cred.SystemCredentialsPath(opts.FsRoot))
	if err != nil {
		return C.CString(errorToJSON(err))
	}

	jsn, _ := json.Marshal(&creds)
	return C.CString(string(jsn))
}

//export update_system
func update_system(clientParams, distroTarget *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))

	api := connect.NewWrappedAPI(opts)
	if err := api.KeepAlive(); err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString("{}")
}

//export deactivate_system
func deactivate_system(clientParams *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	err := connect.Deregister(api, opts)
	if err != nil {
		return C.CString(errorToJSON(err))
	}

	return C.CString("{}")
}

//export credentials
func credentials(path *C.char) *C.char {
	creds, err := cred.ReadCredentials(C.GoString(path))
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, _ := json.Marshal(&creds)
	return C.CString(string(jsn))
}

//export create_credentials_file
func create_credentials_file(login, password, token, path *C.char) *C.char {
	credPath := C.GoString(path)

	if !filepath.IsAbs(credPath) {
		credPath = filepath.Join(cred.DefaultCredentialsDir, credPath)
	}

	err := cred.CreateCredentials(
		C.GoString(login), C.GoString(password), C.GoString(token), credPath)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString("") // TODO need more consistent return path
}

//export curlrc_credentials
func curlrc_credentials() *C.char {
	// NOTE: errors are ignored to match original
	creds, _ := cred.ReadCurlrcCredentials()
	jsn, _ := json.Marshal(&creds)
	return C.CString(string(jsn))
}

//export show_product
func show_product(clientParams, product *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))

	var productQuery registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &productQuery)
	if err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}

	wrapper := connect.NewWrappedAPI(opts)
	productData, err := registration.FetchProductInfo(wrapper.GetConnection(), productQuery.Identifier, productQuery.Version, productQuery.Arch)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(productData)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export activate_product
func activate_product(clientParams, product, email *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	var p registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &p)
	if err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	service, err := connect.ActivateProduct(api.GetConnection(), opts.Token, p)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(service)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export activated_products
func activated_products(clientParams *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	products, err := connect.ActivatedProducts(api.GetConnection())
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(products)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export deactivate_product
func deactivate_product(clientParams, product *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	var p registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &p)
	if err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}

	metadata, pr, err := registration.Deactivate(api.GetConnection(), p.Identifier, p.Version, p.Arch)
	if err != nil {
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
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export get_config
func get_config(path *C.char) *C.char {
	opts, _ := connect.ReadFromConfiguration(C.GoString(path))
	jsn, err := json.Marshal(opts)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export write_config
func write_config(clientParams *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))

	err := opts.SaveAsConfiguration()
	if err != nil {
		return C.CString(errorToJSON(err))
	}
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
		util.Debug.SetOutput(callbackWriter{llDebug})
	}

	// Read the options from the default configuration path and merge the
	// provided clientParams into as well.
	opts, _ := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	_ = json.Unmarshal([]byte(clientParams), opts)

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

	if ae, ok := err.(connect.APIError); ok {
		s.ErrType = "APIError"
		s.Code = ae.Code
		s.Message = ae.Message
	} else if uerr, ok := err.(*url.Error); ok {
		ierr := errors.Unwrap(err)
		if uerr.Timeout() {
			s.ErrType = "Timeout"
			s.Message = ierr.Error()
		} else if ce, ok := ierr.(x509.CertificateInvalidError); ok {
			s.ErrType = "SSLError"
			s.Message = ierr.Error()
			s.Data = certToPEM(ce.Cert)
			s.Code = sslErrorMap[int(ce.Reason)]
		} else if ce, ok := ierr.(x509.UnknownAuthorityError); ok {
			s.ErrType = "SSLError"
			s.Message = ierr.Error()
			s.Data = certToPEM(ce.Cert)
			// this could be:
			// 18 (X509_V_ERR_DEPTH_ZERO_SELF_SIGNED_CERT),
			// 19 (X509_V_ERR_SELF_SIGNED_CERT_IN_CHAIN) or
			// 20 (X509_V_ERR_UNABLE_TO_GET_ISSUER_CERT_LOCALLY)
			s.Code = 19 // this seems to best match original behavior
		} else if ce, ok := ierr.(*tls.CertificateVerificationError); ok {
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
			s.ErrType = "SSLError"
			s.Message = ierr.Error()
			// ruby version doesn't have this but it might be useful
			s.Data = certToPEM(ce.Certificate)
		} else if _, ok := ierr.(*net.OpError); ok {
			s.ErrType = "NetError"
			s.Message = ierr.Error()
		} else {
			util.Debug.Printf("url.Error: %T: %v", ierr, err)
			s.Message = err.Error()
		}
	} else if je, ok := err.(connect.JSONError); ok {
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
		s.Message = err.Error()
	}

	jsn, _ := json.Marshal(&s)
	return string(jsn)
}

//export getstatus
func getstatus(format *C.char) *C.char {
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
		return C.CString(err.Error())
	}
	return C.CString(output)
}

//export update_certificates
func update_certificates() *C.char {
	// NOTE: this is no longer relevant, but we keep it for
	// backwards-compatibility.
	return C.CString("{}")
}

//export reload_certificates
func reload_certificates() *C.char {
	// NOTE: this is no longer relevant, but we keep it for
	// backwards-compatibility.
	return C.CString("{}")
}

//export list_installer_updates
func list_installer_updates(clientParams, product *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	var productQuery registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &productQuery)
	if err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	repos, err := connect.InstallerUpdates(api.GetConnection(), productQuery)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(repos)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export system_migrations
func system_migrations(clientParams, products *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	installed := make([]registration.Product, 0)
	err := json.Unmarshal([]byte(C.GoString(products)), &installed)
	if err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	migrations, err := connect.ProductMigrations(api.GetConnection(), installed)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(migrations)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export offline_system_migrations
func offline_system_migrations(clientParams, products, targetBaseProduct *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	installed := make([]registration.Product, 0)
	err := json.Unmarshal([]byte(C.GoString(products)), &installed)
	if err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	var target registration.Product
	if err := json.Unmarshal([]byte(C.GoString(targetBaseProduct)), &target); err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	migrations, err := connect.OfflineProductMigrations(api.GetConnection(), installed, target)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(migrations)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export upgrade_product
func upgrade_product(clientParams, product *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))

	var prod registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &prod)
	if err != nil {
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
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(service)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export synchronize
func synchronize(clientParams, products *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	prods := make([]registration.Product, 0)
	err := json.Unmarshal([]byte(C.GoString(products)), &prods)
	if err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}
	activated, err := connect.SyncProducts(api.GetConnection(), prods)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(activated)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export system_activations
func system_activations(clientParams *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	activations, err := registration.FetchActivations(api.GetConnection())
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(activations)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export search_package
func search_package(clientParams, product, query *C.char) *C.char {
	opts := loadConfig(C.GoString(clientParams))
	api := connect.NewWrappedAPI(opts)

	var p registration.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &p)
	if err != nil {
		return C.CString(errorToJSON(connect.JSONError{Err: err}))
	}

	results, err := search.Package(api.GetConnection(), C.GoString(query), p.ToTriplet())
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(results)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

func main() {}
