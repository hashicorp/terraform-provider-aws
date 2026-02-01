// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/youmark/pkcs8"
)

func loadClientCertificate(certFile, keyFile, passphrase string) (tls.Certificate, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to read client certificate file %q: %w", certFile, err)
	}

	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to read client private key file %q: %w", keyFile, err)
	}

	var cert tls.Certificate
	if passphrase != "" {
		cert, err = tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			keyBlock, rest := pem.Decode(keyPEM)
			if keyBlock == nil || len(rest) > 0 {
				return tls.Certificate{}, fmt.Errorf("failed to decode PEM block: %w", err)
			}
			var decryptedDER []byte
			var decryptErr error

			// PKCS#8 encrypted keys
			privateKey, pkcs8Err := pkcs8.ParsePKCS8PrivateKey(keyBlock.Bytes, []byte(passphrase))
			if pkcs8Err == nil {
				decryptedDER, decryptErr = pkcs8.ConvertPrivateKeyToPKCS8(privateKey)
				if decryptErr != nil {
					return tls.Certificate{}, fmt.Errorf("failed to convert private key to PKCS#8 format: %w", decryptErr)
				}
			} else {
				// PKCS#5 v1.0 encrypted keys - using deprecated functions for backward compatibility
				//nolint:staticcheck // SA1019: Using deprecated x509.IsEncryptedPEMBlock for backward compatibility
				if x509.IsEncryptedPEMBlock(keyBlock) {
					//nolint:staticcheck // SA1019: Using deprecated x509.DecryptPEMBlock for backward compatibility
					decryptedDER, decryptErr = x509.DecryptPEMBlock(keyBlock, []byte(passphrase))
					if decryptErr != nil {
						return tls.Certificate{}, fmt.Errorf(
							"failed to decrypt private key with provided passphrase: PKCS#8 error: %w, PKCS#5 error: %w", pkcs8Err, decryptErr,
						)
					}
				} else {
					return tls.Certificate{}, fmt.Errorf("decryption failed: %w", pkcs8Err)
				}
			}
			decryptedPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: decryptedDER})
			cert, err = tls.X509KeyPair(certPEM, decryptedPEM)
		}
	} else {
		cert, err = tls.X509KeyPair(certPEM, keyPEM)
	}

	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse client certificate and key: %w", err)
	}

	return cert, nil
}

func (c *Config) validateMTLSConfig() error {
	if c.ClientCertificate == "" {
		return nil
	}
	if c.ClientPrivateKey == "" {
		return fmt.Errorf("client_private_key must be provided when client_certificate is set")
	}
	if _, err := os.Stat(c.ClientCertificate); err != nil {
		return fmt.Errorf("client certificate file %q: %w", c.ClientCertificate, err)
	}
	if _, err := os.Stat(c.ClientPrivateKey); err != nil {
		return fmt.Errorf("client private key file %q: %w", c.ClientPrivateKey, err)
	}
	return nil
}

func (c *Config) configureHTTPClientMTLS(awsConfig *aws.Config) error {
	cert, err := loadClientCertificate(c.ClientCertificate, c.ClientPrivateKey, c.ClientPrivateKeyPassphrase)
	if err != nil {
		return err
	}

	var caCertPool *x509.CertPool
	if c.CustomCABundle != "" {
		caCertPool, err = c.loadCustomCABundle()
		if err != nil {
			return err
		}
	}

	sdkHTTPClient := awsConfig.HTTPClient
	if sdkHTTPClient == nil {
		return fmt.Errorf("SDK HTTP client is unexpectedly nil")
	}

	switch client := sdkHTTPClient.(type) {
	case *http.Client:
		err := c.configureStandardHTTPClient(client, cert, caCertPool)
		if err != nil {
			return err
		}
	case *awshttp.BuildableClient:
		err := c.configureBuildableHTTPClient(awsConfig, client, cert, caCertPool)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported SDK HTTP client type: %T", sdkHTTPClient)
	}

	return nil
}

func (c *Config) configureStandardHTTPClient(httpClient *http.Client, cert tls.Certificate, caCertPool *x509.CertPool) error {
	var transport *http.Transport
	if httpClient.Transport != nil {
		if t, ok := httpClient.Transport.(*http.Transport); ok {
			transport = t.Clone()
		} else {
			return fmt.Errorf("HTTP transport is not a standard *http.Transport, got type %T", httpClient.Transport)
		}
	} else {
		return fmt.Errorf("HTTP client has no transport configured")
	}

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}

	if transport.TLSClientConfig.Certificates == nil {
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	} else {
		transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, cert)
	}

	if caCertPool != nil {
		transport.TLSClientConfig.RootCAs = caCertPool
	}

	httpClient.Transport = transport
	return nil
}

func (c *Config) configureBuildableHTTPClient(awsConfig *aws.Config, buildableClient *awshttp.BuildableClient, cert tls.Certificate, caCertPool *x509.CertPool) error {
	transportOption := func(transport *http.Transport) {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}

		if transport.TLSClientConfig.Certificates == nil {
			transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
		} else {
			transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, cert)
		}

		if caCertPool != nil {
			transport.TLSClientConfig.RootCAs = caCertPool
		}
	}

	awsConfig.HTTPClient = buildableClient.WithTransportOptions(transportOption)
	return nil
}

func (c *Config) loadCustomCABundle() (*x509.CertPool, error) {
	caBundleData, err := os.ReadFile(c.CustomCABundle)
	if err != nil {
		return nil, fmt.Errorf("failed to read custom CA bundle file %q: %w", c.CustomCABundle, err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caBundleData) {
		return nil, fmt.Errorf("failed to parse certificates from custom CA bundle file %q", c.CustomCABundle)
	}

	return caCertPool, nil
}
