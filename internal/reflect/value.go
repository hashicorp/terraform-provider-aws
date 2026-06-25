// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"reflect"
)

// ValueCanElem returns whether the value can be dereferenced with Elem().
func ValueCanElem(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Pointer || k == reflect.Interface
}
