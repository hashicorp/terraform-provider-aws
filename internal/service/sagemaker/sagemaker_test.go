// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.SageMakerServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"is not supported in region",
		"is not supported for the chosen region",
		"No available licenses were found",
	)
}

// Tests are serialized as SageMaker Domain resources are limited to 1 per account by default.
// SageMaker UserProfile and App depend on the Domain resources and as such are also part of the serialized test suite.
// SageMaker Workteam tests must also be serialized
func TestAccSageMaker_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"App": {
			acctest.CtBasic:         testAccApp_basic,
			acctest.CtDisappears:    testAccApp_disappears,
			"tags":                  testAccApp_tags,
			"resourceSpec":          testAccApp_resourceSpec,
			"resourceSpecLifecycle": testAccApp_resourceSpecLifecycle,
			"space":                 testAccApp_space,
			"decodeAppId":           testAccDecodeAppID,
		},
		"Domain": {
			acctest.CtBasic:                            testAccDomain_basic,
			acctest.CtDisappears:                       testAccDomain_tags,
			"tags":                                     testAccDomain_disappears,
			"tensorboardAppSettings":                   testAccDomain_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage":          testAccDomain_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":                 testAccDomain_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_customImage":     testAccDomain_kernelGatewayAppSettings_customImage,
			"kernelGatewayAppSettings_lifecycleConfig": testAccDomain_kernelGatewayAppSettings_lifecycleConfig,
			"kernelGatewayAppSettings_defaultResourceAndCustomImage":  testAccDomain_kernelGatewayAppSettings_defaultResourceSpecAndCustomImage,
			"jupyterServerAppSettings":                                testAccDomain_jupyterServerAppSettings,
			"codeEditorAppSettings":                                   testAccDomain_codeEditorAppSettings,
			"codeEditorAppSettings_customImage":                       testAccDomain_codeEditorAppSettings_customImage,
			"codeEditorAppSettings_defaultResourceSpecAndCustomImage": testAccDomain_codeEditorAppSettings_defaultResourceSpecAndCustomImage,
			"jupyterLabAppSettings":                                   testAccDomain_jupyterLabAppSettings,
			"jupyterLabAppSettingsAppLifecycle":                       testAccDomain_jupyterLabAppSettingsAppLifecycle,
			"jupyterLabAppSettingsEMRSettings":                        testAccDomain_jupyterLabAppSettingsEMRSettings,
			"jupyterLabAppSettingsBuiltInLifecycle":                   testAccDomain_jupyterLabAppSettingsBuiltInLifecycle,
			"kms":                                                     testAccDomain_kms,
			"defaultUserSettingsSecurityGroupUpdated":                 testAccDomain_defaultUserSettingsSecurityGroupUpdated,
			"sharingSettings":                                         testAccDomain_sharingSettings,
			"defaultUserSettingsUpdated":                              testAccDomain_defaultUserSettingsUpdated,
			"canvas":                                                  testAccDomain_canvasAppSettings,
			"modelRegisterSettings":                                   testAccDomain_modelRegisterSettings,
			"generativeAi":                                            testAccDomain_generativeAiSettings,
			"emrServerlessSettings":                                   testAccDomain_emrServerlessSettings,
			"identityProviderOauthSettings":                           testAccDomain_identityProviderOAuthSettings,
			"directDeploySettings":                                    testAccDomain_directDeploySettings,
			"kendraSettings":                                          testAccDomain_kendraSettings,
			"workspaceSettings":                                       testAccDomain_workspaceSettings,
			"domainSettings":                                          testAccDomain_domainSettings,
			"domainSettingsDockerSettingsUpdated":                     testAccDomain_domainSettingsDockerSettingsUpdated,
			"rSessionAppSettings":                                     testAccDomain_rSessionAppSettings,
			"rStudioServerProAppSettings":                             testAccDomain_rStudioServerProAppSettings,
			"rStudioServerProDomainSettings":                          testAccDomain_rStudioServerProDomainSettings,
			"rStudioDomainDisabledNetworkUpdate":                      testAccDomain_rStudioDomainDisabledNetworkUpdate,
			"spaceSettingsKernelGatewayAppSettings":                   testAccDomain_spaceSettingsKernelGatewayAppSettings,
			"spaceSettingsJupyterLabAppSettings":                      testAccDomain_spaceSettingsJupyterLabAppSettings,
			"spaceSettingsSpaceStorageSettings":                       testAccDomain_spaceSettingsSpaceStorageSettings,
			"spaceSettingsCustomPOSIXUserConfig":                      testAccDomain_spaceSettingsCustomPOSIXUserConfig,
			"spaceSettingsCustomFileSystemConfigs":                    testAccDomain_spaceSettingsCustomFileSystemConfigs,
			"code":                                                    testAccDomain_jupyterServerAppSettings_code,
			"efs":                                                     testAccDomain_efs,
			"posix":                                                   testAccDomain_posix,
			"spaceStorageSettings":                                    testAccDomain_spaceStorageSettings,
			"studioWebPortalSettings_hiddenAppTypes":                  testAccDomain_studioWebPortalSettings_hiddenAppTypes,
			"studioWebPortalSettings_hiddenInstanceTypes":             testAccDomain_studioWebPortalSettings_hiddenInstanceTypes,
			"studioWebPortalSettings_hiddenMlTools":                   testAccDomain_studioWebPortalSettings_hiddenMlTools,
		},
		"FlowDefinition": {
			acctest.CtBasic:                  testAccFlowDefinition_basic,
			acctest.CtDisappears:             testAccFlowDefinition_disappears,
			"tags":                           testAccFlowDefinition_tags,
			"HumanLoopConfigPublicWorkforce": testAccFlowDefinition_humanLoopConfig_publicWorkforce,
			"HumanLoopRequestSource":         testAccFlowDefinition_humanLoopRequestSource,
		},
		"Space": {
			acctest.CtBasic:            testAccSpace_basic,
			acctest.CtDisappears:       testAccSpace_tags,
			"tags":                     testAccSpace_disappears,
			"kernelGatewayAppSettings": testAccSpace_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_lifecycleConfig": testAccSpace_kernelGatewayAppSettings_lifecycleconfig,
			"kernelGatewayAppSettings_imageConfig":     testAccSpace_kernelGatewayAppSettings_imageconfig,
			"jupyterServerAppSettings":                 testAccSpace_jupyterServerAppSettings,
			"jupyterLabAppSettings":                    testAccSpace_jupyterLabAppSettings,
			"jupyterLabAppSettingsLifecycle":           testAccSpace_jupyterLabAppSettingsAppLifecycle,
			"codeEditorAppSettings":                    testAccSpace_codeEditorAppSettings,
			"storageSettings":                          testAccSpace_storageSettings,
			"customFileSystem":                         testAccSpace_customFileSystem,
		},
		"UserProfile": {
			acctest.CtBasic:                            testAccUserProfile_basic,
			acctest.CtDisappears:                       testAccUserProfile_tags,
			"tags":                                     testAccUserProfile_disappears,
			"tensorboardAppSettings":                   testAccUserProfile_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage":          testAccUserProfile_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":                 testAccUserProfile_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_lifecycleConfig": testAccUserProfile_kernelGatewayAppSettings_lifecycleconfig,
			"kernelGatewayAppSettings_imageConfig":     testAccUserProfile_kernelGatewayAppSettings_imageconfig,
			"codeEditorAppSettings_customImage":        testAccUserProfile_codeEditorAppSettings_customImage,
			"jupyterServerAppSettings":                 testAccUserProfile_jupyterServerAppSettings,
			"studioWebPortalSettings_hiddenAppTypes":   testAccUserProfile_studioWebPortalSettings_hiddenAppTypes,
			"studioWebPortalSettings_hiddenMlTools":    testAccUserProfile_studioWebPortalSettings_hiddenMlTools,
		},
		"Workforce": {
			acctest.CtDisappears: testAccWorkforce_disappears,
			"CognitoConfig":      testAccWorkforce_cognitoConfig,
			"OidcConfig":         testAccWorkforce_oidcConfig,
			"OidcConfig_full":    testAccWorkforce_oidcConfig_full,
			"SourceIpConfig":     testAccWorkforce_sourceIPConfig,
			"VPC":                testAccWorkforce_vpc,
		},
		"Workteam": {
			acctest.CtDisappears:        testAccWorkteam_disappears,
			"tags":                      testAccWorkteam_tags,
			"CognitoConfig":             testAccWorkteam_cognitoConfig,
			"CognitoOmitWorkforceName":  testAccWorkteam_cognitoOmitWorkforceName,
			"NotificationConfig":        testAccWorkteam_notificationConfig,
			"WorkerAccessConfiguration": testAccWorkteam_workerAccessConfiguration,
			"OidcConfig":                testAccWorkteam_oidcConfig,
		},
		"Servicecatalog": {
			acctest.CtBasic: testAccServicecatalogPortfolioStatus_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
