// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	"fmt"
	"strings"

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
		"Expected type to be "+expected,
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

func withPath(d diag.Diagnostic, path cty.Path) diag.Diagnostic {
	d.AttributePath = path
	return d
}

// newAttributeConflictsError is included for use with NewAttributeConflictsWillBeError.
// The typical behavior is covered using the schema ConflictsWith parameter.
func newAttributeConflictsError(path, otherPath cty.Path) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		path,
		"Invalid Attribute Combination",
		fmt.Sprintf("Attribute %q cannot be specified when %q is specified.",
			PathString(path),
			PathString(otherPath),
		),
	)
}

// NewAttributeConflictsWhenError returns an error diagnostic indicating that the attribute at the given path cannot be
// specified when the attribute at otherPath has the given value.
func NewAttributeConflictsWhenError(path, otherPath cty.Path, otherValue string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		path,
		"Invalid Attribute Combination",
		fmt.Sprintf("Attribute %q cannot be specified when %q is %q.",
			PathString(path),
			PathString(otherPath),
			otherValue,
		),
	)
}

// NewAttributeRequiredWhenError returns an error diagnostic indicating that the attribute at neededPath is required when the
// attribute at otherPath has the given value.
func NewAttributeRequiredWhenError(neededPath, otherPath cty.Path, value string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		otherPath,
		"Invalid Attribute Combination",
		fmt.Sprintf("Attribute %q must be specified when %q is %q.",
			PathString(neededPath),
			PathString(otherPath),
			value,
		),
	)
}

// NewAttributeConflictsWillBeError returns a warning diagnostic indicating that the attribute at the given path cannot be
// specified when the attribute at otherPath is set.
// This is intended to be used for situations where the conflict will become an error in a future release.
func NewAttributeConflictsWillBeError(path, otherPath cty.Path) diag.Diagnostic {
	return willBeError(
		newAttributeConflictsError(path, otherPath),
	)
}

// NewAttributeConflictsWhenWillBeError returns a warning diagnostic indicating that the attribute at the given path cannot be
// specified when the attribute at otherPath has the given value.
// This is intended to be used for situations where the conflict will become an error in a future release.
func NewAttributeConflictsWhenWillBeError(path, otherPath cty.Path, otherValue string) diag.Diagnostic {
	return willBeError(
		NewAttributeConflictsWhenError(path, otherPath, otherValue),
	)
}

func PathString(path cty.Path) string {
	var buf strings.Builder
	for i, step := range path {
		switch x := step.(type) {
		case cty.GetAttrStep:
			if i != 0 {
				buf.WriteString(".")
			}
			buf.WriteString(x.Name)
		case cty.IndexStep:
			val := x.Key
			typ := val.Type()
			var s string
			switch {
			case typ == cty.String:
				s = val.AsString()
			case typ == cty.Number:
				num := val.AsBigFloat()
				s = num.String()
			default:
				s = fmt.Sprintf("<unexpected index: %s>", typ.FriendlyName())
			}
			buf.WriteString(fmt.Sprintf("[%s]", s))
		default:
			if i != 0 {
				buf.WriteString(".")
			}
			buf.WriteString(fmt.Sprintf("<unexpected step: %[1]T %[1]v>", x))
		}
	}
	return buf.String()
}

func errorToWarning(d diag.Diagnostic) diag.Diagnostic {
	d.Severity = diag.Warning
	return d
}

func willBeError(d diag.Diagnostic) diag.Diagnostic {
	d.Detail += "\n\nThis will be an error in a future release."
	return errorToWarning(d)
}
