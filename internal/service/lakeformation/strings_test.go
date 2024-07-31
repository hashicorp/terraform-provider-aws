// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"testing"

	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
)

func TestStringSlicesEqualIgnoreOrder(t *testing.T) {
	t.Parallel()

	equal := []interface{}{
		[]interface{}{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
		},
		[]interface{}{
			[]string{"b", "a", "c"},
			[]string{"a", "b", "c"},
		},
		[]interface{}{
			[]string{"apple", "carrot", "tomato"},
			[]string{"tomato", "apple", "carrot"},
		},
		[]interface{}{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Barrier", "Application", "Donut", "Chilly"},
		},
	}
	for _, v := range equal {
		if !tflakeformation.StringSlicesEqualIgnoreOrder(v.([]interface{})[0].([]string), v.([]interface{})[1].([]string)) {
			t.Fatalf("%v should be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}

	notEqual := []interface{}{
		[]interface{}{
			[]string{"c", "b", "c"},
			[]string{"a", "b", "c"},
		},
		[]interface{}{
			[]string{"b", "a", "c"},
			[]string{"a", "bread", "c"},
		},
		[]interface{}{
			[]string{"apple", "carrot", "tomato"},
			[]string{"tomato", "apple"},
		},
		[]interface{}{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Barrier", "Applications", "Donut", "Chilly"},
		},
		[]interface{}{
			[]string{},
			[]string{"Barrier", "Applications", "Donut", "Chilly"},
		},
	}
	for _, v := range notEqual {
		if tflakeformation.StringSlicesEqualIgnoreOrder(v.([]interface{})[0].([]string), v.([]interface{})[1].([]string)) {
			t.Fatalf("%v should not be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}
}
