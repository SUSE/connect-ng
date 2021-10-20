require 'singleton'
require 'ffi'

module SUSE
  module Connect
    class GlobalLogger
      include Singleton

      attr_accessor :log

      # log levels
      LL_DEBUG = 1
      LL_INFO = 2
      LL_WARNING = 3
      LL_ERROR = 4
      LL_FATAL = 5

      def initialize()
        GoConnect.set_log_callback(LogLine)
      end

      LogLine = ::FFI::Function.new(:void, [:int, :string]) do |level, message|
        log = GlobalLogger.instance.log
        case level
        when LL_DEBUG
          log.debug(message)
        when LL_INFO
          log.info(message)
        when LL_WARNING
          log.warn(message)
        when LL_ERROR
          log.error(message)
        when LL_FATAL
          log.fatal(message)
        else
          raise 'unknown log level, msg=#{message}'
        end
      end
    end
  end
end
