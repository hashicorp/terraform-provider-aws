// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Package lazyvalidation provides Terraform Plugin SDKv2 validation functions that defer
// regular expression compilation until the first time a value is validated, instead of
// compiling at package initialization.
//
// The package lives in the `lazy` directory but is named `lazyvalidation` so that it is
// distinguishable from other `validation` packages when imported.
package lazyvalidation

import (
	"fmt"
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// StringMatch is a lazy equivalent of `validation.StringMatch`.
// `pattern` is compiled on first use, not when the validator is created.
// If `message` is not empty, it is used in the error returned on a mismatch.
func StringMatch(pattern, message string) schema.SchemaValidateFunc {
	return func(i any, k string) ([]string, []error) {
		v, ok := i.(string)
		if !ok {
			return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
		}

		re, matched, err := matchString(pattern, v)
		switch {
		case err != nil:
			return nil, []error{fmt.Errorf("failed to compile regex pattern %q: %v", pattern, err)}
		case !matched && message != "":
			return nil, []error{fmt.Errorf("invalid value for %s (%s)", k, message)}
		case !matched:
			return nil, []error{fmt.Errorf("expected value of %s to match regular expression %q, got %q", k, re.String(), v)}
		}
		return nil, nil
	}
}

// StringMatchDiag is a lazy equivalent of `validation.StringMatchDiag`.
// `pattern` is compiled on first use, not when the validator is created.
// If `message` is not empty, it is used in the diagnostic returned on a mismatch.
func StringMatchDiag(pattern, message string) schema.SchemaValidateDiagFunc {
	return func(i any, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		v, ok := i.(string)
		if !ok {
			return append(diags, errs.NewIncorrectValueTypeAttributeError(path, "string"))
		}

		re, matched, err := matchString(pattern, v)
		switch {
		case err != nil:
			return append(diags, errs.NewAttributeErrorDiagnostic(
				path,
				"Invalid Validator Configuration",
				"An unexpected error occurred while validating the attribute value. "+
					"This is always an error in the provider. "+
					"Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("The regex pattern %q failed to compile: %v", pattern, err),
			))
		case !matched && message != "":
			return append(diags, errs.NewInvalidValueAttributeError(
				path,
				"Value is invalid: "+message,
			))
		case !matched:
			return append(diags, errs.NewInvalidValueAttributeError(
				path,
				fmt.Sprintf("Expected value to match regular expression %q, got %q", re.String(), v),
			))
		}
		return diags
	}
}

// matchString reports whether `v` matches `pattern`.
// `err` is non-nil only if `pattern` fails to compile, in which case `re` is nil.
func matchString(pattern, v string) (re *regexp.Regexp, matched bool, err error) {
	re, err = regexacheCompile(pattern)
	if err != nil {
		return nil, false, err
	}
	return re, re.MatchString(v), nil
}

// regexacheCompile compiles `pattern` using `regexache`, which caches compiled expressions.
// `regeaxache` does not provide a `Compile` function that returns an error, so similate it here.
// `regexache.MustCompile` panics on an invalid pattern, so the panic is recovered and converted to an error.
// It never returns a nil `*regexp.Regexp` without a non-nil error.
func regexacheCompile(pattern string) (re *regexp.Regexp, err error) {
	defer func() {
		if r := recover(); r != nil {
			re = nil
			// Get the error message directly instead of relying on parsing the panic value
			if _, err = regexp.Compile(pattern); err == nil {
				// `regexache` panicked on a pattern the standard library accepts.
				// This can only happen if `regexache` has certain environment variables set.
				// Don't return a nil `*regexp.Regexp` without an error.
				err = fmt.Errorf("regexache failed to compile pattern: %v", r)
			}
		}
	}()

	return regexache.MustCompile(pattern), nil
}
