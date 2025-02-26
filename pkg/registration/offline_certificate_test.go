package registration_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/stretchr/testify/assert"
)

func TestOfflineCertificateFromSuccess(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/valid.cert")
	reader := bytes.NewReader(rawCert)

	cert, err := registration.OfflineCertificateFrom(reader)

	assert.NoError(err)

	assert.Equal("1", cert.Version)
	assert.Equal("RSA", cert.Cipher)
	assert.Equal("SHA256", cert.Hash)
}

func TestOfflineCertificateFromMalformedCertificate(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/malformed.cert")
	reader := bytes.NewReader(rawCert)

	cert, err := registration.OfflineCertificateFrom(reader)

	assert.ErrorContains(err, "base64: illegal base64 data")
	assert.Nil(cert)
}

func TestOfflineCertificateIsValidTrue(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/valid.cert")
	reader := bytes.NewReader(rawCert)

	cert, err := registration.OfflineCertificateFrom(reader)
	assert.NoError(err)

	valid, err := cert.IsValid()
	assert.NoError(err)
	assert.True(valid)
}

func TestOfflineCertificateIsValidFalse(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/invalid.cert")
	reader := bytes.NewReader(rawCert)

	cert, err := registration.OfflineCertificateFrom(reader)
	assert.NoError(err)

	valid, err := cert.IsValid()
	assert.NoError(err)
	assert.False(valid)
}

func TestOfflineCertificateIsValidMalformedSignature(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/malformed-signature.cert")
	reader := bytes.NewReader(rawCert)

	cert, err := registration.OfflineCertificateFrom(reader)
	assert.NoError(err)

	valid, err := cert.IsValid()
	assert.False(valid)
	assert.ErrorContains(err, "base64: illegal base64 data")
}

func TestOfflineCertificateExtractPayloadSuccess(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/valid.cert")
	reader := bytes.NewReader(rawCert)
	expectedExpirationDate := time.Date(2026, time.January, 27, 11, 53, 51, 223000000, time.UTC)

	cert, err := registration.OfflineCertificateFrom(reader)
	assert.NoError(err)

	payload, extractErr := cert.ExtractPayload()
	assert.NoError(extractErr)

	assert.Equal("SCC_384deae18e324233b20de20c87b89df7", payload.Login)
	assert.Equal("https://rnch1.dev.company.net/index", payload.Information["server_url"].(string))
	assert.Equal("RANCHER-X86-ALPHA", payload.Subscription.ProductClasses[0].Name)
	assert.Equal(expectedExpirationDate, payload.Subscription.ExpiresAt)
}

func TestOfflineCertificateExtractPayloadInvalidJSON(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/invalid-json-payload.cert")
	reader := bytes.NewReader(rawCert)

	cert, err := registration.OfflineCertificateFrom(reader)
	assert.NoError(err)

	payload, extractErr := cert.ExtractPayload()

	assert.ErrorContains(extractErr, "json: invalid character")
	assert.Nil(payload)
}

func TestOfflineCertificateRegcodeMatches(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/valid.cert")
	reader := bytes.NewReader(rawCert)

	cert, readErr := registration.OfflineCertificateFrom(reader)
	assert.NoError(readErr)

	matches, matchErr := cert.RegcodeMatches("some-scc-regcode")
	assert.NoError(matchErr)
	assert.True(matches)

	notMatches, matchErr := cert.RegcodeMatches("INVALID")
	assert.NoError(matchErr)
	assert.False(notMatches)
}

func TestOfflineCertificateUUIDMatches(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/valid.cert")
	reader := bytes.NewReader(rawCert)

	cert, readErr := registration.OfflineCertificateFrom(reader)
	assert.NoError(readErr)

	matches, matchErr := cert.UUIDMatches("3a4d46b4-0b06-488f-8d20-a931d398d357")
	assert.NoError(matchErr)
	assert.True(matches)

	notMatches, matchErr := cert.UUIDMatches("INVALID")
	assert.NoError(matchErr)
	assert.False(notMatches)
}

func TestOfflineCertificateProductClassIncluded(t *testing.T) {
	assert := assert.New(t)

	rawCert := fixture(t, "pkg/registration/offline_certificate/valid.cert")
	reader := bytes.NewReader(rawCert)

	cert, readErr := registration.OfflineCertificateFrom(reader)
	assert.NoError(readErr)

	shouldBeIncluded, matchErr := cert.ProductClassIncluded("RANCHER-X86")
	assert.NoError(matchErr)

	shouldNotBeIncluded, matchErr := cert.ProductClassIncluded("SLES-X86")
	assert.NoError(matchErr)

	assert.True(shouldBeIncluded)
	assert.False(shouldNotBeIncluded)
}
