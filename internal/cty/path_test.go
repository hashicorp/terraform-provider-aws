// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty_test

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
)

type pathSafeApplyTestCase struct {
	path          cty.Path
	value         cty.Value
	expectedValue cty.Value
	expectedFound bool
	expectedErr   error
}

func TestPathSafeApply_NilValue(t *testing.T) {
	t.Parallel()

	tests := map[string]pathSafeApplyTestCase{
		"attr path": {
			path:          cty.GetAttrPath("some_attribute"),
			value:         cty.NilVal,
			expectedValue: cty.NilVal,
			expectedFound: false,
			expectedErr:   nil,
		},
		"string index path": {
			path:          cty.IndexStringPath("some_key"),
			value:         cty.NilVal,
			expectedValue: cty.NilVal,
			expectedFound: false,
			expectedErr:   nil,
		},
		"number index path": {
			path:          cty.IndexIntPath(0),
			value:         cty.NilVal,
			expectedValue: cty.NilVal,
			expectedFound: false,
			expectedErr:   nil,
		},
		"attr path multiple steps": {
			path:          cty.GetAttrPath("some_attribute").IndexInt(0),
			value:         cty.NilVal,
			expectedValue: cty.NilVal,
			expectedFound: false,
			expectedErr:   nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualValue, actualFound, err := tfcty.PathSafeApply(tt.path, tt.value)

			if err != nil {
				if tt.expectedErr == nil {
					t.Errorf("PathSafeApply() failed: %q", err)
				} else if err.Error() != tt.expectedErr.Error() {
					t.Errorf("PathSafeApply() failed: expected error %q, got %q", tt.expectedErr, err)
				}
			} else if tt.expectedErr != nil {
				t.Fatal("PathSafeApply() succeeded unexpectedly")
			}

			if !actualValue.IsNull() {
				t.Errorf("unexpected value %v, expected Null", actualValue)
			}

			if actualFound != tt.expectedFound {
				t.Errorf("unexpected found %t, expected %t", actualFound, tt.expectedFound)
			}
		})
	}
}

func TestPathSafeApply_AttrAndIndex(t *testing.T) {
	t.Parallel()

	emptyVal := cty.ObjectVal(map[string]cty.Value{
		"some_attribute": cty.ListValEmpty(cty.String),
	})

	listVal := cty.ObjectVal(map[string]cty.Value{
		"some_attribute": cty.ListVal([]cty.Value{
			cty.StringVal("test"),
		}),
	})

	tests := map[string]pathSafeApplyTestCase{
		"empty list": {
			path:          cty.GetAttrPath("some_attribute").IndexInt(0),
			value:         emptyVal,
			expectedValue: cty.NilVal,
			expectedFound: false,
			expectedErr:   nil,
		},
		"list with value": {
			path:          cty.GetAttrPath("some_attribute").IndexInt(0),
			value:         listVal,
			expectedValue: cty.StringVal("test"),
			expectedFound: true,
			expectedErr:   nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualValue, actualFound, err := tfcty.PathSafeApply(tt.path, tt.value)

			if err != nil {
				if tt.expectedErr == nil {
					t.Errorf("PathSafeApply() failed: %q", err)
				} else if err.Error() != tt.expectedErr.Error() {
					t.Errorf("PathSafeApply() failed: expected error %q, got %q", tt.expectedErr, err)
				}
			} else if tt.expectedErr != nil {
				t.Fatal("PathSafeApply() succeeded unexpectedly")
			}

			if tt.expectedValue.IsNull() {
				if !actualValue.IsNull() {
					t.Errorf("unexpected value %v, expected Null", actualValue)
				}
			} else if !tt.expectedValue.RawEquals(actualValue) {
				t.Errorf("unexpected value %v, expected %v", actualValue, tt.expectedValue)
			}

			if actualFound != tt.expectedFound {
				t.Errorf("unexpected found %t, expected %t", actualFound, tt.expectedFound)
			}
		})
	}
}
