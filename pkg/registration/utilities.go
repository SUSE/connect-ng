package registration

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

var (
	//go:embed scc_public_key.pem
	publicKey []byte
)

func decodeBase64(input []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(input)))
	size, err := base64.StdEncoding.Decode(decoded, input)

	if err != nil {
		return nil, fmt.Errorf("base64: %s", err)
	}

	// We need to cut at the end since it is possible to include
	// gibberish after len(input)
	return decoded[:size], nil
}

func sccPublicKey() (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKey)

	// PKCS#8 compatible
	key, parseErr := x509.ParsePKIXPublicKey(block.Bytes)

	if parseErr != nil {
		return nil, fmt.Errorf("public key: %s", parseErr)
	}

	if public, ok := key.(*rsa.PublicKey); ok {
		return public, nil
	}

	return nil, fmt.Errorf("public key: SCC public key can not be parsed. This is a bug")
}

func calcSHA256From(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))

	return hex.EncodeToString(hash.Sum(nil))
}
