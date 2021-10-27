module SUSE
  module Toolkit
    module ShimUtils
      def _process_result(ptr)
        jsn_out = _consume_str(ptr)
        result = JSON.parse(jsn_out, object_class: OpenStruct)
        _check_error(result)
        result
      end

      def _consume_str(ptr)
        s = ptr.get_string(0)
        Stdio.free(ptr)
        return s
      end

      def _check_error(r)
        # errors come as OpenStruct or hash
        r = r.to_h if r.is_a?(OpenStruct)
        return unless r.is_a?(Hash) && r.key?(:err_type)
        case r[:err_type]
        when "APIError"
          error = SUSE::Connect::ApiError.new(OpenStruct.new(r))
          raise error, error.message
        when "MalformedSccCredentialsFile"
          raise SUSE::Connect::MalformedSccCredentialsFile, r[:message]
        when "MissingCredentialsFile"
          raise SUSE::Connect::MissingSccCredentialsFile, r[:message]
        else
          raise r[:message] if r.key?(:message)
          raise r.to_s
        end
      end
    end
  end
end
