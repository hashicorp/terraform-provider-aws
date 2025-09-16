// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
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

			decryptedDER, decryptErr := x509.DecryptPEMBlock(keyBlock, []byte(passphrase))
			if decryptErr != nil {
				return tls.Certificate{}, fmt.Errorf("failed to decrypt private key with provided passphrase: %w", decryptErr)
			}

			decryptedPEM := pem.EncodeToMemory(&pem.Block{Type: keyBlock.Type, Bytes: decryptedDER})
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

func (c *Config) configureHTTPClientMTLS(awsConfig *aws.Config, ctx context.Context) error {
	cert, err := loadClientCertificate(c.ClientCertificate, c.ClientPrivateKey, c.ClientPrivateKeyPassphrase)
	if err != nil {
		return err
	}

	sdkHTTPClient := awsConfig.HTTPClient
	if sdkHTTPClient == nil {
		return fmt.Errorf("SDK HTTP client is unexpectedly nil")
	}

	switch client := sdkHTTPClient.(type) {
	case *http.Client:
		err := c.configureStandardHTTPClient(client, cert, ctx)
		if err != nil {
			return err
		}
	case *awshttp.BuildableClient:
		err := c.configureBuildableHTTPClient(awsConfig, client, cert, ctx)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported SDK HTTP client type: %T", sdkHTTPClient)
	}

	return nil
}

func (c *Config) configureStandardHTTPClient(httpClient *http.Client, cert tls.Certificate, ctx context.Context) error {
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

	if c.CustomCABundle != "" {
		err := c.configureCustomCABundle(transport.TLSClientConfig)
		if err != nil {
			return fmt.Errorf("failed to configure custom CA bundle: %w", err)
		}
	}

	httpClient.Transport = transport
	return nil
}

func (c *Config) configureBuildableHTTPClient(awsConfig *aws.Config, buildableClient *awshttp.BuildableClient, cert tls.Certificate, ctx context.Context) error {
	baseTransport := buildableClient.GetTransport()

	if baseTransport.TLSClientConfig == nil {
		baseTransport.TLSClientConfig = &tls.Config{}
	}

	if baseTransport.TLSClientConfig.Certificates == nil {
		baseTransport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	} else {
		baseTransport.TLSClientConfig.Certificates = append(baseTransport.TLSClientConfig.Certificates, cert)
	}

	if c.CustomCABundle != "" {
		err := c.configureCustomCABundle(baseTransport.TLSClientConfig)
		if err != nil {
			return fmt.Errorf("failed to configure custom CA bundle: %w", err)
		}
	}

	// Create a new standard http.Client with the configured transport
	// This ensures the mTLS configuration is applied directly to the HTTP client
	httpClient := &http.Client{
		Transport: baseTransport,
		Timeout:   buildableClient.GetTimeout(),
	}

	awsConfig.HTTPClient = httpClient
	return nil
}

func (c *Config) configureCustomCABundle(tlsConfig *tls.Config) error {
	caBundleData, err := os.ReadFile(c.CustomCABundle)
	if err != nil {
		return fmt.Errorf("failed to read custom CA bundle file %q: %w", c.CustomCABundle, err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caBundleData) {
		return fmt.Errorf("failed to parse certificates from custom CA bundle file %q", c.CustomCABundle)
	}

	tlsConfig.RootCAs = caCertPool
	return nil
}
