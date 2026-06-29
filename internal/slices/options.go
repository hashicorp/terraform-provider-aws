// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package slices

type finderOptions[T any] struct {
	filter           Predicate[T]
	returnFirstMatch bool
}

func NewFinderOptions[T any](optFns ...FinderOptionsFunc[T]) finderOptions[T] {
	var opts finderOptions[T]
	for _, fn := range optFns {
		fn(&opts)
	}
	return opts
}

func (o *finderOptions[T]) ReturnFirstMatch() bool {
	return o.returnFirstMatch
}

func (o *finderOptions[T]) Filter() Predicate[T] {
	return o.filter
}

type FinderOptionsFunc[T any] func(*finderOptions[T])

// WithReturnFirstMatch is a finder option to enable paginated operations to short
// circuit after the first filter match
//
// This option should only be used when only a single match will ever be returned
// from the specified filter.
func WithReturnFirstMatch[T any]() FinderOptionsFunc[T] {
	return func(o *finderOptions[T]) {
		o.returnFirstMatch = true
	}
}

func WithFilter[T any](filter Predicate[T]) FinderOptionsFunc[T] {
	return func(o *finderOptions[T]) {
		o.filter = filter
	}
}
