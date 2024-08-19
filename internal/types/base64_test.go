// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import "testing"

func TestBase64EncodeOnce(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		blob []byte
		s    string
	}{
		{[]byte("data should be encoded"), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
		{[]byte("ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
	} {
		s := Base64EncodeOnce(tc.blob)
		if got, want := s, tc.s; got != want {
			t.Errorf("Base64EncodeOnce(%q) = %v, want %v", tc.blob, got, want)
		}
	}
}

func TestIsBase64Encoded(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		s     string
		valid bool
	}{
		{"ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA==", true},
		{"ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA==%%", false},
		{"123456789012", true},
		{"", true},
	} {
		ok := IsBase64Encoded(tc.s)
		if got, want := ok, tc.valid; got != want {
			t.Errorf("IsBase64Encoded(%q) = %v, want %v", tc.s, got, want)
		}
	}
}
