require 'ffi'

module SUSEConnect
  extend FFI::Library
  ffi_lib './libsuseconnect.so'

  attach_function :getstatus, [:string], :string
end

puts SUSEConnect.getstatus('json')

