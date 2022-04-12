require 'json'
require 'fiddle/import'
require 'suse/toolkit/shim_utils'

module GoConnect
  extend Fiddle::Importer
  dlload 'libsuseconnect.so'
  typealias 'string', 'char*'

  #callback type: void log_line(int, string)
  extern 'void set_log_callback(void*)'
  extern 'void free_string(string)'
  extern 'string announce_system(string, string)'
  extern 'string update_system(string, string)'
  extern 'string credentials(string)'
  extern 'string create_credentials_file(string, string, string)'
  extern 'string curlrc_credentials()'
  extern 'string show_product(string, string)'
  extern 'string activated_products(string)'
  extern 'string activate_product(string, string, string)'
  extern 'string get_config(string)'
  extern 'string write_config(string)'
  extern 'string update_certificates()'
  extern 'string reload_certificates()'
  extern 'string list_installer_updates(string, string)'
  extern 'string system_migrations(string, string)'
  extern 'string offline_system_migrations(string, string, string)'
  extern 'string upgrade_product(string, string)'
  extern 'string synchronize(string, string)'
  extern 'string system_activations(string)'
  extern 'string search_package(string, string, string)'
end

module SUSE
  module Connect
    class YaST
      DEFAULT_CONFIG_FILE = SUSE::Connect::Config::DEFAULT_CONFIG_FILE
      DEFAULT_URL = SUSE::Connect::Config::DEFAULT_URL
      DEFAULT_CREDENTIALS_DIR = "/etc/zypp/credentials.d"
      GLOBAL_CREDENTIALS_FILE = "/etc/zypp/credentials.d/SCCcredentials"
      SERVER_CERT_FILE = SUSE::Connect::SSLCertificate::SERVER_CERT_FILE
      UPDATE_CERTIFICATES = "/usr/sbin/update-ca-certificates"

      class << self
        include SUSE::Toolkit::ShimUtils

        # Announces the system to SCC / the registration server.
        # Expects a token / regcode to identify the correct subscription.
        # Additionally, distro_target should be set to avoid calls to Zypper.
        # Returns the system credentials from SCC.
        #
        # @param [Hash] client_params parameters to override SUSEConnect config
        # @param [String] distro_target desired distro target
        #
        # @return [Array <String>] SCC / system credentials - login and password tuple
        def announce_system(client_params = {}, distro_target = nil)
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          _process_result(GoConnect.announce_system(jsn_params, distro_target)).credentials
        end

        # Updates the systems hardware info on the server
        # @param [Hash] client_params parameters to override SUSEConnect config
        # @param [String] distro_target desired distro target
        def update_system(client_params = {}, distro_target = nil)
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          _process_result(GoConnect.update_system(jsn_params, distro_target))
        end

        # Activates a product on SCC / the registration server.
        # Expects product parameter to identify the product.
        # Requires a token / regcode except for free products/extensions.
        # Returns a service object for the activated product.
        #
        # @param [OpenStruct] product with identifier, arch and version defined
        # @param [Hash] client_params parameters to override SUSEConnect config
        # @param [String] email email to which this activation should be connected to
        #
        # @return [OpenStruct] Service object as openstruct
        def activate_product(product, client_params = {}, email = nil)
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          jsn_product = JSON.generate(product.to_h)
          _process_result(GoConnect.activate_product(jsn_params, jsn_product, email))
        end

        # Upgrades a product on SCC / the registration server.
        # Expects product parameter to identify the product.
        # Token / regcode is not required. The new product needs to be available to the regcode the old
        # product was registered with, or be a free product.
        # Returns a service object for the new activated product.
        #
        # @param [OpenStruct] product with identifier, arch and version defined
        # @param [Hash] client_params parameters to override SUSEConnect config
        #
        # @return [OpenStruct] Service object as openstruct
        def upgrade_product(product, client_params = {})
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          jsn_product = JSON.generate(product.to_h)
          _process_result(GoConnect.upgrade_product(jsn_params, jsn_product))
        end

        # Downgrades a product on SCC / the registration server.
        # Expects product parameter to identify the product.
        # Token / regcode is not required. The new product needs to be available to the regcode the old
        # product was registered with, or be a free product.
        # Returns a service object for the new activated product.
        #
        # @param [OpenStruct] product with identifier, arch and version defined
        # @param [Hash] client_params parameters to override SUSEConnect config
        #
        # @return [OpenStruct] Service object as openstruct
        alias_method :downgrade_product, :upgrade_product

        # Synchronize activated system products with registration server.
        # This will remove obsolete activations on the server after all installed products went through a downgrade().
        #
        # @param [OpenStruct] products - list of activated system products with identifier, arch and version defined
        # @param [Hash] client_params parameters to override SUSEConnect config
        def synchronize(products, client_params = {})
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          jsn_products = JSON.generate(products.map(&:to_h))
          _process_result(GoConnect.synchronize(jsn_params, jsn_products))
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
        # Accepts a parameter product, which scopes the result set down to all
        # products for the system that are extensions to the specified product.
        # Gets the list from SCC and returns them.
        #
        # @param [OpenStruct] product to list extensions for
        # @param [Hash] client_params parameters to override SUSEConnect config
        #
        # @return [OpenStruct] {Product} from registration server with all extensions included
        def show_product(product, client_params = {})
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          jsn_product = JSON.generate(product.to_h)
          _process_result(GoConnect.show_product(jsn_params, jsn_product))
        end

        # Lists all available online migration paths for a given list of products.
        # Accepts an array of products, and returns an array of possible
        # migration paths. A migration path is a list of products that may
        # be upgraded.
        #
        # @param [Array <OpenStruct>] the list of currently installed {Product}s in the system
        # @param [Hash] client_params parameters to override SUSEConnect config
        #
        # @return [Array <Array <OpenStruct>>] the list of possible migration paths for the given {Product}s,
        #   where a migration path is an array of OpenStruct objects with the attributes
        #   identifier, arch, version, and release_type
        def system_migrations(products, client_params = {})
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          jsn_products = JSON.generate(products.map(&:to_h))
          _process_result(GoConnect.system_migrations(jsn_params, jsn_products))
        end

        # Lists all available offline migration paths for a given list of products.
        # Accepts an array of products, and returns an array of possible
        # migration paths. A migration path is a list of products that may
        # be upgraded.
        #
        # @param installed_products [Array <OpenStruct>] the list of currently installed {Product}s in the system
        # @param target_base_product [OpenStruct] the {Product} that the system wants to upgrade to
        # @param [Hash] client_params parameters to override SUSEConnect config
        #
        # @return [Array <Array <OpenStruct>>] the list of possible migration paths for the given {Product}s,
        #   where a migration path is an array of OpenStruct objects with the attributes
        #   identifier, arch, version, and release_type
        def system_offline_migrations(installed_products, target_base_product, client_params = {})
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          jsn_products = JSON.generate(installed_products.map(&:to_h))
          jsn_target = JSON.generate(target_base_product.to_h)
          _process_result(GoConnect.offline_system_migrations(jsn_params, jsn_products, jsn_target))
        end

        # List available Installer-Updates repositories for the given product
        #
        # @param [OpenStruct] list repositories for this product
        # @param [Hash] client_params parameters to override SUSEConnect config
        #
        # @return [Array <OpenStruct>] list of Installer-Updates repositories
        def list_installer_updates(product, client_params = {})
          _set_verify_callback(client_params[:verify_callback])
          jsn_params = JSON.generate(client_params)
          jsn_product = JSON.generate(product.to_h)
          _process_result(GoConnect.list_installer_updates(jsn_params, jsn_product))
        end

        # Writes the config file with the given parameters, overwriting any existing contents
        # Attributes not defined in client_params will not be modified
        # @param [Hash] client_params parameters to override SUSEConnect config
        def write_config(client_params = {})
          jsn_params = JSON.generate(client_params)
          _process_result(GoConnect.write_config(jsn_params))
        end

        # Adds given certificate to trusted
        # @param certificate [OpenSSL::X509::Certificate]
        def import_certificate(certificate)
          SUSE::Connect::SSLCertificate.import(certificate)
        end

        # Provides SHA-1 fingerprint of given certificate
        # @param certificate [OpenSSL::X509::Certificate]
        def cert_sha1_fingerprint(certificate)
          SUSE::Connect::SSLCertificate.sha1_fingerprint(certificate)
        end

        # Provides SHA-256 fingerprint of given certificate
        # @param certificate [OpenSSL::X509::Certificate]
        def cert_sha256_fingerprint(certificate)
          SUSE::Connect::SSLCertificate.sha256_fingerprint(certificate)
        end

        # Provides access to current system status in terms of activated products
        # @param [Hash] client_params parameters to override SUSEConnect config
        def status(client_params)
          _set_verify_callback(client_params[:verify_callback])
          Status.new(client_params)
        end
      end
    end
  end
end
