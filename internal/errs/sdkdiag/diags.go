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
func DiagnosticsError(diags diag.Diagnostics) (errs error) {
	if !diags.HasError() {
		return
	}

	for _, d := range Errors(diags) {
		errs = multierror.Append(errs, errors.New(DiagnosticString(d)))
	}

	return
}

// DiagnosticString formats a Diagnostic
// If there is no `Detail`, only prints summary, otherwise prints both
func DiagnosticString(d diag.Diagnostic) string {
	if d.Detail == "" {
		return d.Summary
	}
	return fmt.Sprintf("%s\n\n%s", d.Summary, d.Detail)
}
