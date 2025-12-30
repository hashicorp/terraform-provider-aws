// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package fwdiag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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

func NewCreatingResourceIDErrorDiagnostic(err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Creating Resource ID",
		err.Error(),
	)
}

func NewParsingResourceIDErrorDiagnostic(err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Parsing Resource ID",
		err.Error(),
	)
}

func NewResourceNotFoundWarningDiagnostic(err error) diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		"AWS resource not found during refresh",
		"Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original error: "+err.Error(),
	)
}

func NewListResultErrorDiagnostic(err error) list.ListResult {
	return list.ListResult{
		Diagnostics: diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Error Listing Remote Resources",
				err.Error(),
			),
		},
	}
}

func NewListResultSDKDiagnostics(diags sdkdiag.Diagnostics) list.ListResult {
	return list.ListResult{
		Diagnostics: FromSDKDiagnostics(diags),
	}
}

func FromSDKDiagnostics(diags sdkdiag.Diagnostics) diag.Diagnostics {
	return tfslices.ApplyToAll(diags, FromSDKDiagnostic)
}

func FromSDKDiagnostic(d sdkdiag.Diagnostic) diag.Diagnostic {
	switch d.Severity {
	case sdkdiag.Error:
		return diag.NewErrorDiagnostic(
			d.Summary,
			d.Detail,
		)
	case sdkdiag.Warning:
		return diag.NewWarningDiagnostic(
			d.Summary,
			d.Detail,
		)
	}
	return nil
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
