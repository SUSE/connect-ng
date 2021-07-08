#!/usr/bin/python3

# Based on https://www.ardanlabs.com/blog/2020/07/extending-python-with-go.html
# It has a note about using free().

import ctypes

so = ctypes.cdll.LoadLibrary('./out/libsuseconnect.so')
getstatus = so.getstatus
getstatus.argtypes = [ctypes.c_char_p]
getstatus.restype = ctypes.c_void_p
free = so.free
free.argtypes = [ctypes.c_void_p]
ptr = getstatus('json'.encode('utf-8'))
out = ctypes.string_at(ptr)
print(out.decode('utf-8'))
free(ptr)
