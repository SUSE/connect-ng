require 'suse/toolkit/shim_utils'

class SUSE::Toolkit::CurlrcDotfile
  include SUSE::Toolkit::ShimUtils

  def initialize
    @creds = _process_result(GoConnect.curlrc_credentials())
  end

  def username
    @creds.username
  end

  def password
    @creds.password
  end
end
