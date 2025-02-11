// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

// ChangeOption is a type alias for a functional option that modifies ChangeOptions
type ChangeOption func(*ChangeOptions)

// ChangeOptions holds configuration for calculating plan changes
type ChangeOptions struct {
	IgnoredFields []string
}

// WithIgnoredField specifies a field name to be ignored when calculating plan changes
func WithIgnoredField(fieldName string) ChangeOption {
	return func(o *ChangeOptions) {
		o.IgnoredFields = append(o.IgnoredFields, fieldName)
	}
}

// NewChangeOptions initializes ChangeOptions with the provided options
func NewChangeOptions(options ...ChangeOption) *ChangeOptions {
	opts := &ChangeOptions{
		IgnoredFields: make([]string, 0),
	}

	for _, opt := range options {
		opt(opts)
	}

	return opts
}
