package registration

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// The information extracted from an offline registration certificate
// To further inspect payload and signature check [IsValid] and [ExtractPayload].
type OfflineCertificate struct {
	Version          string `json:"version"`
	Cipher           string `json:"cipher"`
	Hash             string `json:"hash"`
	EncodedPayload   string `json:"payload"`
	EncodedSignature string `json:"signature"`
}

// The information supplied and validated by the payload. The [RegcodeHash] can be used
// to determine if the certificate hold has the knowledge over the registration code.
//
// Information holds arbitrary information about the registered system and might or might
// not have the relevant information. Make sure to check if the key exists and prepare for
// type casting if necessary.
type OfflinePayload struct {
	Login        string         `json:"login"`
	Password     string         `json:"password"`
	Subscription Subscription   `json:"subscription"`
	RegcodeHash  string         `json:"regcode_sha256"`
	Information  map[string]any `json:"information"`
}

// Reads an offline registration certificate from a read object
//
// An error indicates the certificate was probably malformed
func OfflineCertificateFrom(reader io.Reader) (*OfflineCertificate, error) {
	certificate := OfflineCertificate{}

	raw, readErr := io.ReadAll(reader)

	if readErr != nil {
		return nil, fmt.Errorf("read certificate: %s", readErr)
	}

	decoded, decodeErr := decodeBase64(raw)

	if decodeErr != nil {
		return nil, decodeErr
	}

	if err := json.Unmarshal(decoded, &certificate); err != nil {
		return nil, fmt.Errorf("json error: %s", err)
	}

	return &certificate, nil
}

// Checks if the provided offline registration certificate is valid using
// the public RSA key to validate the included signature
func (cert *OfflineCertificate) IsValid() (bool, error) {
	key, keyErr := sCCPublicKey()

	if keyErr != nil {
		return false, keyErr
	}

	digest := sha256.Sum256([]byte(cert.EncodedPayload))
	signature, sigErr := cert.Signature()

	if sigErr != nil {
		return false, sigErr
	}

	options := &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA256,
	}

	verifyErr := rsa.VerifyPSS(key, crypto.SHA256, digest[:], signature, options)

	return verifyErr == nil, nil
}

// Extracts the payload from an offline registration certificate
func (cert *OfflineCertificate) ExtractPayload() (*OfflinePayload, error) {
	payload := OfflinePayload{}
	raw, decodeErr := decodeBase64([]byte(cert.EncodedPayload))

	if decodeErr != nil {
		return nil, decodeErr
	}

	if jsonErr := json.Unmarshal(raw, &payload); jsonErr != nil {
		return nil, fmt.Errorf("json: %s", jsonErr)
	}

	return &payload, nil
}

// Provides the signature encoded in the offline registration certificate which
// can be used to validate the base64 encoded payload in the [OfflineCertificate] structure.
func (cert *OfflineCertificate) Signature() ([]byte, error) {
	return decodeBase64([]byte(strings.TrimSpace(cert.EncodedSignature)))
}
