// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

var _ knownvalue.Check = listNotEmpty{}

type listNotEmpty struct{}

func (v listNotEmpty) CheckValue(other any) error {
	otherVal, ok := other.([]any)
	if !ok {
		return fmt.Errorf("expected []any value for ListNotEmpty check, got: %T", other)
	}

	if len(otherVal) == 0 {
		return errors.New("expected non-empty list for ListNotEmpty check")
	}

	return nil
}

// String returns the string representation of the value.
func (v listNotEmpty) String() string {
	return "stuff"
}

// ListNotEmpty returns a Check for asserting that a list is not empty.
func ListNotEmpty() listNotEmpty {
	return listNotEmpty{}
}
