// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

type mockWriteOnlyAttrGetter struct {
	value cty.Value
}

func (m *mockWriteOnlyAttrGetter) GetRawConfigAt(_ cty.Path) (cty.Value, diag.Diagnostics) {
	return m.value, nil
}

func (m *mockWriteOnlyAttrGetter) Id() string {
	return "id"
}

func TestGetWriteOnlyValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input       cty.Path
		inputType   cty.Type
		setPath     cty.Path
		value       cty.Value
		expectError bool
	}{
		"valid value": {
			input:     cty.GetAttrPath("test_path"),
			inputType: cty.String,
			value:     cty.StringVal("test_value"),
		},
		"invalid type": {
			input:       cty.GetAttrPath("test_path"),
			inputType:   cty.String,
			value:       cty.BoolVal(true),
			expectError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := mockWriteOnlyAttrGetter{
				value: testCase.value,
			}
			_, diags := flex.GetWriteOnlyValue(&m, testCase.input, testCase.inputType)

			if testCase.expectError && !diags.HasError() {
				t.Fatalf("expected error, got none")
			}
		})
	}
}
