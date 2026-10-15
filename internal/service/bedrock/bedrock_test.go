// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBedrock_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		// AWS has a default quota of 10 model evaluation jobs per account. Running
		// all tests in parallel exceeds this quota and triggers cascading failures,
		// especially in list tests.
		"EvaluationJob": {
			acctest.CtBasic: testAccEvaluationJob_basic,
			"optional":      testAccEvaluationJob_optional,
			"skipDestroy":   testAccEvaluationJob_skipDestroy,
			"tags":          testAccBedrockEvaluationJob_tagsSerial,
			"Identity":      testAccBedrockEvaluationJob_identitySerial,
			"List":          testAccEvaluationJob_listSerial,
		},
		// Model customization has a non-adjustable maximum concurrency of 2
		"CustomModel": {
			acctest.CtBasic:                         testAccCustomModel_basic,
			acctest.CtDisappears:                    testAccCustomModel_disappears,
			"tags":                                  testAccBedrockCustomModel_tagsSerial,
			"kmsKey":                                testAccCustomModel_kmsKey,
			"validationDataConfig":                  testAccCustomModel_validationDataConfig,
			"validationDataConfigWaitForCompletion": testAccCustomModel_validationDataConfigWaitForCompletion,
			"vpcConfig":                             testAccCustomModel_vpcConfig,
			"singularDataSourceBasic":               testAccCustomModelDataSource_basic,
			"pluralDataSourceBasic":                 testAccCustomModelsDataSource_basic,
			"Identity":                              testAccBedrockCustomModel_identitySerial,
		},
		"ModelInvocationLoggingConfiguration": {
			acctest.CtBasic:      testAccModelInvocationLoggingConfiguration_basic,
			acctest.CtDisappears: testAccModelInvocationLoggingConfiguration_disappears,
			"upgradeV6.0.0":      testAccModelInvocationLoggingConfiguration_upgrade_V6_0_0,
			"Identity":           testAccBedrockModelInvocationLoggingConfiguration_identitySerial,
		},
		"FoundationModelAgreement": {
			acctest.CtBasic:      testAccBedrockFoundationModelAgreement_basic,
			acctest.CtDisappears: testAccBedrockFoundationModelAgreement_disappears,
			"Identity":           testAccBedrockFoundationModelAgreement_identitySerial,
		},
		"UseCaseForModelAccess": {
			acctest.CtBasic: testAccBedrockUseCaseForModelAccess_basic,
			"createImport":  testAccBedrockUseCaseForModelAccess_createImport,
			"Identity":      testAccBedrockUseCaseForModelAccess_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
