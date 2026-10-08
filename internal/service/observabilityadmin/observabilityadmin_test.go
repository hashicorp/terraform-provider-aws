// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccObservabilityAdmin_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"S3TableIntegration": {
			acctest.CtBasic:      testAccS3TableIntegration_basic,
			acctest.CtDisappears: testAccS3TableIntegration_disappears,
			"kmsKey":             testAccS3TableIntegration_kmsKey,
			"tags":               testAccS3TableIntegration_tags,
			"Identity":           testAccObservabilityAdminS3TableIntegration_identitySerial,
		},
		"TelemetryEnrichment": {
			acctest.CtBasic:      testAccTelemetryEnrichment_basic,
			acctest.CtDisappears: testAccTelemetryEnrichment_disappears,
			"Identity":           testAccObservabilityAdminTelemetryEnrichment_identitySerial,
		},
		"TelemetryEvaluation": {
			acctest.CtBasic:      testAccTelemetryEvaluation_basic,
			acctest.CtDisappears: testAccTelemetryEvaluation_disappears,
			"Identity":           testAccObservabilityAdminTelemetryEvaluation_identitySerial,
		},
		"TelemetryEvaluationForOrganization": {
			acctest.CtBasic:      testAccTelemetryEvaluationForOrganization_basic,
			acctest.CtDisappears: testAccTelemetryEvaluationForOrganization_disappears,
			"Identity":           testAccObservabilityAdminTelemetryEvaluationForOrganization_identitySerial,
		},
		"TelemetryRule": {
			acctest.CtBasic:       testAccTelemetryRule_basic,
			acctest.CtDisappears:  testAccTelemetryRule_disappears,
			"tags":                testAccTelemetryRule_tags,
			"vpcFlowLogs":         testAccTelemetryRule_vpcFlowLogs,
			"allRegions":          testAccTelemetryRule_allRegions,
			"regions":             testAccTelemetryRule_regions,
			"update":              testAccTelemetryRule_update,
			"Identity":            testAccObservabilityAdminTelemetryRule_identitySerial,
			"ListBasic":           testAccTelemetryRule_List_basic,
			"ListIncludeResource": testAccTelemetryRule_List_includeResource,
			"ListRegionOverride":  testAccTelemetryRule_List_regionOverride,
		},
		"TelemetryRuleForOrganization": {
			acctest.CtBasic:       testAccTelemetryRuleForOrganization_basic,
			acctest.CtDisappears:  testAccTelemetryRuleForOrganization_disappears,
			"tags":                testAccTelemetryRuleForOrganization_tags,
			"vpcFlowLogs":         testAccTelemetryRuleForOrganization_vpcFlowLogs,
			"allRegions":          testAccTelemetryRuleForOrganization_allRegions,
			"Identity":            testAccObservabilityAdminTelemetryRuleForOrganization_identitySerial,
			"ListBasic":           testAccTelemetryRuleForOrganization_List_basic,
			"ListIncludeResource": testAccTelemetryRuleForOrganization_List_includeResource,
			"ListRegionOverride":  testAccTelemetryRuleForOrganization_List_regionOverride,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
