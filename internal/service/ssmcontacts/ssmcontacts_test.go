// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// SSMContacts resources depend on a replication set existing and
// only one replication set resource can be active at once, so we must have serialised tests
func TestAccSSMContacts_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Contact Resource Tests": {
			"basic":             testContact_basic,
			"disappears":        testContact_disappears,
			"updateAlias":       testContact_updateAlias,
			"updateDisplayName": testContact_updateDisplayName,
			"updateTags":        testContact_updateTags,
			"updateType":        testContact_updateType,
		},
		"Contact Data Source Tests": {
			"basic": testContactDataSource_basic,
		},
		"Contact Channel Resource Tests": {
			"basic":           testContactChannel_basic,
			"contactId":       testContactChannel_contactID,
			"deliveryAddress": testContactChannel_deliveryAddress,
			"disappears":      testContactChannel_disappears,
			"name":            testContactChannel_name,
			"type":            testContactChannel_type,
		},
		"Contact Channel Data Source Tests": {
			"basic": testContactChannelDataSource_basic,
		},
		"Plan Resource Tests": {
			"basic":                   testPlan_basic,
			"disappears":              testPlan_disappears,
			"updateChannelTargetInfo": testPlan_updateChannelTargetInfo,
			"updateContactId":         testPlan_updateContactId,
			"updateContactTargetInfo": testPlan_updateContactTargetInfo,
			"updateDurationInMinutes": testPlan_updateDurationInMinutes,
			"updateStages":            testPlan_updateStages,
			"updateTargets":           testPlan_updateTargets,
		},
		"Plan Data Source Tests": {
			"basic":             testPlanDataSource_basic,
			"channelTargetInfo": testPlanDataSource_channelTargetInfo,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
