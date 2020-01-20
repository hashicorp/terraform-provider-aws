package validation

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// IntBetween returns a SchemaValidateFunc which tests if the provided value
// is of type int and is between min and max (inclusive)
func IntBetween(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		if v < min || v > max {
			es = append(es, fmt.Errorf("expected %s to be in the range (%d - %d), got %d", k, min, max, v))
			return
		}

		return
	}
}

// IntAtLeast returns a SchemaValidateFunc which tests if the provided value
// is of type int and is at least min (inclusive)
func IntAtLeast(min int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		if v < min {
			es = append(es, fmt.Errorf("expected %s to be at least (%d), got %d", k, min, v))
			return
		}

		return
	}
}

// IntAtMost returns a SchemaValidateFunc which tests if the provided value
// is of type int and is at most max (inclusive)
func IntAtMost(max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		if v > max {
			es = append(es, fmt.Errorf("expected %s to be at most (%d), got %d", k, max, v))
			return
		}

		return
	}
}

// IntInSlice returns a SchemaValidateFunc which tests if the provided value
// is of type int and matches the value of an element in the valid slice
func IntInSlice(valid []int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be an integer", k))
			return
		}

		for _, validInt := range valid {
			if v == validInt {
				return
			}
		}

		es = append(es, fmt.Errorf("expected %s to be one of %v, got %d", k, valid, v))
		return
	}
}
