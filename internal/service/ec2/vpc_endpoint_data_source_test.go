package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCEndpointDataSource_gatewayBasic(t *testing.T) {
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_gatewayBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "service_name", resourceName, "service_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_byID(t *testing.T) {
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_byID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "service_name", resourceName, "service_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_byFilter(t *testing.T) {
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_byFilter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "service_name", resourceName, "service_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_byTags(t *testing.T) {
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_byTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "service_name", resourceName, "service_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_gatewayWithRouteTableAndTags(t *testing.T) {
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_gatewayRouteTableAndTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "service_name", resourceName, "service_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccVPCEndpointDataSource_interface(t *testing.T) {
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_interface(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "cidr_blocks.#", resourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_entry.#", resourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_options.#", resourceName, "dns_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_enabled", resourceName, "private_dns_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "requester_managed", resourceName, "requester_managed"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "service_name", resourceName, "service_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "vpc_id"),
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
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

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
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

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
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

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
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

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
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

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
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ec2"
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
