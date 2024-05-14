// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
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
				a:        acctest.CtValue1,
				b:        acctest.CtValue2,
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
		"map": {
			"equal": {
				a: map[string]any{
					acctest.CtKey1: names.AttrValue,
					acctest.CtKey2: map[string]any{
						"key3": names.AttrValue,
					},
				},
				b: map[string]any{
					acctest.CtKey1: names.AttrValue,
					acctest.CtKey2: map[string]any{
						"key3": names.AttrValue,
					},
				},
				expected: true,
			},
			"not equal": {
				a: map[string]any{
					acctest.CtKey1: names.AttrValue,
					acctest.CtKey2: map[string]any{
						"key3": acctest.CtValue1,
					},
				},
				b: map[string]any{
					acctest.CtKey1: names.AttrValue,
					acctest.CtKey2: map[string]any{
						"key3": acctest.CtValue2,
					},
				},
				expected: false,
			},
			"proposed null": {
				a: map[string]any{
					acctest.CtKey1: names.AttrValue,
					acctest.CtKey2: map[string]any{
						"key3": names.AttrValue,
					},
				},
				b: map[string]any{
					acctest.CtKey1: nil,
					acctest.CtKey2: map[string]any{
						"key3": names.AttrValue,
					},
				},
				expected: true,
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

					if tfdms.TaskSettingsEqual(test.a, test.b) != test.expected {
						t.Fatalf("expected %v, got %v", test.expected, !test.expected)
					}
				})
			}
		})
	}
}
