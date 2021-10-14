package ec2_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func testAccClientVPNNetworkAssociation_basic(t *testing.T) {
	var assoc ec2.TargetNetwork
	var group ec2.SecurityGroup
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigBasic(rStr),
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
	rStr := sdkacctest.RandString(5)
	resourceNames := []string{"aws_ec2_client_vpn_network_association.test", "aws_ec2_client_vpn_network_association.test2"}
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceNames := []string{"aws_subnet.test", "aws_subnet.test2"}
	vpcResourceName := "aws_vpc.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigMultipleSubnets(rStr),
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
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigBasic(rStr),
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
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"
	securityGroup1ResourceName := "aws_security_group.test1"
	securityGroup2ResourceName := "aws_security_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationTwoSecurityGroups(rStr),
				Check: resource.ComposeTestCheckFunc(
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
				Config: testAccEc2ClientVpnNetworkAssociationOneSecurityGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(resourceName, &assoc2),
					testAccCheckDefaultSecurityGroupExists(securityGroup1ResourceName, &group21),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					testAccCheckClientVPNNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group21),
				),
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

		resp, _ := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		for _, v := range resp.ClientVpnTargetNetworks {
			if *v.AssociationId == rs.Primary.ID && !(*v.Status.Code == ec2.AssociationStatusCodeDisassociated) {
				return fmt.Errorf("[DESTROY ERROR] Client VPN network association (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckClientVPNNetworkAssociationExists(name string, assoc *ec2.TargetNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		resp, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Error reading Client VPN network association (%s): %w", rs.Primary.ID, err)
		}

		for _, a := range resp.ClientVpnTargetNetworks {
			if *a.AssociationId == rs.Primary.ID && !(*a.Status.Code == ec2.AssociationStatusCodeDisassociated) {
				*assoc = *a
				return nil
			}
		}

		return fmt.Errorf("Client VPN network association (%s) not found", rs.Primary.ID)
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

		return tfec2.ClientVPNNetworkAssociationCreateID(rs.Primary.Attributes["client_vpn_endpoint_id"], rs.Primary.ID), nil
	}
}

func testAccEc2ClientVpnNetworkAssociationConfigBasic(rName string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test.id
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationConfigMultipleSubnets(rName string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test.id
}

resource "aws_ec2_client_vpn_network_association" "test2" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test2.id
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationTwoSecurityGroups(rName string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test.id
  security_groups        = [aws_security_group.test1.id, aws_security_group.test2.id]
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}

resource "aws_security_group" "test1" {
  name        = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  name        = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationOneSecurityGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test.id
  security_groups        = [aws_security_group.test1.id]
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}

resource "aws_security_group" "test1" {
  name        = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  name        = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationVpcBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-%[1]s"
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%[1]s"
  }
}

resource "aws_subnet" "test2" {
  availability_zone       = data.aws_availability_zones.available.names[1]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%[1]s-2"
  }
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationAcmCertificateBase() string {
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}
