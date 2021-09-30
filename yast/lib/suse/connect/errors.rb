module SUSE
  module Connect

    # simplified version of the original for demonstration
    class ApiError < StandardError

      def initialize(response)
        @response = response
      end

      def code
        @response["code"]
      end

      def message
        @response["message"]
      end

    end
  end
end
