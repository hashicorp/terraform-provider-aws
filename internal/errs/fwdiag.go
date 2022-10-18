package errs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func NewResourceNotFoundWarningDiagnostic(err error) diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		"AWS resource not found during refresh",
		fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original error: %s", err.Error()),
	)
}
