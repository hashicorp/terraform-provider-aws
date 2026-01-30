// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty

import (
	"iter"

	"github.com/hashicorp/go-cty/cty"
)

// ValueElements returns an iterator over index-value pairs of the receiver
// which must be a collection type, a tuple type, or an object type. If called
// on a value of any other type, this function will panic.
//
// The value must be Known and non-Null, or this function will panic.
//
// If the receiver is of a list type, the returned keys will be of type Number
// and the values will be of the list's element type.
//
// If the receiver is of a map type, the returned keys will be of type String
// and the value will be of the map's element type. Elements are passed in
// ascending lexicographical order by key.
//
// If the receiver is of a set type, each element is returned as both the
// key and the value, since set members are their own identity.
//
// If the receiver is of a tuple type, the returned keys will be of type Number
// and the value will be of the corresponding element's type.
//
// If the receiver is of an object type, the returned keys will be of type
// String and the value will be of the corresponding attributes's type.
func ValueElements(v cty.Value) iter.Seq2[cty.Value, cty.Value] {
	return func(yield func(cty.Value, cty.Value) bool) {
		it := v.ElementIterator()
		for it.Next() {
			if !yield(it.Element()) {
				return
			}
		}
	}
}

// ValueElementValues returns an iterator over the values of the receiver
// which must be a collection type, a tuple type, or an object type. If called
// on a value of any other type, this function will panic.
//
// The value must be Known and non-Null, or this function will panic.
func ValueElementValues(v cty.Value) iter.Seq[cty.Value] {
	return func(yield func(cty.Value) bool) {
		it := v.ElementIterator()
		for it.Next() {
			_, v := it.Element()
			if !yield(v) {
				return
			}
		}
	}
}
