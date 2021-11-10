require 'suse/toolkit/shim_utils'

module SUSE
  module Connect
    # Enable connect and zypper extensions/scripts to search packages for a
    # certain product
    class PackageSearch
      class << self
        include SUSE::Toolkit::ShimUtils

        # Search packages depending on the product and its extension/module
        # tree.
        #
        # @param query [String] package to search
        # @param product [SUSE::Connect::Zypper::Product] product to base search on
        # @param config_params [<Hash>] overwrites from the config file
        #
        # @return [Array< <Hash>>] Returns all matched packages or an empty array if no matches where found
        def search(query, product: nil, config_params: {})
          # NOTE: product and config_params above are named parameters unlike the rest of
          #       Connect interface
          _set_verify_callback(config_params[:verify_callback])
          jsn_params = JSON.generate(config_params)
          jsn_product = JSON.generate(product.to_h)
          _process_result(GoConnect.search_package(jsn_params, jsn_product, query))
        end
      end
    end
  end
end
