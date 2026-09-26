// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Network Security Manager resources are account-level policy objects, and
// the acceptance tests all run in the current account, so they run serially.
func TestAccNetworkSecurityManager_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Policy": {
			acctest.CtBasic:      testAccPolicy_basic,
			acctest.CtDisappears: testAccPolicy_disappears,
			"full":               testAccPolicy_full,
			"shieldAdvanced":     testAccPolicy_shieldAdvanced,
			"wafConfig":          testAccPolicy_wafConfig,
			"priority":           testAccPolicy_priority,
			"associations":       testAccPolicy_associations,
			"associationsMax":    testAccPolicy_associationsMax,
			"publish":            testAccPolicy_publish,
			"description":        testAccPolicy_description,
			"updateToken":        testAccPolicy_updateToken,
			"replace":            testAccPolicy_replace,
			"recreate":           testAccPolicy_recreate,
			"referenceReplace":   testAccPolicy_referenceReplace,
			"draftReferences":    testAccPolicy_draftReferences,
			"validation":         testAccPolicy_validation,
			"Identity":           testAccNetworkSecurityManagerPolicy_identitySerial,
			"tags":               testAccNetworkSecurityManagerPolicy_tagsSerial,
		},
		"Rule": {
			acctest.CtBasic:      testAccRule_basic,
			acctest.CtDisappears: testAccRule_disappears,
			"full":               testAccRule_full,
			"configurationTypes": testAccRule_configurationTypes,
			"inspection":         testAccRule_inspection,
			"publish":            testAccRule_publish,
			"description":        testAccRule_description,
			"normalization":      testAccRule_normalization,
			"updateToken":        testAccRule_updateToken,
			"replace":            testAccRule_replace,
			"recreate":           testAccRule_recreate,
			"validation":         testAccRule_validation,
			"Identity":           testAccNetworkSecurityManagerRule_identitySerial,
			"tags":               testAccNetworkSecurityManagerRule_tagsSerial,
		},
		"Template": {
			acctest.CtBasic:      testAccTemplate_basic,
			acctest.CtDisappears: testAccTemplate_disappears,
			"full":               testAccTemplate_full,
			"rules":              testAccTemplate_rules,
			"publish":            testAccTemplate_publish,
			"description":        testAccTemplate_description,
			"updateToken":        testAccTemplate_updateToken,
			"replace":            testAccTemplate_replace,
			"recreate":           testAccTemplate_recreate,
			"ruleReplace":        testAccTemplate_ruleReplace,
			"validation":         testAccTemplate_validation,
			"Identity":           testAccNetworkSecurityManagerTemplate_identitySerial,
			"tags":               testAccNetworkSecurityManagerTemplate_tagsSerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
