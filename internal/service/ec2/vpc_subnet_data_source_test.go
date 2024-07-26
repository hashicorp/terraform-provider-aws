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

func TestAccVPCSubnetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandIntRange(0, 256)
	cidr := fmt.Sprintf("172.%d.123.0/24", rInt)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	snResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	ds1ResourceName := "data.aws_subnet.by_id"
	ds2ResourceName := "data.aws_subnet.by_cidr"
	ds3ResourceName := "data.aws_subnet.by_tag"
	ds4ResourceName := "data.aws_subnet.by_vpc"
	ds5ResourceName := "data.aws_subnet.by_filter"
	ds6ResourceName := "data.aws_subnet.by_az_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetDataSourceConfig_basic(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrID, snResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrOwnerID, snResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrAvailabilityZone, snResourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds1ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(ds1ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrARN, snResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrID, snResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrOwnerID, snResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrAvailabilityZone, snResourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds2ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(ds2ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds2ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrARN, snResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds3ResourceName, names.AttrID, snResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds3ResourceName, names.AttrOwnerID, snResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds3ResourceName, names.AttrAvailabilityZone, snResourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds3ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(ds3ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds3ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds3ResourceName, names.AttrARN, snResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds4ResourceName, names.AttrID, snResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds4ResourceName, names.AttrOwnerID, snResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds4ResourceName, names.AttrAvailabilityZone, snResourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(ds4ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds4ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds4ResourceName, names.AttrARN, snResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds5ResourceName, names.AttrID, snResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds5ResourceName, names.AttrOwnerID, snResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds5ResourceName, names.AttrAvailabilityZone, snResourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(ds5ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds5ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds5ResourceName, names.AttrARN, snResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds6ResourceName, names.AttrID, snResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds6ResourceName, names.AttrOwnerID, snResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds6ResourceName, names.AttrAvailabilityZone, snResourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(ds6ResourceName, names.AttrCIDRBlock, cidr),
					resource.TestCheckResourceAttr(ds6ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds6ResourceName, names.AttrARN, snResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "outpost_arn", snResourceName, "outpost_arn"),
				),
			},
		},
	})
}

func TestAccVPCSubnetDataSource_ipv6ByIPv6Filter(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetDataSourceConfig_ipv6(rName, rInt),
			},
			{
				Config: testAccVPCSubnetDataSourceConfig_ipv6Filter(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block_association_id"),
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block"),
				),
			},
		},
	})
}

func TestAccVPCSubnetDataSource_ipv6ByIPv6CIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetDataSourceConfig_ipv6(rName, rInt),
			},
			{
				Config: testAccVPCSubnetDataSourceConfig_ipv6IPv6CIDRBlock(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block_association_id"),
				),
			},
		},
	})
}

func TestAccVPCSubnetDataSource_enableLniAtDeviceIndex(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dsResourceName := "data.aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetDataSourceConfig_enableLniAtDeviceIndex(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsResourceName, "enable_lni_at_device_index", acctest.Ct1),
				),
			},
		},
	})
}

func testAccVPCSubnetDataSourceConfig_basic(rName string, rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "172.%[2]d.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet" "by_id" {
  id = aws_subnet.test.id
}

data "aws_subnet" "by_cidr" {
  vpc_id     = aws_subnet.test.vpc_id
  cidr_block = aws_subnet.test.cidr_block
}

data "aws_subnet" "by_tag" {
  vpc_id = aws_subnet.test.vpc_id

  tags = {
    Name = aws_subnet.test.tags["Name"]
  }
}

data "aws_subnet" "by_vpc" {
  vpc_id = aws_subnet.test.vpc_id
}

data "aws_subnet" "by_filter" {
  filter {
    name   = "vpc-id"
    values = [aws_subnet.test.vpc_id]
  }
}

data "aws_subnet" "by_az_id" {
  vpc_id               = aws_subnet.test.vpc_id
  availability_zone_id = aws_subnet.test.availability_zone_id
}
`, rName, rInt)
}

func testAccVPCSubnetDataSourceConfig_enableLniAtDeviceIndex(rName string, deviceIndex int) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone          = data.aws_outposts_outpost.test.availability_zone
  cidr_block                 = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  enable_lni_at_device_index = %[2]d
  outpost_arn                = data.aws_outposts_outpost.test.arn
  vpc_id                     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet" "test" {
  id = aws_subnet.test.id
}
`, rName, deviceIndex)
}

func testAccVPCSubnetDataSourceConfig_ipv6(rName string, rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block                       = "172.%[2]d.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}
`, rName, rInt)
}

func testAccVPCSubnetDataSourceConfig_ipv6Filter(rName string, rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block                       = "172.%[2]d.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet" "by_ipv6_cidr" {
  filter {
    name   = "ipv6-cidr-block-association.ipv6-cidr-block"
    values = [aws_subnet.test.ipv6_cidr_block]
  }
}
`, rName, rInt)
}

func testAccVPCSubnetDataSourceConfig_ipv6IPv6CIDRBlock(rName string, rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block                       = "172.%[2]d.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet" "by_ipv6_cidr" {
  ipv6_cidr_block = aws_subnet.test.ipv6_cidr_block
}
`, rName, rInt)
}
