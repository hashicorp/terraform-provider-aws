// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// only one replication set resource can be active at once, so we must have serialised tests
func TestAccSSMIncidents_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Replication Set Resource Tests": {
			acctest.CtBasic:    testReplicationSet_basic,
			"updateDefaultKey": testReplicationSet_updateRegionsWithoutCMK,
			"updateCMK":        testReplicationSet_updateRegionsWithCMK,
			"updateTags":       testReplicationSet_updateTags,
			"updateEmptyTags":  testReplicationSet_updateEmptyTags,
			"disappears":       testReplicationSet_disappears,
		},
		"Replication Set Data Source Tests": {
			acctest.CtBasic: testReplicationSetDataSource_basic,
		},
		"Response Plan Resource Tests": {
			acctest.CtBasic:          testResponsePlan_basic,
			"update":                 testResponsePlan_updateRequiredFields,
			"updateTags":             testResponsePlan_updateTags,
			"updateEmptyTags":        testResponsePlan_updateEmptyTags,
			"disappears":             testResponsePlan_disappears,
			"incidentTemplateFields": testResponsePlan_incidentTemplateOptionalFields,
			"displayName":            testResponsePlan_displayName,
			"chatChannel":            testResponsePlan_chatChannel,
			"engagement":             testResponsePlan_engagement,
			names.AttrAction:         testResponsePlan_action,
		},
		"Response Plan Data Source Tests": {
			acctest.CtBasic: testResponsePlanDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
