// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53domains

import "testing"

func TestDigestFromLegacyDNSSECKeyID(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in   string
		want string
	}{
		"legacy DS form is reduced to digest": {
			in:   "DS:12345-13-2-abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			want: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		"already-digest is unchanged (idempotent)": {
			in:   "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			want: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		"empty is unchanged": {
			in:   "",
			want: "",
		},
		"malformed DS prefix with too few segments is left alone": {
			in:   "DS:12345-13",
			want: "DS:12345-13",
		},
		"digest containing hyphen-like chars from a longer split keeps trailing segment": {
			in:   "DS:1-2-3-DEADBEEF",
			want: "DEADBEEF",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := digestFromLegacyDNSSECKeyID(tc.in)
			if got != tc.want {
				t.Fatalf("digestFromLegacyDNSSECKeyID(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
