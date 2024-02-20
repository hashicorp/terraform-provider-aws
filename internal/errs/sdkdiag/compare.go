// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkdiag

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// Comparer is a Comparer function for use with cmp.Diff to compare two diag.Diagnostic values
func Comparer(l, r diag.Diagnostic) bool {
	if l.Severity != r.Severity {
		return false
	}
	if l.Summary != r.Summary {
		return false
	}
	if l.Detail != r.Detail {
		return false
	}

	lp := l.AttributePath
	rp := r.AttributePath
	if len(lp) != len(rp) {
		return false
	}
	return lp.Equals(rp)
}
