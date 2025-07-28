// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

type EmptyResultError struct {
	LastRequest any
}

var ErrEmptyResult = &EmptyResultError{}

func NewEmptyResultError(lastRequest any) error {
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

func (e *EmptyResultError) As(target any) bool {
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
	LastRequest any
}

var ErrTooManyResults = &TooManyResultsError{}

func NewTooManyResultsError(count int, lastRequest any) error {
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

func (e *TooManyResultsError) As(target any) bool {
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

// foundFunc is function that returns false if the specified value causes a `NotFound` error to be returned.
type foundFunc[T any] tfslices.Predicate[*T]

// AssertMaybeSingleValueResult returns the single non-nil value in the specified slice, or `None` if the slice is empty.
// Returns a `NotFound` error otherwise.
func AssertMaybeSingleValueResult[T any](a []T) (option.Option[T], error) {
	if l := len(a); l == 0 {
		return option.None[T](), nil
	} else if l > 1 {
		return nil, NewTooManyResultsError(l, nil)
	}

	return option.Some(a[0]), nil
}

// AssertSingleValueResult returns a pointer to the single value in the specified slice of values.
// Returns a `NotFound` error otherwise.
func AssertSingleValueResult[T any](a []T, fs ...foundFunc[T]) (*T, error) {
	if l := len(a); l == 0 {
		return nil, NewEmptyResultError(nil)
	} else if l > 1 {
		return nil, NewTooManyResultsError(l, nil)
	} else {
		v := &a[0]
		for _, f := range fs {
			if !f(v) {
				return nil, NewEmptyResultError(nil)
			}
		}
		return v, nil
	}
}

// AssertFirstValueResult returns a pointer to the first value in the specified slice of values.
// Returns a `NotFound` error otherwise.
func AssertFirstValueResult[T any](a []T) (*T, error) {
	if l := len(a); l == 0 {
		return nil, NewEmptyResultError(nil)
	}
	return &a[0], nil
}
