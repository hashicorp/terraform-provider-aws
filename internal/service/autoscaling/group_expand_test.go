// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"testing"
)

func TestExpandCapacityReservationSpecification(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input map[string]any
	}{
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "empty",
			input: map[string]any{},
		},
		{
			name: "empty capacity_reservation_target with nil element (#43008)",
			input: map[string]any{
				"capacity_reservation_target": []any{nil},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := expandCapacityReservationSpecification(tc.input)

			if tc.input == nil && got != nil {
				t.Fatalf("expected nil, got %#v", got)
			}
		})
	}
}
