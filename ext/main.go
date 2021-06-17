package main

import (
	"C"
	"github.com/SUSE/connect-ng/connect"
)

//export getstatus
func getstatus(format *C.char) *C.char {
	connect.CFG.Load()
	gFormat := C.GoString(format)
	return C.CString(connect.GetProductStatuses(gFormat))
}

func main() {}
