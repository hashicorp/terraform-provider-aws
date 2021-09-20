package aws

import (
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// RandSSHKeyPair generates a public and private SSH key pair. The public key is
// returned in OpenSSH format, and the private key is PEM encoded.
// Copied from github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest,
//  with the addition of the key size
func RandSSHKeyPairSize(keySize int, comment string) (string, string, error) {
	privateKey, privateKeyPEM, err := genPrivateKey(keySize)
	if err != nil {
		return "", "", err
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	keyMaterial := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(publicKey)))
	return fmt.Sprintf("%s %s", keyMaterial, comment), privateKeyPEM, nil
}

func genPrivateKey(keySize int) (*rsa.PrivateKey, string, error) {
	privateKey, err := rsa.GenerateKey(crand.Reader, keySize)
	if err != nil {
		return nil, "", err
	}

	privateKeyPEM, err := pemEncode(x509.MarshalPKCS1PrivateKey(privateKey), "RSA PRIVATE KEY")
	if err != nil {
		return nil, "", err
	}

	return privateKey, privateKeyPEM, nil
}

func pemEncode(b []byte, block string) (string, error) {
	var buf bytes.Buffer
	pb := &pem.Block{Type: block, Bytes: b}
	if err := pem.Encode(&buf, pb); err != nil {
		return "", err
	}

	return buf.String(), nil
}
