package ec2_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Due to the nature of byoip cidrs, we have each possible test represented as a single test with
// multiple steps in order to share the dependencies

// IPAM IPv6 BYOIP Tests
func TestAccIPAM_byoipIPv6(t *testing.T) {
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
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCIPv6CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: ipv4IPAMBYOIPIPv6DefaultNetmask(p, m, s),
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
				Config: testAccIPAMConfig_ipv6BYOIPCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config: testAccIPAMConfig_ipv6BYOIPExplicitNetmask(p, m, s, netmaskLength),
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
				Config: testAccIPAMConfig_ipv6BYOIPCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config:   testAccIPAMConfig_ipv6BYOIPExplicitCIDR(p, m, s, ipv6CidrVPC),
				SkipFunc: testAccIPAMConfig_ipv6BYOIPSkipExplicitCIDR(t, ipv6CidrVPC),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ipv6CidrVPC),
				),
			},
			// disassociate ipv6
			{
				Config:   testAccIPAMConfig_ipv6BYOIPCIDRBase(p, m, s),
				SkipFunc: testAccIPAMConfig_ipv6BYOIPSkipExplicitCIDR(t, ipv6CidrVPC),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			// aws_vpc_ipv6_cidr_block_association
			{
				Config: testAccIPAMConfig_byoipIPv6CIDRBlockAssociationIPAMDefaultNetmask(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, strconv.Itoa(netmaskLength)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccIPAMConfig_ipv6BYOIPCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc)),
				// vpc will still have association id because its based on the aws_vpc_ipv6_cidr_block_association resource
			},
			{
				Config: testAccIPAMConfig_ipv6BYOIPCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config: testAccIPAMConfig_byoipIPv6CIDRBlockAssociationIPAMExplicitNetmask(p, m, s, netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, strconv.Itoa(netmaskLength)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccIPAMConfig_ipv6BYOIPCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc)),
				// vpc will still have association id because its based on the aws_vpc_ipv6_cidr_block_association resource
			},
			{
				Config: testAccIPAMConfig_ipv6BYOIPCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config:   testAccIPAMConfig_byoipIPv6CIDRBlockAssociationIPAMExplicitCIDR(p, m, s, ipv6CidrAssoc),
				SkipFunc: testAccIPAMConfig_ipv6BYOIPSkipExplicitCIDR(t, ipv6CidrAssoc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					resource.TestCheckResourceAttr(assocName, "ipv6_cidr_block", ipv6CidrAssoc),
				),
			},
		},
	})
}

func testAccIPAMConfig_ipv6BYOIPSkipExplicitCIDR(t *testing.T, ipv6CidrVPC string) func() (bool, error) {
	return func() (bool, error) {
		if ipv6CidrVPC != "" {
			return false, nil
		}
		t.Log("Skipping step: Environment variable IPAM_BYOIP_IPV6_EXPLICIT_CIDR_VPC or IPAM_BYOIP_IPV6_EXPLICIT_CIDR_ASSOCIATE must be set.")
		return true, nil
	}
}

func testAccIPAMConfig_ipv6BYOIPCIDRBase(cidr, msg, signature string) string {
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

func ipv4IPAMBYOIPIPv6DefaultNetmask(cidr, msg, signature string) string {
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

func testAccIPAMConfig_ipv6BYOIPExplicitNetmask(cidr, msg, signature string, netmask int) string {
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

func testAccIPAMConfig_ipv6BYOIPExplicitCIDR(cidr, msg, signature, vpcCidr string) string {
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

func testAccIPAMConfig_byoipIPv6CIDRBlockAssociationIPAMDefaultNetmask(cidr, msg, signature string) string {
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

func testAccIPAMConfig_byoipIPv6CIDRBlockAssociationIPAMExplicitNetmask(cidr, msg, signature string, netmask int) string {
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

func testAccIPAMConfig_byoipIPv6CIDRBlockAssociationIPAMExplicitCIDR(cidr, msg, signature, vpcCidr string) string {
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
