// Copyright IBM Corp. 2014, 2026
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
		"ContactResource": {
			acctest.CtBasic:      testAccContact_basic,
			acctest.CtDisappears: testAccContact_disappears,
			"updateAlias":        testAccContact_updateAlias,
			"updateDisplayName":  testAccContact_updateDisplayName,
			"tags":               testAccSSMContactsContact_tagsSerial,
			"updateType":         testAccContact_updateType,
			"identity":           testAccSSMContactsContact_identitySerial,
		},
		"ContactDataSource": {
			acctest.CtBasic: testAccContactDataSource_basic,
			"tags":          testAccSSMContactsContactDataSource_tagsSerial,
		},
		"ContactChannelResource": {
			acctest.CtBasic:      testAccContactChannel_basic,
			"contactId":          testAccContactChannel_contactID,
			"deliveryAddress":    testAccContactChannel_deliveryAddress,
			acctest.CtDisappears: testAccContactChannel_disappears,
			acctest.CtName:       testAccContactChannel_name,
			"type":               testAccContactChannel_type,
			"identity":           testAccSSMContactsContactChannel_identitySerial,
		},
		"ContactChannelDataSource": {
			acctest.CtBasic: testAccContactChannelDataSource_basic,
		},
		"PlanResource": {
			acctest.CtBasic:           testAccPlan_basic,
			acctest.CtDisappears:      testAccPlan_disappears,
			"updateChannelTargetInfo": testAccPlan_updateChannelTargetInfo,
			"updateContactId":         testAccPlan_updateContactId,
			"updateContactTargetInfo": testAccPlan_updateContactTargetInfo,
			"updateDurationInMinutes": testAccPlan_updateDurationInMinutes,
			"updateStages":            testAccPlan_updateStages,
			"updateTargets":           testAccPlan_updateTargets,
		},
		"PlanDataSource": {
			acctest.CtBasic:     testAccPlanDataSource_basic,
			"channelTargetInfo": testAccPlanDataSource_channelTargetInfo,
		},
		"RotationResource": {
			acctest.CtBasic:          testAccRotation_basic,
			acctest.CtDisappears:     testAccRotation_disappears,
			"update":                 testAccRotation_updateRequiredFields,
			"startTime":              testAccRotation_startTime,
			"contactIds":             testAccRotation_contactIds,
			"recurrence":             testAccRotation_recurrence,
			"tags":                   testAccSSMContactsRotation_tagsSerial,
			"identity":               testAccSSMContactsRotation_identitySerial,
			"identityRegionOverride": testAccSSMContactsRotation_Identity_regionOverride,
		},
		"RotationDataSource": {
			acctest.CtBasic:   testAccRotationDataSource_basic,
			"dailySettings":   testAccRotationDataSource_dailySettings,
			"monthlySettings": testAccRotationDataSource_monthlySettings,
			"tags":            testAccSSMContactsRotationDataSource_tagsSerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
