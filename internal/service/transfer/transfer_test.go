// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccTransfer_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Access": {
			acctest.CtDisappears: testAccAccess_disappears,
			"EFSBasic":           testAccAccess_efs_basic,
			"S3Basic":            testAccAccess_s3_basic,
			"S3Policy":           testAccAccess_s3_policy,
		},
		"Agreement": {
			acctest.CtBasic:      testAccAgreement_basic,
			acctest.CtDisappears: testAccAgreement_disappears,
			"tags":               testAccAgreement_tags,
		},
		"Server": {
			acctest.CtBasic:                   testAccServer_basic,
			acctest.CtDisappears:              testAccServer_disappears,
			"tags":                            testAccServer_tags,
			"APIGateway":                      testAccServer_apiGateway,
			"APIGatewayForceDestroy":          testAccServer_apiGateway_forceDestroy,
			"AuthenticationLoginBanners":      testAccServer_authenticationLoginBanners,
			"DataSourceBasic":                 testAccServerDataSource_basic,
			"DataSourceServiceManaged":        testAccServerDataSource_Service_managed,
			"DataSourceAPIGateway":            testAccServerDataSource_apigateway,
			"DirectoryService":                testAccServer_directoryService,
			"Domain":                          testAccServer_domain,
			"ForceDestroy":                    testAccServer_forceDestroy,
			"HostKey":                         testAccServer_hostKey,
			"LambdaFunction":                  testAccServer_lambdaFunction,
			"Protocols":                       testAccServer_protocols,
			"ProtocolDetails":                 testAccServer_protocolDetails,
			"S3StorageOptions":                testAccServer_s3StorageOptions,
			"SecurityPolicy":                  testAccServer_securityPolicy,
			"SecurityPolicyFIPS":              testAccServer_securityPolicyFIPS,
			"SftpAuthenticationMethods":       testAccServer_identityProviderType_sftpAuthenticationMethods,
			"UpdateSftpAuthenticationMethods": testAccServer_updateIdentityProviderType_sftpAuthenticationMethods,
			"StructuredLogDestinations":       testAccServer_structuredLogDestinations,
			"UpdateEndpointTypePublicToVPC":   testAccServer_updateEndpointType_publicToVPC,
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
			acctest.CtBasic:      testAccSSHKey_basic,
			acctest.CtDisappears: testAccSSHKey_disappears,
		},
		"Tag": {
			acctest.CtBasic:      testAccTag_basic,
			acctest.CtDisappears: testAccTag_disappears,
			"Value":              testAccTag_value,
			"System":             testAccTag_system,
		},
		"User": {
			acctest.CtBasic:         testAccUser_basic,
			acctest.CtDisappears:    testAccUser_disappears,
			"tags":                  testAccUser_tags,
			"HomeDirectoryMappings": testAccUser_homeDirectoryMappings,
			"ModifyWithOptions":     testAccUser_modifyWithOptions,
			"Posix":                 testAccUser_posix,
			"UserNameValidation":    testAccUser_UserName_Validation,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
