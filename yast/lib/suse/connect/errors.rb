module SUSE
  module Connect

    class MalformedSccCredentialsFile < StandardError; end
    class MissingSccCredentialsFile < StandardError; end

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
