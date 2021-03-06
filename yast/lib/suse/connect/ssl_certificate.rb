module SUSE
  module Connect
    class SSLCertificate
      extend SUSE::Toolkit::ShimUtils
      extend SUSE::Connect::Logger

      # where to save the imported certificate
      SERVER_CERT_FILE = '/usr/share/pki/trust/anchors/registration_server.pem'

      # compute SHA1 fingerprint of a certificate
      # @param cert [OpenSSL::X509::Certificate] the certificate
      # @return [String] fingerprint in "AB:CD:EF:..." format
      def self.sha1_fingerprint(cert)
        format_digest(OpenSSL::Digest::SHA1, cert)
      end

      # compute SHA256 fingerprint of a certificate
      # @param cert [OpenSSL::X509::Certificate] the certificate
      # @return [String] fingerprint in "AB:CD:EF:..." format
      def self.sha256_fingerprint(cert)
        format_digest(OpenSSL::Digest::SHA256, cert)
      end

      # import the SSL certificate into the system
      # @see https://github.com/openSUSE/ca-certificates
      # @param cert [OpenSSL::X509::Certificate] the certificate
      def self.import(cert)
        log.debug "Writing a SSL certificate to #{SERVER_CERT_FILE} file..."
        log.warn 'The certificate file already exists, rewriting...' if File.exist?(SERVER_CERT_FILE)
        File.write(SERVER_CERT_FILE, cert.to_pem)

        # update the symlinks
        _process_result(GoConnect.update_certificates())
      end

      # reload internal CA cert pool
      def self.reload
        log.debug "Reloading Golang CA cert pool..."
        _process_result(GoConnect.reload_certificates())
      end

      # @param digest_class [Class] target digest class (e.g. OpenSSL::Digest::SHA1)
      # @param cert [OpenSSL::X509::Certificate] the certificate
      # @return [String] digest in "AB:CD:EF:..." format
      def self.format_digest(digest_class, cert)
        digest_class.new(cert.to_der).to_s.upcase.scan(/../).join(':')
      end
    end
  end
end
