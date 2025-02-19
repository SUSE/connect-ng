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
	assert.Equal("3a4d46b4-0b06-488f-8d20-a931d398d357", payload.Information["uuid"].(string))
	assert.Equal("cpe:/o:suse:rancher:2.4.8", payload.Subscription.Products[0].CPE)
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
