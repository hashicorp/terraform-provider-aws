// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpointDataSource_gatewayBasic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_gatewayBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrServiceName, resourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(datasourceName, "service_region", resourceName, "service_region"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_byID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_byID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrServiceName, resourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_byFilter(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_byFilter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrServiceName, resourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_byTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_byTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrServiceName, resourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_gatewayWithRouteTableAndTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_gatewayRouteTableAndTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrServiceName, resourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_interface(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_interface(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrServiceName, resourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_resourceConfigurationPrivateDNS(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_resourceConfigurationDNSOptions(rName, "SPECIFIED_DOMAINS_ONLY", []string{"example1.com", "example2.com"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.0.private_dns_preference", resourceName, "dns_options.0.private_dns_preference"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.0.private_dns_specified_domains", resourceName, "dns_options.0.private_dns_specified_domains"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
				),
			},
		},
	})
}

func testAccVPCEndpointDataSourceConfig_gatewayBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint" "test" {
  vpc_id         = aws_vpc.test.id
  service_name   = aws_vpc_endpoint.test.service_name
  service_region = data.aws_region.current.region
  state          = "available"
}
`, rName)
}

func testAccVPCEndpointDataSourceConfig_byID(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint" "test" {
  id = aws_vpc_endpoint.test.id
}
`, rName)
}

func testAccVPCEndpointDataSourceConfig_byFilter(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint" "test" {
  filter {
    name   = "vpc-endpoint-id"
    values = [aws_vpc_endpoint.test.id]
  }
}
`, rName)
}

func testAccVPCEndpointDataSourceConfig_byTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2"
    Key3 = "Value3"
  }
}

data "aws_vpc_endpoint" "test" {
  vpc_id = aws_vpc_endpoint.test.vpc_id

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2"
    Key3 = "Value3"
  }
}
`, rName)
}

func testAccVPCEndpointDataSourceConfig_gatewayRouteTableAndTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  route_table_ids = [
    aws_route_table.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = aws_vpc_endpoint.test.service_name
  state        = "available"
}
`, rName)
}

func testAccVPCEndpointDataSourceConfig_interface(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = aws_vpc.test.cidr_block
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  vpc_endpoint_type   = "Interface"
  service_name        = "com.amazonaws.${data.aws_region.current.region}.ec2"
  private_dns_enabled = false

  subnet_ids = [
    aws_subnet.test.id,
  ]

  security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = aws_vpc_endpoint.test.service_name
  state        = "available"
}
`, rName))
}
func testAccVPCEndpointDataSourceConfig_resourceConfigurationDNSOptions(rName, privateDNSPreference string, privateDNSSpecifiedDomains []string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_vpc" "endpoint" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "endpoint" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.endpoint.id
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  resource_configuration_arn = aws_vpclattice_resource_configuration.test.arn
  subnet_ids                 = [aws_subnet.endpoint.id]
  vpc_endpoint_type          = "Resource"
  vpc_id                     = aws_vpc.endpoint.id

  private_dns_enabled = true

  dns_options {
    private_dns_preference        = %[2]q
    private_dns_specified_domains = ["%[3]s"]
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.endpoint.id
  vpc_endpoint_type = aws_vpc_endpoint.test.vpc_endpoint_type
  state             = "available"
}
`, rName, privateDNSPreference, strings.Join(privateDNSSpecifiedDomains, `", "`)))
}
