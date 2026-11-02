// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

type NestedObjectOfOptionsFunc[T any] func(*nestedObjectOfOptions[T])

type nestedObjectOfOptions[T any] struct {
	SemanticEqualityFunc semanticEqualityFunc[T]
}

func WithSemanticEqualityFunc[T any](f semanticEqualityFunc[T]) NestedObjectOfOptionsFunc[T] {
	return func(o *nestedObjectOfOptions[T]) {
		o.SemanticEqualityFunc = f
	}
}

func newNestedObjectOfOptions[T any](optFns ...NestedObjectOfOptionsFunc[T]) nestedObjectOfOptions[T] {
	var opts nestedObjectOfOptions[T]
	for _, fn := range optFns {
		fn(&opts)
	}
	return opts
}
