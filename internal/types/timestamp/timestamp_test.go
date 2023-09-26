// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestamp

import "testing"

func TestValidateOnceADayWindowFormat(t *testing.T) {
	t.Parallel()
	type tc struct {
		value       string
		expectError bool
	}
	tests := map[string]tc{
		"invalid hour": {
			value:       "24:00-25:00",
			expectError: true,
		},
		"invalid minute": {
			value:       "04:00-04:60",
			expectError: true,
		},
		"valid": {
			value: "04:00-05:00",
		},
		"empty": {
			value: "",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ts := New(test.value)
			err := ts.ValidateOnceADayWindowFormat()

			if err == nil && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if err != nil && !test.expectError {
				t.Fatalf("got unexpected error: %s", err)
			}
		})
	}
}

func TestValidateOnceAWeekWindowFormat(t *testing.T) {
	t.Parallel()
	type tc struct {
		value       string
		expectError bool
	}
	tests := map[string]tc{
		"invalid day of week": {
			value:       "san:04:00-san:05:00",
			expectError: true,
		},
		"invalid hour": {
			value:       "sun:24:00-san:25:00",
			expectError: true,
		},
		"invalid minute": {
			value:       "sun:04:00-sun:04:60",
			expectError: true,
		},
		"valid": {
			value: "sun:04:00-sun:05:00",
		},
		"case insensitive day": {
			value: "Sun:04:00-Sun:05:00",
		},
		"empty": {
			value: "",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ts := New(test.value)
			err := ts.ValidateOnceAWeekWindowFormat()

			if err == nil && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if err != nil && !test.expectError {
				t.Fatalf("got unexpected error: %s", err)
			}
		})
	}
}

func TestValidateUTCFormat(t *testing.T) {
	t.Parallel()
	type tc struct {
		value       string
		expectError bool
	}
	tests := map[string]tc{
		"invalid no TZ": {
			value:       "2015-03-07 23:45:00",
			expectError: true,
		},
		"invalid date order": {
			value:       "27-03-2019 23:45:00",
			expectError: true,
		},
		"invalid format": {
			value:       "Mon, 02 Jan 2006 15:04:05 -0700",
			expectError: true,
		},
		"valid": {
			value: "2006-01-02T15:04:05Z",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ts := New(test.value)
			err := ts.ValidateUTCFormat()

			if err == nil && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if err != nil && !test.expectError {
				t.Fatalf("got unexpected error: %s", err)
			}
		})
	}
}
