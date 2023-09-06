// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nullable

import (
	"errors"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
)

func TestNullableBool(t *testing.T) {
	t.Parallel()

	cases := []struct {
		val           string
		expectNull    bool
		expectedValue bool
		expectedErr   error
	}{
		{
			val:           "true",
			expectNull:    false,
			expectedValue: true,
		},
		{
			val:           "false",
			expectNull:    false,
			expectedValue: false,
		},
		{
			val:           "1",
			expectNull:    false,
			expectedValue: true,
		},
		{
			val:           "0",
			expectNull:    false,
			expectedValue: false,
		},
		{
			val:           "",
			expectNull:    true,
			expectedValue: false,
		},
		{
			val:           "A",
			expectNull:    false,
			expectedValue: false,
			expectedErr:   strconv.ErrSyntax,
		},
	}

	for i, tc := range cases {
		v := Bool(tc.val)

		if null := v.IsNull(); null != tc.expectNull {
			t.Fatalf("expected test case %d IsNull to return %t, got %t", i, null, tc.expectNull)
		}

		value, null, err := v.Value()
		if value != tc.expectedValue {
			t.Fatalf("expected test case %d Value to be %t, got %t", i, tc.expectedValue, value)
		}
		if null != tc.expectNull {
			t.Fatalf("expected test case %d Value null flag to be %t, got %t", i, tc.expectNull, null)
		}
		if tc.expectedErr == nil && err != nil {
			t.Fatalf("expected test case %d to succeed, got error %s", i, err)
		}
		if tc.expectedErr != nil {
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected test case %d to have error matching \"%s\", got %s", i, tc.expectedErr, err)
			}
		}
	}
}

func TestValidationBool(t *testing.T) {
	t.Parallel()

	runValidationTestCases(t, []testCase{
		{
			val: "true",
			f:   ValidateTypeStringNullableBool,
		},
		{
			val:             "1",
			f:               ValidateTypeStringNullableBool,
			expectedWarning: regexache.MustCompile(`^\w+: the use of values other than "true" and "false" is deprecated and will be removed in a future version of the provider$`),
		},
		{
			val:         "A",
			f:           ValidateTypeStringNullableBool,
			expectedErr: regexache.MustCompile(`^\w+: cannot parse 'A' as boolean: .+$`),
		},
		{
			val:         1,
			f:           ValidateTypeStringNullableBool,
			expectedErr: regexache.MustCompile(`^expected type of \w+ to be string$`),
		},
	})
}

func TestDiffSuppressBool(t *testing.T) {
	t.Parallel()

	cases := []struct {
		old, new   string
		equivalent bool
	}{
		// Truthy values
		{
			old:        "true",
			new:        "1",
			equivalent: true,
		},
		{
			old:        "1",
			new:        "true",
			equivalent: true,
		},

		// Falsey values
		{
			old:        "false",
			new:        "0",
			equivalent: true,
		},
		{
			old:        "0",
			new:        "false",
			equivalent: true,
		},

		// Truthy -> Falsey
		{
			old:        "true",
			new:        "false",
			equivalent: false,
		},
		{
			old:        "true",
			new:        "0",
			equivalent: false,
		},
		{
			old:        "1",
			new:        "false",
			equivalent: false,
		},
		{
			old:        "1",
			new:        "0",
			equivalent: false,
		},

		// Falsey -> Truthy
		{
			old:        "false",
			new:        "true",
			equivalent: false,
		},
		{
			old:        "false",
			new:        "1",
			equivalent: false,
		},
		{
			old:        "0",
			new:        "true",
			equivalent: false,
		},
		{
			old:        "0",
			new:        "1",
			equivalent: false,
		},

		// Null -> Truthy
		{
			old:        "",
			new:        "true",
			equivalent: false,
		},
		{
			old:        "",
			new:        "1",
			equivalent: false,
		},

		// Null -> Falsey
		{
			old:        "",
			new:        "false",
			equivalent: false,
		},
		{
			old:        "",
			new:        "0",
			equivalent: false,
		},

		// Truthy -> Null
		{
			old:        "true",
			new:        "",
			equivalent: false,
		},
		{
			old:        "1",
			new:        "",
			equivalent: false,
		},

		// Falsey -> Null
		{
			old:        "false",
			new:        "",
			equivalent: false,
		},
		{
			old:        "0",
			new:        "",
			equivalent: false,
		},

		// Null => Null
		{
			old:        "",
			new:        "",
			equivalent: true,
		},
	}

	for i, tc := range cases {
		v := DiffSuppressNullableBool("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !v {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && v {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}

func TestDiffSuppressBoolFalseAsNull(t *testing.T) {
	t.Parallel()

	cases := []struct {
		old, new   string
		equivalent bool
	}{
		// Falsey -> Null
		{
			old:        "false",
			new:        "",
			equivalent: true,
		},
		{
			old:        "0",
			new:        "",
			equivalent: true,
		},

		// Null -> Falsey
		{
			old:        "",
			new:        "false",
			equivalent: true,
		},
		{
			old:        "",
			new:        "0",
			equivalent: true,
		},

		// Null => Null
		{
			old:        "",
			new:        "",
			equivalent: true,
		},

		// Truthy values
		{
			old:        "true",
			new:        "1",
			equivalent: false,
		},
		{
			old:        "1",
			new:        "true",
			equivalent: false,
		},

		// Falsey values
		{
			old:        "false",
			new:        "0",
			equivalent: false,
		},
		{
			old:        "0",
			new:        "false",
			equivalent: false,
		},

		// Truthy -> Falsey
		{
			old:        "true",
			new:        "false",
			equivalent: false,
		},
		{
			old:        "true",
			new:        "0",
			equivalent: false,
		},
		{
			old:        "1",
			new:        "false",
			equivalent: false,
		},
		{
			old:        "1",
			new:        "0",
			equivalent: false,
		},

		// Falsey -> Truthy
		{
			old:        "false",
			new:        "true",
			equivalent: false,
		},
		{
			old:        "false",
			new:        "1",
			equivalent: false,
		},
		{
			old:        "0",
			new:        "true",
			equivalent: false,
		},
		{
			old:        "0",
			new:        "1",
			equivalent: false,
		},

		// Null -> Truthy
		{
			old:        "",
			new:        "true",
			equivalent: false,
		},
		{
			old:        "",
			new:        "1",
			equivalent: false,
		},

		// Truthy -> Null
		{
			old:        "true",
			new:        "",
			equivalent: false,
		},
		{
			old:        "1",
			new:        "",
			equivalent: false,
		},
	}

	for i, tc := range cases {
		v := DiffSuppressNullableBoolFalseAsNull("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !v {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && v {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}
