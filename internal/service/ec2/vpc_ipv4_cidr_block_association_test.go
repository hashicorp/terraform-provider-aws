package ec2_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccVPCIPv4CIDRBlockAssociation_basic(t *testing.T) {
	var associationSecondary, associationTertiary ec2.VpcCidrBlockAssociation

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCIPv4CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists("aws_vpc_ipv4_cidr_block_association.secondary_cidr", &associationSecondary),
					testAccCheckAdditionalVPCIPv4CIDRBlock(&associationSecondary, "172.2.0.0/16"),
					testAccCheckVPCIPv4CIDRBlockAssociationExists("aws_vpc_ipv4_cidr_block_association.tertiary_cidr", &associationTertiary),
					testAccCheckAdditionalVPCIPv4CIDRBlock(&associationTertiary, "170.2.0.0/16"),
				),
			},
			{
				ResourceName:      "aws_vpc_ipv4_cidr_block_association.secondary_cidr",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_vpc_ipv4_cidr_block_association.tertiary_cidr",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCIPv4CIDRBlockAssociation_IpamBasic(t *testing.T) {
	var associationSecondary ec2.VpcCidrBlockAssociation
	netmaskLength := "28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCIPv4CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationIpam(netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists("aws_vpc_ipv4_cidr_block_association.secondary_cidr", &associationSecondary),
					testAccCheckVPCAssociationCIDRPrefix(&associationSecondary, netmaskLength),
				),
			},
		},
	})
}

func TestAccVPCIPv4CIDRBlockAssociation_IpamBasicExplicitCIDR(t *testing.T) {
	var associationSecondary ec2.VpcCidrBlockAssociation
	cidr := "172.2.0.32/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCIPv4CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationIpamExplicitCIDR(cidr),
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

		// Try to find the VPC
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.Attributes["vpc_id"])},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err == nil {
			vpc := resp.Vpcs[0]

			for _, ipv4Association := range vpc.CidrBlockAssociationSet {
				if *ipv4Association.AssociationId == rs.Primary.ID {
					return fmt.Errorf("VPC CIDR block association still exists")
				}
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidVpcID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckVPCIPv4CIDRBlockAssociationExists(n string, association *ec2.VpcCidrBlockAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.Attributes["vpc_id"])},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 {
			return fmt.Errorf("VPC not found")
		}

		vpc := resp.Vpcs[0]
		found := false
		for _, cidrAssociation := range vpc.CidrBlockAssociationSet {
			if *cidrAssociation.AssociationId == rs.Primary.ID {
				*association = *cidrAssociation
				found = true
			}
		}

		if !found {
			return fmt.Errorf("VPC CIDR block association not found")
		}

		return nil
	}
}

const testAccVPCIPv4CIDRBlockAssociationConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-ipv4-cidr-block-association"
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  vpc_id     = aws_vpc.foo.id
  cidr_block = "172.2.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "tertiary_cidr" {
  vpc_id     = aws_vpc.foo.id
  cidr_block = "170.2.0.0/16"
}
`

func testAccVPCIPv4CIDRBlockAssociationIpam(netmaskLength string) string {
	return testAccVpcIpamBase + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-ipv4-cidr-block-association"
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = %[1]q
  vpc_id              = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`, netmaskLength)
}

func testAccVPCIPv4CIDRBlockAssociationIpamExplicitCIDR(cidr string) string {
	return testAccVpcIpamBase + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-ipv4-cidr-block-association"
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr_block        = %[1]q
  vpc_id            = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`, cidr)
}
