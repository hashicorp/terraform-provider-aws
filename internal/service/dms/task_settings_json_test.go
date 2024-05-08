// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestTaskSettingsEqual(t *testing.T) {
	t.Parallel()

	tests := map[string]map[string]struct {
		a, b     any
		expected bool
	}{
		"bool": {
			"both true": {
				a:        true,
				b:        true,
				expected: true,
			},
			"not equal": {
				a:        true,
				b:        false,
				expected: false,
			},
			"both null": {
				a:        nil,
				b:        nil,
				expected: true,
			},
			"true proposed null": {
				a:        true,
				b:        nil,
				expected: true,
			},
			"false proposed null": {
				a:        false,
				b:        nil,
				expected: true,
			},
			"null proposed true": {
				a:        nil,
				b:        true,
				expected: false,
			},
			"null proposed false": {
				a:        nil,
				b:        false,
				expected: false,
			},
		},
		"float64": {
			"equal": {
				a:        float64(1),
				b:        float64(1),
				expected: true,
			},
			"not equal": {
				a:        float64(1),
				b:        float64(2),
				expected: false,
			},
			"proposed null": {
				a:        float64(1),
				b:        nil,
				expected: true,
			},
			"null proposed value": {
				a:        nil,
				b:        float64(1),
				expected: false,
			},
		},
		"string": {
			"equal": {
				a:        names.AttrValue,
				b:        names.AttrValue,
				expected: true,
			},
			"not equal": {
				a:        "value1",
				b:        "value2",
				expected: false,
			},
			"both null": {
				a:        nil,
				b:        nil,
				expected: true,
			},
			"proposed null": {
				a:        names.AttrValue,
				b:        nil,
				expected: true,
			},
			"null proposed value": {
				a:        nil,
				b:        names.AttrValue,
				expected: false,
			},
		},
	}

	for name, typeTest := range tests {
		name, typeTest := name, typeTest
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for name, test := range typeTest {
				name, test := name, test
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					if taskSettingsEqual(test.a, test.b) != test.expected {
						t.Fatalf("expected %v, got %v", test.expected, !test.expected)
					}
				})
			}
		})
	}
}
