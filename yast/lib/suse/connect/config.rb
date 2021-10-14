require 'ostruct'
require 'suse/toolkit/shim_utils'

module SUSE
  module Connect
    class Config < OpenStruct
      include SUSE::Toolkit::ShimUtils

      DEFAULT_CONFIG_FILE = '/etc/SUSEConnect'
      DEFAULT_URL = 'https://scc.suse.com'

      def initialize(file = DEFAULT_CONFIG_FILE)
        super()
        jsn_out = _consume_str(GoConnect.get_config(file))
        cfg = JSON.parse(jsn_out)
        cfg.each { |k, v| self[k] = v }
      end
    end
  end
end
