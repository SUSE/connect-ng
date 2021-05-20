#!/usr/bin/ruby

require 'ffi'

# see https://github.com/ffi/ffi/issues/467#issuecomment-159370223

module Stdio
  extend FFI::Library
  ffi_lib FFI::Platform::LIBC
  attach_function :free, [ :pointer ], :void
end

module SUSEConnect
  extend FFI::Library
  ffi_lib './libsuseconnect.so'

  attach_function :getstatus, [:string], :pointer
end

p_out = SUSEConnect.getstatus('json')
puts p_out.get_string(0)
Stdio.free(p_out)
