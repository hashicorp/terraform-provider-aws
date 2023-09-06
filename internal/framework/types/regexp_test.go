// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestRegexpTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected fwtypes.Regexp
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.RegexpNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.RegexpUnknown(),
		},
		"valid Regexp": {
			val:      tftypes.NewValue(tftypes.String, `\w+`),
			expected: fwtypes.RegexpValue(regexache.MustCompile(`\w+`)),
		},
		"invalid Regexp": {
			val:      tftypes.NewValue(tftypes.String, `(`),
			expected: fwtypes.RegexpUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.RegexpType.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if !test.expected.Equal(val) {
				t.Errorf("unexpected diff: wanted %q, got %q", test.expected.String(), val.String())
			}
		})
	}
}

func TestRegexpTypeValidate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         tftypes.Value
		expectError bool
	}
	tests := map[string]testCase{
		"not a string": {
			val:         tftypes.NewValue(tftypes.Bool, true),
			expectError: true,
		},
		"unknown string": {
			val: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"null string": {
			val: tftypes.NewValue(tftypes.String, nil),
		},
		"valid string": {
			val: tftypes.NewValue(tftypes.String, `\w+`),
		},
		"invalid string": {
			val:         tftypes.NewValue(tftypes.String, `(`),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			diags := fwtypes.RegexpType.Validate(ctx, test.val, path.Root("test"))

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %#v", diags)
			}
		})
	}
}

func TestRegexpToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		regexp   fwtypes.Regexp
		expected types.String
	}{
		"value": {
			regexp:   fwtypes.RegexpValue(regexache.MustCompile(`\w+`)),
			expected: types.StringValue(`\w+`),
		},
		"null": {
			regexp:   fwtypes.RegexpNull(),
			expected: types.StringNull(),
		},
		"unknown": {
			regexp:   fwtypes.RegexpUnknown(),
			expected: types.StringUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			s, _ := test.regexp.ToStringValue(ctx)

			if !test.expected.Equal(s) {
				t.Fatalf("expected %#v to equal %#v", s, test.expected)
			}
		})
	}
}
