package main

import (
	"C"
	"encoding/json"
	"fmt"
	"github.com/SUSE/connect-ng/internal/connect"
	"os"
)

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
		jsn, _ := json.Marshal(&s)
		return string(jsn)
	}
	// TODO other error types
	return fmt.Sprintf(`{"err_type": nil, message": "%s", "code": 0}`, err.Error())
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
