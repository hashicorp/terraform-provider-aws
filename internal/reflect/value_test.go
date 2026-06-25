// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"reflect"
	"testing"
)

func TestCanElem(t *testing.T) {
	t.Parallel()

	// Kind() == Interface: take the addressable element of a *interface{}.
	// reflect.ValueOf(x) always unwraps to the concrete type, so you can't get
	// an Interface-kind Value directly.
	var iface any = 42
	ifaceVal := reflect.ValueOf(&iface).Elem()

	tests := map[string]struct {
		v    reflect.Value
		want bool
	}{
		"pointer":            {v: reflect.ValueOf(new(int)), want: true},
		"typed nil pointer":  {v: reflect.ValueOf((*int)(nil)), want: true},
		"interface":          {v: ifaceVal, want: true},
		"int":                {v: reflect.ValueOf(42), want: false},
		"string":             {v: reflect.ValueOf("x"), want: false},
		"struct":             {v: reflect.ValueOf(struct{ X int }{}), want: false},
		"slice":              {v: reflect.ValueOf([]int{1}), want: false},
		"map":                {v: reflect.ValueOf(map[string]int{}), want: false},
		"chan":               {v: reflect.ValueOf(make(chan int)), want: false},
		"func":               {v: reflect.ValueOf(func() {}), want: false},
		"array":              {v: reflect.ValueOf([2]int{}), want: false},
		"invalid zero Value": {v: reflect.Value{}, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := CanElem(tc.v); got != tc.want {
				t.Errorf("CanElem(kind=%s) = %v, want %v", tc.v.Kind(), got, tc.want)
			}
		})
	}
}
