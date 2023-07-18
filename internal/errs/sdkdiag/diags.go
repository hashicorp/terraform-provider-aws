// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkdiag

import (
	"errors"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// Errors returns all the Diagnostic in Diagnostics that are SeverityError.
// Equivalent to terraform-plugin-framework/diag/diags.Errors()
func Errors(diags diag.Diagnostics) diag.Diagnostics {
	return tfslices.Filter(diags, severityFilter(diag.Error))
}

// Warnings returns all the Diagnostic in Diagnostics that are SeverityWarning.
// Equivalent to terraform-plugin-framework/diag/diags.Warnings()
func Warnings(diags diag.Diagnostics) diag.Diagnostics {
	return tfslices.Filter(diags, severityFilter(diag.Warning))
}

func severityFilter(s diag.Severity) tfslices.FilterFunc[diag.Diagnostic] {
	return func(d diag.Diagnostic) bool {
		return d.Severity == s
	}
}

// DiagnosticsError returns an error containing all Diagnostic with SeverityError
func DiagnosticsError(diags diag.Diagnostics) error {
	if !diags.HasError() {
		return nil
	}

	errDiags := Errors(diags)

	if len(errDiags) == 1 {
		return diagnosticError(errDiags[0])
	}

	var errs error
	for _, d := range errDiags {
		errs = multierror.Append(errs, diagnosticError(d))
	}

	return errs
}

func diagnosticError(diag diag.Diagnostic) error {
	return errors.New(DiagnosticString(diag))
}

// DiagnosticString formats a Diagnostic
// If there is no `Detail`, only prints summary, otherwise prints both
func DiagnosticString(d diag.Diagnostic) string {
	if d.Detail == "" {
		return d.Summary
	}
	return fmt.Sprintf("%s\n\n%s", d.Summary, d.Detail)
}
