// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// nameValidator returns the ValidateFunc applied to the `name` attribute of the
// named `field_to_match` block, e.g. "single_header" or "single_query_argument".
func nameValidator(t *testing.T, block string) schema.SchemaValidateFunc {
	t.Helper()

	elem, ok := fieldToMatchBaseSchema().SchemaMap()[block].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("%s Elem is not a *schema.Resource", block)
	}

	validate := elem.SchemaMap()[names.AttrName].ValidateFunc
	if validate == nil {
		t.Fatalf("%s.name has no ValidateFunc", block)
	}

	return validate
}

func TestFieldToMatchSingleHeaderName(t *testing.T) {
	t.Parallel()

	validate := nameValidator(t, "single_header")

	testCases := map[string]struct {
		name    string
		wantErr bool
	}{
		"simple": {
			name: "user-agent",
		},
		"underscore": {
			name: "x_custom_header",
		},
		// The AWS WAF API documents SingleHeader.Name as `.*\S.*`, and the
		// CloudFront console creates rules using dotted pseudo-header names.
		"dotted pseudo-header": {
			name: "httprequest.headers.2.value",
		},
		"maximum length": {
			name: strings.Repeat("a", 64),
		},
		"minimum length": {
			name: "a",
		},
		"empty": {
			name:    "",
			wantErr: true,
		},
		"only whitespace": {
			name:    "   ",
			wantErr: true,
		},
		"too long": {
			name:    strings.Repeat("a", 65),
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, errs := validate(testCase.name, names.AttrName)

			if got, want := len(errs) > 0, testCase.wantErr; got != want {
				t.Errorf("got error = %t (%v), want error = %t", got, errs, want)
			}
		})
	}
}

func TestFieldToMatchSingleQueryArgumentName(t *testing.T) {
	t.Parallel()

	validate := nameValidator(t, "single_query_argument")

	testCases := map[string]struct {
		name    string
		wantErr bool
	}{
		"simple": {
			name: "session-id",
		},
		"dotted": {
			name: "user.id",
		},
		"maximum length": {
			name: strings.Repeat("a", 30),
		},
		"empty": {
			name:    "",
			wantErr: true,
		},
		"only whitespace": {
			name:    "  ",
			wantErr: true,
		},
		"too long": {
			name:    strings.Repeat("a", 31),
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, errs := validate(testCase.name, names.AttrName)

			if got, want := len(errs) > 0, testCase.wantErr; got != want {
				t.Errorf("got error = %t (%v), want error = %t", got, errs, want)
			}
		})
	}
}
