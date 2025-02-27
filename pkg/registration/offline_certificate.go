package registration

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// The information extracted from an offline registration certificate
// To further inspect payload and signature check [IsValid] and [ExtractPayload].
type OfflineCertificate struct {
	Version          string `json:"version"`
	Cipher           string `json:"cipher"`
	Hash             string `json:"hash"`
	EncodedPayload   string `json:"payload"`
	EncodedSignature string `json:"signature"`

	*OfflinePayload
}

// The information supplied and validated by the payload. The [RegcodeHash] can be used
// to determine if the certificate hold has the knowledge over the registration code.
//
// Information holds arbitrary information about the registered system and might or might
// not have the relevant information. Make sure to check if the key exists and prepare for
// type casting if necessary.
type OfflinePayload struct {
	Login            string           `json:"login"`
	Password         string           `json:"password"`
	SubscriptionInfo SubscriptionInfo `json:"subscription"`
	HashedRegcode    string           `json:"hashed_regcode"`
	HashedUUID       string           `json:"hashed_uuid"`
	Information      map[string]any   `json:"information"`
}

// Reads an offline registration certificate from a read object
//
// An error indicates the certificate was probably malformed
func OfflineCertificateFrom(reader io.Reader) (*OfflineCertificate, error) {
	certificate := &OfflineCertificate{}

	raw, readErr := io.ReadAll(reader)

	if readErr != nil {
		return nil, fmt.Errorf("read certificate: %s", readErr)
	}

	decoded, decodeErr := decodeBase64(raw)

	if decodeErr != nil {
		return nil, decodeErr
	}

	if err := json.Unmarshal(decoded, certificate); err != nil {
		return nil, fmt.Errorf("json error: %s", err)
	}

	return certificate, nil
}

// Checks if the provided offline registration certificate is valid using
// the public RSA key to validate the included signature
func (cert *OfflineCertificate) IsValid() (bool, error) {
	key, keyErr := sccPublicKey()

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
	if cert.OfflinePayload != nil {
		return cert.OfflinePayload, nil
	}

	payload := &OfflinePayload{}
	raw, decodeErr := decodeBase64([]byte(cert.EncodedPayload))

	if decodeErr != nil {
		return nil, decodeErr
	}

	if jsonErr := json.Unmarshal(raw, payload); jsonErr != nil {
		return nil, fmt.Errorf("json: %s", jsonErr)
	}

	cert.OfflinePayload = payload
	return payload, nil
}

// Provides the signature encoded in the offline registration certificate which
// can be used to validate the base64 encoded payload in the [OfflineCertificate] structure.
func (cert *OfflineCertificate) Signature() ([]byte, error) {
	return decodeBase64([]byte(strings.TrimSpace(cert.EncodedSignature)))
}

// Validates a certificate by providing the subscriptions registration code. Only the
// rightful owner of the subscription has this information and therefore this can be leveraged
// to ensure the certificate is indeed used by the subscription owner
func (cert *OfflineCertificate) RegcodeMatches(regcode string) (bool, error) {
	payload, extractErr := cert.ExtractPayload()

	if extractErr != nil {
		return false, extractErr
	}

	return calcSHA256From(regcode) == payload.HashedRegcode, nil
}

// Validates a certificate by providing the system UUID. This validates that the UUID
// while generating the offline registration certificate is the same as the provided.
func (cert *OfflineCertificate) UUIDMatches(uuid string) (bool, error) {
	payload, extractErr := cert.ExtractPayload()

	if extractErr != nil {
		return false, extractErr
	}

	return calcSHA256From(uuid) == payload.HashedUUID, nil
}

// Checks if the certificate includes a certain product class and therefore is eligible for
// a certain product.
// Note: This does not check for a specific version but rather only the product.
func (cert *OfflineCertificate) ProductClassIncluded(name string) (bool, error) {
	payload, extractErr := cert.ExtractPayload()

	if extractErr != nil {
		return false, extractErr
	}

	for _, class := range payload.SubscriptionInfo.ProductClasses {
		if class.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (cert *OfflineCertificate) ExpiresAt() (time.Time, error) {
	payload, extractErr := cert.ExtractPayload()

	if extractErr != nil {
		return time.Now(), extractErr
	}

	return payload.SubscriptionInfo.ExpiresAt, nil
}
