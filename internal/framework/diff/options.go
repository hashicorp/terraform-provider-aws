// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diff

// ChangeOptionsFunc is a type alias for a changeOptions functional option
type ChangeOptionsFunc func(*changeOptions)

type changeOptions struct {
	ignoredFieldNames []string
}

// WithException specifies a field name to be ignored when calculating plan changes
func WithException(s string) ChangeOptionsFunc {
	return func(o *changeOptions) {
		o.ignoredFieldNames = append(o.ignoredFieldNames, s)
	}
}

func initChangeOptions(options []ChangeOptionsFunc) *changeOptions {
	o := changeOptions{
		ignoredFieldNames: defaultIgnoredFieldNames,
	}

	for _, opt := range options {
		opt(&o)
	}

	return &o
}

var defaultIgnoredFieldNames = []string{
	"Tags",
	"TagsAll",
	"Timeouts",
}
