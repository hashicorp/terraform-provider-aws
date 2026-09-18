// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"testing"
)

func TestQueueContinuousTargetOccurrence(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		assumeNoPropagationDelay bool
		defaultOccurrence        int
		want                     int
	}{
		"default create occurrence on real AWS": {
			assumeNoPropagationDelay: false,
			defaultOccurrence:        queueAttributesPropagatedContinuousTargetOccurrence,
			want:                     queueAttributesPropagatedContinuousTargetOccurrence,
		},
		"default delete occurrence on real AWS": {
			assumeNoPropagationDelay: false,
			defaultOccurrence:        queueDeletedContinuousTargetOccurrence,
			want:                     queueDeletedContinuousTargetOccurrence,
		},
		"collapsed create occurrence when no propagation delay": {
			assumeNoPropagationDelay: true,
			defaultOccurrence:        queueAttributesPropagatedContinuousTargetOccurrence,
			want:                     1,
		},
		"collapsed delete occurrence when no propagation delay": {
			assumeNoPropagationDelay: true,
			defaultOccurrence:        queueDeletedContinuousTargetOccurrence,
			want:                     1,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := queueContinuousTargetOccurrence(testCase.assumeNoPropagationDelay, testCase.defaultOccurrence); got != testCase.want {
				t.Errorf("queueContinuousTargetOccurrence(%t, %d) = %d, want %d", testCase.assumeNoPropagationDelay, testCase.defaultOccurrence, got, testCase.want)
			}
		})
	}
}
