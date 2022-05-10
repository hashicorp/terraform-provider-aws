package transfer_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(transfer.EndpointsID, testAccErrorCheckSkip)

}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Invalid server type: PUBLIC",
		"InvalidServiceName: The Vpc Endpoint Service",
	)
}

func testAccServer_basic(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	iamRoleResourceName := "aws_iam_role.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "endpoint", "server.transfer", regexp.MustCompile(`s-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "function", ""),
					resource.TestCheckNoResourceAttr(resourceName, "host_key"),
					resource.TestCheckResourceAttrSet(resourceName, "host_key_fingerprint"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "invocation_role", ""),
					resource.TestCheckResourceAttr(resourceName, "logging_role", ""),
					resource.TestCheckResourceAttr(resourceName, "post_authentication_login_banner", ""),
					resource.TestCheckResourceAttr(resourceName, "pre_authentication_login_banner", ""),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "SFTP"),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "url", ""),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "domain", "S3"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "endpoint", "server.transfer", regexp.MustCompile(`s-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "function", ""),
					resource.TestCheckNoResourceAttr(resourceName, "host_key"),
					resource.TestCheckResourceAttrSet(resourceName, "host_key_fingerprint"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "invocation_role", ""),
					resource.TestCheckResourceAttrPair(resourceName, "logging_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "SFTP"),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "url", ""),
				),
			},
		},
	})
}

func testAccServer_domain(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDomainConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "domain", "EFS"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_disappears(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tftransfer.ResourceServer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccServer_securityPolicy(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerSecurityPolicyConfig("TransferSecurityPolicy-2020-06"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2020-06"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerSecurityPolicyConfig("TransferSecurityPolicy-2018-11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
				),
			},
		},
	})
}

func testAccServer_vpc(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerVPCUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccServerVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccServer_vpcAddressAllocationIDs(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	eip1ResourceName := "aws_eip.test.0"
	eip2ResourceName := "aws_eip.test.1"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCAddressAllocationIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerVPCAddressAllocationIdsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccServerVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccServer_vpcSecurityGroupIDs(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	securityGroup1ResourceName := "aws_security_group.test"
	securityGroup2ResourceName := "aws_security_group.test2"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCSecurityGroupIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerVPCSecurityGroupIdsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccServer_vpcAddressAllocationIds_securityGroupIDs(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	eip1ResourceName := "aws_eip.test.0"
	eip2ResourceName := "aws_eip.test.1"
	securityGroup1ResourceName := "aws_security_group.test"
	securityGroup2ResourceName := "aws_security_group.test2"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCAddressAllocationIdsSecurityGroupIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerVPCAddressAllocationIdsSecurityGroupIdsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccServer_updateEndpointType_publicToVPC(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				Config: testAccServerVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_publicToVPC_addressAllocationIDs(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				Config: testAccServerVPCAddressAllocationIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_vpcEndpointToVPC(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				Config: testAccServerVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_vpcEndpointToVPC_addressAllocationIDs(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
				),
			},
			{
				Config: testAccServerVPCAddressAllocationIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_vpcEndpointToVPC_securityGroupIDs(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				Config: testAccServerVPCSecurityGroupIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_updateEndpointType_vpcToPublic(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_protocols(t *testing.T) {
	var s transfer.DescribedServer
	var ca acmpca.CertificateAuthority
	resourceName := "aws_transfer_server.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerProtocolsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &s),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTP"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			// We need to create and activate the CA before issuing a certificate.
			{
				Config: testAccServerRootCAConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(&ca),
				),
			},
			{
				Config: testAccServerProtocolsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &s),
					resource.TestCheckResourceAttrPair(resourceName, "certificate", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTP"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTPS"),
				),
			},
			{
				Config: testAccServerProtocolsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(&ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccServer_apiGateway(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerAPIGatewayIdentityProviderTypeConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_role", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_apiGateway_forceDestroy(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerAPIGatewayIdentityProviderTypeConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_role", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_directoryService(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDirectoryServiceIdentityProviderTypeConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "AWS_DIRECTORY_SERVICE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_forceDestroy(t *testing.T) {
	var s transfer.DescribedServer
	var u transfer.DescribedUser
	var k transfer.SshPublicKey
	resourceName := "aws_transfer_server.test"
	userResourceName := "aws_transfer_user.test"
	sshKeyResourceName := "aws_transfer_ssh_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerForceDestroyConfig(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &s),
					testAccCheckUserExists(userResourceName, &u),
					testAccCheckSSHKeyExists(sshKeyResourceName, &k),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccServer_hostKey(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	hostKey := "test-fixtures/transfer-ssh-rsa-key"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerHostKeyConfig(hostKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "host_key_fingerprint", "SHA256:Z2pW9sPKDD/T34tVfCoolsRcECNTlekgaKvDn9t+9sg="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccServer_vpcEndpointID(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerVPCEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_endpoint_id", vpcEndpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.vpc_id", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccServer_lambdaFunction(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerLambdaFunctionIdentityProviderTypeConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttrPair(resourceName, "function", "aws_lambda_function.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_authenticationLoginBanners(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDisplayBannersConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "post_authentication_login_banner", "This system is for the use of authorized users only - post"),
					resource.TestCheckResourceAttr(resourceName, "pre_authentication_login_banner", "This system is for the use of authorized users only - pre"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccServer_workflowDetails(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerWorkflowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.0.on_upload.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_upload.0.execution_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_upload.0.workflow_id", "aws_transfer_workflow.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccServerWorkflowConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow_details.0.on_upload.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_upload.0.execution_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_details.0.on_upload.0.workflow_id", "aws_transfer_workflow.test2", "id"),
				),
			},
		},
	})
}

func testAccCheckServerExists(n string, v *transfer.DescribedServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Server ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn

		output, err := tftransfer.FindServerByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckServerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_server" {
			continue
		}

		_, err := tftransfer.FindServerByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Transfer Server %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn

	input := &transfer.ListServersInput{}

	_, err := conn.ListServers(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccServerBaseVPCConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_subnet" "test2" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true

  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_eip" "test" {
  count = 2

  vpc = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerBaseLoggingRoleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoCloudWatchLogs",
    "Effect": "Allow",
    "Action": [
      "logs:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}
`, rName)
}

func testAccServerBaseAPIGatewayConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_name        = "test"
  description       = %[1]q
  stage_description = %[1]q

  variables = {
    "a" = "2"
  }
}
`, rName)
}

func testAccServerBasicConfig() string {
	return `
resource "aws_transfer_server" "test" {}
`
}

func testAccServerDisplayBannersConfig() string {
	return `
resource "aws_transfer_server" "test" {
  pre_authentication_login_banner  = "This system is for the use of authorized users only - pre"
  post_authentication_login_banner = "This system is for the use of authorized users only - post"
}
`
}

func testAccServerDomainConfig() string {
	return `
resource "aws_transfer_server" "test" {
  domain = "EFS"
}
`
}

func testAccServerSecurityPolicyConfig(policy string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  security_policy_name = %[1]q
}
`, policy)
}

func testAccServerUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseLoggingRoleConfig(rName),
		`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role           = aws_iam_role.test.arn
}
`)
}

func testAccServerAPIGatewayIdentityProviderTypeConfig(rName string, forceDestroy bool) string {
	return acctest.ConfigCompose(
		testAccServerBaseAPIGatewayConfig(rName),
		testAccServerBaseLoggingRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn

  force_destroy = %[1]t
}
`, forceDestroy))
}

func testAccServerDirectoryServiceIdentityProviderTypeConfig(rName string, forceDestroy bool) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		testAccServerBaseDirectoryServiceConfig(rName),
		testAccServerBaseLoggingRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "AWS_DIRECTORY_SERVICE"
  directory_id           = aws_directory_service_directory.test.id
  logging_role           = aws_iam_role.test.arn

  force_destroy = %[1]t
}
`, forceDestroy))
}

func testAccServerBaseDirectoryServiceConfig(rName string) string {
	return `
resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"

  vpc_settings {
    vpc_id = aws_vpc.test.id

    subnet_ids = [
      aws_subnet.test.id,
      aws_subnet.test2.id
    ]
  }
}
`
}

func testAccServerForceDestroyConfig(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoS3",
    "Effect": "Allow",
    "Action": [
      "s3:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = %[1]q
  role      = aws_iam_role.test.arn
}

resource "aws_transfer_ssh_key" "test" {
  server_id = aws_transfer_server.test.id
  user_name = aws_transfer_user.test.user_name
  body      = "%[2]s"
}
`, rName, publicKey)
}

func testAccServerVPCEndpointConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "transfer.server"
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  vpc_endpoint_type = "Interface"
  service_name      = data.aws_vpc_endpoint_service.test.service_name

  security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_transfer_server" "test" {
  endpoint_type = "VPC_ENDPOINT"

  endpoint_details {
    vpc_endpoint_id = aws_vpc_endpoint.test.id
  }
}
`, rName))
}

func testAccServerVPCConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    vpc_id = aws_vpc.test.id
  }
}
`)
}

func testAccServerVPCUpdateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }
}
`)
}

func testAccServerVPCAddressAllocationIdsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[0].id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`)
}

func testAccServerVPCAddressAllocationIdsUpdateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[1].id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`)
}

func testAccServerVPCAddressAllocationIdsSecurityGroupIdsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[0].id]
    security_group_ids     = [aws_security_group.test.id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`)
}

func testAccServerVPCAddressAllocationIdsSecurityGroupIdsUpdateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[1].id]
    security_group_ids     = [aws_security_group.test2.id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`, rName))
}

func testAccServerVPCSecurityGroupIdsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    security_group_ids = [aws_security_group.test.id]
    vpc_id             = aws_vpc.test.id
  }
}
`)
}

func testAccServerVPCSecurityGroupIdsUpdateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    security_group_ids = [aws_security_group.test2.id]
    vpc_id             = aws_vpc.test.id
  }
}
`, rName))
}

func testAccServerHostKeyConfig(hostKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  host_key = file(%[1]q)
}
`, hostKey)
}

func testAccServerProtocolsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		testAccServerBaseAPIGatewayConfig(rName),
		testAccServerBaseLoggingRoleConfig(rName),
		`
resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
  protocols              = ["FTP"]

  endpoint_type = "VPC"
  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }
}
`)
}

func testAccServerRootCAConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "%[1]s.com"
    }
  }
}
`, rName)
}

func testAccServerProtocolsUpdateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccServerBaseVPCConfig(rName),
		testAccServerBaseAPIGatewayConfig(rName),
		testAccServerBaseLoggingRoleConfig(rName),
		testAccServerRootCAConfig(rName),
		fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = "test.%[1]s.com"
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
  protocols              = ["FTP", "FTPS"]
  certificate            = aws_acm_certificate.test.arn

  endpoint_type = "VPC"
  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }
}
`, rName))
}

func testAccServerLambdaFunctionIdentityProviderTypeConfig(rName string, forceDestroy bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		testAccServerBaseLoggingRoleConfig(rName+"-logging"),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs14.x"
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "AWS_LAMBDA"
  function               = aws_lambda_function.test.arn
  logging_role           = aws_iam_role.test.arn

  force_destroy = %[2]t
}
`, rName, forceDestroy))
}

func testAccServerWorkflowConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_transfer_workflow" "test" {
  steps {
    delete_step_details {
      name                 = "test"
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}

resource "aws_transfer_server" "test" {
  workflow_details {
    on_upload {
      execution_role = aws_iam_role.test.arn
      workflow_id    = aws_transfer_workflow.test.id
    }
  }
}
`, rName)
}

func testAccServerWorkflowConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_transfer_workflow" "test" {
  steps {
    delete_step_details {
      name                 = "test"
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}

resource "aws_transfer_workflow" "test2" {
  steps {
    delete_step_details {
      name                 = "test"
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}

resource "aws_transfer_server" "test" {
  workflow_details {
    on_upload {
      execution_role = aws_iam_role.test.arn
      workflow_id    = aws_transfer_workflow.test2.id
    }
  }
}
`, rName)
}
