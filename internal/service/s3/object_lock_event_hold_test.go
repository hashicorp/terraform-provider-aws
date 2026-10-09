// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"testing"
)

func TestValidateEventHoldDuration(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		hasDuration bool
		configured  string
		state       attrState
		effective   string
		wantErr     bool
	}{
		"no duration is always fine": {
			hasDuration: false,
			state:       attrAbsent,
			effective:   "",
		},
		"no duration with a released hold": {
			hasDuration: false,
			state:       attrAbsent,
			effective:   "OFF",
		},
		"duration with a configured ON": {
			hasDuration: true,
			configured:  "ON",
			state:       attrSet,
		},
		"duration with a configured OFF": {
			hasDuration: true,
			configured:  "OFF",
			state:       attrSet,
			wantErr:     true,
		},
		// A hold supplied by another resource is not resolved during the first
		// plan. Rejecting it there would fail a configuration that is valid.
		"duration with an unknown hold defers": {
			hasDuration: true,
			state:       attrUnknown,
		},
		"duration with an unknown hold defers even with state": {
			hasDuration: true,
			state:       attrUnknown,
			effective:   "OFF",
		},
		"duration with no hold anywhere": {
			hasDuration: true,
			state:       attrAbsent,
			effective:   "",
			wantErr:     true,
		},
		// Inherited from a bucket default_event_hold.
		"duration with an inherited ON": {
			hasDuration: true,
			state:       attrAbsent,
			effective:   "ON",
		},
		// The hold was released, so the duration would be dropped silently.
		"duration with a released hold in state": {
			hasDuration: true,
			state:       attrAbsent,
			effective:   "OFF",
			wantErr:     true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := validateEventHoldDuration(testCase.hasDuration, testCase.configured, testCase.state, testCase.effective)

			if got, want := err != nil, testCase.wantErr; got != want {
				t.Fatalf("error %t, want %t: %v", got, want, err)
			}
		})
	}
}
