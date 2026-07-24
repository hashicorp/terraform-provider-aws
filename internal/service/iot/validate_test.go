// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"testing"
)

func Test_validTopicRuleCloudWatchMetricTimestamp(t *testing.T) {
	t.Parallel()

	validCases := []string{
		"${timestamp()}",
		"${device_time}",
		"${topic(3)}",
		"${get(thing.attributes, 'time')}",
	}

	for _, v := range validCases {
		_, errors := validTopicRuleCloudWatchMetricTimestamp(v, "metric_timestamp")
		if len(errors) != 0 {
			t.Errorf("expected %q to be valid, got errors: %v", v, errors)
		}
	}

	invalidCases := []string{
		"not-a-template",
		"2023-01-01T00:00:00Z",
		"${unclosed",
		"just text",
		"",
	}

	for _, v := range invalidCases {
		_, errors := validTopicRuleCloudWatchMetricTimestamp(v, "metric_timestamp")
		if len(errors) == 0 {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}
