// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lazystringvalidator

import (
	"context"
	"fmt"
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatorfuncerr"
	fwstringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = regexMatchesValidator{}

var _ function.StringParameterValidator = regexMatchesValidator{}

type regexMatchesValidator struct {
	regexp  string
	message string
}

func (validator regexMatchesValidator) Description(_ context.Context) string {
	if validator.message != "" {
		return validator.message
	}
	return fmt.Sprintf("value must match regular expression '%s'", validator.regexp)
}

func (validator regexMatchesValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (v regexMatchesValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	re, err := regexacheCompile(v.regexp)
	if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidValidatorUsageDiagnostic(
			request.Path,
			"RegexMatches",
			v.invalidUsageMessage(err),
		))
		return
	}

	fwstringvalidator.RegexMatches(re, v.message).ValidateString(ctx, request, response)
}

func (v regexMatchesValidator) ValidateParameterString(ctx context.Context, request function.StringParameterValidatorRequest, response *function.StringParameterValidatorResponse) {
	if request.Value.IsNull() || request.Value.IsUnknown() {
		return
	}

	re, err := regexacheCompile(v.regexp)
	if err != nil {
		response.Error = validatorfuncerr.InvalidValidatorUsageFuncError(
			request.ArgumentPosition,
			"RegexMatches",
			v.invalidUsageMessage(err),
		)
		return
	}

	fwstringvalidator.RegexMatches(re, v.message).ValidateParameterString(ctx, request, response)
}

func (validator regexMatchesValidator) invalidUsageMessage(err error) string {
	return fmt.Sprintf("The regular expression %q is invalid: %v", validator.regexp, err)
}

// RegexMatches is a lazy equivalent of `stringvalidator.RegexMatches`.
// `pattern` is compiled on first use, not when the validator is created.
// Null (unconfigured) and unknown (known after apply) values are skipped.
// Optionally an error message can be provided to return something friendlier
// than "value must match regular expression 'regexp'".
func RegexMatches(regexp, message string) regexMatchesValidator {
	return regexMatchesValidator{
		regexp:  regexp,
		message: message,
	}
}

// regexacheCompile compiles `pattern` using `regexache`, which caches compiled expressions.
// `regexache` does not provide a `Compile` function that returns an error, so simulate it here.
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
