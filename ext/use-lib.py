#!/usr/bin/python3

import ctypes

so = ctypes.cdll.LoadLibrary('./libsuseconnect.so')
getstatus = so.getstatus
getstatus.argtypes = [ctypes.c_char_p]
getstatus.restype = ctypes.c_void_p
free = so.free
free.argtypes = [ctypes.c_void_p]
ptr = getstatus('json'.encode('utf-8'))
out = ctypes.string_at(ptr)
print(out.decode('utf-8'))
free(ptr)
