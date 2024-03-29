// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"strings"
) // StateTrimSpace is a StateFunc that trims extraneous whitespace from strings.
// This prevents differences caused by an API canonicalizing a string with a
// trailing newline character removed.
func StateTrimSpace(v interface{}) string {
	s, ok := v.(string)

	if !ok {
		return ""
	}

	return strings.TrimSpace(s)
}
