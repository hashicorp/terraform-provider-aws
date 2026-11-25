// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package conns_test

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"slices"
	"testing"
	"time"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"golang.org/x/crypto/cryptobyte"
)

// TLS protocol constants (IANA TLS parameter registries).
const (
	tlsRecordHeaderLength       = 5
	tlsRecordTypeHandshake      = 22
	tlsHandshakeTypeClientHello = 1
	tlsClientRandomLength       = 32

	tlsExtensionSupportedGroups         uint16 = 10 // RFC 8422, Section 5.1.1
	tlsExtensionSignatureAlgorithmsCert uint16 = 50 // RFC 8446, Section 4.2.3

	// ML-DSA signature schemes (draft-ietf-tls-mldsa). The tls.MLDSA* constants
	// only exist from Go 1.27, so the IANA codepoints are used directly.
	tlsSignatureSchemeMLDSA44 uint16 = 0x0904
	tlsSignatureSchemeMLDSA65 uint16 = 0x0905
	tlsSignatureSchemeMLDSA87 uint16 = 0x0906
)

const clientHelloCaptureTimeout = 10 * time.Second

// TestHTTPClientClientHelloMiddleboxCompatibility verifies that the TLS ClientHello sent
// by the HTTP client used for AWS API calls does not advertise post-quantum algorithms
// known to be dropped by TLS-inspecting middleboxes (AWS Network Firewall, corporate
// proxies and firewalls). Such drops surface as "net/http: TLS handshake timeout" errors
// while validating provider credentials against STS.
//
// The ClientHello content is determined by the Go toolchain the provider is built with
// and by the godebug directives in go.mod, so this test guards against regressions when
// either one changes:
//   - X25519MLKEM768 key exchange is disabled via "godebug tlsmlkem=0" (#39311, #41740).
//   - ML-DSA must not be advertised in signature_algorithms_cert. Go 1.27.0 through
//     1.27.3 advertise it; Go 1.26 and Go 1.27.4+ do not (https://go.dev/issue/82536).
func TestHTTPClientClientHelloMiddleboxCompatibility(t *testing.T) {
	t.Parallel()

	clientHello := captureClientHello(t, awshttp.NewBuildableClient())

	if len(clientHello.supportedGroups) == 0 {
		t.Fatal("ClientHello has no supported_groups extension; the capture or the parser is broken")
	}

	if slices.Contains(clientHello.supportedGroups, uint16(tls.X25519MLKEM768)) {
		t.Errorf("ClientHello supported_groups advertises X25519MLKEM768 (0x%04x), which AWS Network Firewall is known to drop; keep \"godebug tlsmlkem=0\" in go.mod", uint16(tls.X25519MLKEM768))
	}

	for _, scheme := range clientHello.signatureAlgorithmsCert {
		switch scheme {
		case tlsSignatureSchemeMLDSA44, tlsSignatureSchemeMLDSA65, tlsSignatureSchemeMLDSA87:
			t.Errorf("ClientHello signature_algorithms_cert advertises ML-DSA (0x%04x), which TLS-inspecting middleboxes are known to drop (https://go.dev/issue/82536); build with Go 1.26 or Go 1.27.4+", scheme)
		}
	}
}

type clientHello struct {
	supportedGroups         []uint16
	signatureAlgorithmsCert []uint16
}

// captureClientHello performs a TLS handshake attempt from client against a local
// listener that records the first TLS record and then closes the connection.
func captureClientHello(t *testing.T, client interface {
	Do(*http.Request) (*http.Response, error)
}) clientHello {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listening: %s", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Errorf("closing listener: %s", err)
		}
	})

	records := make(chan []byte, 1)
	errs := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errs <- err
			return
		}
		defer conn.Close() //nolint:errcheck // The connection is intentionally abandoned once the ClientHello has been captured.

		if err := conn.SetReadDeadline(time.Now().Add(clientHelloCaptureTimeout)); err != nil {
			errs <- err
			return
		}

		record, err := readTLSRecord(conn)
		if err != nil {
			errs <- err
			return
		}
		records <- record
	}()

	ctx, cancel := context.WithTimeout(t.Context(), clientHelloCaptureTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+listener.Addr().String(), nil)
	if err != nil {
		t.Fatalf("creating request: %s", err)
	}

	// The listener never answers the handshake, so the request is expected to fail.
	if response, err := client.Do(request); err == nil {
		response.Body.Close()
		t.Fatal("expected the TLS handshake against the capturing listener to fail")
	}

	select {
	case record := <-records:
		return parseClientHello(t, record)
	case err := <-errs:
		t.Fatalf("capturing ClientHello: %s", err)
	case <-ctx.Done():
		t.Fatalf("capturing ClientHello: %s", ctx.Err())
	}

	return clientHello{}
}

// readTLSRecord reads a single TLS record and returns its fragment (RFC 8446, Section 5.1).
func readTLSRecord(r io.Reader) ([]byte, error) {
	header := make([]byte, tlsRecordHeaderLength)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	if header[0] != tlsRecordTypeHandshake {
		return nil, &net.OpError{Op: "read", Err: io.ErrUnexpectedEOF}
	}

	fragment := make([]byte, binary.BigEndian.Uint16(header[3:5]))
	if _, err := io.ReadFull(r, fragment); err != nil {
		return nil, err
	}

	return fragment, nil
}

// parseClientHello extracts the extensions of interest from a ClientHello handshake
// message (RFC 8446, Section 4.1.2).
func parseClientHello(t *testing.T, message []byte) clientHello {
	t.Helper()

	var (
		s                 = cryptobyte.String(message)
		handshakeType     uint8
		body              cryptobyte.String
		legacyVersion     uint16
		random            []byte
		sessionID         cryptobyte.String
		cipherSuites      cryptobyte.String
		compression       cryptobyte.String
		extensions        cryptobyte.String
		extensionType     uint16
		extensionData     cryptobyte.String
		parsedClientHello clientHello
	)

	if !s.ReadUint8(&handshakeType) || handshakeType != tlsHandshakeTypeClientHello || !s.ReadUint24LengthPrefixed(&body) {
		t.Fatal("parsing ClientHello: not a ClientHello handshake message")
	}

	if !body.ReadUint16(&legacyVersion) ||
		!body.ReadBytes(&random, tlsClientRandomLength) ||
		!body.ReadUint8LengthPrefixed(&sessionID) ||
		!body.ReadUint16LengthPrefixed(&cipherSuites) ||
		!body.ReadUint8LengthPrefixed(&compression) ||
		!body.ReadUint16LengthPrefixed(&extensions) {
		t.Fatal("parsing ClientHello: malformed message body")
	}

	for !extensions.Empty() {
		if !extensions.ReadUint16(&extensionType) || !extensions.ReadUint16LengthPrefixed(&extensionData) {
			t.Fatal("parsing ClientHello: malformed extension")
		}

		switch extensionType {
		case tlsExtensionSupportedGroups:
			parsedClientHello.supportedGroups = readUint16List(t, extensionData)
		case tlsExtensionSignatureAlgorithmsCert:
			parsedClientHello.signatureAlgorithmsCert = readUint16List(t, extensionData)
		}
	}

	return parsedClientHello
}

// readUint16List reads a 16-bit length-prefixed list of 16-bit values.
func readUint16List(t *testing.T, s cryptobyte.String) []uint16 {
	t.Helper()

	var list cryptobyte.String
	if !s.ReadUint16LengthPrefixed(&list) {
		t.Fatal("parsing ClientHello: malformed list")
	}

	var values []uint16
	for !list.Empty() {
		var value uint16
		if !list.ReadUint16(&value) {
			t.Fatal("parsing ClientHello: malformed list element")
		}
		values = append(values, value)
	}

	return values
}
