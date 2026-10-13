// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

var _ knownvalue.Check = setNotEmpty{}

type setNotEmpty struct{}

func (v setNotEmpty) CheckValue(other any) error {
	otherVal, ok := other.([]any)
	if !ok {
		return fmt.Errorf("expected []any value for SetNotEmpty check, got: %T", other)
	}

	if len(otherVal) == 0 {
		return errors.New("expected non-empty set for SetNotEmpty check")
	}

	return nil
}

// String returns the string representation of the value.
func (v setNotEmpty) String() string {
	return "non-empty set"
}

// SetNotEmpty returns a Check for asserting that a set is not empty.
func SetNotEmpty() setNotEmpty {
	return setNotEmpty{}
}
