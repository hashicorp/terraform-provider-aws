// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"testing"
	"time"
)

func TestSuppressEquivalentRoundedTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		old        string
		new        string
		layout     string
		d          time.Duration
		equivalent bool
	}{
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:13.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: true,
		},
		{
			old:        "2024-04-19T23:01:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: true,
		},
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: false,
		},
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Hour,
			equivalent: true,
		},
	}

	for i, tc := range testCases {
		value := SuppressEquivalentRoundedTime(tc.layout, tc.d)("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}
