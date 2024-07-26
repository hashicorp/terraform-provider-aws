// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rInt1 := sdkacctest.RandIntRange(1, 128)
	rInt2 := sdkacctest.RandIntRange(128, 254)
	cidr := fmt.Sprintf("10.%d.%d.0/28", rInt1, rInt2)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	vpcResourceName := "aws_vpc.test"
	ds1ResourceName := "data.aws_vpc.by_id"
	ds2ResourceName := "data.aws_vpc.by_cidr"
	ds3ResourceName := "data.aws_vpc.by_tag"
	ds4ResourceName := "data.aws_vpc.by_filter"

	resource.ParallelTest(t, resource.TestCase{
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDataSourceConfig_cidrBlockAssociationsMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cidr_block_associations.#", acctest.Ct2),
				),
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
