// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.CloudTrailServiceID, testAccErrorCheckSkip)
}

// testAccErrorCheckSkip skips CloudTrail tests that have error messages indicating unsupported features
func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"AccessDeniedException:",
	)
}

func TestAccCloudTrail_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"OrganizationDelegatedAdminAccount": {
			acctest.CtBasic:      testAccOrganizationDelegatedAdminAccount_basic,
			acctest.CtDisappears: testAccOrganizationDelegatedAdminAccount_disappears,
		},
		"Trail": {
			acctest.CtBasic:               testAccTrail_basic,
			"cloudwatch":                  testAccTrail_cloudWatch,
			"enableLogging":               testAccTrail_enableLogging,
			"globalServiceEvents":         testAccTrail_globalServiceEvents,
			"multiRegion":                 testAccTrail_multiRegion,
			"organization":                testAccTrail_organization,
			"logValidation":               testAccTrail_logValidation,
			"kmsKey":                      testAccTrail_kmsKey,
			"snsTopicNameBasic":           testAccTrail_snsTopicNameBasic,
			"snsTopicNameAlternateRegion": testAccTrail_snsTopicNameAlternateRegion,
			"tags":                        testAccTrail_tags,
			"eventSelector":               testAccTrail_eventSelector,
			"eventSelectorDynamoDB":       testAccTrail_eventSelectorDynamoDB,
			"eventSelectorExclude":        testAccTrail_eventSelectorExclude,
			"insightSelector":             testAccTrail_insightSelector,
			"advancedEventSelector":       testAccTrail_advancedEventSelector,
			acctest.CtDisappears:          testAccTrail_disappears,
			"migrateV0":                   testAccTrail_migrateV0,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
