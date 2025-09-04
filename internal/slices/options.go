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

// WithReturnFirstMatch is a finder option to enable paginated operations to short
// circuit after the first filter match
//
// This option should only be used when only a single match will ever be returned
// from the specified filter.
var WithReturnFirstMatch = func(o *FinderOptions) {
	o.returnFirstMatch = true
}
