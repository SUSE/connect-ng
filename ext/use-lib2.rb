#!/usr/bin/ruby

require 'ffi'

module SUSEConnect
  extend FFI::Library
  ffi_lib './libsuseconnect.so'

  attach_function :getstatus, [:string], :pointer
  attach_function :free, [ :pointer ], :void
end

p_out = SUSEConnect.getstatus('json')
puts p_out.get_string(0)
SUSEConnect.free(p_out)
