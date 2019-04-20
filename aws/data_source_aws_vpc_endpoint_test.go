package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsVpcEndpoint_gatewayBasic(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint.s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_gatewayBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttrSet(datasourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttr(datasourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(datasourceName, "requester_managed", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_byId(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint.s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_byId,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttrSet(datasourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttr(datasourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(datasourceName, "requester_managed", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_gatewayWithRouteTable(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint.s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_gatewayWithRouteTable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttrSet(datasourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttr(datasourceName, "route_table_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(datasourceName, "requester_managed", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_interface(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint.ec2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_interface,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_type", "Interface"),
					resource.TestCheckNoResourceAttr(datasourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(datasourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "network_interface_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(datasourceName, "requester_managed", "false"),
				),
			},
		},
	})
}

const testAccDataSourceAwsVpcEndpointConfig_gatewayBasic = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-data-source-gw-basic"
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "${aws_vpc_endpoint.s3.service_name}"
  state = "available"
}
`

const testAccDataSourceAwsVpcEndpointConfig_byId = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-data-source-by-id"
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_vpc_endpoint" "s3" {
  id = "${aws_vpc_endpoint.s3.id}"
}
`

const testAccDataSourceAwsVpcEndpointConfig_gatewayWithRouteTable = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-data-source-with-route-table"
  }
}

resource "aws_route_table" "rt" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
  route_table_ids = ["${aws_route_table.rt.id}"]
}

data "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "${aws_vpc_endpoint.s3.service_name}"
  state = "available"
}
`

const testAccDataSourceAwsVpcEndpointConfig_interface = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-data-source-interface"
  }
}

resource "aws_subnet" "sn" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${aws_vpc.foo.cidr_block}"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-vpc-endpoint-data-source-interface"
  }
}

resource "aws_security_group" "sg" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc_endpoint" "ec2" {
  vpc_id = "${aws_vpc.foo.id}"
  vpc_endpoint_type = "Interface"
  service_name = "com.amazonaws.us-west-2.ec2"
  subnet_ids = ["${aws_subnet.sn.id}"]
  security_group_ids = ["${aws_security_group.sg.id}"]
  private_dns_enabled = false
}

data "aws_vpc_endpoint" "ec2" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "${aws_vpc_endpoint.ec2.service_name}"
  state = "available"
}
`
