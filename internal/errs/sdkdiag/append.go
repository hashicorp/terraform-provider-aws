// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkdiag

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func AppendWarningf(diags diag.Diagnostics, format string, a ...any) diag.Diagnostics {
	return append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  fmt.Sprintf(format, a...),
	})
}

func AppendErrorf(diags diag.Diagnostics, format string, a ...any) diag.Diagnostics {
	return append(diags, diag.Errorf(format, a...)...) // nosemgrep:ci.semgrep.pluginsdk.avoid-diag_Errorf
}

func AppendFromErr(diags diag.Diagnostics, err error) diag.Diagnostics {
	if err == nil {
		return diags
	}
	return append(diags, diag.FromErr(err)...) // nosemgrep:ci.semgrep.pluginsdk.avoid-diag_FromErr
}

func WrapDiagsf(orig diag.Diagnostics, format string, a ...any) diag.Diagnostics {
	if len(orig) == 0 {
		return orig
	}

	msg := fmt.Sprintf(format, a...)
	return tfslices.ApplyToAll(orig, func(d diag.Diagnostic) diag.Diagnostic {
		return diag.Diagnostic{
			Severity:      d.Severity,
			Summary:       fmt.Sprintf("%s: %s", msg, d.Summary),
			Detail:        d.Detail,
			AttributePath: d.AttributePath,
		}
	})
}
