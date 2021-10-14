require 'singleton'

module SUSE
  module Connect
    class GlobalLogger
      include Singleton

      attr_accessor :log
    end
  end
end
