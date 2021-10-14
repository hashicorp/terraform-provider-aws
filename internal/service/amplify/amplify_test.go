package amplify_test

import (
	"testing"
	"time"
)

// Serialize to limit API rate-limit exceeded errors.
func TestAccAWSAmplify_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"App": {
			"basic":                    testAccApp_basic,
			"disappears":               testAccApp_disappears,
			"Tags":                     testAccApp_Tags,
			"AutoBranchCreationConfig": testAccApp_AutoBranchCreationConfig,
			"BasicAuthCredentials":     testAccApp_BasicAuthCredentials,
			"BuildSpec":                testAccApp_BuildSpec,
			"CustomRules":              testAccApp_CustomRules,
			"Description":              testAccApp_Description,
			"EnvironmentVariables":     testAccApp_EnvironmentVariables,
			"IamServiceRole":           testAccApp_IAMServiceRole,
			"Name":                     testAccApp_Name,
			"Repository":               testAccApp_Repository,
		},
		"BackendEnvironment": {
			"basic":                         testAccBackendEnvironment_basic,
			"disappears":                    testAccBackendEnvironment_disappears,
			"DeploymentArtifacts_StackName": testAccBackendEnvironment_DeploymentArtifacts_StackName,
		},
		"Branch": {
			"basic":                testAccBranch_basic,
			"disappears":           testAccBranch_disappears,
			"Tags":                 testAccBranch_Tags,
			"BasicAuthCredentials": testAccBranch_BasicAuthCredentials,
			"EnvironmentVariables": testAccBranch_EnvironmentVariables,
			"OptionalArguments":    testAccBranch_OptionalArguments,
		},
		"DomainAssociation": {
			"basic":      testAccDomainAssociation_basic,
			"disappears": testAccDomainAssociation_disappears,
			"update":     testAccDomainAssociation_update,
		},
		"Webhook": {
			"basic":      testAccWebhook_basic,
			"disappears": testAccWebhook_disappears,
			"update":     testAccWebhook_update,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
					// Explicitly sleep between tests.
					time.Sleep(5 * time.Second)
				})
			}
		})
	}
}
