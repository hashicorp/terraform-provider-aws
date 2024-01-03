// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"testing"
)

func TestBase64Encode(t *testing.T) {
	t.Parallel()

	for _, tt := range base64encodingTests {
		out := Base64Encode(tt.in)
		if out != tt.out {
			t.Errorf("Base64Encode(%s) => %s, want %s", tt.in, out, tt.out)
		}
	}
}

var base64encodingTests = []struct {
	in  []byte
	out string
}{
	// normal encoding case
	{[]byte("data should be encoded"), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
	// base64 encoded input should result in no change of output
	{[]byte("ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
}
