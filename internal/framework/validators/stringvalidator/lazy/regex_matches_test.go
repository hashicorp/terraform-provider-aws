// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lazystringvalidator

import (
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/YakDriver/regexache"
	fwstringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringValidator is satisfied by both this package's validator and the upstream `stringvalidator.RegexMatches`.
type stringValidator interface {
	validator.String
	function.StringParameterValidator
}

func validateString(t *testing.T, v stringValidator, val types.String) diag.Diagnostics {
	t.Helper()

	request := validator.StringRequest{
		Path:           path.Root("test"),
		PathExpression: path.MatchRoot("test"),
		ConfigValue:    val,
	}
	response := validator.StringResponse{}
	v.ValidateString(t.Context(), request, &response)

	return response.Diagnostics
}

func validateParameterString(t *testing.T, v stringValidator, val types.String) *function.FuncError {
	t.Helper()

	request := function.StringParameterValidatorRequest{
		ArgumentPosition: 1,
		Value:            val,
	}
	response := function.StringParameterValidatorResponse{}
	v.ValidateParameterString(t.Context(), request, &response)

	return response.Error
}

// TestRegexMatchesValidator is ported from the upstream `TestRegexMatchesValidator`.
func TestRegexMatchesValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		regexp      string
		expectError bool
	}
	tests := map[string]testCase{
		"unknown String": {
			val:    types.StringUnknown(),
			regexp: `^o[j-l]?$`,
		},
		"null String": {
			val:    types.StringNull(),
			regexp: `^o[j-l]?$`,
		},
		"valid String": {
			val:    types.StringValue("ok"),
			regexp: `^o[j-l]?$`,
		},
		"invalid String": {
			val:         types.StringValue("not ok"),
			regexp:      `^o[j-l]?$`,
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run("ValidateString - "+name, func(t *testing.T) {
			t.Parallel()

			diags := validateString(t, RegexMatches(test.regexp, ""), test.val)

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", diags)
			}
		})

		t.Run("ValidateParameterString - "+name, func(t *testing.T) {
			t.Parallel()

			err := validateParameterString(t, RegexMatches(test.regexp, ""), test.val)

			if err == nil && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if err != nil && !test.expectError {
				t.Fatalf("got unexpected error: %s", err)
			}
		})
	}
}

// TestRegexMatchesValidator_ParityWithUpstream verifies that the lazy validator behaves
// identically to `stringvalidator.RegexMatches` for every valid pattern.
func TestRegexMatchesValidator_ParityWithUpstream(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val     types.String
		pattern string
		message string
	}
	tests := map[string]testCase{
		"unknown":                  {val: types.StringUnknown(), pattern: `^o[j-l]?$`},
		"null":                     {val: types.StringNull(), pattern: `^o[j-l]?$`},
		"match":                    {val: types.StringValue("ok"), pattern: `^o[j-l]?$`},
		"mismatch":                 {val: types.StringValue("not ok"), pattern: `^o[j-l]?$`},
		"match with message":       {val: types.StringValue("ok"), pattern: `^o[j-l]?$`, message: "must be o, oj, ok, or ol"},
		"mismatch with message":    {val: types.StringValue("not ok"), pattern: `^o[j-l]?$`, message: "must be o, oj, ok, or ol"},
		"empty string matches":     {val: types.StringValue(""), pattern: `^$`},
		"empty string mismatches":  {val: types.StringValue(""), pattern: `^.+$`},
		"unanchored match":         {val: types.StringValue("foobar"), pattern: `foo`},
		"unanchored mismatch":      {val: types.StringValue("bar"), pattern: `foo`},
		"anchored partial":         {val: types.StringValue("foobar"), pattern: `^foo$`},
		"pattern with quote":       {val: types.StringValue("bar"), pattern: `^"foo"$`},
		"value with special chars": {val: types.StringValue("a\"b\\c\nd"), pattern: `^x$`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lazy := RegexMatches(test.pattern, test.message)
			upstream := fwstringvalidator.RegexMatches(regexache.MustCompile(test.pattern), test.message) // nosemgrep: ci.semgrep.framework.use-lazystringvalidator-regex-matches // Compares lazy behavior with the upstream validator.

			ctx := t.Context()
			if got, want := lazy.Description(ctx), upstream.Description(ctx); got != want {
				t.Errorf("Description: got %q, want %q", got, want)
			}
			if got, want := lazy.MarkdownDescription(ctx), upstream.MarkdownDescription(ctx); got != want {
				t.Errorf("MarkdownDescription: got %q, want %q", got, want)
			}

			if got, want := validateString(t, lazy, test.val), validateString(t, upstream, test.val); !got.Equal(want) {
				t.Errorf("ValidateString diagnostics differ:\n got: %v\nwant: %v", got, want)
			}

			if got, want := validateParameterString(t, lazy, test.val), validateParameterString(t, upstream, test.val); !got.Equal(want) {
				t.Errorf("ValidateParameterString error differs:\n got: %v\nwant: %v", got, want)
			}
		})
	}
}

func TestRegexMatchesValidator_Description(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		pattern string
		message string
		want    string
	}{
		"default": {
			pattern: `^o[j-l]?$`,
			want:    "value must match regular expression '^o[j-l]?$'",
		},
		"custom message": {
			pattern: `^o[j-l]?$`,
			message: "must be o, oj, ok, or ol",
			want:    "must be o, oj, ok, or ol",
		},
		// The description is built from the pattern text, so it must not compile the pattern.
		"invalid pattern, default": {
			pattern: `(`,
			want:    "value must match regular expression '('",
		},
		"invalid pattern, custom message": {
			pattern: `(`,
			message: "friendly message",
			want:    "friendly message",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := RegexMatches(test.pattern, test.message)

			if got := v.Description(t.Context()); got != test.want {
				t.Errorf("Description: got %q, want %q", got, test.want)
			}
			if got := v.MarkdownDescription(t.Context()); got != test.want {
				t.Errorf("MarkdownDescription: got %q, want %q", got, test.want)
			}
		})
	}
}

func TestRegexMatchesValidator_CustomMessage(t *testing.T) {
	t.Parallel()

	const message = "must be o, oj, ok, or ol"
	v := RegexMatches(`^o[j-l]?$`, message)

	diags := validateString(t, v, types.StringValue("not ok"))
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d: %v", len(diags), diags)
	}
	if !strings.Contains(diags[0].Detail(), message) {
		t.Errorf("expected diagnostic detail %q to contain %q", diags[0].Detail(), message)
	}

	err := validateParameterString(t, v, types.StringValue("not ok"))
	if err == nil {
		t.Fatal("expected error, got no error")
	}
	if !strings.Contains(err.Text, message) {
		t.Errorf("expected error text %q to contain %q", err.Text, message)
	}
}

// Constructing the validator, or asking for its description, must neither compile the pattern nor panic.
func TestRegexMatchesValidator_InvalidPatternIsNotCompiledAtConstruction(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	v := RegexMatches(`(`, "")
	_ = v.Description(t.Context())
	_ = v.MarkdownDescription(t.Context())
}

func TestRegexMatchesValidator_InvalidPattern(t *testing.T) {
	t.Parallel()

	const pattern = `(`
	_, compileErr := regexp.Compile(pattern)
	if compileErr == nil {
		t.Fatalf("expected %q to be an invalid pattern", pattern)
	}
	wantDetail := `The regular expression "(" is invalid: ` + compileErr.Error()

	t.Run("ValidateString", func(t *testing.T) {
		t.Parallel()

		diags := validateString(t, RegexMatches(pattern, ""), types.StringValue("bar"))

		if len(diags) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d: %v", len(diags), diags)
		}
		d := diags[0]
		if got, want := d.Severity(), diag.SeverityError; got != want {
			t.Errorf("Severity: got %v, want %v", got, want)
		}
		if got, want := d.Summary(), "Invalid Validator Usage"; got != want {
			t.Errorf("Summary: got %q, want %q", got, want)
		}
		if got := d.Detail(); !strings.Contains(got, `"RegexMatches"`) || !strings.Contains(got, wantDetail) {
			t.Errorf("Detail: got %q, want it to contain %q and %q", got, `"RegexMatches"`, wantDetail)
		}
		if withPath, ok := d.(diag.DiagnosticWithPath); !ok {
			t.Errorf("expected a diagnostic with a path, got %T", d)
		} else if got, want := withPath.Path(), path.Root("test"); !got.Equal(want) {
			t.Errorf("Path: got %s, want %s", got, want)
		}
	})

	t.Run("ValidateParameterString", func(t *testing.T) {
		t.Parallel()

		err := validateParameterString(t, RegexMatches(pattern, ""), types.StringValue("bar"))

		if err == nil {
			t.Fatal("expected error, got no error")
		}
		if !strings.Contains(err.Text, `"RegexMatches"`) || !strings.Contains(err.Text, wantDetail) {
			t.Errorf("Text: got %q, want it to contain %q and %q", err.Text, `"RegexMatches"`, wantDetail)
		}
		if err.FunctionArgument == nil || *err.FunctionArgument != 1 {
			t.Errorf("FunctionArgument: got %v, want 1", err.FunctionArgument)
		}
	})

	// A custom message only changes the description of a mismatch, not the invalid-pattern error.
	t.Run("custom message does not mask error", func(t *testing.T) {
		t.Parallel()

		diags := validateString(t, RegexMatches(pattern, "friendly message"), types.StringValue("bar"))

		if len(diags) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d: %v", len(diags), diags)
		}
		if got := diags[0].Detail(); !strings.Contains(got, wantDetail) {
			t.Errorf("Detail: got %q, want it to contain %q", got, wantDetail)
		}
	})
}

// The validators may be called concurrently on first use.
// Run with `-race` to detect data races.
func TestRegexMatchesValidator_ConcurrentFirstUse(t *testing.T) {
	t.Parallel()

	const goroutines = 32

	t.Run("valid pattern", func(t *testing.T) {
		t.Parallel()

		v := RegexMatches(`^concurrent-valid-[0-9]+$`, "")

		var wg sync.WaitGroup
		errs := make(chan string, goroutines*2)
		for range goroutines {
			wg.Go(func() {
				if diags := validateString(t, v, types.StringValue("concurrent-valid-42")); diags.HasError() {
					errs <- "ValidateString: " + diags.Errors()[0].Detail()
				}
				if err := validateParameterString(t, v, types.StringValue("concurrent-valid-42")); err != nil {
					errs <- "ValidateParameterString: " + err.Text
				}
			})
		}
		wg.Wait()
		close(errs)

		for e := range errs {
			t.Error(e)
		}
	})

	t.Run("invalid pattern", func(t *testing.T) {
		t.Parallel()

		v := RegexMatches(`(concurrent-invalid`, "")

		var wg sync.WaitGroup
		errs := make(chan string, goroutines*2)
		for range goroutines {
			wg.Go(func() {
				if diags := validateString(t, v, types.StringValue("x")); !diags.HasError() {
					errs <- "ValidateString: expected error, got no error"
				}
				if err := validateParameterString(t, v, types.StringValue("x")); err == nil {
					errs <- "ValidateParameterString: expected error, got no error"
				}
			})
		}
		wg.Wait()
		close(errs)

		for e := range errs {
			t.Error(e)
		}
	})
}

func TestRegexacheCompile(t *testing.T) {
	t.Parallel()

	t.Run("valid pattern", func(t *testing.T) {
		t.Parallel()

		re, err := regexacheCompile(`^foo[0-9]+$`)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if re == nil {
			t.Fatal("expected non-nil regexp")
		}
		if !re.MatchString("foo123") {
			t.Errorf("expected %q to match foo123", re)
		}
		if re.MatchString("bar") {
			t.Errorf("expected %q not to match bar", re)
		}
	})

	t.Run("invalid pattern returns stdlib error", func(t *testing.T) {
		t.Parallel()

		const pattern = `(`

		re, err := regexacheCompile(pattern)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if re != nil {
			t.Errorf("expected nil regexp, got %q", re)
		}

		_, wantErr := regexp.Compile(pattern)
		if err.Error() != wantErr.Error() {
			t.Errorf("expected error %q, got %q", wantErr, err)
		}
	})

	t.Run("repeated calls return same compiled regexp", func(t *testing.T) {
		t.Parallel()

		if !regexache.IsCachingEnabled() {
			t.Skip("regexache caching is disabled")
		}

		const pattern = `^[a-z]+$`

		first, err := regexacheCompile(pattern)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		second, err := regexacheCompile(pattern)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if first != second {
			t.Errorf("expected the same *regexp.Regexp for repeated calls, got %p and %p", first, second)
		}
		if !first.MatchString("abc") {
			t.Errorf("expected %q to match abc", first)
		}
	})
}
