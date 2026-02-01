// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/youmark/pkcs8"
)

type testCA struct {
	cert    *x509.Certificate
	key     *rsa.PrivateKey
	certPEM []byte
}

type testClientCert struct {
	certPEM []byte
	keyPEM  []byte
	tlsCert tls.Certificate
}

func createTestCA(t *testing.T) *testCA {
	t.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate CA private key: %v", err)
	}

	// CA certificate template
	caTemplate := &x509.Certificate{
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

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("Failed to create CA certificate: %v", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		t.Fatalf("Failed to parse CA certificate: %v", err)
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertDER,
	})

	return &testCA{
		cert:    caCert,
		key:     caKey,
		certPEM: caCertPEM,
	}
}

func (ca *testCA) createClientCert(t *testing.T, encrypted bool, passphrase string) *testClientCert {
	t.Helper()

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate client private key: %v", err)
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:  []string{"Test Client"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"San Francisco"},
			CommonName:    "test-client",
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

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, ca.cert, &clientKey.PublicKey, ca.key)
	if err != nil {
		t.Fatalf("Failed to create client certificate: %v", err)
	}

	clientCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertDER,
	})

	clientKeyDER, err := x509.MarshalPKCS8PrivateKey(clientKey)
	if err != nil {
		t.Fatalf("Failed to marshal client private key: %v", err)
	}

	var clientKeyPEM []byte
	if encrypted && passphrase != "" {
		//nolint:staticcheck // SA1019: Using deprecated x509.EncryptPEMBlock for testing backward compatibility
		encryptedBlock, err := x509.EncryptPEMBlock(rand.Reader, "PRIVATE KEY", clientKeyDER, []byte(passphrase), x509.PEMCipherAES256)
		if err != nil {
			t.Fatalf("Failed to encrypt private key: %v", err)
		}
		clientKeyPEM = pem.EncodeToMemory(encryptedBlock)
	} else {
		clientKeyPEM = pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: clientKeyDER,
		})
	}

	var tlsCert tls.Certificate
	if !encrypted || passphrase == "" {
		tlsCert, err = tls.X509KeyPair(clientCertPEM, clientKeyPEM)
		if err != nil {
			t.Fatalf("Failed to create client TLS certificate: %v", err)
		}
	}

	return &testClientCert{
		certPEM: clientCertPEM,
		keyPEM:  clientKeyPEM,
		tlsCert: tlsCert,
	}
}

func (ca *testCA) createClientCertPKCS8(t *testing.T, passphrase string) *testClientCert {
	t.Helper()

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate client private key: %v", err)
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:  []string{"Test Client PKCS8"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"San Francisco"},
			CommonName:    "test-client-pkcs8",
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

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, ca.cert, &clientKey.PublicKey, ca.key)
	if err != nil {
		t.Fatalf("Failed to create client certificate: %v", err)
	}

	clientCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertDER,
	})

	opts := &pkcs8.Opts{
		Cipher: pkcs8.AES256CBC,
		KDFOpts: pkcs8.PBKDF2Opts{
			SaltSize:       16,
			IterationCount: 10000,
			HMACHash:       crypto.SHA256,
		},
	}

	clientKeyDER, err := pkcs8.MarshalPrivateKey(clientKey, []byte(passphrase), opts)
	if err != nil {
		t.Fatalf("Failed to encrypt private key with PKCS#8: %v", err)
	}

	clientKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "ENCRYPTED PRIVATE KEY",
		Bytes: clientKeyDER,
	})

	return &testClientCert{
		certPEM: clientCertPEM,
		keyPEM:  clientKeyPEM,
		tlsCert: tls.Certificate{},
	}
}

func (ca *testCA) createServerCert(t *testing.T) *testClientCert {
	t.Helper()

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate server private key: %v", err)
	}

	serverTemplate := &x509.Certificate{
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

	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, ca.cert, &serverKey.PublicKey, ca.key)
	if err != nil {
		t.Fatalf("Failed to create server certificate: %v", err)
	}

	serverCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertDER,
	})

	serverKeyDER, err := x509.MarshalPKCS8PrivateKey(serverKey)
	if err != nil {
		t.Fatalf("Failed to marshal server private key: %v", err)
	}

	serverKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: serverKeyDER,
	})

	tlsCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatalf("Failed to create server TLS certificate: %v", err)
	}

	return &testClientCert{
		certPEM: serverCertPEM,
		keyPEM:  serverKeyPEM,
		tlsCert: tlsCert,
	}
}

func (ca *testCA) certPool() *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(ca.cert)
	return pool
}

func writeTestCertFiles(t *testing.T, ca *testCA, clientCert *testClientCert) (caCertFile, clientCertFile, clientKeyFile string) {
	t.Helper()

	tempDir := t.TempDir()

	caCertFile = filepath.Join(tempDir, "ca.pem")
	if err := os.WriteFile(caCertFile, ca.certPEM, 0644); err != nil {
		t.Fatalf("Failed to write CA certificate file: %v", err)
	}

	clientCertFile = filepath.Join(tempDir, "client.pem")
	if err := os.WriteFile(clientCertFile, clientCert.certPEM, 0644); err != nil {
		t.Fatalf("Failed to write client certificate file: %v", err)
	}

	clientKeyFile = filepath.Join(tempDir, "client-key.pem")
	if err := os.WriteFile(clientKeyFile, clientCert.keyPEM, 0600); err != nil {
		t.Fatalf("Failed to write client private key file: %v", err)
	}

	return caCertFile, clientCertFile, clientKeyFile
}

type MockDynamoDBMTLSServer struct {
	server *httptest.Server
	url    string
	mux    *http.ServeMux
}

func newMockDynamoDBMTLSServer(t *testing.T, ca *testCA) *MockDynamoDBMTLSServer {
	t.Helper()

	serverCert := ca.createServerCert(t)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-Amz-Target")
		if target == "" {
			http.Error(w, "Missing X-Amz-Target header", http.StatusBadRequest)
			return
		}

		switch target {
		case "DynamoDB_20120810.ListTables":
			handleMockListTablesAcc(t, w)
		case "DynamoDB_20120810.DescribeTable":
			handleMockDescribeTableAcc(t, w, r)
		default:
			t.Errorf("Mock DynamoDB Server - Unsupported operation: %s", target)
			w.Header().Set("Content-Type", "application/x-amz-json-1.0")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("{}")); err != nil {
				t.Fatalf("failed to write empty JSON: %v", err)
			}
		}
	})

	server := httptest.NewUnstartedServer(mux)

	caCertPool := ca.certPool()
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverCert.tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	server.StartTLS()

	return &MockDynamoDBMTLSServer{
		server: server,
		url:    server.URL,
		mux:    mux,
	}
}

func (s *MockDynamoDBMTLSServer) Close() {
	if s.server != nil {
		s.server.Close()
	}
}

func handleMockListTablesAcc(t *testing.T, w http.ResponseWriter) {
	t.Helper()

	response := map[string]any{
		"TableNames": []string{
			"terraform-mtls-test-table-1",
			"terraform-mtls-test-table-2",
		},
	}

	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		t.Errorf("Mock DynamoDB Server - Failed to encode ListTables response: %v", err)
	}
}

func handleMockDescribeTableAcc(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Mock DynamoDB Server - Failed to read DescribeTable request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		t.Errorf("Mock DynamoDB Server - Failed to parse DescribeTable JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tableName, ok := request["TableName"].(string)
	if !ok {
		tableName = "terraform-mtls-test-table"
	}

	response := map[string]any{
		"Table": map[string]any{
			"TableName":   tableName,
			"TableStatus": "ACTIVE",
			"AttributeDefinitions": []map[string]any{
				{
					"AttributeName": "id",
					"AttributeType": "S",
				},
			},
			"KeySchema": []map[string]any{
				{
					"AttributeName": "id",
					"KeyType":       "HASH",
				},
			},
			"ProvisionedThroughput": map[string]any{
				"ReadCapacityUnits":  int64(5),
				"WriteCapacityUnits": int64(5),
			},
		},
	}

	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		t.Errorf("Mock DynamoDB Server - Failed to encode DescribeTable response: %v", err)
	}
}

func TestAccProvider_DynamoDBMTLS(t *testing.T) {
	ca := createTestCA(t)
	clientCert := ca.createClientCert(t, false, "")

	caCertFile, clientCertFile, clientKeyFile := writeTestCertFiles(t, ca, clientCert)

	mockServer := newMockDynamoDBMTLSServer(t, ca)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderDynamoDBMTLSConfig(mockServer.url, clientCertFile, clientKeyFile, caCertFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.#", "2"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.0", "terraform-mtls-test-table-1"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.1", "terraform-mtls-test-table-2"),
				),
			},
		},
	})
}

func testAccProviderDynamoDBMTLSConfig(dynamodbEndpoint, clientCert, clientKey, caCert string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  
  max_retries = 1

  client_certificate                = %[2]q
  client_private_key               = %[3]q
  client_private_key_passphrase    = ""
  custom_ca_bundle                 = %[4]q
  
  endpoints {
    dynamodb = %[1]q
  }
}

data "aws_dynamodb_tables" "test" {
  # This will call ListTables on our mock mTLS server
}

output "table_names" {
  value = data.aws_dynamodb_tables.test.names
}
`, dynamodbEndpoint, clientCert, clientKey, caCert)
}

func TestAccProvider_DynamoDBMTLSFailsWithoutCert(t *testing.T) {
	ca := createTestCA(t)
	clientCert := ca.createClientCert(t, false, "")

	caCertFile, _, _ := writeTestCertFiles(t, ca, clientCert)

	mockServer := newMockDynamoDBMTLSServer(t, ca)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderDynamoDBMTLSConfigNoCert(mockServer.url, caCertFile),
				ExpectError: regexache.MustCompile("tls|certificate"),
			},
		},
	})
}

func testAccProviderDynamoDBMTLSConfigNoCert(dynamodbEndpoint, caCert string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  
  max_retries = 1
  
  custom_ca_bundle = %[2]q
  
  endpoints {
    dynamodb = %[1]q
  }
}

data "aws_dynamodb_tables" "test" {
}

output "table_names" {
  value = data.aws_dynamodb_tables.test.names
}

`, dynamodbEndpoint, caCert)
}

func TestAccProvider_DynamoDBMTLSWithEncryptedKey(t *testing.T) {
	ca := createTestCA(t)
	passphrase := "test-passphrase-123"
	clientCert := ca.createClientCert(t, true, passphrase)

	caCertFile, clientCertFile, clientKeyFile := writeTestCertFiles(t, ca, clientCert)

	mockServer := newMockDynamoDBMTLSServer(t, ca)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderDynamoDBMTLSConfigWithPassphrase(mockServer.url, clientCertFile, clientKeyFile, passphrase, caCertFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.#", "2"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.0", "terraform-mtls-test-table-1"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.1", "terraform-mtls-test-table-2"),
				),
			},
		},
	})
}

func TestAccProvider_DynamoDBMTLSWithPKCS8EncryptedKey(t *testing.T) {
	ca := createTestCA(t)
	passphrase := "pkcs8-test-passphrase-456"
	clientCert := ca.createClientCertPKCS8(t, passphrase)

	caCertFile, clientCertFile, clientKeyFile := writeTestCertFiles(t, ca, clientCert)

	mockServer := newMockDynamoDBMTLSServer(t, ca)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderDynamoDBMTLSConfigWithPassphrase(mockServer.url, clientCertFile, clientKeyFile, passphrase, caCertFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.#", "2"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.0", "terraform-mtls-test-table-1"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.1", "terraform-mtls-test-table-2"),
				),
			},
		},
	})
}

func TestAccProvider_DynamoDBMTLSFailsWithoutCABundle(t *testing.T) {
	ca := createTestCA(t)
	clientCert := ca.createClientCert(t, false, "")

	_, clientCertFile, clientKeyFile := writeTestCertFiles(t, ca, clientCert)

	mockServer := newMockDynamoDBMTLSServer(t, ca)
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderDynamoDBMTLSConfigNoCA(mockServer.url, clientCertFile, clientKeyFile),
				ExpectError: regexache.MustCompile("tls|certificate|x509"),
			},
		},
	})
}

func testAccProviderDynamoDBMTLSConfigWithPassphrase(dynamodbEndpoint, clientCert, clientKey, passphrase, caCert string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  
  max_retries = 1

  client_certificate                = %[2]q
  client_private_key               = %[3]q
  client_private_key_passphrase    = %[4]q
  custom_ca_bundle                 = %[5]q
  
  endpoints {
    dynamodb = %[1]q
  }
}

data "aws_dynamodb_tables" "test" {
  # This will call ListTables on our mock mTLS server with encrypted key
}

output "table_names" {
  value = data.aws_dynamodb_tables.test.names
}
`, dynamodbEndpoint, clientCert, clientKey, passphrase, caCert)
}

func TestAccProvider_DynamoDBMTLSWithEncryptedKeyEnvVars(t *testing.T) {
	ca := createTestCA(t)
	passphrase := "test_120jvb!9_-passphrase#@"
	clientCert := ca.createClientCert(t, true, passphrase)

	caCertFile, clientCertFile, clientKeyFile := writeTestCertFiles(t, ca, clientCert)

	mockServer := newMockDynamoDBMTLSServer(t, ca)
	defer mockServer.Close()

	t.Setenv("TF_AWS_CLIENT_CERTIFICATE_PATH", clientCertFile)
	t.Setenv("TF_AWS_CLIENT_PRIVATE_KEY_PATH", clientKeyFile)
	t.Setenv("TF_AWS_CLIENT_PRIVATE_KEY_PASSPHRASE", passphrase)
	t.Setenv("AWS_CA_BUNDLE", caCertFile)

	resource.Test(t, resource.TestCase{
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderDynamoDBMTLSConfigEnvVars(mockServer.url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.#", "2"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.0", "terraform-mtls-test-table-1"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_tables.test", "names.1", "terraform-mtls-test-table-2"),
				),
			},
		},
	})
}

func testAccProviderDynamoDBMTLSConfigNoCA(dynamodbEndpoint, clientCert, clientKey string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  
  max_retries = 1

  client_certificate = %[2]q
  client_private_key = %[3]q
  
  endpoints {
    dynamodb = %[1]q
  }
}

data "aws_dynamodb_tables" "test" {
}
`, dynamodbEndpoint, clientCert, clientKey)
}

func testAccProviderDynamoDBMTLSConfigEnvVars(dynamodbEndpoint string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  
  max_retries = 1

  endpoints {
    dynamodb = %[1]q
  }
}

data "aws_dynamodb_tables" "test" {
  # This will call ListTables on our mock mTLS server
}

output "table_names" {
  value = data.aws_dynamodb_tables.test.names
}
`, dynamodbEndpoint)
}
