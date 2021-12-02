package ec2_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
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

// Due to the nature of byoip cidrs, we have each possible test represented as a single test with
// multiple steps in order to share the dependencies

// IPAM IPv6 BYOIP Tests
func TestAccVPCIpam_ByoipIPv6(t *testing.T) {
	if os.Getenv("IPAM_BYOIP_IPV6_MESSAGE") == "" || os.Getenv("IPAM_BYOIP_IPV6_SIGNATURE") == "" || os.Getenv("IPAM_BYOIP_IPV6_PROVISIONED_CIDR") == "" {
		t.Skip("Environment variable IPAM_BYOIP_IPV6_MESSAGE, IPAM_BYOIP_IPV6_SIGNATURE, or IPAM_BYOIP_IPV6_PROVISIONED_CIDR is not set")
	}

	var m string
	var s string
	var p string
	var ipv6CidrVPC string
	var ipv6CidrAssoc string

	// test passing an explicit CIDR to aws_vpc
	if os.Getenv("IPAM_BYOIP_IPV6_EXPLICIT_CIDR_VPC") != "" {
		ipv6CidrVPC = os.Getenv("IPAM_BYOIP_IPV6_EXPLICIT_CIDR_VPC")
	}

	// test passing an explicit CIDR to aws_vpc_ipv6_cidr_block_association
	if os.Getenv("IPAM_BYOIP_IPV6_EXPLICIT_CIDR_ASSOCIATE") != "" {
		ipv6CidrAssoc = os.Getenv("IPAM_BYOIP_IPV6_EXPLICIT_CIDR_ASSOCIATE")
	}

	m = os.Getenv("IPAM_BYOIP_IPV6_MESSAGE")
	s = os.Getenv("IPAM_BYOIP_IPV6_SIGNATURE")
	p = os.Getenv("IPAM_BYOIP_IPV6_PROVISIONED_CIDR")

	resourceName := "aws_vpc.test"
	assocName := "aws_vpc_ipv6_cidr_block_association.test"
	var vpc ec2.Vpc
	var associationIPv6 ec2.VpcIpv6CidrBlockAssociation
	netmaskLength := 56

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCIPv6CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: ipv4VPCIpamByoipIPv6DefaultNetmask(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckNoResourceAttr(resourceName, "ipv6_netmask_length"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config: testAccVPCIpamIPv6ByoipExplicitNetmask(p, m, s, netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", strconv.Itoa(netmaskLength)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
				),
			},
			// // disassociate ipv6
			{
				Config: testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config:   testAccVPCIpamIPv6ByoipExplicitCIDR(p, m, s, ipv6CidrVPC),
				SkipFunc: testAccVPCIpamIPv6ByoipSkipExplicitCidr(t, ipv6CidrVPC),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ipv6CidrVPC),
				),
			},
			// disassociate ipv6
			{
				Config:   testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				SkipFunc: testAccVPCIpamIPv6ByoipSkipExplicitCidr(t, ipv6CidrVPC),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			// aws_vpc_ipv6_cidr_block_association
			{
				Config: testAccVPCIpamByoipIPv6CIDRBlockAssociationIpamDefaultNetmask(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, strconv.Itoa(netmaskLength)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc)),
				// vpc will still have association id because its based on the aws_vpc_ipv6_cidr_block_association resource
			},
			{
				Config: testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config: testAccVPCIpamByoipIPv6CIDRBlockAssociationIpamExplicitNetmask(p, m, s, netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, strconv.Itoa(netmaskLength)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc)),
				// vpc will still have association id because its based on the aws_vpc_ipv6_cidr_block_association resource
			},
			{
				Config: testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config:   testAccVPCIpamByoipIPv6CIDRBlockAssociationIpamExplicitCIDR(p, m, s, ipv6CidrAssoc),
				SkipFunc: testAccVPCIpamIPv6ByoipSkipExplicitCidr(t, ipv6CidrAssoc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					resource.TestCheckResourceAttr(assocName, "ipv6_cidr_block", ipv6CidrAssoc),
				),
			},
		},
	})
}

func testAccCheckVPCIPv6CIDRBlockAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipv6_cidr_block_association" {
			continue
		}

		// Try to find the VPC
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.Attributes["vpc_id"])},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err == nil {
			vpc := resp.Vpcs[0]

			for _, ipv6Association := range vpc.Ipv6CidrBlockAssociationSet {
				if *ipv6Association.AssociationId == rs.Primary.ID {
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

func testAccCheckVPCIPv6CIDRBlockAssociationExists(n string, association *ec2.VpcIpv6CidrBlockAssociation) resource.TestCheckFunc {
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
		for _, ipv6CidrAssociation := range vpc.Ipv6CidrBlockAssociationSet {
			if *ipv6CidrAssociation.AssociationId == rs.Primary.ID {
				*association = *ipv6CidrAssociation
				found = true
			}
		}

		if !found {
			return fmt.Errorf("VPC CIDR block association not found")
		}

		return nil
	}
}

func testAccCheckVPCAssociationIPv6CIDRPrefix(association *ec2.VpcIpv6CidrBlockAssociation, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if strings.Split(aws.StringValue(association.Ipv6CidrBlock), "/")[1] != expected {
			return fmt.Errorf("Bad cidr prefix: %s", aws.StringValue(association.Ipv6CidrBlock))
		}

		return nil
	}
}

func testAccVPCIpamIPv6ByoipSkipExplicitCidr(t *testing.T, ipv6CidrVPC string) func() (bool, error) {
	return func() (bool, error) {
		if ipv6CidrVPC != "" {
			return false, nil
		}
		t.Log("Skipping step: Environment variable IPAM_BYOIP_IPV6_EXPLICIT_CIDR_VPC or IPAM_BYOIP_IPV6_EXPLICIT_CIDR_ASSOCIATE must be set.")
		return true, nil
	}
}

func testAccVPCIpamIPv6ByoipCIDRBase(cidr, msg, signature string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv6"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  publicly_advertisable             = false
  aws_service                       = "ec2"
  allocation_default_netmask_length = 56
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
	`, cidr, msg, signature)
}

func ipv4VPCIpamByoipIPv6DefaultNetmask(cidr, msg, signature string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv6"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  publicly_advertisable             = false
  aws_service                       = "ec2"
  allocation_default_netmask_length = 56
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  ipv6_ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr_block        = "10.0.0.0/16"
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
	`, cidr, msg, signature)
}

func testAccVPCIpamIPv6ByoipExplicitNetmask(cidr, msg, signature string, netmask int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv6"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  publicly_advertisable             = false
  aws_service                       = "ec2"
  allocation_default_netmask_length = 56
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = %[4]d
  cidr_block          = "10.0.0.0/16"
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
	`, cidr, msg, signature, netmask)
}

func testAccVPCIpamIPv6ByoipExplicitCIDR(cidr, msg, signature, vpcCidr string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv6"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  publicly_advertisable             = false
  aws_service                       = "ec2"
  allocation_default_netmask_length = 56
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  ipv6_ipam_pool_id = aws_vpc_ipam_pool.test.id
  ipv6_cidr_block   = %[4]q
  cidr_block        = "10.0.0.0/16"
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
	`, cidr, msg, signature, vpcCidr)
}

func testAccVPCIpamByoipIPv6CIDRBlockAssociationIpamDefaultNetmask(cidr, msg, signature string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv6"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  publicly_advertisable             = false
  aws_service                       = "ec2"
  allocation_default_netmask_length = 56
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv6_cidr_block_association" "test" {
  ipv6_ipam_pool_id = aws_vpc_ipam_pool.test.id
  vpc_id            = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
	`, cidr, msg, signature)
}

func testAccVPCIpamByoipIPv6CIDRBlockAssociationIpamExplicitNetmask(cidr, msg, signature string, netmask int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv6"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  publicly_advertisable             = false
  aws_service                       = "ec2"
  allocation_default_netmask_length = 56
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv6_cidr_block_association" "test" {
  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = %[4]d
  vpc_id              = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
	`, cidr, msg, signature, netmask)
}

func testAccVPCIpamByoipIPv6CIDRBlockAssociationIpamExplicitCIDR(cidr, msg, signature, vpcCidr string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv6"
  ipam_scope_id                     = aws_vpc_ipam.test.public_default_scope_id
  locale                            = data.aws_region.current.name
  publicly_advertisable             = false
  aws_service                       = "ec2"
  allocation_default_netmask_length = 56
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  cidr_authorization_context {
    message   = %[2]q
    signature = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv6_cidr_block_association" "test" {
  ipv6_ipam_pool_id = aws_vpc_ipam_pool.test.id
  ipv6_cidr_block   = %[4]q
  vpc_id            = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
	`, cidr, msg, signature, vpcCidr)
}
