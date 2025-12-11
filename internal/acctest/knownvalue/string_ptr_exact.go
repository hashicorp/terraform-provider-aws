// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

var _ knownvalue.Check = stringPtrExact[string]{}

type stringPtrExact[T ~string] struct {
	value *T
}

func (v stringPtrExact[T]) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for StringPtrExact check, got: %T", other)
	}

	if otherVal != string(*v.value) {
		return fmt.Errorf("expected value %s for StringPtrExact check, got: %s", *v.value, otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v stringPtrExact[T]) String() string {
	return string(*v.value)
}

// StringExact returns a Check for asserting equality between the
// supplied string and a value passed to the CheckValue method.
func StringPtrExact[T ~string](value *T) stringPtrExact[T] {
	if value == nil {
		panic("value must not be nil")
	}
	return stringPtrExact[T]{
		value: value,
	}
}
