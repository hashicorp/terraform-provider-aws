// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nullable

import (
	"errors"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
)

func TestNullableInt_64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		val           string
		expectNull    bool
		expectedValue int64
		expectedErr   error
	}{
		{
			val:           "1",
			expectNull:    false,
			expectedValue: 1,
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
		{
			val:           "-1",
			expectNull:    false,
			expectedValue: -1,
		},
		{
			val:           "2147483647", // max int32
			expectNull:    false,
			expectedValue: 2147483647,
		},
		{
			val:           "2147483648", // max int32
			expectNull:    false,
			expectedValue: 2147483648,
		},
		{
			val:           "9223372036854775807", // max int64
			expectNull:    false,
			expectedValue: 9223372036854775807,
		},
		{
			val:           "9223372036854775808", // max int64 + 1
			expectNull:    false,
			expectedValue: 0,
			expectedErr:   strconv.ErrRange,
		},
	}

	for i, tc := range cases {
		v := Int(tc.val)

		if null := v.IsNull(); null != tc.expectNull {
			t.Fatalf("expected test case %d IsNull to return %t, got %t", i, null, tc.expectNull)
		}

		value, null, err := v.ValueInt64()
		if value != tc.expectedValue {
			t.Fatalf("expected test case %d Value to be %d, got %d", i, tc.expectedValue, value)
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

func TestNullableInt_32(t *testing.T) {
	t.Parallel()

	cases := []struct {
		val           string
		expectNull    bool
		expectedValue int32
		expectedErr   error
	}{
		{
			val:           "1",
			expectNull:    false,
			expectedValue: 1,
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
		{
			val:           "-1",
			expectNull:    false,
			expectedValue: -1,
		},
		{
			val:           "2147483647", // max int32
			expectNull:    false,
			expectedValue: 2147483647,
		},
		{
			val:         "2147483648", // max int32
			expectNull:  false,
			expectedErr: strconv.ErrRange,
		},
	}

	for i, tc := range cases {
		v := Int(tc.val)

		if null := v.IsNull(); null != tc.expectNull {
			t.Fatalf("expected test case %d IsNull to return %t, got %t", i, null, tc.expectNull)
		}

		value, null, err := v.ValueInt32()
		if value != tc.expectedValue {
			t.Fatalf("expected test case %d Value to be %d, got %d", i, tc.expectedValue, value)
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

func TestValidationInt(t *testing.T) {
	t.Parallel()

	runValidationTestCases(t, []testCase{
		{
			val: "1",
			f:   ValidateTypeStringNullableInt,
		},
		{
			val:         "A",
			f:           ValidateTypeStringNullableInt,
			expectedErr: regexache.MustCompile(`^\w+: cannot parse 'A' as int: .*`),
		},
		{
			val:         1,
			f:           ValidateTypeStringNullableInt,
			expectedErr: regexache.MustCompile(`^expected type of \w+ to be string`),
		},
	})
}

func TestValidationIntAtLeast(t *testing.T) {
	t.Parallel()

	runValidationTestCases(t, []testCase{
		{
			val: "1",
			f:   ValidateTypeStringNullableIntAtLeast(1),
		},
		{
			val: "1",
			f:   ValidateTypeStringNullableIntAtLeast(0),
		},
		{
			val:         "1",
			f:           ValidateTypeStringNullableIntAtLeast(2),
			expectedErr: regexache.MustCompile(`expected \w+ to be at least \(2\), got 1`),
		},
		{
			val:         1,
			f:           ValidateTypeStringNullableIntAtLeast(2),
			expectedErr: regexache.MustCompile(`expected type of \w+ to be string`),
		},
	})
}
