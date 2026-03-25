// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkdiag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
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

func severityFilter(s diag.Severity) tfslices.Predicate[diag.Diagnostic] {
	return func(d diag.Diagnostic) bool {
		return d.Severity == s
	}
}

func DiagnosticError(diag diag.Diagnostic) error {
	return errors.New(DiagnosticString(diag))
}

// DiagnosticsError returns an error containing all Diagnostic with SeverityError
func DiagnosticsError(diags diag.Diagnostics) error {
	var errs []error

	for _, d := range Errors(diags) {
		errs = append(errs, DiagnosticError(d))
	}

	return errors.Join(errs...)
}

// DiagnosticString formats a Diagnostic
// If there is no `Detail`, only prints summary, otherwise prints both
func DiagnosticString(d diag.Diagnostic) string {
	var buf strings.Builder

	fmt.Fprint(&buf, d.Summary)
	if d.Detail != "" {
		fmt.Fprintf(&buf, "\n\n%s", d.Detail)
	}
	if len(d.AttributePath) > 0 {
		fmt.Fprintf(&buf, "\n\nPath: %s", pathString(d.AttributePath))
	}

	return buf.String()
}

func pathString(path cty.Path) string {
	var buf strings.Builder
	for i, step := range path {
		switch x := step.(type) {
		case cty.GetAttrStep:
			if i != 0 {
				buf.WriteString(".")
			}
			buf.WriteString(x.Name)
		case cty.IndexStep:
			var s string
			switch val := x.Key; val.Type() {
			case cty.String:
				s = val.AsString()
			case cty.Number:
				num := val.AsBigFloat()
				s = num.String()
			default:
				s = fmt.Sprintf("<unexpected index: %s>", val.Type().FriendlyName())
			}
			fmt.Fprintf(&buf, "[%s]", s)
		default:
			if i != 0 {
				buf.WriteString(".")
			}
			fmt.Fprintf(&buf, "<unexpected step: %[1]T %[1]v>", x)
		}
	}
	return buf.String()
}

// DiagnosticsString formats a Diagnostics
func DiagnosticsString(diags diag.Diagnostics) string {
	var buf strings.Builder

	for _, d := range diags {
		fmt.Fprintln(&buf, DiagnosticString(d))
	}

	return buf.String()
}

func NewResourceNotFoundWarningDiagnostic(err error) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "AWS resource not found during refresh",
		Detail:   "Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original error: " + err.Error(),
	}
}
