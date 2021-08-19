package aws

import (
	"testing"
)

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
			"basic":                                testAccAWSSagemakerDomain_basic,
			"disappears":                           testAccAWSSagemakerDomain_tags,
			"tags":                                 testAccAWSSagemakerDomain_disappears,
			"tensorboardAppSettings":               testAccAWSSagemakerDomain_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage":      testAccAWSSagemakerDomain_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":             testAccAWSSagemakerDomain_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_customImage": testAccAWSSagemakerDomain_kernelGatewayAppSettings_customImage,
			"jupyterServerAppSettings":             testAccAWSSagemakerDomain_jupyterServerAppSettings,
			"kms":                                  testAccAWSSagemakerDomain_kms,
			"securityGroup":                        testAccAWSSagemakerDomain_securityGroup,
			"sharingSettings":                      testAccAWSSagemakerDomain_sharingSettings,
		},
		"UserProfile": {
			"basic":                           testAccAWSSagemakerUserProfile_basic,
			"disappears":                      testAccAWSSagemakerUserProfile_tags,
			"tags":                            testAccAWSSagemakerUserProfile_disappears,
			"tensorboardAppSettings":          testAccAWSSagemakerUserProfile_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage": testAccAWSSagemakerUserProfile_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":        testAccAWSSagemakerUserProfile_kernelGatewayAppSettings,
			"jupyterServerAppSettings":        testAccAWSSagemakerUserProfile_jupyterServerAppSettings,
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
