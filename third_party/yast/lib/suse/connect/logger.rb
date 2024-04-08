require 'logger'
require 'singleton'
require 'fiddle'

module SUSE
  module Connect
    # Singleton log instance used by SUSE::Connect::Logger module
    #
    # @example Set own logger
    #   GlobalLogger.instance.log = ::Logger.new($stderr)
    #
    # Used by YaST already, do not refactor without consulting them!
    # Passing the YaST logger for writing the log to /var/log/YaST2/y2log (#log=)
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
        @log = ::Logger.new($stderr)
        GoConnect.set_log_callback(LogLine)
      end

      LogLine = Class.new(Fiddle::Closure) {
        def call(level, message)
          log = GlobalLogger.instance.log
          message = message.to_s
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
      }.new(Fiddle::TYPE_VOID, [Fiddle::TYPE_INT, Fiddle::TYPE_VOIDP])
    end
    # Module provides access to specific logging. To set logging see GlobalLogger.
    #
    # @example Add logging to class
    #   class A
    #     include ::SUSE::Connect::Logger
    #
    #     def self.f
    #       log.info "self f"
    #     end
    #
    #     def a
    #       log.debug "a"
    #     end
    #   end
    module Logger
      def log
        GlobalLogger.instance.log
      end

      def self.included(base)
        base.extend self
      end
    end
  end
end
