package acctest

import (
	"strings"
	"testing"
)

func TestTLSRSAPrivateKeyPEM(t *testing.T) {
	key := TLSRSAPrivateKeyPEM(2048)

	if !strings.Contains(key, pemBlockTypeRsaPrivateKey) {
		t.Errorf("key does not contain RSA PRIVATE KEY: %s", key)
	}
}

func TestTLSRSAPublicKeyPEM(t *testing.T) {
	privateKey := TLSRSAPrivateKeyPEM(2048)
	publicKey := TLSRSAPublicKeyPEM(privateKey)

	if !strings.Contains(publicKey, pemBlockTypePublicKey) {
		t.Errorf("key does not contain PUBLIC KEY: %s", publicKey)
	}
}

func TestTLSRSAX509LocallySignedCertificatePEM(t *testing.T) {
	caKey := TLSRSAPrivateKeyPEM(2048)
	caCertificate := TLSRSAX509SelfSignedCACertificatePEM(caKey)
	key := TLSRSAPrivateKeyPEM(2048)
	certificate := TLSRSAX509LocallySignedCertificatePEM(caKey, caCertificate, key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("certificate does not contain CERTIFICATE: %s", certificate)
	}
}

func TestTLSRSAX509SelfSignedCACertificatePEM(t *testing.T) {
	caKey := TLSRSAPrivateKeyPEM(2048)
	caCertificate := TLSRSAX509SelfSignedCACertificatePEM(caKey)

	if !strings.Contains(caCertificate, pemBlockTypeCertificate) {
		t.Errorf("CA certificate does not contain CERTIFICATE: %s", caCertificate)
	}
}

func TestTLSRSAX509SelfSignedCertificatePEM(t *testing.T) {
	key := TLSRSAPrivateKeyPEM(2048)
	certificate := TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("certificate does not contain CERTIFICATE: %s", certificate)
	}
}

func TestTLSRSAX509CertificateRequestPEM(t *testing.T) {
	csr, key := TLSRSAX509CertificateRequestPEM(2048, "example.com")

	if !strings.Contains(csr, pemBlockTypeCertificateRequest) {
		t.Errorf("certificate does not contain CERTIFICATE REQUEST: %s", csr)
	}

	if !strings.Contains(key, pemBlockTypeRsaPrivateKey) {
		t.Errorf("certificate does not contain RSA PRIVATE KEY: %s", key)
	}
}
