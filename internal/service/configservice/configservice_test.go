// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConfigService_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"ConfigRule": {
			acctest.CtBasic:      testAccConfigRule_basic,
			"ownerAws":           testAccConfigRule_ownerAWS,
			"customlambda":       testAccConfigRule_customlambda,
			"customPolicy":       testAccConfigRule_ownerPolicy,
			"evaluationMode":     testAccConfigRule_evaluationMode,
			"scopeTagKey":        testAccConfigRule_Scope_TagKey,
			"scopeTagKeyEmpty":   testAccConfigRule_Scope_TagKey_Empty,
			"scopeTagValue":      testAccConfigRule_Scope_TagValue,
			"tags":               testAccConfigRule_tags,
			acctest.CtDisappears: testAccConfigRule_disappears,
		},
		"ConfigurationRecorderStatus": {
			acctest.CtBasic:      testAccConfigurationRecorderStatus_basic,
			"startEnabled":       testAccConfigurationRecorderStatus_startEnabled,
			acctest.CtDisappears: testAccConfigurationRecorderStatus_disappears,
		},
		"ConfigurationRecorder": {
			acctest.CtBasic:      testAccConfigurationRecorder_basic,
			"allParams":          testAccConfigurationRecorder_allParams,
			"recordStrategy":     testAccConfigurationRecorder_recordStrategy,
			acctest.CtDisappears: testAccConfigurationRecorder_disappears,
		},
		"ConformancePack": {
			acctest.CtBasic:             testAccConformancePack_basic,
			acctest.CtDisappears:        testAccConformancePack_disappears,
			"updateName":                testAccConformancePack_updateName,
			"inputParameters":           testAccConformancePack_inputParameters,
			"S3Delivery":                testAccConformancePack_S3Delivery,
			"S3Template":                testAccConformancePack_S3Template,
			"S3TemplateAndTemplateBody": testAccConformancePack_S3TemplateAndTemplateBody,
			"updateInputParameters":     testAccConformancePack_updateInputParameters,
			"updateS3Delivery":          testAccConformancePack_updateS3Delivery,
			"updateS3Template":          testAccConformancePack_updateS3Template,
			"updateTemplateBody":        testAccConformancePack_updateTemplateBody,
		},
		"DeliveryChannel": {
			acctest.CtBasic:      testAccDeliveryChannel_basic,
			"allParams":          testAccDeliveryChannel_allParams,
			acctest.CtDisappears: testAccDeliveryChannel_disappears,
		},
		"OrganizationConformancePack": {
			acctest.CtBasic:         testAccOrganizationConformancePack_basic,
			acctest.CtDisappears:    testAccOrganizationConformancePack_disappears,
			"excludedAccounts":      testAccOrganizationConformancePack_excludedAccounts,
			"updateName":            testAccOrganizationConformancePack_updateName,
			"inputParameters":       testAccOrganizationConformancePack_inputParameters,
			"S3Delivery":            testAccOrganizationConformancePack_S3Delivery,
			"S3Template":            testAccOrganizationConformancePack_S3Template,
			"updateInputParameters": testAccOrganizationConformancePack_updateInputParameters,
			"updateS3Delivery":      testAccOrganizationConformancePack_updateS3Delivery,
			"updateS3Template":      testAccOrganizationConformancePack_updateS3Template,
			"updateTemplateBody":    testAccOrganizationConformancePack_updateTemplateBody,
		},
		"OrganizationCustomPolicyRule": {
			acctest.CtBasic:      testAccOrganizationCustomPolicyRule_basic,
			acctest.CtDisappears: testAccOrganizationCustomPolicyRule_disappears,
			"policyText":         testAccOrganizationCustomPolicyRule_PolicyText,
		},
		"OrganizationCustomRule": {
			acctest.CtBasic:             testAccOrganizationCustomRule_basic,
			acctest.CtDisappears:        testAccOrganizationCustomRule_disappears,
			"errorHandling":             testAccOrganizationCustomRule_errorHandling,
			"Description":               testAccOrganizationCustomRule_Description,
			"ExcludedAccounts":          testAccOrganizationCustomRule_ExcludedAccounts,
			"InputParameters":           testAccOrganizationCustomRule_InputParameters,
			"LambdaFunctionArn":         testAccOrganizationCustomRule_lambdaFunctionARN,
			"MaximumExecutionFrequency": testAccOrganizationCustomRule_MaximumExecutionFrequency,
			"ResourceIdScope":           testAccOrganizationCustomRule_ResourceIdScope,
			"ResourceTypesScope":        testAccOrganizationCustomRule_ResourceTypesScope,
			"TagKeyScope":               testAccOrganizationCustomRule_TagKeyScope,
			"TagValueScope":             testAccOrganizationCustomRule_TagValueScope,
			"TriggerTypes":              testAccOrganizationCustomRule_TriggerTypes,
		},
		"OrganizationManagedRule": {
			acctest.CtBasic:             testAccOrganizationManagedRule_basic,
			acctest.CtDisappears:        testAccOrganizationManagedRule_disappears,
			"errorHandling":             testAccOrganizationManagedRule_errorHandling,
			"Description":               testAccOrganizationManagedRule_Description,
			"ExcludedAccounts":          testAccOrganizationManagedRule_ExcludedAccounts,
			"InputParameters":           testAccOrganizationManagedRule_InputParameters,
			"MaximumExecutionFrequency": testAccOrganizationManagedRule_MaximumExecutionFrequency,
			"ResourceIdScope":           testAccOrganizationManagedRule_ResourceIdScope,
			"ResourceTypesScope":        testAccOrganizationManagedRule_ResourceTypesScope,
			"RuleIdentifier":            testAccOrganizationManagedRule_RuleIdentifier,
			"TagKeyScope":               testAccOrganizationManagedRule_TagKeyScope,
			"TagValueScope":             testAccOrganizationManagedRule_TagValueScope,
		},
		"RemediationConfiguration": {
			acctest.CtBasic:      testAccRemediationConfiguration_basic,
			"basicBackward":      testAccRemediationConfiguration_basicBackwardCompatible,
			acctest.CtDisappears: testAccRemediationConfiguration_disappears,
			"updates":            testAccRemediationConfiguration_updates,
			"values":             testAccRemediationConfiguration_values,
		},
		"RetentionConfiguration": {
			acctest.CtBasic:      testAccRetentionConfiguration_basic,
			acctest.CtDisappears: testAccRetentionConfiguration_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 15*time.Second)
}
