package sdkdiag

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func AppendWarningf(diags diag.Diagnostics, format string, a ...any) diag.Diagnostics {
	return append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  fmt.Sprintf(format, a...),
	})
}

func AppendErrorf(diags diag.Diagnostics, format string, a ...any) diag.Diagnostics {
	return append(diags, diag.Errorf(format, a...)...)
}

func AppendFromErr(diags diag.Diagnostics, err error) diag.Diagnostics {
	if err == nil {
		return diags
	}
	return append(diags, diag.FromErr(err)...)
}
