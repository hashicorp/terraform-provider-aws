// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDevOpsGuru_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"EventSourcesConfig": {
			"basic":      testAccEventSourcesConfig_basic,
			"disappears": testAccEventSourcesConfig_disappears,
		},
		// A maxiumum of 2 notification channels can be configured at once, so
		// serialize tests for safety.
		"NotificationChannel": {
			"basic":      testAccNotificationChannel_basic,
			"disappears": testAccNotificationChannel_disappears,
			"filters":    testAccNotificationChannel_filters,
		},
		"ResourceCollection": {
			"basic":            testAccResourceCollection_basic,
			"cloudformation":   testAccResourceCollection_cloudformation,
			"disappears":       testAccResourceCollection_disappears,
			"tags":             testAccResourceCollection_tags,
			"tagsAllResources": testAccResourceCollection_tagsAllResources,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
