// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package option

import (
	"errors"

	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

var (
	errMissingValue = errors.New("missing value")
)

type Option[T any] []T

const (
	value = iota
)

// Some returns an Option containing a value.
func Some[T any](v T) Option[T] {
	return Option[T]{
		value: v,
	}
}

// None returns an Option with no value.
func None[T any]() Option[T] {
	return nil
}

// IsNone returns whether the Option has no value.
func (o Option[T]) IsNone() bool {
	return o == nil
}

// IsSome returns whether the Option has a value.
func (o Option[T]) IsSome() bool {
	return o != nil
}

// MustUnwrap returns the contained value or an error.
func (o Option[T]) Unwrap() (T, error) {
	if o.IsNone() {
		var zero T
		return zero, errMissingValue
	}
	return o[value], nil
}

// MustUnwrap returns the contained value or panics.
func (o Option[T]) MustUnwrap() T {
	return errs.Must(o.Unwrap())
}

// UnwrapOr returns the contained value or the specified default.
func (o Option[T]) UnwrapOr(v T) T {
	return o.UnwrapOrElse(func() T {
		return v
	})
}

// UnwrapOrDefault returns the contained value or the default value for T.
func (o Option[T]) UnwrapOrDefault() T {
	return o.UnwrapOrElse(func() T {
		var v T
		return v
	})
}

// UnwrapOrElse returns the contained value or computes a value from f.
func (o Option[T]) UnwrapOrElse(f func() T) T {
	if o.IsNone() {
		return f()
	}
	return o[value]
}
