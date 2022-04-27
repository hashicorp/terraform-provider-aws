package lakeformation_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
)

func TestStringSlicesEqualIgnoreOrder(t *testing.T) {
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
		if !tflakeformation.StringSlicesEqualIgnoreOrder(aws.StringSlice(v.([]interface{})[0].([]string)), aws.StringSlice(v.([]interface{})[1].([]string))) {
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
		if tflakeformation.StringSlicesEqualIgnoreOrder(aws.StringSlice(v.([]interface{})[0].([]string)), aws.StringSlice(v.([]interface{})[1].([]string))) {
			t.Fatalf("%v should not be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}
}

func TestStringSlicesEqual(t *testing.T) {
	equal := []interface{}{
		[]interface{}{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
		},
		[]interface{}{
			[]string{"b", "a", "c"},
			[]string{"b", "a", "c"},
		},
		[]interface{}{
			[]string{"apple", "carrot", "tomato"},
			[]string{"apple", "carrot", "tomato"},
		},
		[]interface{}{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Application", "Barrier", "Chilly", "Donut"},
		},
		[]interface{}{
			[]string{},
			[]string{},
		},
	}
	for _, v := range equal {
		if !tflakeformation.StringSlicesEqual(aws.StringSlice(v.([]interface{})[0].([]string)), aws.StringSlice(v.([]interface{})[1].([]string))) {
			t.Fatalf("%v should be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}

	notEqual := []interface{}{
		[]interface{}{
			[]string{"a", "b", "c"},
			[]string{"a", "b"},
		},
		[]interface{}{
			[]string{"a", "b", "c"},
			[]string{"b", "a", "c"},
		},
		[]interface{}{
			[]string{"apple", "carrot", "tomato"},
			[]string{"apple", "carrot", "tomato", "zucchini"},
		},
		[]interface{}{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Application", "Barrier", "Chilly", "Donuts"},
		},
		[]interface{}{
			[]string{},
			[]string{"Application", "Barrier", "Chilly", "Donuts"},
		},
	}
	for _, v := range notEqual {
		if tflakeformation.StringSlicesEqual(aws.StringSlice(v.([]interface{})[0].([]string)), aws.StringSlice(v.([]interface{})[1].([]string))) {
			t.Fatalf("%v should not be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}
}
