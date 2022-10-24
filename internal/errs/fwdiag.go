package errs

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func NewDiagnosticsError(diags diag.Diagnostics) error {
	if !diags.HasError() {
		return nil
	}

	var errs *multierror.Error

	for _, v := range diags.Errors() {
		errs = multierror.Append(errs, fmt.Errorf("%s\n\n%s", v.Summary(), v.Detail()))
	}

	return errs.ErrorOrNil()
}

func NewResourceNotFoundWarningDiagnostic(err error) diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		"AWS resource not found during refresh",
		fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original error: %s", err.Error()),
	)
}
