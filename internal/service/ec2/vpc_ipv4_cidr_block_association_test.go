package ec2_test

import (
	"fmt"
	"strings"
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

func TestAccVPCIPv4CIDRBlockAssociation_basic(t *testing.T) {
	var associationSecondary, associationTertiary ec2.VpcCidrBlockAssociation
	resource1Name := "aws_vpc_ipv4_cidr_block_association.secondary_cidr"
	resource2Name := "aws_vpc_ipv4_cidr_block_association.tertiary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCIPv4CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists(resource1Name, &associationSecondary),
					testAccCheckAdditionalVPCIPv4CIDRBlock(&associationSecondary, "172.2.0.0/16"),
					testAccCheckVPCIPv4CIDRBlockAssociationExists(resource2Name, &associationTertiary),
					testAccCheckAdditionalVPCIPv4CIDRBlock(&associationTertiary, "170.2.0.0/16"),
				),
			},
			{
				ResourceName:      resource1Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCIPv4CIDRBlockAssociation_disappears(t *testing.T) {
	var associationSecondary, associationTertiary ec2.VpcCidrBlockAssociation
	resource1Name := "aws_vpc_ipv4_cidr_block_association.secondary_cidr"
	resource2Name := "aws_vpc_ipv4_cidr_block_association.tertiary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCIPv4CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists(resource1Name, &associationSecondary),
					testAccCheckVPCIPv4CIDRBlockAssociationExists(resource2Name, &associationTertiary),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPCIPv4CIDRBlockAssociation(), resource1Name),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCIPv4CIDRBlockAssociation_IpamBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var associationSecondary ec2.VpcCidrBlockAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCIPv4CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationIpam(rName, 28),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists("aws_vpc_ipv4_cidr_block_association.secondary_cidr", &associationSecondary),
					testAccCheckVPCAssociationCIDRPrefix(&associationSecondary, "28"),
				),
			},
		},
	})
}

func TestAccVPCIPv4CIDRBlockAssociation_IpamBasicExplicitCIDR(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var associationSecondary ec2.VpcCidrBlockAssociation
	cidr := "172.2.0.32/28"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCIPv4CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationIpamExplicitCIDR(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists("aws_vpc_ipv4_cidr_block_association.secondary_cidr", &associationSecondary),
					testAccCheckAdditionalVPCIPv4CIDRBlock(&associationSecondary, cidr)),
			},
		},
	})
}

func testAccCheckAdditionalVPCIPv4CIDRBlock(association *ec2.VpcCidrBlockAssociation, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		CIDRBlock := association.CidrBlock
		if *CIDRBlock != expected {
			return fmt.Errorf("Bad CIDR: %s", *association.CidrBlock)
		}

		return nil
	}
}

func testAccCheckVPCAssociationCIDRPrefix(association *ec2.VpcCidrBlockAssociation, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if strings.Split(aws.StringValue(association.CidrBlock), "/")[1] != expected {
			return fmt.Errorf("Bad cidr prefix: %s", aws.StringValue(association.CidrBlock))
		}

		return nil
	}
}

func testAccCheckVPCIPv4CIDRBlockAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipv4_cidr_block_association" {
			continue
		}

		_, _, err := tfec2.FindVPCCIDRBlockAssociationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 VPC IPv4 CIDR Block Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVPCIPv4CIDRBlockAssociationExists(n string, v *ec2.VpcCidrBlockAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC IPv4 CIDR Block Association is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, _, err := tfec2.FindVPCCIDRBlockAssociationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCIPv4CIDRBlockAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "172.2.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "tertiary_cidr" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "170.2.0.0/16"
}
`, rName)
}

func testAccVPCIPv4CIDRBlockAssociationIpam(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVpcIPv4IPAMConfigBase(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = %[2]d
  vpc_id              = aws_vpc.test.id

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, netmaskLength))
}

func testAccVPCIPv4CIDRBlockAssociationIpamExplicitCIDR(rName, cidr string) string {
	return acctest.ConfigCompose(testAccVpcIPv4IPAMConfigBase(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr_block        = %[2]q
  vpc_id            = aws_vpc.test.id

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, cidr))
}
