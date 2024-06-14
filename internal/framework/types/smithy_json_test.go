// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestSmithyJSONTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.SmithyJSONNull[smithyjson.JSONStringer](),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.SmithyJSONUnknown[smithyjson.JSONStringer](),
		},
		"valid SmithyJSON": {
			val:      tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			expected: fwtypes.SmithyJSONValue[smithyjson.JSONStringer](`{"test": "value"}`, nil), // lintignore:AWSAT003,AWSAT005
		},
		"invalid SmithyJSON": {
			val:      tftypes.NewValue(tftypes.String, "not ok"),
			expected: fwtypes.SmithyJSONUnknown[smithyjson.JSONStringer](),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.SmithyJSONType[smithyjson.JSONStringer]{}.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSmithyJSONValidateAttribute(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         fwtypes.SmithyJSON[smithyjson.JSONStringer]
		expectError bool
	}{
		"null value": {
			val: fwtypes.SmithyJSONNull[smithyjson.JSONStringer](),
		},
		"unknown value": {
			val: fwtypes.SmithyJSONUnknown[smithyjson.JSONStringer](),
		},
		"valid SmithyJSON": { // lintignore:AWSAT003,AWSAT005
			val: fwtypes.SmithyJSONValue[smithyjson.JSONStringer](`{"test": "value"}`, nil), // lintignore:AWSAT003,AWSAT005
		},
		"invalid SmithyJSON": {
			val:         fwtypes.SmithyJSONValue[smithyjson.JSONStringer]("not ok", nil),
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

type testJSONDocument struct {
	Value any
}

func newTestJSONDocument(v any) smithyjson.JSONStringer {
	return &testJSONDocument{Value: v}
}

func (m *testJSONDocument) UnmarshalSmithyDocument(v interface{}) error {
	data, err := json.Marshal(m.Value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (m *testJSONDocument) MarshalSmithyDocument() ([]byte, error) {
	return json.Marshal(m.Value)
}

func TestSmithyJSONValueInterface(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         fwtypes.SmithyJSON[smithyjson.JSONStringer]
		expected    smithyjson.JSONStringer
		expectError bool
	}{
		"null value": {
			val: fwtypes.SmithyJSONNull[smithyjson.JSONStringer](),
		},
		"unknown value": {
			val: fwtypes.SmithyJSONUnknown[smithyjson.JSONStringer](),
		},
		"valid SmithyJSON": { // lintignore:AWSAT003,AWSAT005
			val: fwtypes.SmithyJSONValue[smithyjson.JSONStringer](`{"test": "value"}`, newTestJSONDocument), // lintignore:AWSAT003,AWSAT005
			expected: &testJSONDocument{
				Value: map[string]any{
					"test": "value",
				},
			},
		},
		"invalid SmithyJSON": {
			val:         fwtypes.SmithyJSONValue[smithyjson.JSONStringer]("not ok", newTestJSONDocument), // lintignore:AWSAT003,AWSAT005
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s, err := test.val.ValueInterface()
			gotErr := err.HasError()

			if gotErr != test.expectError {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, test.expectError)
			}

			if gotErr {
				if !test.expectError {
					t.Errorf("err = %q", err)
				}
			} else if diff := cmp.Diff(s, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
