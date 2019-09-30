package aws

import (
	"strings"
	"testing"
)

func TestTlsRsaPrivateKeyPem(t *testing.T) {
	key := tlsRsaPrivateKeyPem(2048)

	if !strings.Contains(key, pemBlockTypeRsaPrivateKey) {
		t.Errorf("key does not contain RSA PRIVATE KEY: %s", key)
	}
}

func TestTlsRsaPublicKeyPem(t *testing.T) {
	privateKey := tlsRsaPrivateKeyPem(2048)
	publicKey := tlsRsaPublicKeyPem(privateKey)

	if !strings.Contains(publicKey, pemBlockTypePublicKey) {
		t.Errorf("key does not contain PUBLIC KEY: %s", publicKey)
	}
}

func TestTlsRsaX509CertificatePem(t *testing.T) {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("key does not contain CERTIFICATE: %s", certificate)
	}
}
