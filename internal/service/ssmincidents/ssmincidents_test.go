// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// only one replication set resource can be active at once, so we must have serialised tests
func TestAccSSMIncidents_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Replication Set Resource Tests": {
			acctest.CtBasic:      testAccReplicationSet_basic,
			"updateDefaultKey":   testAccReplicationSet_updateRegionsWithoutCMK,
			"updateCMK":          testAccReplicationSet_updateRegionsWithCMK,
			"updateTags":         testAccReplicationSet_updateTags,
			"updateEmptyTags":    testAccReplicationSet_updateEmptyTags,
			acctest.CtDisappears: testAccReplicationSet_disappears,
		},
		"Replication Set Data Source Tests": {
			acctest.CtBasic: testAccReplicationSetDataSource_basic,
		},
		"Response Plan Resource Tests": {
			acctest.CtBasic:          testAccResponsePlan_basic,
			"update":                 testAccResponsePlan_updateRequiredFields,
			"updateTags":             testAccResponsePlan_updateTags,
			"updateEmptyTags":        testAccResponsePlan_updateEmptyTags,
			acctest.CtDisappears:     testAccResponsePlan_disappears,
			"incidentTemplateFields": testAccResponsePlan_incidentTemplateOptionalFields,
			"displayName":            testAccResponsePlan_displayName,
			"chatChannel":            testAccResponsePlan_chatChannel,
			"engagement":             testAccResponsePlan_engagement,
			"action":                 testAccResponsePlan_action,
		},
		"Response Plan Data Source Tests": {
			acctest.CtBasic: testAccResponsePlanDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
