// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLogs_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Delivery": {
			acctest.CtBasic:          testAccDelivery_basic,
			acctest.CtDisappears:     testAccDelivery_disappears,
			"cloudFrontDistribution": testAccDelivery_cloudFrontDistribution,
			"tags":                   testAccDelivery_tags,
			"update":                 testAccDelivery_update,
			"updateRecordFieldsNoS3": testAccDelivery_updateRecordFieldsNoS3,
			"Identity":               testAccLogsDelivery_identitySerial,
		},
		"DeliverySource": {
			acctest.CtBasic:      testAccDeliverySource_basic,
			acctest.CtDisappears: testAccDeliverySource_disappears,
			"tags":               testAccDeliverySource_tags,
			"Identity":           testAccLogsDeliverySource_identitySerial,
		},
		"S3TableIntegrationSource": {
			acctest.CtBasic:        testAccS3TableIntegrationSource_basic,
			acctest.CtDisappears:   testAccS3TableIntegrationSource_disappears,
			"Identity":             testAccLogsS3TableIntegrationSource_identitySerial,
			"List_basic":           testAccS3TableIntegrationSource_List_basic,
			"List_includeResource": testAccS3TableIntegrationSource_List_includeResource,
			"List_regionOverride":  testAccS3TableIntegrationSource_List_regionOverride,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
