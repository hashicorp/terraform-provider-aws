// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import "slices"

var (
	DefaultIgnoredFieldNames = []string{
		"Tags", // Resource tags are handled separately.
	}
)

// AutoFlexOptionsFunc is a type alias for an autoFlexer functional option.
type AutoFlexOptionsFunc func(*AutoFlexOptions)

// AutoFlexOptions stores configurable options for an auto-flattener or expander.
type AutoFlexOptions struct {
	// fieldNamePrefix specifies a common prefix which may be applied to one
	// or more fields on an AWS data structure
	fieldNamePrefix string

	// fieldNameSuffix specifies a common suffix which may be applied to one
	// or more fields on an AWS data structure
	fieldNameSuffix string

	// ignoredFieldNames stores names which expanders and flatteners will
	// not read from or write to
	ignoredFieldNames []string
}

// WithFieldNamePrefix specifies a prefix to be accounted for when
// matching field names between Terraform and AWS data structures
//
// Use this option to improve fuzzy matching of field names during AutoFlex
// expand/flatten operations.
func WithFieldNamePrefix(s string) AutoFlexOptionsFunc {
	return func(o *AutoFlexOptions) {
		o.fieldNamePrefix = s
	}
}

// WithFieldNameSuffix specifies a suffix to be accounted for when
// matching field names between Terraform and AWS data structures
//
// Use this option to improve fuzzy matching of field names during AutoFlex
// expand/flatten operations.
func WithFieldNameSuffix(s string) AutoFlexOptionsFunc {
	return func(o *AutoFlexOptions) {
		o.fieldNameSuffix = s
	}
}

// WithIgnoredFieldNamesAppend appends to the list of ignored field names
//
// Use this option to preserve preexisting items in the ignored fields list.
func WithIgnoredFieldNamesAppend(s string) AutoFlexOptionsFunc {
	return func(o *AutoFlexOptions) {
		o.ignoredFieldNames = append(o.ignoredFieldNames, s)
	}
}

// WithIgnoredFieldNames sets the list of ignored field names
//
// Use this option to fully overwrite the ignored fields list. To preseve
// preexisting items, use WithIgnoredFieldNamesAppend instead.
func WithIgnoredFieldNames(fields []string) AutoFlexOptionsFunc {
	return func(o *AutoFlexOptions) {
		o.ignoredFieldNames = fields
	}
}

// WithNoIgnoredFieldNames empties the list of ignored field names
func WithNoIgnoredFieldNames() AutoFlexOptionsFunc {
	return func(o *AutoFlexOptions) {
		o.ignoredFieldNames = []string{}
	}
}

// isIgnoredField returns true if s is in the list of ignored field names
func (o *AutoFlexOptions) isIgnoredField(s string) bool {
	return slices.Contains(o.ignoredFieldNames, s)
}
