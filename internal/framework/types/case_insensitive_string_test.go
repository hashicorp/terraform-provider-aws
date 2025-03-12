// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestCaseInsensitiveStringSemanticEquals(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val1, val2 fwtypes.CaseInsensitiveString
		equals     bool
	}
	tests := map[string]testCase{
		"both lowercase, equal": {
			val1:   fwtypes.CaseInsensitiveStringValue("thursday"),
			val2:   fwtypes.CaseInsensitiveStringValue("thursday"),
			equals: true,
		},
		"both uppercase, equal": {
			val1:   fwtypes.CaseInsensitiveStringValue("THURSDAY"),
			val2:   fwtypes.CaseInsensitiveStringValue("THURSDAY"),
			equals: true,
		},
		"first uppercase, second lowercase, equal": {
			val1:   fwtypes.CaseInsensitiveStringValue("THURSDAY"),
			val2:   fwtypes.CaseInsensitiveStringValue("thursday"),
			equals: true,
		},
		"first lowercase, second uppercase, equal": {
			val1:   fwtypes.CaseInsensitiveStringValue("thursday"),
			val2:   fwtypes.CaseInsensitiveStringValue("THURSDAY"),
			equals: true,
		},
		"not equal": {
			val1:   fwtypes.CaseInsensitiveStringValue("Thursday"),
			val2:   fwtypes.CaseInsensitiveStringValue("Friday"),
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
