// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDevOpsGuru_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"EventSourcesConfig": {
			acctest.CtBasic:      testAccEventSourcesConfig_basic,
			acctest.CtDisappears: testAccEventSourcesConfig_disappears,
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
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DevOpsGuruClient(ctx)

	_, err := conn.DescribeAccountHealth(ctx, &devopsguru.DescribeAccountHealthInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
