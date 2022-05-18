package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccClientVPNNetworkAssociation_basic(t *testing.T) {
	var assoc ec2.TargetNetwork
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_network_association.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test1"
	vpcResourceName := "aws_vpc.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(resourceName, &assoc),
					resource.TestMatchResourceAttr(resourceName, "association_id", regexp.MustCompile("^cvpn-assoc-[a-z0-9]+$")),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					testAccCheckDefaultSecurityGroupExists(defaultSecurityGroupResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					testAccCheckClientVPNNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccClientVPNNetworkAssociationImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccClientVPNNetworkAssociation_multipleSubnets(t *testing.T) {
	var assoc ec2.TargetNetwork
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceNames := []string{"aws_ec2_client_vpn_network_association.test1", "aws_ec2_client_vpn_network_association.test2"}
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceNames := []string{"aws_subnet.test1", "aws_subnet.test2"}
	vpcResourceName := "aws_vpc.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigMultipleSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(resourceNames[0], &assoc),
					resource.TestMatchResourceAttr(resourceNames[0], "association_id", regexp.MustCompile("^cvpn-assoc-[a-z0-9]+$")),
					resource.TestMatchResourceAttr(resourceNames[1], "association_id", regexp.MustCompile("^cvpn-assoc-[a-z0-9]+$")),
					resource.TestCheckResourceAttrPair(resourceNames[0], "id", resourceNames[0], "association_id"),
					resource.TestCheckResourceAttrPair(resourceNames[0], "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceNames[0], "subnet_id", subnetResourceNames[0], "id"),
					resource.TestCheckResourceAttrPair(resourceNames[1], "subnet_id", subnetResourceNames[1], "id"),
					testAccCheckDefaultSecurityGroupExists(defaultSecurityGroupResourceName, &group),
					resource.TestCheckResourceAttr(resourceNames[0], "security_groups.#", "1"),
					testAccCheckClientVPNNetworkAssociationSecurityGroupID(resourceNames[0], "security_groups.*", &group),
					resource.TestCheckResourceAttrPair(resourceNames[0], "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceNames[0],
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccClientVPNNetworkAssociationImportStateIdFunc(resourceNames[0]),
			},
			{
				ResourceName:      resourceNames[1],
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccClientVPNNetworkAssociationImportStateIdFunc(resourceNames[1]),
			},
		},
	})
}

func testAccClientVPNNetworkAssociation_disappears(t *testing.T) {
	var assoc ec2.TargetNetwork
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_network_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(resourceName, &assoc),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceClientVPNNetworkAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClientVPNNetworkAssociation_securityGroups(t *testing.T) {
	var assoc1, assoc2 ec2.TargetNetwork
	var group11, group12, group21 ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_network_association.test"
	securityGroup1ResourceName := "aws_security_group.test1"
	securityGroup2ResourceName := "aws_security_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigTwoSecurityGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(resourceName, &assoc1),
					testAccCheckDefaultSecurityGroupExists(securityGroup1ResourceName, &group11),
					testAccCheckDefaultSecurityGroupExists(securityGroup2ResourceName, &group12),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					testAccCheckClientVPNNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group11),
					testAccCheckClientVPNNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group12),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccClientVPNNetworkAssociationImportStateIdFunc(resourceName),
			},
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigOneSecurityGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(resourceName, &assoc2),
					testAccCheckDefaultSecurityGroupExists(securityGroup1ResourceName, &group21),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					testAccCheckClientVPNNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group21),
				),
			},
		},
	})
}

func testAccClientVPNNetworkAssociation_securityGroupsOnEndpoint(t *testing.T) {
	var assoc ec2.TargetNetwork
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_network_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigTwoSecurityGroupsOnEndpoint(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(resourceName, &assoc),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccClientVPNNetworkAssociationImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccCheckClientVPNNetworkAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_network_association" {
			continue
		}

		_, err := tfec2.FindClientVPNNetworkAssociationByIDs(conn, rs.Primary.ID, rs.Primary.Attributes["client_vpn_endpoint_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Client VPN Network Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckClientVPNNetworkAssociationExists(name string, v *ec2.TargetNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Client VPN Network Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindClientVPNNetworkAssociationByIDs(conn, rs.Primary.ID, rs.Primary.Attributes["client_vpn_endpoint_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClientVPNNetworkAssociationSecurityGroupID(name, key string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckTypeSetElemAttr(name, key, aws.StringValue(group.GroupId))(s)
	}
}

func testAccClientVPNNetworkAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["client_vpn_endpoint_id"], rs.Primary.ID), nil
	}
}

func testAccEc2ClientVpnNetworkAssociationConfigBase(rName string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfig(rName),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test1" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  availability_zone       = data.aws_availability_zones.available.names[1]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationConfigBasic(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnNetworkAssociationConfigBase(rName), `
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test1.id
}
`)
}

func testAccEc2ClientVpnNetworkAssociationConfigMultipleSubnets(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnNetworkAssociationConfigBase(rName), `
resource "aws_ec2_client_vpn_network_association" "test1" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test1.id
}

resource "aws_ec2_client_vpn_network_association" "test2" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test2.id
}
`)
}

func testAccEc2ClientVpnNetworkAssociationConfigTwoSecurityGroups(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnNetworkAssociationConfigBase(rName), fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test1.id
  security_groups        = [aws_security_group.test1.id, aws_security_group.test2.id]
}

resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationConfigOneSecurityGroup(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnNetworkAssociationConfigBase(rName), fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test1.id
  security_groups        = [aws_security_group.test1.id]
}

resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationConfigTwoSecurityGroupsOnEndpoint(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigSecurityGroups(rName, 2), `
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}
`)
}
