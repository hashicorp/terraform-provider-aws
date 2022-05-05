package transfer_test

import (
	"testing"
)

func TestAccTransfer_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Access": {
			"disappears": testAccAccess_disappears,
			"EFSBasic":   testAccAccess_efs_basic,
			"S3Basic":    testAccAccess_s3_basic,
			"S3Policy":   testAccAccess_s3_policy,
		},
		"Server": {
			"basic":                         testAccServer_basic,
			"disappears":                    testAccServer_disappears,
			"APIGateway":                    testAccServer_apiGateway,
			"APIGatewayForceDestroy":        testAccServer_apiGateway_forceDestroy,
			"AuthenticationLoginBanners":    testAccServer_authenticationLoginBanners,
			"DirectoryService":              testAccServer_directoryService,
			"Domain":                        testAccServer_domain,
			"ForceDestroy":                  testAccServer_forceDestroy,
			"HostKey":                       testAccServer_hostKey,
			"LambdaFunction":                testAccServer_lambdaFunction,
			"Protocols":                     testAccServer_protocols,
			"SecurityPolicy":                testAccServer_securityPolicy,
			"UpdateEndpointTypePublicToVPC": testAccServer_updateEndpointType_publicToVPC,
			"UpdateEndpointTypePublicToVPCAddressAllocationIDs":      testAccServer_updateEndpointType_publicToVPC_addressAllocationIDs,
			"UpdateEndpointTypeVPCEndpointToVPC":                     testAccServer_updateEndpointType_vpcEndpointToVPC,
			"UpdateEndpointTypeVPCEndpointToVPCAddressAllocationIDs": testAccServer_updateEndpointType_vpcEndpointToVPC_addressAllocationIDs,
			"UpdateEndpointTypeVPCEndpointToVPCSecurityGroupIDs":     testAccServer_updateEndpointType_vpcEndpointToVPC_securityGroupIDs,
			"UpdateEndpointTypeVPCToPublic":                          testAccServer_updateEndpointType_vpcToPublic,
			"VPC":                                                    testAccServer_vpc,
			"VPCAddressAllocationIDs":                                testAccServer_vpcAddressAllocationIDs,
			"VPCAddressAllocationIDsSecurityGroupIDs":                testAccServer_vpcAddressAllocationIds_securityGroupIDs,
			"VPCEndpointID":                                          testAccServer_vpcEndpointID,
			"VPCSecurityGroupIDs":                                    testAccServer_vpcSecurityGroupIDs,
			"Workflow":                                               testAccServer_workflowDetails,
		},
		"SSHKey": {
			"basic": testAccSSHKey_basic,
		},
		"User": {
			"basic":                 testAccUser_basic,
			"disappears":            testAccUser_disappears,
			"HomeDirectoryMappings": testAccUser_homeDirectoryMappings,
			"ModifyWithOptions":     testAccUser_modifyWithOptions,
			"Posix":                 testAccUser_posix,
			"UserNameValidation":    testAccUser_UserName_Validation,
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
