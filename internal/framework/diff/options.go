// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diff

type ChangeOptionsFunc func(*changeOptions)

type changeOptions struct {
	ignoredFieldNames []string
}

func initChangeOptions() *changeOptions {
	return &changeOptions{
		ignoredFieldNames: defaultIgnoredFieldNames,
	}
}

var defaultIgnoredFieldNames = []string{
	"Tags",
	"TagsAll",
	"Timeouts",
}

// WithException specifies a field name to be ignored when calculating plan changes
func WithException(s string) ChangeOptionsFunc {
	return func(o *changeOptions) {
		o.ignoredFieldNames = append(o.ignoredFieldNames, s)
	}
}
