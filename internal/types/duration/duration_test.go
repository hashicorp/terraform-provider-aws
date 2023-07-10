// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package duration

import (
	"errors"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		input       string
		expected    Duration
		expectedErr error
	}{
		// Invalid
		"empty": {
			input:       "",
			expectedErr: ErrSyntax,
		},
		"P only": {
			input:       "P",
			expectedErr: ErrSyntax,
		},

		// Single
		"years only": {
			input:       "P1Y",
			expected:    Duration{years: 1},
			expectedErr: nil,
		},
		"months only": {
			input:       "P3M",
			expected:    Duration{months: 3},
			expectedErr: nil,
		},
		"days only": {
			input:       "P30D",
			expected:    Duration{days: 30},
			expectedErr: nil,
		},

		// Multiple
		"years months": {
			input:       "P1Y2M",
			expected:    Duration{years: 1, months: 2},
			expectedErr: nil,
		},
		"years days": {
			input:       "P2Y15D",
			expected:    Duration{years: 2, days: 15},
			expectedErr: nil,
		},
		"years months days": {
			input:       "P2Y1M10D",
			expected:    Duration{years: 2, months: 1, days: 10},
			expectedErr: nil,
		},

		"zero years": {
			input:       "P0Y1M10D",
			expected:    Duration{months: 1, days: 10},
			expectedErr: nil,
		},

		"insensitive": {
			input:       "p2y1M10d",
			expected:    Duration{years: 2, months: 1, days: 10},
			expectedErr: nil,
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			duration, err := Parse(tc.input)

			if tc.expectedErr == nil && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error matching \"%s\", got %s", tc.expectedErr, err)
				}
			}

			if !duration.equal(tc.expected) {
				t.Errorf("expected %q, got %q", tc.expected, duration)
			}
		})
	}
}

func TestSub(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tz, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		t.Fatalf(err.Error())
	}

	testcases := map[string]struct {
		startTime time.Time
		duration  Duration
		expected  time.Time
		hoursDiff int
	}{
		"zero": {
			startTime: now,
			duration:  Duration{},
			expected:  now,
		},
		"regular": {
			startTime: now,
			duration:  Duration{years: 1, months: 2, days: 3},
			expected:  now.AddDate(-1, -2, -3),
		},

		"month": {
			startTime: time.Date(2022, 3, 29, 0, 0, 0, 0, tz),
			duration:  Duration{months: 1},
			expected:  time.Date(2022, 3, 1, 0, 0, 0, 0, tz),
		},
		"leap year month": {
			startTime: time.Date(2020, 3, 29, 0, 0, 0, 0, tz),
			duration:  Duration{months: 1},
			expected:  time.Date(2020, 2, 29, 0, 0, 0, 0, tz),
		},

		"day": {
			startTime: time.Date(2022, 4, 14, 12, 0, 0, 0, tz),
			duration:  Duration{days: 3},
			expected:  time.Date(2022, 4, 11, 12, 0, 0, 0, tz),
			hoursDiff: 3 * 24,
		},
		"daylight saving day": {
			startTime: time.Date(2022, 3, 14, 12, 0, 0, 0, tz),
			duration:  Duration{days: 3},
			expected:  time.Date(2022, 3, 11, 12, 0, 0, 0, tz),
			hoursDiff: 3*24 - 1,
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := Sub(tc.startTime, tc.duration)

			if !actual.Equal(tc.expected) {
				t.Fatalf("expected %s, got %s", tc.expected, actual)
			}

			if tc.hoursDiff != 0 {
				diff := tc.startTime.Sub(tc.expected)
				if diff.Hours() != float64(tc.hoursDiff) {
					t.Fatalf("diff expected %d, got %f", tc.hoursDiff, diff.Hours())
				}
			}
		})
	}
}
