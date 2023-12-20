// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nullable

import (
	"errors"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
)

func TestNullableFloat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		val           string
		expectNull    bool
		expectedValue float64
		expectedErr   error
	}{
		{
			val:           "1",
			expectNull:    false,
			expectedValue: 1,
		},
		{
			val:           "1.1",
			expectNull:    false,
			expectedValue: 1.1,
		},
		{
			val:           "0",
			expectNull:    false,
			expectedValue: 0,
		},
		{
			val:           "",
			expectNull:    true,
			expectedValue: 0,
		},
		{
			val:           "A",
			expectNull:    false,
			expectedValue: 0,
			expectedErr:   strconv.ErrSyntax,
		},
	}

	for i, tc := range cases {
		v := Float(tc.val)

		if null := v.IsNull(); null != tc.expectNull {
			t.Fatalf("expected test case %d IsNull to return %t, got %t", i, null, tc.expectNull)
		}

		value, null, err := v.Value()
		if value != tc.expectedValue {
			t.Fatalf("expected test case %d Value to be %f, got %f", i, tc.expectedValue, value)
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

func TestValidationFloat(t *testing.T) {
	t.Parallel()

	runValidationTestCases(t, []testCase{
		{
			val: "1",
			f:   ValidateTypeStringNullableFloat,
		},
		{
			val: "1.1",
			f:   ValidateTypeStringNullableFloat,
		},
		{
			val: "0",
			f:   ValidateTypeStringNullableFloat,
		},
		{
			val:         "A",
			f:           ValidateTypeStringNullableFloat,
			expectedErr: regexache.MustCompile(`^\w+: cannot parse 'A' as float: .+$`),
		},
		{
			val:         1,
			f:           ValidateTypeStringNullableFloat,
			expectedErr: regexache.MustCompile(`^expected type of \w+ to be string$`),
		},
	})
}
