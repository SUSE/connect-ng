module SUSE
  module Connect
    require 'suse/connect/errors'
    require 'suse/connect/logger'
    require 'suse/connect/config'
    require 'suse/connect/ssl_certificate'
    require 'suse/connect/status'
    require 'suse/connect/yast'
  end
  module Toolkit
    require 'suse/toolkit/curlrc_dotfile'
  end
end
