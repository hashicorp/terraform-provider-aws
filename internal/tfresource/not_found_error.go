// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type EmptyResultError struct {
	LastRequest interface{}
}

var ErrEmptyResult = &EmptyResultError{}

func NewEmptyResultError(lastRequest interface{}) error {
	return &EmptyResultError{
		LastRequest: lastRequest,
	}
}

func (e *EmptyResultError) Error() string {
	return "empty result"
}

func (e *EmptyResultError) Is(err error) bool {
	_, ok := err.(*EmptyResultError)
	return ok
}

func (e *EmptyResultError) As(target interface{}) bool {
	t, ok := target.(**retry.NotFoundError)
	if !ok {
		return false
	}

	*t = &retry.NotFoundError{
		Message:     e.Error(),
		LastRequest: e.LastRequest,
	}

	return true
}

type TooManyResultsError struct {
	Count       int
	LastRequest interface{}
}

var ErrTooManyResults = &TooManyResultsError{}

func NewTooManyResultsError(count int, lastRequest interface{}) error {
	return &TooManyResultsError{
		Count:       count,
		LastRequest: lastRequest,
	}
}

func (e *TooManyResultsError) Error() string {
	return fmt.Sprintf("too many results: wanted 1, got %d", e.Count)
}

func (e *TooManyResultsError) Is(err error) bool {
	_, ok := err.(*TooManyResultsError)
	return ok
}

func (e *TooManyResultsError) As(target interface{}) bool {
	t, ok := target.(**retry.NotFoundError)
	if !ok {
		return false
	}

	*t = &retry.NotFoundError{
		Message:     e.Error(),
		LastRequest: e.LastRequest,
	}

	return true
}

// SingularDataSourceFindError returns a standard error message for a singular data source's non-nil resource find error.
func SingularDataSourceFindError(resourceType string, err error) error {
	if NotFound(err) {
		if errors.Is(err, &TooManyResultsError{}) {
			return fmt.Errorf("multiple %[1]ss matched; use additional constraints to reduce matches to a single %[1]s", resourceType)
		}

		return fmt.Errorf("no matching %[1]s found", resourceType)
	}

	return fmt.Errorf("reading %s: %w", resourceType, err)
}

func AssertSinglePtrResult[T any](a []*T) (*T, error) {
	if l := len(a); l == 0 {
		return nil, NewEmptyResultError(nil)
	} else if l > 1 {
		return nil, NewTooManyResultsError(l, nil)
	} else if a[0] == nil {
		return nil, NewEmptyResultError(nil)
	}
	return a[0], nil
}

func AssertSingleValueResult[T any](a []T) (*T, error) {
	if l := len(a); l == 0 {
		return nil, NewEmptyResultError(nil)
	} else if l > 1 {
		return nil, NewTooManyResultsError(l, nil)
	}
	return &a[0], nil
}
