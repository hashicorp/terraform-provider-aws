package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(sagemaker.EndpointsID, testAccErrorCheckSkipSagemaker)
}

func testAccErrorCheckSkipSagemaker(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"is not supported in region",
		"is not supported for the chosen region",
	)
}

// Tests are serialized as SagmMaker Domain resources are limited to 1 per account by default.
// SageMaker UserProfile and App depend on the Domain resources and as such are also part of the serialized test suite.
// Sagemaker Workteam tests must also be serialized
func TestAccAWSSagemaker_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"App": {
			"basic":        testAccAWSSagemakerApp_basic,
			"disappears":   testAccAWSSagemakerApp_tags,
			"tags":         testAccAWSSagemakerApp_disappears,
			"resourceSpec": testAccAWSSagemakerApp_resourceSpec,
		},
		"Domain": {
			"basic":                                    testAccAWSSagemakerDomain_basic,
			"disappears":                               testAccAWSSagemakerDomain_tags,
			"tags":                                     testAccAWSSagemakerDomain_disappears,
			"tensorboardAppSettings":                   testAccAWSSagemakerDomain_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage":          testAccAWSSagemakerDomain_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":                 testAccAWSSagemakerDomain_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_customImage":     testAccAWSSagemakerDomain_kernelGatewayAppSettings_customImage,
			"kernelGatewayAppSettings_lifecycleConfig": testAccAWSSagemakerDomain_kernelGatewayAppSettings_lifecycleConfig,
			"jupyterServerAppSettings":                 testAccAWSSagemakerDomain_jupyterServerAppSettings,
			"kms":                                      testAccAWSSagemakerDomain_kms,
			"securityGroup":                            testAccAWSSagemakerDomain_securityGroup,
			"sharingSettings":                          testAccAWSSagemakerDomain_sharingSettings,
		},
		"FlowDefinition": {
			"basic":                          testAccAWSSagemakerFlowDefinition_basic,
			"disappears":                     testAccAWSSagemakerFlowDefinition_disappears,
			"HumanLoopConfigPublicWorkforce": testAccAWSSagemakerFlowDefinition_humanLoopConfig_publicWorkforce,
			"HumanLoopRequestSource":         testAccAWSSagemakerFlowDefinition_humanLoopRequestSource,
			"Tags":                           testAccAWSSagemakerFlowDefinition_tags,
		},
		"UserProfile": {
			"basic":                           testAccAWSSagemakerUserProfile_basic,
			"disappears":                      testAccAWSSagemakerUserProfile_tags,
			"tags":                            testAccAWSSagemakerUserProfile_disappears,
			"tensorboardAppSettings":          testAccAWSSagemakerUserProfile_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage": testAccAWSSagemakerUserProfile_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":        testAccAWSSagemakerUserProfile_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_lifecycleConfig": testAccAWSSagemakerUserProfile_kernelGatewayAppSettings_lifecycleconfig,
			"jupyterServerAppSettings":                 testAccAWSSagemakerUserProfile_jupyterServerAppSettings,
		},
		"Workforce": {
			"disappears":     testAccAWSSagemakerWorkforce_disappears,
			"CognitoConfig":  testAccAWSSagemakerWorkforce_cognitoConfig,
			"OidcConfig":     testAccAWSSagemakerWorkforce_oidcConfig,
			"SourceIpConfig": testAccAWSSagemakerWorkforce_sourceIpConfig,
		},
		"Workteam": {
			"disappears":         testAccAWSSagemakerWorkteam_disappears,
			"CognitoConfig":      testAccAWSSagemakerWorkteam_cognitoConfig,
			"NotificationConfig": testAccAWSSagemakerWorkteam_notificationConfig,
			"OidcConfig":         testAccAWSSagemakerWorkteam_oidcConfig,
			"Tags":               testAccAWSSagemakerWorkteam_tags,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
