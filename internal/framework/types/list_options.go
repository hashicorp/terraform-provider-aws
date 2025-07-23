// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

type NestedObjectOfOption[T any] func(*NestedObjectOfOptions[T])

type NestedObjectOfOptions[T any] struct {
	SemanticEqualityFunc semanticEqualityFunc[T]
}

func WithSemanticEqualityFunc[T any](f semanticEqualityFunc[T]) NestedObjectOfOption[T] {
	return func(o *NestedObjectOfOptions[T]) {
		o.SemanticEqualityFunc = f
	}
}

func newNestedObjectOfOptions[T any](options ...NestedObjectOfOption[T]) *NestedObjectOfOptions[T] {
	opts := &NestedObjectOfOptions[T]{}

	for _, opt := range options {
		opt(opts)
	}

	return opts
}
