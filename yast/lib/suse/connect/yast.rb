require 'json'
require 'ffi'

module Stdio
  extend FFI::Library
  ffi_lib FFI::Platform::LIBC
  attach_function :free, [ :pointer ], :void
end

module GoConnect
  extend FFI::Library
  ffi_lib '../out/libsuseconnect.so'

  attach_function :announce_system, [:string, :string], :pointer
end

module SUSE
  module Connect
    class YaST

        class << self

          # Announces the system to SCC / the registration server.
          # Expects a token / regcode to identify the correct subscription.
          # Additionally, distro_target should be set to avoid calls to Zypper.
          # Returns the system credentials from SCC.
          #
          # @param [Hash] client_params parameters to instantiate {Client}
          # @param [String] distro_target desired distro target
          #
          # @return [Array <String>] SCC / system credentials - login and password tuple
          def announce_system(client_params = {}, distro_target = nil)
            jsn_params = JSON.generate(client_params)
            jsn_out = _consume_str(GoConnect.announce_system(jsn_params, distro_target))
            result = JSON.parse(jsn_out)
            if result.key?("err_type")
              if result["err_type"] == "APIError"
                error = SUSE::Connect::ApiError.new(result)
                raise error, error.message
              end
              # check other errors
            end
            result["credentials"]
          end

          private

          def _consume_str(ptr)
            s = ptr.get_string(0)
            Stdio.free(ptr)
            return s
          end

        end
    end
  end
end
