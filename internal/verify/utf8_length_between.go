// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"fmt"
	"unicode/utf8"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func StringUTF8LenBetween(minVal, maxVal int) schema.SchemaValidateDiagFunc {
	return func(v any, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		value, ok := v.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("expected a string, got %T", v),
				Detail:   fmt.Sprintf("expected a string, got %T", v),
			})
			return diags
		}

		utf8Length := utf8.RuneCountInString(value)
		if utf8Length < minVal || utf8Length > maxVal {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid character length",
				Detail:   fmt.Sprintf("expected length of pattern to be between %d and %d UTF-8 characters, got %d", minVal, maxVal, utf8Length),
			})
		}

		return diags
	}
}
