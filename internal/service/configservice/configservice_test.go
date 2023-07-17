// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConfigService_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Config": {
			"basic":            testAccConfigRule_basic,
			"ownerAws":         testAccConfigRule_ownerAws,
			"customlambda":     testAccConfigRule_customlambda,
			"customPolicy":     testAccConfigRule_ownerPolicy,
			"scopeTagKey":      testAccConfigRule_Scope_TagKey,
			"scopeTagKeyEmpty": testAccConfigRule_Scope_TagKey_Empty,
			"scopeTagValue":    testAccConfigRule_Scope_TagValue,
			"tags":             testAccConfigRule_tags,
			"disappears":       testAccConfigRule_disappears,
		},
		"ConfigurationRecorderStatus": {
			"basic":        testAccConfigurationRecorderStatus_basic,
			"startEnabled": testAccConfigurationRecorderStatus_startEnabled,
			"importBasic":  testAccConfigurationRecorderStatus_importBasic,
		},
		"ConfigurationRecorder": {
			"basic":          testAccConfigurationRecorder_basic,
			"allParams":      testAccConfigurationRecorder_allParams,
			"recordStrategy": testAccConfigurationRecorder_recordStrategy,
			"disappears":     testAccConfigurationRecorder_disappears,
		},
		"ConformancePack": {
			"basic":                     testAccConformancePack_basic,
			"disappears":                testAccConformancePack_disappears,
			"forceNew":                  testAccConformancePack_forceNew,
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
			"basic":       testAccDeliveryChannel_basic,
			"allParams":   testAccDeliveryChannel_allParams,
			"importBasic": testAccDeliveryChannel_importBasic,
		},
		"OrganizationConformancePack": {
			"basic":                 testAccOrganizationConformancePack_basic,
			"disappears":            testAccOrganizationConformancePack_disappears,
			"excludedAccounts":      testAccOrganizationConformancePack_excludedAccounts,
			"forceNew":              testAccOrganizationConformancePack_forceNew,
			"inputParameters":       testAccOrganizationConformancePack_inputParameters,
			"S3Delivery":            testAccOrganizationConformancePack_S3Delivery,
			"S3Template":            testAccOrganizationConformancePack_S3Template,
			"updateInputParameters": testAccOrganizationConformancePack_updateInputParameters,
			"updateS3Delivery":      testAccOrganizationConformancePack_updateS3Delivery,
			"updateS3Template":      testAccOrganizationConformancePack_updateS3Template,
			"updateTemplateBody":    testAccOrganizationConformancePack_updateTemplateBody,
		},
		"OrganizationCustomRule": {
			"basic":                     testAccOrganizationCustomRule_basic,
			"disappears":                testAccOrganizationCustomRule_disappears,
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
			"basic":                     testAccOrganizationManagedRule_basic,
			"disappears":                testAccOrganizationManagedRule_disappears,
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
			"basic":             testAccRemediationConfiguration_basic,
			"basicBackward":     testAccRemediationConfiguration_basicBackwardCompatible,
			"disappears":        testAccRemediationConfiguration_disappears,
			"migrateParameters": testAccRemediationConfiguration_migrateParameters,
			"recreates":         testAccRemediationConfiguration_recreates,
			"updates":           testAccRemediationConfiguration_updates,
			"values":            testAccRemediationConfiguration_values,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
