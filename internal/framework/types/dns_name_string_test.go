// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestDNSNameStringSemanticEquals(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val1, val2 fwtypes.DNSNameString
		equals     bool
	}
	tests := map[string]testCase{
		"no trailing dot, equal": {
			val1:   fwtypes.DNSNameStringValue("www.example.com"),
			val2:   fwtypes.DNSNameStringValue("www.example.com"),
			equals: true,
		},
		"both trailing dot, equal": {
			val1:   fwtypes.DNSNameStringValue("www.example.com."),
			val2:   fwtypes.DNSNameStringValue("www.example.com."),
			equals: true,
		},
		"first trailing dot, equal": {
			val1:   fwtypes.DNSNameStringValue("www.example.com."),
			val2:   fwtypes.DNSNameStringValue("www.example.com"),
			equals: true,
		},
		"second trailing dot, equal": {
			val1:   fwtypes.DNSNameStringValue("www.example.com"),
			val2:   fwtypes.DNSNameStringValue("www.example.com."),
			equals: true,
		},
		"first upper, equal": {
			val1:   fwtypes.DNSNameStringValue("WWW.EXAMPLE.COM"),
			val2:   fwtypes.DNSNameStringValue("www.example.com"),
			equals: true,
		},
		"second upper, equal": {
			val1:   fwtypes.DNSNameStringValue("www.example.com"),
			val2:   fwtypes.DNSNameStringValue("WWW.EXAMPLE.COM"),
			equals: true,
		},
		"both upper, equal": {
			val1:   fwtypes.DNSNameStringValue("WWW.EXAMPLE.COM"),
			val2:   fwtypes.DNSNameStringValue("WWW.EXAMPLE.COM"),
			equals: true,
		},
		"not equal": {
			val1:   fwtypes.DNSNameStringValue("www.example.com"),
			val2:   fwtypes.DNSNameStringValue("www.otherexample.com"),
			equals: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			equals, _ := test.val1.StringSemanticEquals(ctx, test.val2)
			if got, want := equals, test.equals; got != want {
				t.Errorf("StringSemanticEquals(%q, %q) = %v, want %v", test.val1, test.val2, got, want)
			}
		})
	}
}
