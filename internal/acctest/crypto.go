package acctest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	pemBlockTypeCertificate        = `CERTIFICATE`
	pemBlockTypeRsaPrivateKey      = `RSA PRIVATE KEY`
	pemBlockTypeEcPrivateKey       = `EC PRIVATE KEY`
	pemBlockTypePublicKey          = `PUBLIC KEY`
	pemBlockTypeCertificateRequest = `CERTIFICATE REQUEST`
	pemBlockProcTypeEncrypted      = `Proc-Type: 4,ENCRYPTED`
)

var tlsX509CertificateSerialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128) //nolint:gomnd

// TLSRSAPrivateKeyPEM generates a RSA private key PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func TLSRSAPrivateKeyPEM(bits int) string {
	key, err := rsa.GenerateKey(rand.Reader, bits)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	block := &pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(key),
		Type:  pemBlockTypeRsaPrivateKey,
	}

	return string(pem.EncodeToMemory(block))
}

// TLSPrivateKey returns the private key from a RSA or ECDSA PEM string.
func TLSPrivateKey(keyPem string) interface{} {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	if keyBlock.Type == pemBlockTypeRsaPrivateKey {
		key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			//lintignore:R009
			panic(err)
		}
		return key
	} else if keyBlock.Type == pemBlockTypeEcPrivateKey {
		key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			//lintignore:R009
			panic(err)
		}
		return key
	} else {
		panic(fmt.Errorf("unsupported PEM block %s", keyBlock.Type))
	}
}

// TLSRSAPublicKeyPEM generates a RSA public key PEM string.
// Replaced by TLSPublicKeyPEM

// TLSPublicKeyPEM generates a public key PEM string from a RSA or ECDSA private key.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: public_key_pem = "%[1]s"
func TLSPublicKeyPEM(keyPem string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	privateKey := TLSPrivateKey(keyPem)

	var publicKeyBytes []byte
	var err error

	if keyBlock.Type == pemBlockTypeRsaPrivateKey {
		publicKeyBytes, err = x509.MarshalPKIXPublicKey(&privateKey.(*rsa.PrivateKey).PublicKey)
	} else if keyBlock.Type == pemBlockTypeEcPrivateKey {
		publicKeyBytes, err = x509.MarshalPKIXPublicKey(&privateKey.(*ecdsa.PrivateKey).PublicKey)
	} else {
		panic(fmt.Errorf("unsupported PEM block %s", keyBlock.Type))
	}

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	block := &pem.Block{
		Bytes: publicKeyBytes,
		Type:  pemBlockTypePublicKey,
	}

	return string(pem.EncodeToMemory(block))
}

// TLSRSAX509LocallySignedCertificatePEM generates a local CA x509 certificate PEM string.
// Replaced by TLSX509LocallySignedCertificatePEM

// TLSX509LocallySignedCertificatePEM generates a local CA x509 certificate PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: certificate_pem = "%[1]s"
func TLSX509LocallySignedCertificatePEM(caKeyPem, caCertificatePem, keyPem, commonName string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))
	caCertificateBlock, _ := pem.Decode([]byte(caCertificatePem))

	caCertificate, err := x509.ParseCertificate(caCertificateBlock.Bytes)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		//lintignore:R009
		panic(err)
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

	caKey := TLSPrivateKey(caKeyPem)

	key := TLSPrivateKey(keyPem)

	var certificateBytes []byte

	if keyBlock.Type == pemBlockTypeRsaPrivateKey {
		certificateBytes, err = x509.CreateCertificate(rand.Reader, certificate, caCertificate, &key.(*rsa.PrivateKey).PublicKey, caKey)
	} else if keyBlock.Type == pemBlockTypeEcPrivateKey {
		certificateBytes, err = x509.CreateCertificate(rand.Reader, certificate, caCertificate, &key.(*ecdsa.PrivateKey).PublicKey, caKey)
	} else {
		panic(fmt.Errorf("unsupported PEM block %s", keyBlock.Type))
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  pemBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// TLSRSAX509SelfSignedCACertificatePEM generates a x509 CA certificate PEM string.
// Replaced by TLSX509SelfSignedCACertificatePEM

// TLSX509SelfSignedCACertificatePEM generates a x509 CA certificate PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: root_certificate_pem = "%[1]s"
func TLSX509SelfSignedCACertificatePEM(keyPem string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key := TLSPrivateKey(keyPem)
	var publicKeyBytes []byte
	var err error

	if keyBlock.Type == pemBlockTypeRsaPrivateKey {
		publicKeyBytes, err = x509.MarshalPKIXPublicKey(&key.(*rsa.PrivateKey).PublicKey)
	} else if keyBlock.Type == pemBlockTypeEcPrivateKey {
		publicKeyBytes, err = x509.MarshalPKIXPublicKey(&key.(*ecdsa.PrivateKey).PublicKey)
	} else {
		panic(fmt.Errorf("unsupported PEM block %s", keyBlock.Type))
	}

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	publicKeyBytesSha1 := sha1.Sum(publicKeyBytes)

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		//lintignore:R009
		panic(err)
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

	var certificateBytes []byte

	if keyBlock.Type == pemBlockTypeRsaPrivateKey {
		certificateBytes, err = x509.CreateCertificate(rand.Reader, certificate, certificate, &key.(*rsa.PrivateKey).PublicKey, key)
	} else if keyBlock.Type == pemBlockTypeEcPrivateKey {
		certificateBytes, err = x509.CreateCertificate(rand.Reader, certificate, certificate, &key.(*ecdsa.PrivateKey).PublicKey, key)
	} else {
		panic(fmt.Errorf("unsupported PEM block %s", keyBlock.Type))
	}

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  pemBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// TLSRSAX509SelfSignedCertificatePEM generates a x509 certificate PEM string.
// Replaced by TLSX509SelfSignedCertificatePEM

// TLSX509SelfSignedCertificatePEM generates a x509 certificate PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func TLSX509SelfSignedCertificatePEM(keyPem, commonName string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key := TLSPrivateKey(keyPem)

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		//lintignore:R009
		panic(err)
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

	var certificateBytes []byte
	if keyBlock.Type == pemBlockTypeRsaPrivateKey {
		certificateBytes, err = x509.CreateCertificate(rand.Reader, certificate, certificate, &key.(*rsa.PrivateKey).PublicKey, key)
	} else if keyBlock.Type == pemBlockTypeEcPrivateKey {
		certificateBytes, err = x509.CreateCertificate(rand.Reader, certificate, certificate, &key.(*ecdsa.PrivateKey).PublicKey, key)
	} else {
		panic(fmt.Errorf("unsupported PEM block %s", keyBlock.Type))
	}

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  pemBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// TLSRSAX509CertificateRequestPEM generates a x509 certificate request PEM string
// Replaced by TLSX509CertificateRequestPEM

// TLSX509CertificateRequestPEM generates a x509 certificate request PEM string
// and a RSA or ECDSA private key PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: certificate_signing_request_pem = "%[1]s" private_key_pem = "%[2]s"
func TLSX509CertificateRequestPEM(publicKeyAlgorithm x509.PublicKeyAlgorithm, keyBits int, commonName string) (string, string) {
	var keyBytes interface{}
	keyBlock := &pem.Block{}
	var signatureAlgorithm x509.SignatureAlgorithm
	var err error

	if publicKeyAlgorithm == x509.RSA {
		keyBytes, err = rsa.GenerateKey(rand.Reader, keyBits)
		if err != nil {
			//lintignore:R009
			panic(err)
		}
		keyBlock = &pem.Block{
			Bytes: x509.MarshalPKCS1PrivateKey(keyBytes.(*rsa.PrivateKey)),
			Type:  pemBlockTypeRsaPrivateKey,
		}
		signatureAlgorithm = x509.SHA256WithRSA
	} else if publicKeyAlgorithm == x509.ECDSA {
		keyBytes, err = ecdsa.GenerateKey(TLSECCurve(keyBits), rand.Reader)
		if err != nil {
			//lintignore:R009
			panic(err)
		}
		privateKey, err := x509.MarshalECPrivateKey(keyBytes.(*ecdsa.PrivateKey))
		if err != nil {
			//lintignore:R009
			panic(err)
		}
		keyBlock = &pem.Block{
			Bytes: privateKey,
			Type:  pemBlockTypeEcPrivateKey,
		}
		signatureAlgorithm = TLSECSignature(keyBits)
	} else {
		panic(fmt.Errorf("unsupported public key algorithm"))
	}

	csr := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"ACME Examples, Inc"},
		},
		SignatureAlgorithm: signatureAlgorithm,
		PublicKeyAlgorithm: publicKeyAlgorithm,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csr, keyBytes)
	if err != nil {
		//lintignore:R009
		panic(err)
	}

	csrBlock := &pem.Block{
		Bytes: csrBytes,
		Type:  pemBlockTypeCertificateRequest,
	}

	return string(pem.EncodeToMemory(csrBlock)), string(pem.EncodeToMemory(keyBlock))
}

// TLSECPublicKeyPEM generates a ECDSA public key PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: public_key_pem = "%[1]s"
func TLSECPublicKeyPEM(keyPem string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	block := &pem.Block{
		Bytes: publicKeyBytes,
		Type:  pemBlockTypePublicKey,
	}

	return string(pem.EncodeToMemory(block))
}

func TLSPEMEscapeNewlines(pem string) string {
	return strings.ReplaceAll(pem, "\n", "\\n")
}

// TLSPEMEncrypt encrypts a PEM block with a passphrase.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func TLSPEMEncrypt(keyPem, passphrase string) string {
	clearBlock, _ := pem.Decode([]byte(keyPem))

	encryptedBlock, err := x509.EncryptPEMBlock(rand.Reader, clearBlock.Type, clearBlock.Bytes, []byte(passphrase), x509.PEMCipherAES256)
	if err != nil {
		panic(err)
	}
	return string(pem.EncodeToMemory(encryptedBlock))
}

// TLSECPrivateKeyPEM generates a ECDSA private key PEM string.
// Wrap with TLSPEMEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func TLSECPrivateKeyPEM(bits int) string {
	key, err := ecdsa.GenerateKey(TLSECCurve(bits), rand.Reader)
	if err != nil {
		//lintignore:R009
		panic(err)
	}

	privateKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		//lintignore:R009
		panic(err)
	}

	block := &pem.Block{
		Bytes: privateKey,
		Type:  pemBlockTypeEcPrivateKey,
	}

	return string(pem.EncodeToMemory(block))
}

// Select an appropriate EC curve for the number of bits
func TLSECCurve(bits int) elliptic.Curve {
	switch bits {
	case 224:
		return elliptic.P224()
	case 256:
		return elliptic.P256()
	case 384:
		return elliptic.P384()
	case 512:
		return elliptic.P521()
	case 521:
		return elliptic.P521()
	default:
		panic(fmt.Errorf("unsupported curve bits %d", bits))
	}
}

// Select an appropriate EC signature for the number of bits
func TLSECSignature(bits int) x509.SignatureAlgorithm {
	switch {
	case bits <= 256:
		return x509.ECDSAWithSHA256
	case bits <= 384:
		return x509.ECDSAWithSHA384
	case bits > 384:
		return x509.ECDSAWithSHA512
	}
	return x509.ECDSAWithSHA256
}

// Aliases for compatibility wit existing RSA function names
func TLSRSAPublicKeyPEM(keyPem string) string {
	return TLSPublicKeyPEM(keyPem)
}

func TLSRSAX509LocallySignedCertificatePEM(caKeyPem, caCertificatePem, keyPem, commonName string) string {
	return TLSX509LocallySignedCertificatePEM(caKeyPem, caCertificatePem, keyPem, commonName)
}

func TLSRSAX509SelfSignedCACertificatePEM(keyPem string) string {
	return TLSX509SelfSignedCACertificatePEM(keyPem)
}

func TLSRSAX509SelfSignedCertificatePEM(keyPem, commonName string) string {
	return TLSX509SelfSignedCertificatePEM(keyPem, commonName)
}

func TLSRSAX509CertificateRequestPEM(keyBits int, commonName string) (string, string) {
	return TLSX509CertificateRequestPEM(x509.RSA, keyBits, commonName)
}
