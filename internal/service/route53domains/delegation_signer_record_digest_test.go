// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"strings"
	"testing"
)

func TestDSDigests(t *testing.T) {
	t.Parallel()

	// Public KSK shared by all Cloudflare-signed zones; the expected digests are the
	// values returned by Route 53 Domains GetDomainDetail / published in the parent zones.
	const cloudflareKSK = "mdsswUyr3DPW132mOi8V9xESWE8jTo0dxCjjnopKl+GqJxpVXckHAeF+KkxLbxILfDLUT0rAK9iUzy1L53eKGQ=="

	testCases := map[string]struct {
		domainName string
		flags      int64
		algorithm  int64
		publicKey  string
		digestType int32
		want       string
		wantErr    bool
	}{
		// RFC 4034, section 5.4 example (SHA-1).
		"rfc4034 example sha1": {
			domainName: "dskey.example.com.",
			flags:      256,
			algorithm:  5,
			publicKey: "AQOeiiR0GOMYkDshWoSKz9Xz" +
				"fwJr1AYtsmx3TGkJaNXVbfi/" +
				"2pHm822aJ5iI9BMzNXxeYCmZ" +
				"DRD99WYwYqUSdjMmmAphXdvx" +
				"egXd/M5+X7OrzKBaMbCVdFLU" +
				"Uh6DhweJBjEVv5f2wwjM9Xzc" +
				"nOf+EPbtG9DMBmADjFDc2w/r" +
				"ljwvFw==",
			digestType: dsDigestTypeSHA1,
			want:       "2BB183AF5F22588179A53B0A98631FAD1A292118",
		},
		"gandi .com sha256": {
			domainName: "urbantz.com",
			flags:      257,
			algorithm:  13,
			publicKey:  cloudflareKSK,
			digestType: dsDigestTypeSHA256,
			want:       "ABA22F65F11E2D8D43BC95B32BCDB6744AA07C4CE38CE4D3E46189760DE699EB",
		},
		"amazon registrar .com sha256": {
			domainName: "urbantz-legacy.com",
			flags:      257,
			algorithm:  13,
			publicKey:  cloudflareKSK,
			digestType: dsDigestTypeSHA256,
			want:       "B8EF62832AA5515654F042AD80B2B5D97EF4D29AA3520333FC4F350FAD01E0B4",
		},
		"gandi .eu sha256": {
			domainName: "urbz.eu",
			flags:      257,
			algorithm:  13,
			publicKey:  cloudflareKSK,
			digestType: dsDigestTypeSHA256,
			want:       "4B25E8FA94850AA97B78AB790CBE126A546305ECF3AD5C861A5366FD7FBEEDCB",
		},
		"owner name is canonicalized": {
			domainName: "UrbZ.EU.",
			flags:      257,
			algorithm:  13,
			publicKey:  cloudflareKSK,
			digestType: dsDigestTypeSHA256,
			want:       "4B25E8FA94850AA97B78AB790CBE126A546305ECF3AD5C861A5366FD7FBEEDCB",
		},
		"invalid public key": {
			domainName: "example.com",
			flags:      257,
			algorithm:  13,
			publicKey:  "not base64!",
			wantErr:    true,
		},
		"invalid flags": {
			domainName: "example.com",
			flags:      70000,
			algorithm:  13,
			publicKey:  cloudflareKSK,
			wantErr:    true,
		},
		"invalid domain name": {
			domainName: "example..com",
			flags:      257,
			algorithm:  13,
			publicKey:  cloudflareKSK,
			wantErr:    true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := dsDigests(testCase.domainName, testCase.flags, testCase.algorithm, testCase.publicKey)

			if testCase.wantErr {
				if err == nil {
					t.Fatal("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if got[testCase.digestType] != testCase.want {
				t.Errorf("digest type %d = %q, want %q", testCase.digestType, got[testCase.digestType], testCase.want)
			}
			if !dsDigestMatches(got, testCase.want) {
				t.Errorf("dsDigestMatches(%q) = false, want true", testCase.want)
			}
			// Matching must be case-insensitive and must fail for other values.
			if !dsDigestMatches(got, strings.ToLower(testCase.want)) {
				t.Errorf("dsDigestMatches(%q) = false, want true (case-insensitive)", strings.ToLower(testCase.want))
			}
			if dsDigestMatches(got, "0000") {
				t.Errorf("dsDigestMatches(%q) = true, want false", "0000")
			}
		})
	}
}
