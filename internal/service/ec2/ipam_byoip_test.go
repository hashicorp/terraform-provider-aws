// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Due to the nature of byoip cidrs, we have each possible test represented as a single test with
// multiple steps in order to share the dependencies

// IPAM IPv6 BYOIP Tests
func TestAccIPAM_byoipIPv6(t *testing.T) {
	ctx := acctest.Context(t)
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
	var vpc awstypes.Vpc
	var associationIPv6 awstypes.VpcIpv6CidrBlockAssociation
	netmaskLength := 56

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv6CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMBYOIPConfig_ipv4IPv6DefaultNetmask(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckNoResourceAttr(resourceName, "ipv6_netmask_length"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexache.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccIPAMBYOIPConfig_ipv6CIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config: testAccIPAMBYOIPConfig_ipv6ExplicitNetmask(p, m, s, netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", strconv.Itoa(netmaskLength)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexache.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
				),
			},
			// // disassociate ipv6
			{
				Config: testAccIPAMBYOIPConfig_ipv6CIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config:   testAccIPAMBYOIPConfig_ipv6ExplicitCIDR(p, m, s, ipv6CidrVPC),
				SkipFunc: testAccIPAMConfig_ipv6BYOIPSkipExplicitCIDR(t, ipv6CidrVPC),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc/vpc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexache.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ipv6CidrVPC),
				),
			},
			// disassociate ipv6
			{
				Config:   testAccIPAMBYOIPConfig_ipv6CIDRBase(p, m, s),
				SkipFunc: testAccIPAMConfig_ipv6BYOIPSkipExplicitCIDR(t, ipv6CidrVPC),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			// aws_vpc_ipv6_cidr_block_association
			{
				Config: testAccIPAMBYOIPConfig_ipv6CIDRBlockAssociationDefaultNetmask(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, strconv.Itoa(netmaskLength)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccIPAMBYOIPConfig_ipv6CIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc)),
				// vpc will still have association id because its based on the aws_vpc_ipv6_cidr_block_association resource
			},
			{
				Config: testAccIPAMBYOIPConfig_ipv6CIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config: testAccIPAMBYOIPConfig_ipv6CIDRBlockAssociationExplicitNetmask(p, m, s, netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, strconv.Itoa(netmaskLength)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccIPAMBYOIPConfig_ipv6CIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc)),
				// vpc will still have association id because its based on the aws_vpc_ipv6_cidr_block_association resource
			},
			{
				Config: testAccIPAMBYOIPConfig_ipv6CIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config:   testAccIPAMBYOIPConfig_ipv6CIDRBlockAssociationExplicitCIDR(p, m, s, ipv6CidrAssoc),
				SkipFunc: testAccIPAMConfig_ipv6BYOIPSkipExplicitCIDR(t, ipv6CidrAssoc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, assocName, &associationIPv6),
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

func testAccIPAMBYOIPConfig_ipv6CIDRBase(cidr, msg, signature string) string {
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

func testAccIPAMBYOIPConfig_ipv4IPv6DefaultNetmask(cidr, msg, signature string) string {
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

func testAccIPAMBYOIPConfig_ipv6ExplicitNetmask(cidr, msg, signature string, netmask int) string {
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

func testAccIPAMBYOIPConfig_ipv6ExplicitCIDR(cidr, msg, signature, vpcCidr string) string {
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

func testAccIPAMBYOIPConfig_ipv6CIDRBlockAssociationDefaultNetmask(cidr, msg, signature string) string {
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

func testAccIPAMBYOIPConfig_ipv6CIDRBlockAssociationExplicitNetmask(cidr, msg, signature string, netmask int) string {
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

func testAccIPAMBYOIPConfig_ipv6CIDRBlockAssociationExplicitCIDR(cidr, msg, signature, vpcCidr string) string {
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
