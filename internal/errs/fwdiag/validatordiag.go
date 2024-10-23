// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwdiag

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// NewAttributeRequiredWhenError returns an error diagnostic indicating that the attribute at neededPath is required when the
// attribute at otherPath has the given value.
func NewAttributeRequiredWhenError(neededPath, otherPath path.Path, value string) diag.Diagnostic {
	return validatordiag.InvalidAttributeCombinationDiagnostic(
		otherPath,
		fmt.Sprintf("Attribute %q must be specified when %q is %q.",
			neededPath.String(),
			otherPath.String(),
			value,
		),
	)
}

// NewAttributeConflictsWhenError returns an error diagnostic indicating that the attribute at the given path cannot be
// specified when the attribute at otherPath has the given value.
func NewAttributeConflictsWhenError(path, otherPath path.Path, otherValue string) diag.Diagnostic {
	return validatordiag.InvalidAttributeCombinationDiagnostic(
		path,
		fmt.Sprintf("Attribute %q cannot be specified when %q is %q.",
			path.String(),
			otherPath.String(),
			otherValue,
		),
	)
}
