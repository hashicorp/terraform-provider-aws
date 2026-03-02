// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDevOpsGuru_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"EventSourcesConfig": {
			acctest.CtBasic:      testAccEventSourcesConfig_basic,
			acctest.CtDisappears: testAccEventSourcesConfig_disappears,
			"Identity":           testAccDevOpsGuruEventSourcesConfig_identitySerial,
		},
		// A maxiumum of 2 notification channels can be configured at once, so
		// serialize tests for safety.
		"NotificationChannel": {
			acctest.CtBasic:      testAccNotificationChannel_basic,
			acctest.CtDisappears: testAccNotificationChannel_disappears,
			"filters":            testAccNotificationChannel_filters,
		},
		"NotificationChannelDataSource": {
			acctest.CtBasic: testAccNotificationChannelDataSource_basic,
		},
		"ResourceCollection": {
			acctest.CtBasic:      testAccResourceCollection_basic,
			"cloudformation":     testAccResourceCollection_cloudformation,
			acctest.CtDisappears: testAccResourceCollection_disappears,
			"tags":               testAccResourceCollection_tags,
			"tagsAllResources":   testAccResourceCollection_tagsAllResources,
		},
		"ResourceCollectionDataSource": {
			acctest.CtBasic: testAccResourceCollectionDataSource_basic,
		},
		"ServiceIntegration": {
			acctest.CtBasic: testAccServiceIntegration_basic,
			"kms":           testAccServiceIntegration_kms,
			"Identity":      testAccDevOpsGuruServiceIntegration_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DevOpsGuruClient(ctx)

	input := devopsguru.DescribeAccountHealthInput{}
	_, err := conn.DescribeAccountHealth(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
