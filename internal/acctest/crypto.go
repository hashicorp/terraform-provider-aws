// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

const (
	PEMBlockTypeCertificate        = `CERTIFICATE`
	PEMBlockTypeCertificateRequest = `CERTIFICATE REQUEST`
	PEMBlockTypeECPrivateKey       = `EC PRIVATE KEY`
	PEMBlockTypeRSAPrivateKey      = `RSA PRIVATE KEY`
	PEMBlockTypePublicKey          = `PUBLIC KEY`
)

var (
	tlsX509CertificateSerialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128) //nolint:gomnd
)

// TLSPEMRemovePublicKeyEncapsulationBoundaries removes public key
// pre and post encapsulation boundaries from a PEM string.
func TLSPEMRemovePublicKeyEncapsulationBoundaries(pem string) string {
	return removePEMEncapsulationBoundaries(pem, PEMBlockTypePublicKey)
}

func removePEMEncapsulationBoundaries(pem, label string) string {
	return strings.ReplaceAll(strings.ReplaceAll(pem, pemPreEncapsulationBoundary(label), ""), pemPostEncapsulationBoundary(label), "")
}

// See https://www.rfc-editor.org/rfc/rfc7468#section-2.
func pemPreEncapsulationBoundary(label string) string {
	return `-----BEGIN ` + label + `-----`
}

func pemPostEncapsulationBoundary(label string) string {
	return `-----END ` + label + `-----`
}

// TLSECDSAPublicKeyPEM generates an ECDSA private key PEM string using the specified elliptic curve.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func TLSECDSAPrivateKeyPEM(t *testing.T, curveName string) string {
	curve := ellipticCurveForName(curveName)

	if curve == nil {
		t.Fatalf("unsupported elliptic curve: %s", curveName)
	}

	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	if err != nil {
		t.Fatal(err)
	}

	bytes, err := x509.MarshalECPrivateKey(key)

	if err != nil {
		t.Fatal(err)
	}

	block := &pem.Block{
		Bytes: bytes,
		Type:  PEMBlockTypeECPrivateKey,
	}

	return string(pem.EncodeToMemory(block))
}

// TLSECDSAPublicKeyPEM generates an ECDSA public key PEM string and fingerprint.
func TLSECDSAPublicKeyPEM(t *testing.T, keyPem string) (string, string) {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)

	if err != nil {
		t.Fatal(err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)

	if err != nil {
		t.Fatal(err)
	}

	block := &pem.Block{
		Bytes: publicKeyBytes,
		Type:  PEMBlockTypePublicKey,
	}

	md5sum := md5.Sum(publicKeyBytes)
	hexarray := make([]string, len(md5sum))
	for i, c := range md5sum {
		hexarray[i] = hex.EncodeToString([]byte{c})
	}

	return string(pem.EncodeToMemory(block)), strings.Join(hexarray, ":")
}

// TLSRSAPrivateKeyPEM generates a RSA private key PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func TLSRSAPrivateKeyPEM(t *testing.T, bits int) string {
	key, err := rsa.GenerateKey(rand.Reader, bits)

	if err != nil {
		t.Fatal(err)
	}

	block := &pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(key),
		Type:  PEMBlockTypeRSAPrivateKey,
	}

	return string(pem.EncodeToMemory(block))
}

// TLSRSAPublicKeyPEM generates a RSA public key PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: public_key_pem = "%[1]s"
func TLSRSAPublicKeyPEM(t *testing.T, keyPem string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		t.Fatal(err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)

	if err != nil {
		t.Fatal(err)
	}

	block := &pem.Block{
		Bytes: publicKeyBytes,
		Type:  PEMBlockTypePublicKey,
	}

	return string(pem.EncodeToMemory(block))
}

// TLSRSAX509LocallySignedCertificatePEM generates a local CA x509 certificate PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: certificate_pem = "%[1]s"
func TLSRSAX509LocallySignedCertificatePEM(t *testing.T, caKeyPem, caCertificatePem, keyPem, commonName string) string {
	caCertificateBlock, _ := pem.Decode([]byte(caCertificatePem))

	caCertificate, err := x509.ParseCertificate(caCertificateBlock.Bytes)

	if err != nil {
		t.Fatal(err)
	}

	caKeyBlock, _ := pem.Decode([]byte(caKeyPem))

	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)

	if err != nil {
		t.Fatal(err)
	}

	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		t.Fatal(err)
	}

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		t.Fatal(err)
	}

	certificate := &x509.Certificate{
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:              time.Now().Add(24 * time.Hour), //nolint:gomnd
		NotBefore:             time.Now(),
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"ACME Examples, Inc"},
		},
	}

	certificateBytes, err := x509.CreateCertificate(rand.Reader, certificate, caCertificate, &key.PublicKey, caKey)

	if err != nil {
		t.Fatal(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  PEMBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// TLSRSAX509SelfSignedCACertificatePEM generates a x509 CA certificate PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: root_certificate_pem = "%[1]s"
func TLSRSAX509SelfSignedCACertificatePEM(t *testing.T, keyPem string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		t.Fatal(err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)

	if err != nil {
		t.Fatal(err)
	}

	publicKeyBytesSha1 := sha1.Sum(publicKeyBytes)

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		t.Fatal(err)
	}

	certificate := &x509.Certificate{
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:              time.Now().Add(24 * time.Hour), //nolint:gomnd
		NotBefore:             time.Now(),
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			CommonName:   "ACME Root CA",
			Organization: []string{"ACME Examples, Inc"},
		},
		SubjectKeyId: publicKeyBytesSha1[:],
	}

	certificateBytes, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &key.PublicKey, key)

	if err != nil {
		t.Fatal(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  PEMBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// TLSRSAX509SelfSignedCACertificateForRolesAnywhereTrustAnchorPEM generates a x509 CA certificate PEM string.
// The CA certificate is suitable for use as an IAM RolesAnywhere Trust Anchor.
// See https://docs.aws.amazon.com/rolesanywhere/latest/userguide/trust-model.html#signature-verification.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: root_certificate_pem = "%[1]s"
func TLSRSAX509SelfSignedCACertificateForRolesAnywhereTrustAnchorPEM(t *testing.T, keyPem string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		t.Fatal(err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)

	if err != nil {
		t.Fatal(err)
	}

	publicKeyBytesSha1 := sha1.Sum(publicKeyBytes)

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		t.Fatal(err)
	}

	certificate := &x509.Certificate{
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		NotAfter:              time.Now().Add(24 * time.Hour), //nolint:gomnd
		NotBefore:             time.Now(),
		SerialNumber:          serialNumber,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		Subject: pkix.Name{
			CommonName:   "ACME Root CA",
			Organization: []string{"ACME Examples, Inc"},
		},
		SubjectKeyId: publicKeyBytesSha1[:],
	}

	certificateBytes, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &key.PublicKey, key)

	if err != nil {
		t.Fatal(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  PEMBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// TLSRSAX509SelfSignedCertificatePEM generates a x509 certificate PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func TLSRSAX509SelfSignedCertificatePEM(t *testing.T, keyPem, commonName string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		t.Fatal(err)
	}

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		t.Fatal(err)
	}

	certificate := &x509.Certificate{
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:              time.Now().Add(24 * time.Hour), //nolint:gomnd
		NotBefore:             time.Now(),
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"ACME Examples, Inc"},
		},
	}

	certificateBytes, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &key.PublicKey, key)

	if err != nil {
		t.Fatal(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  PEMBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// TLSRSAX509CertificateRequestPEM generates a x509 certificate request PEM string
// and a RSA private key PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: certificate_signing_request_pem = "%[1]s" private_key_pem = "%[2]s"
func TLSRSAX509CertificateRequestPEM(t *testing.T, keyBits int, commonName string) (string, string) {
	keyBytes, err := rsa.GenerateKey(rand.Reader, keyBits)

	if err != nil {
		t.Fatal(err)
	}

	csr := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"ACME Examples, Inc"},
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csr, keyBytes)

	if err != nil {
		t.Fatal(err)
	}

	csrBlock := &pem.Block{
		Bytes: csrBytes,
		Type:  PEMBlockTypeCertificateRequest,
	}

	keyBlock := &pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(keyBytes),
		Type:  PEMBlockTypeRSAPrivateKey,
	}

	return string(pem.EncodeToMemory(csrBlock)), string(pem.EncodeToMemory(keyBlock))
}

func TLSPEMEscapeNewlines(pem string) string {
	return strings.ReplaceAll(pem, "\n", "\\n")
}

func TLSPEMRemoveNewlines(pem string) string {
	return strings.ReplaceAll(pem, "\n", "")
}

func ellipticCurveForName(name string) elliptic.Curve {
	switch name {
	case "P-224":
		return elliptic.P224()
	case "P-256":
		return elliptic.P256()
	case "P-384":
		return elliptic.P384()
	case "P-521":
		return elliptic.P521()
	}

	return nil
}
