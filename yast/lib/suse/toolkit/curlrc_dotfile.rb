class SUSE::Toolkit::CurlrcDotfile
#  def initialize
#    @file_location = File.join(Etc.getpwuid.dir, CURLRC_LOCATION)
#  end

  def username
    # TODO:
#    line_with_credentials.match(CURLRC_CREDENTIALS_REGEXP)[1] rescue nil
  end

  def password
    # TODO
#    line_with_credentials.match(CURLRC_CREDENTIALS_REGEXP)[2] rescue nil
  end
end
