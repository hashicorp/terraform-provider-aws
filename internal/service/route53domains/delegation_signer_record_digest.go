// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"crypto/sha1" // nosemgrep: go/sast/internal/crypto/sha1 -- DS digest type 1 is SHA-1 by specification (RFC 4034), used to match registry-returned digests, not for security
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"
)

// DS record digest types (IANA "DNSSEC Delegation Signer (DS) Resource Record (RR) Type Digest Algorithms").
const (
	dsDigestTypeSHA1   int32 = 1
	dsDigestTypeSHA256 int32 = 2
	dsDigestTypeSHA384 int32 = 4
)

// dnskeyProtocol is the fixed value of the DNSKEY RDATA Protocol field (RFC 4034, section 2.1.2).
const dnskeyProtocol = 3

// dsDigests computes the Delegation Signer (DS) record digests of a DNSKEY for every
// digest type that a registry may use, keyed by digest type.
//
// Per RFC 4034 section 5.1.4, the digest is calculated over the canonical wire format of
// the DNSKEY owner name concatenated with the DNSKEY RDATA
// (Flags | Protocol | Algorithm | Public Key). The result is upper-case hexadecimal,
// which is how Route 53 Domains returns `DnssecKey.Digest`.
func dsDigests(domainName string, flags, algorithm int64, publicKey string) (map[int32]string, error) {
	if flags < 0 || flags > 0xFFFF {
		return nil, fmt.Errorf("invalid DNSKEY flags: %d", flags)
	}
	if algorithm < 0 || algorithm > 0xFF {
		return nil, fmt.Errorf("invalid DNSKEY algorithm: %d", algorithm)
	}

	key, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(publicKey), ""))
	if err != nil {
		return nil, fmt.Errorf("decoding DNSKEY public key: %w", err)
	}

	owner, err := canonicalOwnerNameWireFormat(domainName)
	if err != nil {
		return nil, err
	}

	// Owner name in canonical wire format followed by the DNSKEY RDATA.
	data := make([]byte, 0, len(owner)+4+len(key))
	data = append(data, owner...)
	data = binary.BigEndian.AppendUint16(data, uint16(flags))
	data = append(data, dnskeyProtocol, byte(algorithm))
	data = append(data, key...)

	digests := make(map[int32]string, 3)
	for digestType, newHash := range map[int32]func() hash.Hash{
		dsDigestTypeSHA1:   sha1.New, // nosemgrep: go.lang.security.audit.crypto.use_of_weak_crypto.use-of-sha1 -- DS digest type 1 is SHA-1 by specification (RFC 4034)
		dsDigestTypeSHA256: sha256.New,
		dsDigestTypeSHA384: sha512.New384,
	} {
		h := newHash()
		h.Write(data)
		digests[digestType] = strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	}

	return digests, nil
}

// canonicalOwnerNameWireFormat returns the canonical (lower-case, uncompressed) wire
// format of a DNS owner name (RFC 4034, section 6.2).
func canonicalOwnerNameWireFormat(name string) ([]byte, error) {
	name = strings.ToLower(strings.TrimSuffix(name, "."))

	var out []byte
	if name != "" {
		for label := range strings.SplitSeq(name, ".") {
			if len(label) == 0 || len(label) > 63 {
				return nil, fmt.Errorf("invalid DNS label %q in domain name %q", label, name)
			}
			out = append(out, byte(len(label)))
			out = append(out, label...)
		}
	}
	out = append(out, 0) // root label

	if len(out) > 255 {
		return nil, fmt.Errorf("domain name %q exceeds 255 octets in wire format", name)
	}

	return out, nil
}

// dsDigestMatches reports whether any of the computed digests equals digest.
func dsDigestMatches(digests map[int32]string, digest string) bool {
	for _, v := range digests {
		if strings.EqualFold(v, digest) {
			return true
		}
	}
	return false
}
