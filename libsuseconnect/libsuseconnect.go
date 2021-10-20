package main

// #include <stdlib.h>
// typedef void (*logLineFunc)(int level, const char* message);
// void log_bridge_fun(logLineFunc f, int level, const char* message);
import "C"

import (
	"encoding/json"
	"os"
	"unsafe"

	"github.com/SUSE/connect-ng/internal/connect"
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
	connect.Debug.SetOutput(callbackWriter{llDebug})
	connect.Info.SetOutput(callbackWriter{llInfo})
	// TODO: add other levels?
}

//export announce_system
func announce_system(clientParams, distroTarget *C.char) *C.char {
	loadConfig(C.GoString(clientParams))

	login, password, err := connect.AnnounceSystem(C.GoString(distroTarget), "")
	if err != nil {
		return C.CString(errorToJSON(err))
	}

	var res struct {
		Credentials []string `json:"credentials"`
	}
	res.Credentials = []string{login, password}
	jsn, _ := json.Marshal(&res)
	return C.CString(string(jsn))
}

//export credentials
func credentials(path *C.char) *C.char {
	creds, err := connect.ReadCredentials(C.GoString(path))
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, _ := json.Marshal(&creds)
	return C.CString(string(jsn))
}

//export create_credentials_file
func create_credentials_file(login, password, path *C.char) *C.char {
	err := connect.CreateCredentials(
		C.GoString(login), C.GoString(password), C.GoString(path))
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString("") // TODO need more consistent return path
}

//export show_product
func show_product(clientParams, product *C.char) *C.char {
	loadConfig(C.GoString(clientParams))

	var productQuery connect.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &productQuery)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	productData, err := connect.ShowProduct(productQuery)
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
	loadConfig(C.GoString(clientParams))

	var p connect.Product
	err := json.Unmarshal([]byte(C.GoString(product)), &p)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	service, err := connect.ActivateProduct(p, C.GoString(email))
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
	loadConfig(C.GoString(clientParams))

	products, err := connect.ActivatedProducts()
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	jsn, err := json.Marshal(products)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export get_config
func get_config(path *C.char) *C.char {
	var c connect.Config
	c.Path = C.GoString(path)
	c.Load()
	jsn, err := json.Marshal(c)
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString(string(jsn))
}

//export write_config
func write_config(clientParams *C.char) *C.char {
	loadConfig(C.GoString(clientParams))
	err := connect.CFG.Save()
	if err != nil {
		return C.CString(errorToJSON(err))
	}
	return C.CString("{}")
}

func loadConfig(clientParams string) {
	connect.CFG.Load()
	connect.CFG.MergeJSON(clientParams)
	if _, ok := os.LookupEnv("SCCDEBUG"); ok {
		connect.EnableDebug()
	}
}

func errorToJSON(err error) string {
	var s struct {
		ErrType string `json:"err_type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	if ae, ok := err.(connect.APIError); ok {
		s.ErrType = "APIError"
		s.Code = ae.Code
		s.Message = ae.Message
	} else {
		switch err {
		case connect.ErrMalformedSccCredFile:
			s.ErrType = "MalformedSccCredentialsFile"
		case connect.ErrMissingCredentialsFile:
			s.ErrType = "MissingCredentialsFile"
		}
		s.Message = err.Error()
	}

	jsn, _ := json.Marshal(&s)
	return string(jsn)
}

//export getstatus
func getstatus(format *C.char) *C.char {
	connect.CFG.Load()
	gFormat := C.GoString(format)
	output, err := connect.GetProductStatuses(gFormat)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString(output)
}

func main() {}
