// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

type mockWriteOnlyAttrGetter struct {
	path  cty.Path
	value cty.Value
}

func (m *mockWriteOnlyAttrGetter) Get(string) any {
	return nil
}

func (m *mockWriteOnlyAttrGetter) GetRawConfigAt(path cty.Path) (cty.Value, diag.Diagnostics) {
	if !path.Equals(m.path) {
		return cty.NilVal, diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid config Path",
				Detail:        fmt.Sprintf("expected: %v, got: %v", m.path, path),
				AttributePath: path,
			},
		}
	}
	return m.value, nil
}

func (m *mockWriteOnlyAttrGetter) GetRawConfig() cty.Value {
	return cty.MapVal(map[string]cty.Value{
		"has_write_only_value": cty.BoolVal(true),
	})
}

func (m *mockWriteOnlyAttrGetter) Id() string {
	return "id"
}

func TestGetWriteOnlyValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input       cty.Path
		setPath     cty.Path
		inputType   cty.Type
		value       cty.Value
		expectError bool
	}{
		"valid value type": {
			input:     cty.GetAttrPath("test_path"),
			setPath:   cty.GetAttrPath("test_path"),
			inputType: cty.String,
			value:     cty.StringVal("test_value"),
		},
		"invalid value type": {
			input:       cty.GetAttrPath("test_path"),
			setPath:     cty.GetAttrPath("test_path"),
			inputType:   cty.String,
			value:       cty.BoolVal(true),
			expectError: true,
		},
		"invalid path": {
			input:       cty.GetAttrPath("invalid_path"),
			setPath:     cty.GetAttrPath("test_path"),
			inputType:   cty.String,
			value:       cty.StringVal("test_value"),
			expectError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := mockWriteOnlyAttrGetter{
				path:  testCase.setPath,
				value: testCase.value,
			}
			_, diags := flex.GetWriteOnlyValue(&m, testCase.input, testCase.inputType)

			if testCase.expectError && !diags.HasError() {
				t.Fatalf("expected error, got none")
			}
		})
	}
}

func TestGetWriteOnlyStringValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input         cty.Path
		setPath       cty.Path
		value         cty.Value
		expectedValue string
	}{
		"valid value": {
			input:         cty.GetAttrPath("test_path"),
			setPath:       cty.GetAttrPath("test_path"),
			value:         cty.StringVal("test_value"),
			expectedValue: "test_value",
		},
		"value empty string": {
			input:         cty.GetAttrPath("test_path"),
			setPath:       cty.GetAttrPath("test_path"),
			value:         cty.StringVal(""),
			expectedValue: "",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := mockWriteOnlyAttrGetter{
				path:  testCase.setPath,
				value: testCase.value,
			}
			value, diags := flex.GetWriteOnlyStringValue(&m, testCase.input)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if testCase.expectedValue != value {
				t.Fatalf("expected value: %s, got: %s", testCase.expectedValue, value)
			}
		})
	}
}
