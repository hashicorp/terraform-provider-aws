// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

type FinderOptions struct {
	returnFirstMatch bool
}

func NewFinderOptions(optFns []FinderOptionsFunc) *FinderOptions {
	opts := &FinderOptions{}
	for _, fn := range optFns {
		fn(opts)
	}
	return opts
}

func (o *FinderOptions) ReturnFirstMatch() bool {
	return o.returnFirstMatch
}

type FinderOptionsFunc func(*FinderOptions)

func WithReturnFirstMatch() FinderOptionsFunc {
	return func(o *FinderOptions) {
		o.returnFirstMatch = true
	}
}
