// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

const (
	summaryInvalidValue     = "Invalid value"
	summaryInvalidValueType = "Invalid value type"
)

func NewIncorrectValueTypeAttributeError(path cty.Path, expected string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		path,
		summaryInvalidValueType,
		fmt.Sprintf("Expected type to be %s", expected),
	)
}

func NewInvalidValueAttributeErrorf(path cty.Path, format string, a ...any) diag.Diagnostic {
	return NewInvalidValueAttributeError(
		path,
		fmt.Sprintf(format, a...),
	)
}

func NewInvalidValueAttributeError(path cty.Path, detail string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		path,
		summaryInvalidValue,
		detail,
	)
}

func NewAttributeErrorDiagnostic(path cty.Path, summary, detail string) diag.Diagnostic {
	return withPath(
		NewErrorDiagnostic(summary, detail),
		path,
	)
}

func NewAttributeWarningDiagnostic(path cty.Path, summary, detail string) diag.Diagnostic {
	return withPath(
		NewWarningDiagnostic(summary, detail),
		path,
	)
}

func NewErrorDiagnostic(summary, detail string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  summary,
		Detail:   detail,
	}
}

func NewWarningDiagnostic(summary, detail string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  summary,
		Detail:   detail,
	}
}

func FromAttributeError(path cty.Path, err error) diag.Diagnostic {
	return withPath(
		NewErrorDiagnostic(err.Error(), ""),
		path,
	)
}

func withPath(d diag.Diagnostic, path cty.Path) diag.Diagnostic {
	d.AttributePath = path
	return d
}
