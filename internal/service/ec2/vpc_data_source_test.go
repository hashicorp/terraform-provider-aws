// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rInt1 := acctest.RandIntRange(t, 1, 128)
	rInt2 := acctest.RandIntRange(t, 128, 254)
	cidr := fmt.Sprintf("10.%d.%d.0/28", rInt1, rInt2)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	vpcResourceName := "aws_vpc.test"
	ds1ResourceName := "data.aws_vpc.by_id"
	ds2ResourceName := "data.aws_vpc.by_cidr"
	ds3ResourceName := "data.aws_vpc.by_tag"
	ds4ResourceName := "data.aws_vpc.by_filter"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDataSourceConfig_basic(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrARN, vpcResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(ds1ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds1ResourceName, "enable_dns_hostnames", acctest.CtFalse),
					resource.TestCheckResourceAttr(ds1ResourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(ds1ResourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "ipv6_association_id", vpcResourceName, "ipv6_association_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "ipv6_cidr_block", vpcResourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(ds1ResourceName, "ipv6_cidr_block_associations.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(ds1ResourceName, "ipv6_cidr_block_associations.*.association_id", vpcResourceName, "ipv6_association_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "main_route_table_id", vpcResourceName, "main_route_table_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrOwnerID, vpcResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.Name", rName),

					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrOwnerID, vpcResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(ds2ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds2ResourceName, "tags.Name", rName),

					resource.TestCheckResourceAttrPair(ds3ResourceName, names.AttrID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds3ResourceName, names.AttrOwnerID, vpcResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(ds3ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds3ResourceName, "tags.Name", rName),

					resource.TestCheckResourceAttrPair(ds4ResourceName, names.AttrID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds4ResourceName, names.AttrOwnerID, vpcResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(ds4ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds4ResourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccVPCDataSource_CIDRBlockAssociations_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDataSourceConfig_cidrBlockAssociationsMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cidr_block_associations.#", "2"),
				),
			},
		},
	})
}

func TestAccVPCDataSource_IPv6CIDRBlockAssociations_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDataSourceConfig_IPv6CIDRBlockAssociations_amazonProvided(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipv6_cidr_block_associations.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ipv6_cidr_block_associations.*.association_id", "aws_vpc.test", "ipv6_association_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ipv6_cidr_block_associations.*.association_id", "aws_vpc_ipv6_cidr_block_association.test", names.AttrID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("ipv6_cidr_block_associations"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrAssociationID:  knownvalue.NotNull(),
							"ipv6_address_attribute": tfknownvalue.StringExact(awstypes.Ipv6AddressAttributePublic),
							"ipv6_cidr_block":        knownvalue.StringRegexp(regexache.MustCompile(`/56$`)),
							"ipv6_pool":              knownvalue.StringExact("Amazon"),
							"network_border_group":   knownvalue.StringExact(acctest.Region()),
							"ip_source":              tfknownvalue.StringExact(awstypes.IpSourceAmazon),
							names.AttrState:          tfknownvalue.StringExact(awstypes.VpcCidrBlockStateCodeAssociated),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrAssociationID:  knownvalue.NotNull(),
							"ipv6_address_attribute": tfknownvalue.StringExact(awstypes.Ipv6AddressAttributePublic),
							"ipv6_cidr_block":        knownvalue.StringRegexp(regexache.MustCompile(`/56$`)),
							"ipv6_pool":              knownvalue.StringExact("Amazon"),
							"network_border_group":   knownvalue.StringExact(acctest.Region()),
							"ip_source":              tfknownvalue.StringExact(awstypes.IpSourceAmazon),
							names.AttrState:          tfknownvalue.StringExact(awstypes.VpcCidrBlockStateCodeAssociated),
						}),
					})),
				},
			},
		},
	})
}

func TestAccVPCDataSource_IPv6CIDRBlockAssociations_ipam(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDataSourceConfig_IPv6CIDRBlockAssociations_ipam(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipv6_cidr_block_associations.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ipv6_cidr_block_associations.*.association_id", "aws_vpc.test", "ipv6_association_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ipv6_cidr_block_associations.*.association_id", "aws_vpc_ipv6_cidr_block_association.test", names.AttrID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("ipv6_cidr_block_associations"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrAssociationID:  knownvalue.NotNull(),
							"ipv6_address_attribute": tfknownvalue.StringExact(awstypes.Ipv6AddressAttributePublic),
							"ipv6_cidr_block":        knownvalue.StringRegexp(regexache.MustCompile(`/56$`)),
							"ipv6_pool":              knownvalue.StringExact("IPAM Managed"),
							"network_border_group":   knownvalue.StringExact(acctest.Region()),
							"ip_source":              tfknownvalue.StringExact(awstypes.IpSourceAmazon),
							names.AttrState:          tfknownvalue.StringExact(awstypes.VpcCidrBlockStateCodeAssociated),
						}),
					})),
				},
			},
		},
	})
}

func testAccVPCDataSourceConfig_basic(rName, cidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = %[2]q

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc" "by_id" {
  id = aws_vpc.test.id
}

data "aws_vpc" "by_cidr" {
  cidr_block = aws_vpc.test.cidr_block
}

data "aws_vpc" "by_tag" {
  tags = {
    Name = aws_vpc.test.tags["Name"]
  }
}

data "aws_vpc" "by_filter" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }
}
`, rName, cidr)
}

func testAccVPCDataSourceConfig_cidrBlockAssociationsMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "172.0.0.0/16"
}

data "aws_vpc" "test" {
  id = aws_vpc_ipv4_cidr_block_association.test.vpc_id
}
`, rName)
}

func testAccVPCDataSourceConfig_IPv6CIDRBlockAssociations_amazonProvided() string {
	return `
data "aws_vpc" "test" {
  id = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipv6_cidr_block_association.test,
  ]
}

resource "aws_vpc" "test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true
}

resource "aws_vpc_ipv6_cidr_block_association" "test" {
  vpc_id                           = aws_vpc.test.id
  assign_generated_ipv6_cidr_block = true
}
`
}

func testAccVPCDataSourceConfig_IPv6CIDRBlockAssociations_ipam(rName string) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv6(rName), `
data "aws_vpc" "test" {
  id = aws_vpc.test.id
  depends_on = [
    aws_vpc_ipv6_cidr_block_association.test,
  ]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = 56

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}

resource "aws_vpc_ipv6_cidr_block_association" "test" {
  vpc_id = aws_vpc.test.id

  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = 56
}
`)
}
