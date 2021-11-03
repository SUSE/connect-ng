require 'suse/toolkit/shim_utils'

module SUSE
  module Connect
    # The System Status object provides information about the state of currently installed products
    # and subscriptions as known by registration server.
    # At first it collects all installed products from the system, then it gets its `activations`
    # from the registration server. This information is merged and printed out.
    # rubocop:disable ClassLength
    class Status
      include SUSE::Toolkit::ShimUtils

      def initialize(client_params)
        @client_params = client_params
      end

      def activated_products
        jsn_params = JSON.generate(@client_params)
        _process_result(GoConnect.activated_products(jsn_params))
      end

      def activations
        jsn_params = JSON.generate(@client_params)
        _process_result(GoConnect.system_activations(jsn_params))
      end
    end
  end
end
