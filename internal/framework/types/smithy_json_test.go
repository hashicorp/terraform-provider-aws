// Copyright IBM Corp. 2014, 2025
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
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
)

func TestSmithyJSONTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.NewSmithyJSONNull[tfsmithy.JSONStringer](),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.NewSmithyJSONUnknown[tfsmithy.JSONStringer](),
		},
		"valid SmithyJSON": {
			val:      tftypes.NewValue(tftypes.String, `{"test": "value"}`),
			expected: fwtypes.NewSmithyJSONValue[tfsmithy.JSONStringer](`{"test": "value"}`, nil), // lintignore:AWSAT003,AWSAT005
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.SmithyJSONType[tfsmithy.JSONStringer]{}.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if got, want := val, test.expected; !got.Equal(want) {
				t.Errorf("got %T %v, want %T %v", got, got, want, want)
			}
		})
	}
}

func TestSmithyJSONValidateAttribute(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         fwtypes.SmithyJSON[tfsmithy.JSONStringer]
		expectError bool
	}{
		"null value": {
			val: fwtypes.NewSmithyJSONNull[tfsmithy.JSONStringer](),
		},
		"unknown value": {
			val: fwtypes.NewSmithyJSONUnknown[tfsmithy.JSONStringer](),
		},
		"valid SmithyJSON": { // lintignore:AWSAT003,AWSAT005
			val: fwtypes.NewSmithyJSONValue[tfsmithy.JSONStringer](`{"test": "value"}`, nil), // lintignore:AWSAT003,AWSAT005
		},
		"invalid SmithyJSON": {
			val:         fwtypes.NewSmithyJSONValue[tfsmithy.JSONStringer]("not ok", nil),
			expectError: true,
		},
	}

	for name, test := range tests {
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

func newTestJSONDocument(v any) tfsmithy.JSONStringer {
	return &testJSONDocument{Value: v}
}

func (m *testJSONDocument) UnmarshalSmithyDocument(v any) error {
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
		val         fwtypes.SmithyJSON[tfsmithy.JSONStringer]
		expected    tfsmithy.JSONStringer
		expectError bool
	}{
		"null value": {
			val: fwtypes.NewSmithyJSONNull[tfsmithy.JSONStringer](),
		},
		"unknown value": {
			val: fwtypes.NewSmithyJSONUnknown[tfsmithy.JSONStringer](),
		},
		"valid SmithyJSON": { // lintignore:AWSAT003,AWSAT005
			val: fwtypes.NewSmithyJSONValue(`{"test": "value"}`, newTestJSONDocument), // lintignore:AWSAT003,AWSAT005
			expected: &testJSONDocument{
				Value: map[string]any{
					"test": "value",
				},
			},
		},
		"valid SmithyJSON slice": { // lintignore:AWSAT003,AWSAT005
			val: fwtypes.NewSmithyJSONValue(`["value1","value"]`, newTestJSONDocument), // lintignore:AWSAT003,AWSAT005
			expected: &testJSONDocument{
				Value: []any{"value1", "value"},
			},
		},
		"invalid SmithyJSON": {
			val:         fwtypes.NewSmithyJSONValue("not ok", newTestJSONDocument), // lintignore:AWSAT003,AWSAT005
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s, err := test.val.ToSmithyDocument(t.Context())
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
