package acctest

import (
	"crypto/x509"
	"strings"
	"testing"
)

func TestTlsRsaPrivateKeyPem(t *testing.T) {
	key := TLSRSAPrivateKeyPEM(2048)

	if !strings.Contains(key, pemBlockTypeRsaPrivateKey) {
		t.Errorf("key does not contain %s:\n%s", pemBlockTypeRsaPrivateKey, key)
	}
}

func TestTlsRsaPrivateKeyEncrypredPem(t *testing.T) {
	key := TLSRSAPrivateKeyPEM(2048)
	encryptedKey := TLSPEMEncrypt(key, "passphrase")
	if !strings.Contains(encryptedKey, pemBlockTypeRsaPrivateKey) {
		t.Errorf("key does not contain %s:\n%s", pemBlockTypeRsaPrivateKey, key)
	}
	if !strings.Contains(encryptedKey, pemBlockProcTypeEncrypted) {
		t.Errorf("key does not contain %s:\n%s", pemBlockProcTypeEncrypted, key)
	}
}

func TestTlsEcPrivateKeyPem(t *testing.T) {
	key := TLSECPrivateKeyPEM(256)

	if !strings.Contains(key, pemBlockTypeEcPrivateKey) {
		t.Errorf("key does not contain %s:\n%s", pemBlockTypeEcPrivateKey, key)
	}
}

func TestTlsEcPrivateKeyEncryptedPem(t *testing.T) {
	key := TLSECPrivateKeyPEM(256)
	encryptedKey := TLSPEMEncrypt(key, "passphrase")
	if !strings.Contains(encryptedKey, pemBlockTypeEcPrivateKey) {
		t.Errorf("key does not contain %s:\n%s", pemBlockTypeEcPrivateKey, key)
	}
	if !strings.Contains(encryptedKey, pemBlockProcTypeEncrypted) {
		t.Errorf("key does not contain %s:\n%s", pemBlockProcTypeEncrypted, key)
	}
}

func TestTlsRsaPublicKeyPem(t *testing.T) {
	privateKey := TLSRSAPrivateKeyPEM(2048)
	publicKey := TLSPublicKeyPEM(privateKey)

	if !strings.Contains(publicKey, pemBlockTypePublicKey) {
		t.Errorf("key does not contain %s:\n%s", pemBlockTypePublicKey, publicKey)
	}
}

func TestTlsEcPublicKeyPem(t *testing.T) {
	privateKey := TLSECPrivateKeyPEM(256)
	publicKey := TLSPublicKeyPEM(privateKey)

	if !strings.Contains(publicKey, pemBlockTypePublicKey) {
		t.Errorf("key does not contain %s:\n%s", pemBlockTypePublicKey, publicKey)
	}
}

func TestTlsRsaX509LocallySignedCertificatePem(t *testing.T) {
	caKey := TLSRSAPrivateKeyPEM(2048)
	caCertificate := TLSRSAX509SelfSignedCACertificatePEM(caKey)
	key := TLSRSAPrivateKeyPEM(2048)
	certificate := TLSRSAX509LocallySignedCertificatePEM(caKey, caCertificate, key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("certificate does not contain %s:\n%s", pemBlockTypeCertificate, certificate)
	}
}

func TestTlsEcX509LocallySignedCertificatePem(t *testing.T) {
	caKey := TLSECPrivateKeyPEM(384)
	caCertificate := TLSX509SelfSignedCACertificatePEM(caKey)
	key := TLSECPrivateKeyPEM(256)
	certificate := TLSX509LocallySignedCertificatePEM(caKey, caCertificate, key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("certificate does not contain %s:\n%s", pemBlockTypeCertificate, certificate)
	}
}

func TestTlsRsaX509SelfSignedCaCertificatePem(t *testing.T) {
	caKey := TLSRSAPrivateKeyPEM(2048)
	caCertificate := TLSRSAX509SelfSignedCACertificatePEM(caKey)

	if !strings.Contains(caCertificate, pemBlockTypeCertificate) {
		t.Errorf("CA certificate does not contain %s:\n%s", pemBlockTypeCertificate, caCertificate)
	}
}

func TestTlsEcX509SelfSignedCaCertificatePem(t *testing.T) {
	caKey := TLSECPrivateKeyPEM(384)
	caCertificate := TLSX509SelfSignedCACertificatePEM(caKey)

	if !strings.Contains(caCertificate, pemBlockTypeCertificate) {
		t.Errorf("CA certificate does not contain %s:\n%s", pemBlockTypeCertificate, caCertificate)
	}
}

func TestTlsRsaX509SelfSignedCertificatePem(t *testing.T) {
	key := TLSRSAPrivateKeyPEM(2048)
	certificate := TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("certificate does not contain %s:\n%s", pemBlockTypeCertificate, certificate)
	}
}

func TestTlsEcX509SelfSignedCertificatePem(t *testing.T) {
	key := TLSECPrivateKeyPEM(256)
	certificate := TLSX509SelfSignedCertificatePEM(key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("certificate does not contain %s:\n%s", pemBlockTypeCertificate, certificate)
	}
}

func TestTlsRsaX509CertificateRequestPem(t *testing.T) {
	csr, key := TLSRSAX509CertificateRequestPEM(2048, "example.com")

	if !strings.Contains(csr, pemBlockTypeCertificateRequest) {
		t.Errorf("certificate does not contain %s:\n%s", pemBlockTypeCertificateRequest, csr)
	}
	if !strings.Contains(key, pemBlockTypeRsaPrivateKey) {
		t.Errorf("certificate does not contain %s:\n%s", pemBlockTypeRsaPrivateKey, key)
	}
}

func TestTlsEcX509CertificateRequestPem(t *testing.T) {
	csr, key := TLSX509CertificateRequestPEM(x509.ECDSA, 256, "example.com")

	if !strings.Contains(csr, pemBlockTypeCertificateRequest) {
		t.Errorf("certificate does not contain %s:\n%s", pemBlockTypeCertificateRequest, csr)
	}
	if !strings.Contains(key, pemBlockTypeEcPrivateKey) {
		t.Errorf("certificate does not contain %s:\n%s", pemBlockTypeEcPrivateKey, key)
	}
}
