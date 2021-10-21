require 'json'
require 'ffi'
require 'suse/toolkit/shim_utils'

# TODO
# - check if SUSE::Connect::Zypper::Product.determine_release_type() is needed
# - check required Repo fields
# - make sure following code paths are covered by shim:
#     lib/registration/package_search.rb:      SUSE::Connect::PackageSearch.search(text, product: connect_product(product))
#     lib/registration/ssl_certificate.rb:    # @raise Connect::SystemCallError
#     lib/registration/ssl_certificate.rb:      ::SUSE::Connect::YaST.import_certificate(x509_cert)
#     lib/registration/ssl_certificate.rb:    rescue ::SUSE::Connect::SystemCallError => e
#     lib/registration/ssl_certificate.rb:        ::SUSE::Connect::YaST.cert_sha1_fingerprint(x509_cert)
#     lib/registration/ssl_certificate.rb:        ::SUSE::Connect::YaST.cert_sha256_fingerprint(x509_cert)
#     lib/registration/sw_mgmt.rb:        SUSE::Connect::YaST.create_credentials_file(credentials.username,
#     lib/registration/registration.rb:        service = SUSE::Connect::YaST.upgrade_product(product_ident, params)
#     lib/registration/registration.rb:        service = SUSE::Connect::YaST.downgrade_product(product_ident, params)
#     lib/registration/registration.rb:      SUSE::Connect::YaST.synchronize(remote_products, connect_params)
#     lib/registration/registration.rb:      ret = SUSE::Connect::YaST.update_system(connect_params, target_distro)
#     lib/registration/registration.rb:        migrations = SUSE::Connect::YaST.system_migrations(installed_products, connect_params)
#     lib/registration/registration.rb:        migration_paths = SUSE::Connect::YaST.system_offline_migrations(installed_products, target_base_product, connect_params)
#     lib/registration/registration.rb:      updates = SUSE::Connect::YaST.list_installer_updates(remote_product, connect_params)
#     lib/registration/helpers.rb:      cmd = SUSE::Connect::YaST::UPDATE_CERTIFICATES

module Stdio
  extend FFI::Library
  ffi_lib FFI::Platform::LIBC
  attach_function :free, [ :pointer ], :void
end

module GoConnect
  extend FFI::Library
  ffi_lib 'suseconnect'

  callback :log_line, [:int, :string], :void
  attach_function :set_log_callback, [:log_line], :void

  attach_function :announce_system, [:string, :string], :pointer
  attach_function :credentials, [:string], :pointer
  attach_function :create_credentials_file, [:string, :string, :string], :pointer
  attach_function :curlrc_credentials, [], :pointer
  attach_function :show_product, [:string, :string], :pointer
  attach_function :activated_products, [:string], :pointer
  attach_function :activate_product, [:string, :string, :string], :pointer
  attach_function :get_config, [:string], :pointer
  attach_function :write_config, [:string], :pointer
end

module SUSE
  module Connect
    class YaST
      DEFAULT_CONFIG_FILE = SUSE::Connect::Config::DEFAULT_CONFIG_FILE
      DEFAULT_URL = SUSE::Connect::Config::DEFAULT_URL
      DEFAULT_CREDENTIALS_DIR = "/etc/zypp/credentials.d"
      GLOBAL_CREDENTIALS_FILE = "/etc/zypp/credentials.d/SCCcredentials"
      SERVER_CERT_FILE = SUSE::Connect::SSLCertificate::SERVER_CERT_FILE

      class << self
        include SUSE::Toolkit::ShimUtils

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
          _process_result(GoConnect.announce_system(jsn_params, distro_target)).credentials
        end

        # Activates a product on SCC / the registration server.
        # Expects product_ident parameter to be a hash identifying the product.
        # Requires a token / regcode except for free products/extensions.
        # Returns a service object for the activated product.
        #
        # @param [OpenStruct] product with identifier, arch and version defined
        # @param [Hash] client_params parameters to instantiate {Client}
        # @param [String] email email to which this activation should be connected to
        #
        # @return [Service] Service
        def activate_product(product, client_params = {}, email = nil)
          jsn_params = JSON.generate(client_params)
          jsn_product = JSON.generate(product.to_h)
          _process_result(GoConnect.activate_product(jsn_params, jsn_product, email))
        end

        # Reads credentials file.
        # Returns the credentials object with login, password and credentials file
        #
        # @param [String] Path to credentials file - defaults to /etc/zypp/credentials.d/SCCcredentials
        #
        # @return [OpenStruct] Credentials object as openstruct
        def credentials(credentials_file = GLOBAL_CREDENTIALS_FILE)
          _process_result(GoConnect.credentials(credentials_file))
        end

        # Creates the system or zypper service credentials file with given login and password.
        # @param [String] system login - return value of announce_system method
        # @param [String] system password - return value of announce_system method
        # @param [String] credentials_file - defaults to /etc/zypp/credentials.d/SCCcredentials
        def create_credentials_file(login, password, credentials_file = GLOBAL_CREDENTIALS_FILE)
          GoConnect.create_credentials_file(login, password, credentials_file)
        end

        # Lists all available products for a system.
        # Accepts a parameter product_ident, which scopes the result set down to all
        # products for the system that are extensions to the specified product.
        # Gets the list from SCC and returns them.
        #
        # @param [OpenStruct] product to list extensions for
        # @param [Hash] client_params parameters to instantiate {Client}
        #
        # @return [OpenStruct] {Product} from registration server with all extensions included
        def show_product(product, client_params = {})
          jsn_params = JSON.generate(client_params)
          jsn_product = JSON.generate(product.to_h)
          _process_result(GoConnect.show_product(jsn_params, jsn_product))
        end

        # Writes the config file with the given parameters, overwriting any existing contents
        # Only persistent connection parameters (url, insecure) are written by this method
        # Regcode, language, debug etc are not
        # @param [Hash] client_params
        #  - :insecure [Boolean]
        #  - :url [String]
        def write_config(client_params = {})
          jsn_params = JSON.generate(client_params)
          _process_result(GoConnect.write_config(jsn_params))
        end

        # Provides access to current system status in terms of activated products
        # @param [Hash] client_params parameters to instantiate {Client}
        def status(client_params)
          Status.new(client_params)
        end
      end
    end
  end
end
