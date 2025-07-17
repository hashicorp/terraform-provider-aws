// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwdiag

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// WarningInvalidAttributeCombinationDiagnostic returns a warning Diagnostic to be used when a schemavalidator of attributes is invalid.
func WarningInvalidAttributeCombinationDiagnostic(path path.Path, description string) diag.Diagnostic {
	return diag.NewAttributeWarningDiagnostic(
		path,
		"Invalid Attribute Combination",
		description+"\n\nThis will be an error in a future version of the provider",
	)
}
