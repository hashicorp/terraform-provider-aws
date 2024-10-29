// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwdiag

import (
	"errors"
	"fmt"
	"strings"

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
	var buf strings.Builder

	fmt.Fprint(&buf, d.Summary())
	if d.Detail() != "" {
		fmt.Fprintf(&buf, "\n\n%s", d.Detail())
	}
	if withPath, ok := d.(diag.DiagnosticWithPath); ok {
		fmt.Fprintf(&buf, "\n%s", withPath.Path().String())
	}

	return buf.String()
}

func NewResourceNotFoundWarningDiagnostic(err error) diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		"AWS resource not found during refresh",
		"Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original error: "+err.Error(),
	)
}

func AsError[T any](x T, diags diag.Diagnostics) (T, error) {
	return x, DiagnosticsError(diags)
}

// DiagnosticsString formats a Diagnostics
func DiagnosticsString(diags diag.Diagnostics) string {
	var buf strings.Builder

	for _, d := range diags {
		fmt.Fprintln(&buf, DiagnosticString(d))
	}

	return buf.String()
}
