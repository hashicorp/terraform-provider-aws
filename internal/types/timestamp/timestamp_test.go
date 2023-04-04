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
