// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

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

	// ignoredFieldNames stores names which expanders and flatteners will
	// not read from or write to
	ignoredFieldNames []string
}

// NewFieldNamePrefixOptionsFunc specifies a prefix to be accounted for when
// matching field names between Terraform and AWS data structures
//
// Use this option to improve fuzzy matching of field names during AutoFlex
// expand/flatten operations.
func NewFieldNamePrefixOptionsFunc(s string) AutoFlexOptionsFunc {
	return func(o *AutoFlexOptions) {
		o.fieldNamePrefix = s
	}
}

// AddIgnoredField appends s to the list of ignored field names
func (o *AutoFlexOptions) AddIgnoredField(s string) {
	o.ignoredFieldNames = append(o.ignoredFieldNames, s)
}

// SetIgnoredFields replaces the list of ignored field names
//
// To preseve existing items in the list, use the AddIgnoredField
// method instead.
func (o *AutoFlexOptions) SetIgnoredFields(fields []string) {
	o.ignoredFieldNames = fields
}

// IsIgnoredField returns true if s is in the list of ignored field names
func (o *AutoFlexOptions) IsIgnoredField(s string) bool {
	for _, name := range o.ignoredFieldNames {
		if s == name {
			return true
		}
	}
	return false
}
