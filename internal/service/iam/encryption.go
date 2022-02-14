package iam

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/vault/helper/pgpkeys"
)

// retrieveGPGKey returns the PGP key specified as the pgpKey parameter, or queries
// the public key from the keybase or external service if the parameter is a username
// prefixed with the phrase "keybase:" or "external:".
// With "external:", the parameter is expected to be in the form "external:username|url".
// The url is expected to be a valid url to a public key, it can also be a formatable url
// which will use the associated username.
func retrieveGPGKey(pgpKey string) (string, error) {
	const keybasePrefix = "keybase:"
	const externalPrefix = "external:"

	encryptionKey := pgpKey
	if strings.HasPrefix(pgpKey, keybasePrefix) || strings.HasPrefix(pgpKey, externalPrefix) {
		publicKeys, err := pgpkeys.FetchPubkeys([]string{pgpKey})
		if err != nil {
			return "", fmt.Errorf("Error retrieving Public Key for %s: %w", pgpKey, err)
		}
		encryptionKey = publicKeys[pgpKey]
	}

	return encryptionKey, nil
}

// encryptValue encrypts the given value with the given encryption key. Description
// should be set such that errors return a meaningful user-facing response.
func encryptValue(encryptionKey, value, description string) (string, string, error) {
	fingerprints, encryptedValue, err :=
		pgpkeys.EncryptShares([][]byte{[]byte(value)}, []string{encryptionKey})
	if err != nil {
		return "", "", fmt.Errorf("Error encrypting %s: %w", description, err)
	}

	return fingerprints[0], base64.StdEncoding.EncodeToString(encryptedValue[0]), nil
}
