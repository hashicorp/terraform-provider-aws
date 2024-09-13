// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestTLSRSAPrivateKeyPEM(t *testing.T) {
	t.Parallel()

	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)

	if !strings.Contains(key, acctest.PEMBlockTypeRSAPrivateKey) {
		t.Errorf("key does not contain RSA PRIVATE KEY: %s", key)
	}
}

func TestTLSRSAPublicKeyPEM(t *testing.T) {
	t.Parallel()

	privateKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	publicKey := acctest.TLSRSAPublicKeyPEM(t, privateKey)

	if !strings.Contains(publicKey, acctest.PEMBlockTypePublicKey) {
		t.Errorf("key does not contain PUBLIC KEY: %s", publicKey)
	}
}

func TestTLSRSAX509LocallySignedCertificatePEM(t *testing.T) {
	t.Parallel()

	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, "example.com")

	if !strings.Contains(certificate, acctest.PEMBlockTypeCertificate) {
		t.Errorf("certificate does not contain CERTIFICATE: %s", certificate)
	}
}

func TestTLSRSAX509SelfSignedCACertificatePEM(t *testing.T) {
	t.Parallel()

	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)

	if !strings.Contains(caCertificate, acctest.PEMBlockTypeCertificate) {
		t.Errorf("CA certificate does not contain CERTIFICATE: %s", caCertificate)
	}
}

func TestTLSRSAX509SelfSignedCertificatePEM(t *testing.T) {
	t.Parallel()

	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	if !strings.Contains(certificate, acctest.PEMBlockTypeCertificate) {
		t.Errorf("certificate does not contain CERTIFICATE: %s", certificate)
	}
}

func TestTLSRSAX509CertificateRequestPEM(t *testing.T) {
	t.Parallel()

	csr, key := acctest.TLSRSAX509CertificateRequestPEM(t, 2048, "example.com")

	if !strings.Contains(csr, acctest.PEMBlockTypeCertificateRequest) {
		t.Errorf("certificate does not contain CERTIFICATE REQUEST: %s", csr)
	}

	if !strings.Contains(key, acctest.PEMBlockTypeRSAPrivateKey) {
		t.Errorf("certificate does not contain RSA PRIVATE KEY: %s", key)
	}
}

func TestTLSECDSAPublicKeyPEM(t *testing.T) {
	t.Parallel()

	privateKey := acctest.TLSECDSAPrivateKeyPEM(t, "P-384")
	publicKey, _ := acctest.TLSECDSAPublicKeyPEM(t, privateKey)

	if !strings.Contains(publicKey, acctest.PEMBlockTypePublicKey) {
		t.Errorf("key does not contain PUBLIC KEY: %s", publicKey)
	}
}

func TestTLSPEMEscapeNewlines(t *testing.T) {
	t.Parallel()

	input := `
ABCD
12345
`
	want := "\\nABCD\\n12345\\n"

	if got := acctest.TLSPEMEscapeNewlines(input); got != want {
		t.Errorf("got: %s\nwant: %s", got, want)
	}
}

func TestTLSPEMRemovePublicKeyEncapsulationBoundaries(t *testing.T) {
	t.Parallel()

	input := `-----BEGIN PUBLIC KEY-----
ABCD
12345
-----END PUBLIC KEY-----
`
	want := "\nABCD\n12345\n\n"

	if got := acctest.TLSPEMRemovePublicKeyEncapsulationBoundaries(input); got != want {
		t.Errorf("got: %s\nwant: %s", got, want)
	}
}

func TestTLSPEMRemoveNewlines(t *testing.T) {
	t.Parallel()

	input := `
ABCD
12345

`
	want := "ABCD12345"

	if got := acctest.TLSPEMRemoveNewlines(input); got != want {
		t.Errorf("got: %s\nwant: %s", got, want)
	}
}
