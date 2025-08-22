// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"testing"

	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
)

func TestStringSlicesEqualIgnoreOrder(t *testing.T) {
	t.Parallel()

	equal := []any{
		[]any{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
		},
		[]any{
			[]string{"b", "a", "c"},
			[]string{"a", "b", "c"},
		},
		[]any{
			[]string{"apple", "carrot", "tomato"},
			[]string{"tomato", "apple", "carrot"},
		},
		[]any{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Barrier", "Application", "Donut", "Chilly"},
		},
	}
	for _, v := range equal {
		if !tflakeformation.StringSlicesEqualIgnoreOrder(v.([]any)[0].([]string), v.([]any)[1].([]string)) {
			t.Fatalf("%v should be equal: %v", v.([]any)[0].([]string), v.([]any)[1].([]string))
		}
	}

	notEqual := []any{
		[]any{
			[]string{"c", "b", "c"},
			[]string{"a", "b", "c"},
		},
		[]any{
			[]string{"b", "a", "c"},
			[]string{"a", "bread", "c"},
		},
		[]any{
			[]string{"apple", "carrot", "tomato"},
			[]string{"tomato", "apple"},
		},
		[]any{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Barrier", "Applications", "Donut", "Chilly"},
		},
		[]any{
			[]string{},
			[]string{"Barrier", "Applications", "Donut", "Chilly"},
		},
	}
	for _, v := range notEqual {
		if tflakeformation.StringSlicesEqualIgnoreOrder(v.([]any)[0].([]string), v.([]any)[1].([]string)) {
			t.Fatalf("%v should not be equal: %v", v.([]any)[0].([]string), v.([]any)[1].([]string))
		}
	}
}
