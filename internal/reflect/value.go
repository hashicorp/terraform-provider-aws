// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"reflect"
)

// CanElem returns whether the value can be dereferenced with Elem().
func CanElem(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Pointer || k == reflect.Interface
}
