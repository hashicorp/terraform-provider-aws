package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCSubnetDataSource_basic(t *testing.T) {
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetDataSourceConfig(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds1ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds1ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds2ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds2ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds2ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds2ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds3ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds3ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds3ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds3ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds4ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds4ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds4ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds5ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds5ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds5ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "enable_dns64", snResourceName, "enable_dns64"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "enable_resource_name_dns_aaaa_record_on_launch", snResourceName, "enable_resource_name_dns_aaaa_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "enable_resource_name_dns_a_record_on_launch", snResourceName, "enable_resource_name_dns_a_record_on_launch"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "ipv6_native", snResourceName, "ipv6_native"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds6ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds6ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds6ResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "arn", snResourceName, "arn"),
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
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetIPv6DataSourceConfig(rName, rInt),
			},
			{
				Config: testAccSubnetIPv6WithDataSourceFilterDataSourceConfig(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block_association_id"),
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block"),
				),
			},
		},
	})
}

func TestAccVPCSubnetDataSource_ipv6ByIPv6CIDRBlock(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetIPv6DataSourceConfig(rName, rInt),
			},
			{
				Config: testAccSubnetIPv6WithDataSourceIpv6CIDRBlockDataSourceConfig(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block_association_id"),
				),
			},
		},
	})
}

func testAccSubnetDataSourceConfig(rName string, rInt int) string {
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

func testAccSubnetIPv6DataSourceConfig(rName string, rInt int) string {
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

func testAccSubnetIPv6WithDataSourceFilterDataSourceConfig(rName string, rInt int) string {
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

func testAccSubnetIPv6WithDataSourceIpv6CIDRBlockDataSourceConfig(rName string, rInt int) string {
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
