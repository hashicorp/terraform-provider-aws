// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type TestCertificateAuthority struct {
	cert       *x509.Certificate
	privateKey *rsa.PrivateKey
	certPEM    []byte
	keyPEM     []byte
}

type TestClientCertificate struct {
	cert       *x509.Certificate
	privateKey *rsa.PrivateKey
	certPEM    []byte
	keyPEM     []byte
	tlsCert    tls.Certificate
}

func generateTestCA(t *testing.T) *TestCertificateAuthority {
	t.Helper()

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate CA private key: %v", err)
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test CA"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:              []string{"localhost"},
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create CA certificate: %v", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		t.Fatalf("Failed to parse CA certificate: %v", err)
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

	caPrivateKeyDER, err := x509.MarshalPKCS8PrivateKey(caPrivateKey)
	if err != nil {
		t.Fatalf("Failed to marshal CA private key: %v", err)
	}
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: caPrivateKeyDER})

	return &TestCertificateAuthority{
		cert:       caCert,
		privateKey: caPrivateKey,
		certPEM:    caCertPEM,
		keyPEM:     caKeyPEM,
	}
}

func (ca *TestCertificateAuthority) generateClientCertificate(t *testing.T, commonName string) *TestClientCertificate {
	t.Helper()

	clientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate client private key: %v", err)
	}

	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:  []string{"Test Client"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"San Francisco"},
			CommonName:    commonName,
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:    []string{"localhost"},
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, &clientTemplate, ca.cert, &clientPrivateKey.PublicKey, ca.privateKey)
	if err != nil {
		t.Fatalf("Failed to create client certificate: %v", err)
	}

	clientCert, err := x509.ParseCertificate(clientCertDER)
	if err != nil {
		t.Fatalf("Failed to parse client certificate: %v", err)
	}

	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})

	clientPrivateKeyDER, err := x509.MarshalPKCS8PrivateKey(clientPrivateKey)
	if err != nil {
		t.Fatalf("Failed to marshal client private key: %v", err)
	}
	clientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: clientPrivateKeyDER})

	tlsCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	if err != nil {
		t.Fatalf("Failed to create TLS certificate: %v", err)
	}

	return &TestClientCertificate{
		cert:       clientCert,
		privateKey: clientPrivateKey,
		certPEM:    clientCertPEM,
		keyPEM:     clientKeyPEM,
		tlsCert:    tlsCert,
	}
}

func (ca *TestCertificateAuthority) generateServerCertificate(t *testing.T) *TestClientCertificate {
	t.Helper()

	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate server private key: %v", err)
	}

	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization:  []string{"Test Server"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"San Francisco"},
			CommonName:    "localhost",
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:    []string{"localhost"},
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, ca.cert, &serverPrivateKey.PublicKey, ca.privateKey)
	if err != nil {
		t.Fatalf("Failed to create server certificate: %v", err)
	}

	serverCert, err := x509.ParseCertificate(serverCertDER)
	if err != nil {
		t.Fatalf("Failed to parse server certificate: %v", err)
	}

	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})

	serverPrivateKeyDER, err := x509.MarshalPKCS8PrivateKey(serverPrivateKey)
	if err != nil {
		t.Fatalf("Failed to marshal server private key: %v", err)
	}
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: serverPrivateKeyDER})

	tlsCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatalf("Failed to create server TLS certificate: %v", err)
	}

	return &TestClientCertificate{
		cert:       serverCert,
		privateKey: serverPrivateKey,
		certPEM:    serverCertPEM,
		keyPEM:     serverKeyPEM,
		tlsCert:    tlsCert,
	}
}

type MTLSTestServer struct {
	server     *httptest.Server
	ca         *TestCertificateAuthority
	serverCert *TestClientCertificate
	url        string
}

func newMTLSTestServer(t *testing.T, ca *TestCertificateAuthority) *MTLSTestServer {
	t.Helper()

	serverCert := ca.generateServerCertificate(t)

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(ca.cert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert.tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "Client certificate required", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/":
			response := map[string]any{
				"GetCallerIdentityResponse": map[string]any{
					"GetCallerIdentityResult": map[string]any{
						"Account": "123456789012",
						"Arn":     "arn:aws:iam::123456789012:user/test-user",
						"UserId":  "AIDACKCEVSQ6C2EXAMPLE",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}

		case "/health":
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				t.Fatalf("failed to write OK: %v", err)
			}

		default:
			http.NotFound(w, r)
		}
	}))

	server.TLS = tlsConfig
	server.StartTLS()

	return &MTLSTestServer{
		server:     server,
		ca:         ca,
		serverCert: serverCert,
		url:        server.URL,
	}
}

func (s *MTLSTestServer) Close() {
	s.server.Close()
}

func writeCertificateFiles(t *testing.T, ca *TestCertificateAuthority, clientCert *TestClientCertificate) (caCertFile, clientCertFile, clientKeyFile string) {
	t.Helper()

	tempDir := t.TempDir()

	caCertFile = filepath.Join(tempDir, "ca-cert.pem")
	err := os.WriteFile(caCertFile, ca.certPEM, 0600)
	if err != nil {
		t.Fatalf("Failed to write CA certificate file: %v", err)
	}

	clientCertFile = filepath.Join(tempDir, "client-cert.pem")
	err = os.WriteFile(clientCertFile, clientCert.certPEM, 0600)
	if err != nil {
		t.Fatalf("Failed to write client certificate file: %v", err)
	}

	clientKeyFile = filepath.Join(tempDir, "client-key.pem")
	err = os.WriteFile(clientKeyFile, clientCert.keyPEM, 0600)
	if err != nil {
		t.Fatalf("Failed to write client private key file: %v", err)
	}

	return caCertFile, clientCertFile, clientKeyFile
}

func TestMTLSEndToEnd(t *testing.T) {
	t.Parallel()
	ca := generateTestCA(t)

	clientCert := ca.generateClientCertificate(t, "terraform-aws-provider-test")

	caCertFile, clientCertFile, clientKeyFile := writeCertificateFiles(t, ca, clientCert)

	t.Run("successful_mtls_connection", func(t *testing.T) {
		t.Parallel()
		testServer := newMTLSTestServer(t, ca)
		defer testServer.Close()
		config := &Config{
			ClientCertificate:          clientCertFile,
			ClientPrivateKey:           clientKeyFile,
			ClientPrivateKeyPassphrase: "",
			CustomCABundle:             caCertFile,
		}

		awsConfig := &aws.Config{
			HTTPClient: &http.Client{
				Transport: &http.Transport{},
			},
		}

		err := config.configureHTTPClientMTLS(awsConfig)
		if err != nil {
			t.Fatalf("Failed to configure mTLS: %v", err)
		}

		req, err := http.NewRequest(http.MethodGet, testServer.url+"/health", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		httpClient, ok := awsConfig.HTTPClient.(*http.Client)
		if !ok {
			t.Fatalf("Expected HTTP client to be *http.Client, got %T", awsConfig.HTTPClient)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make mTLS request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if string(body) != "OK" {
			t.Fatalf("Expected response 'OK', got %q", string(body))
		}
	})

	t.Run("mtls_connection_without_client_cert_fails", func(t *testing.T) {
		t.Parallel()
		testServer := newMTLSTestServer(t, ca)
		defer testServer.Close()
		client := &http.Client{}

		req, err := http.NewRequest(http.MethodGet, testServer.url+"/health", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = client.Do(req)
		if err == nil {
			t.Fatal("Expected mTLS connection to fail without client certificate, but it succeeded")
		}
	})

	t.Run("mtls_with_wrong_client_cert_fails", func(t *testing.T) {
		t.Parallel()
		testServer := newMTLSTestServer(t, ca)
		defer testServer.Close()
		wrongCA := generateTestCA(t)
		wrongClientCert := wrongCA.generateClientCertificate(t, "wrong-client")

		_, wrongClientCertFile, wrongClientKeyFile := writeCertificateFiles(t, wrongCA, wrongClientCert)

		config := &Config{
			ClientCertificate:          wrongClientCertFile,
			ClientPrivateKey:           wrongClientKeyFile,
			ClientPrivateKeyPassphrase: "",
			CustomCABundle:             caCertFile,
		}

		awsConfig := &aws.Config{
			HTTPClient: &http.Client{
				Transport: &http.Transport{},
			},
		}

		err := config.configureHTTPClientMTLS(awsConfig)
		if err != nil {
			t.Fatalf("Failed to configure mTLS: %v", err)
		}

		req, err := http.NewRequest(http.MethodGet, testServer.url+"/health", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		httpClient, ok := awsConfig.HTTPClient.(*http.Client)
		if !ok {
			t.Fatalf("Expected HTTP client to be *http.Client, got %T", awsConfig.HTTPClient)
		}

		_, err = httpClient.Do(req)
		if err == nil {
			t.Fatal("Expected mTLS connection to fail with wrong client certificate, but it succeeded")
		}
	})
}

func TestMTLSEndToEndWithEncryptedKey(t *testing.T) {
	t.Parallel()
	ca := generateTestCA(t)

	clientCert := ca.generateClientCertificate(t, "terraform-aws-provider-encrypted-test")

	passphrase := "test-passphrase-123"
	keyBlock, _ := pem.Decode(clientCert.keyPEM)
	if keyBlock == nil {
		t.Fatal("Failed to decode private key PEM")
	}

	//nolint:staticcheck // SA1019: Using deprecated x509.EncryptPEMBlock for testing backward compatibility
	encryptedKeyBlock, err := x509.EncryptPEMBlock(rand.Reader, keyBlock.Type, keyBlock.Bytes, []byte(passphrase), x509.PEMCipherAES256)
	if err != nil {
		t.Fatalf("Failed to encrypt private key: %v", err)
	}
	clientCert.keyPEM = pem.EncodeToMemory(encryptedKeyBlock)

	caCertFile, clientCertFile, clientKeyFile := writeCertificateFiles(t, ca, clientCert)

	t.Run("encrypted_key_with_correct_passphrase", func(t *testing.T) {
		t.Parallel()
		testServer := newMTLSTestServer(t, ca)
		defer testServer.Close()
		config := &Config{
			ClientCertificate:          clientCertFile,
			ClientPrivateKey:           clientKeyFile,
			ClientPrivateKeyPassphrase: passphrase,
			CustomCABundle:             caCertFile,
		}

		awsConfig := &aws.Config{
			HTTPClient: &http.Client{
				Transport: &http.Transport{},
			},
		}

		err := config.configureHTTPClientMTLS(awsConfig)
		if err != nil {
			t.Fatalf("Failed to configure mTLS with encrypted key: %v", err)
		}

		req, err := http.NewRequest(http.MethodGet, testServer.url+"/health", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		httpClient, ok := awsConfig.HTTPClient.(*http.Client)
		if !ok {
			t.Fatalf("Expected HTTP client to be *http.Client, got %T", awsConfig.HTTPClient)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make mTLS request with encrypted key: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("encrypted_key_with_wrong_passphrase", func(t *testing.T) {
		t.Parallel()
		testServer := newMTLSTestServer(t, ca)
		defer testServer.Close()
		config := &Config{
			ClientCertificate:          clientCertFile,
			ClientPrivateKey:           clientKeyFile,
			ClientPrivateKeyPassphrase: "wrong-passphrase",
			CustomCABundle:             caCertFile,
		}

		awsConfig := &aws.Config{
			HTTPClient: &http.Client{
				Transport: &http.Transport{},
			},
		}

		err := config.configureHTTPClientMTLS(awsConfig)
		if err == nil {
			t.Fatal("Expected error with wrong passphrase, but configuration succeeded")
		}
	})
}
