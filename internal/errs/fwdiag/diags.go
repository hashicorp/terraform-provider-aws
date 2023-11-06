// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwdiag

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// DiagnosticsError returns an error containing all Diagnostic with SeverityError
func DiagnosticsError(diags diag.Diagnostics) error {
	var errs []error

	for _, d := range diags.Errors() {
		errs = append(errs, errors.New(DiagnosticString(d)))
	}

	return errors.Join(errs...)
}

// DiagnosticString formats a Diagnostic
// If there is no `Detail`, only prints summary, otherwise prints both
func DiagnosticString(d diag.Diagnostic) string {
	if d.Detail() == "" {
		return d.Summary()
	}
	return fmt.Sprintf("%s\n\n%s", d.Summary(), d.Detail())
}

func NewResourceNotFoundWarningDiagnostic(err error) diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		"AWS resource not found during refresh",
		fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original error: %s", err.Error()),
	)
}
