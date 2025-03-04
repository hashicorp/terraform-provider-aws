// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

type ListNestedObjectOfOption[T any] func(*ListNestedObjectOfOptions[T])

type ListNestedObjectOfOptions[T any] struct {
	SemanticEqualityFunc listSemanticEqualityFunc[T]
}

func WithSemanticEqualityFunc[T any](f listSemanticEqualityFunc[T]) ListNestedObjectOfOption[T] {
	return func(o *ListNestedObjectOfOptions[T]) {
		o.SemanticEqualityFunc = f
	}
}

func newListNestedObjectOfOptions[T any](options ...ListNestedObjectOfOption[T]) *ListNestedObjectOfOptions[T] {
	opts := &ListNestedObjectOfOptions[T]{}

	for _, opt := range options {
		opt(opts)
	}

	return opts
}
