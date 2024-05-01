// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestOnceAWeekWindowValidateAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         fwtypes.OnceAWeekWindow
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val: fwtypes.OnceAWeekWindowUnknown(),
		},
		"null": {
			val: fwtypes.OnceAWeekWindowNull(),
		},
		"valid lowercase": {
			val: fwtypes.OnceAWeekWindowValue("thu:07:44-thu:09:44"),
		},
		"valid uppercase": {
			val: fwtypes.OnceAWeekWindowValue("THU:07:44-THU:09:44"),
		},
		"invalid": {
			val:         fwtypes.OnceAWeekWindowValue("thu:25:44-zat:09:88"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			req := xattr.ValidateAttributeRequest{}
			resp := xattr.ValidateAttributeResponse{}

			test.val.ValidateAttribute(ctx, req, &resp)
			if resp.Diagnostics.HasError() != test.expectError {
				t.Errorf("resp.Diagnostics.HasError() = %t, want = %t", resp.Diagnostics.HasError(), test.expectError)
			}
		})
	}
}

func TestOnceAWeekWindowStringSemanticEquals(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val1, val2 fwtypes.OnceAWeekWindow
		equals     bool
	}
	tests := map[string]testCase{
		"both lowercase, equal": {
			val1:   fwtypes.OnceAWeekWindowValue("thu:07:44-thu:09:44"),
			val2:   fwtypes.OnceAWeekWindowValue("thu:07:44-thu:09:44"),
			equals: true,
		},
		"both uppercase, equal": {
			val1:   fwtypes.OnceAWeekWindowValue("THU:07:44-THU:09:44"),
			val2:   fwtypes.OnceAWeekWindowValue("THU:07:44-THU:09:44"),
			equals: true,
		},
		"first uppercase, second lowercase, equal": {
			val1:   fwtypes.OnceAWeekWindowValue("THU:07:44-THU:09:44"),
			val2:   fwtypes.OnceAWeekWindowValue("thu:07:44-thu:09:44"),
			equals: true,
		},
		"first lowercase, second uppercase, equal": {
			val1:   fwtypes.OnceAWeekWindowValue("thu:07:44-thu:09:44"),
			val2:   fwtypes.OnceAWeekWindowValue("THU:07:44-THU:09:44"),
			equals: true,
		},
		"not equal": {
			val1:   fwtypes.OnceAWeekWindowValue("thu:07:44-thu:09:44"),
			val2:   fwtypes.OnceAWeekWindowValue("thu:07:44-fri:11:09"),
			equals: false,
		},
	}

	for name, test := range tests {
		name, test := name, test
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
