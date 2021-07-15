package main

import (
	"C"
	"github.com/SUSE/connect-ng/internal/connect"
)

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
