require 'json'
require 'ffi'

module Stdio
  extend FFI::Library
  ffi_lib FFI::Platform::LIBC
  attach_function :free, [ :pointer ], :void
end

module GoConnect
  extend FFI::Library
  ffi_lib '/root/out/libsuseconnect.so'

  attach_function :announce_system, [:string, :string], :pointer
  attach_function :credentials, [:string], :pointer
  attach_function :create_credentials_file, [:string, :string, :string], :pointer
end

module SUSE
  module Connect
    class YaST
        GLOBAL_CREDENTIALS_FILE = "/etc/zypp/credentials.d/SCCcredentials"
        SERVER_CERT_FILE = SUSE::Connect::SSLCertificate::SERVER_CERT_FILE

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

          # Reads credentials file.
          # Returns the credentials object with login, password and credentials file
          #
          # @param [String] Path to credentials file - defaults to /etc/zypp/credentials.d/SCCcredentials
          #
          # @return [OpenStruct] Credentials object as openstruct
          def credentials(credentials_file = GLOBAL_CREDENTIALS_FILE)
            jsn_out = _consume_str(GoConnect.credentials(credentials_file))
            result = JSON.parse(jsn_out, object_class: OpenStruct)
            if result.err_type == "MalformedSccCredentialsFile"
              raise MalformedSccCredentialsFile, result.message
            elsif result.err_type == "MissingCredentialsFile"
              raise MissingSccCredentialsFile, result.message
            end
            result
          end

          # Creates the system or zypper service credentials file with given login and password.
          # @param [String] system login - return value of announce_system method
          # @param [String] system password - return value of announce_system method
          # @param [String] credentials_file - defaults to /etc/zypp/credentials.d/SCCcredentials
          def create_credentials_file(login, password, credentials_file = GLOBAL_CREDENTIALS_FILE)
            GoConnect.create_credentials_file(login, password, credentials_file)
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
