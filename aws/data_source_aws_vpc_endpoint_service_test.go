package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsVpcEndpointService_gateway(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "service_name", "com.amazonaws.us-west-2.s3"),
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "4"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.2050015877", "us-west-2c"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.3830732582", "us-west-2d"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.3003388505", "s3.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_name", ""),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Gateway"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_interface(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.ec2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "service_name", "com.amazonaws.us-west-2.ec2"),
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "3"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.2050015877", "us-west-2c"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.1880016359", "ec2.us-west-2.vpce.amazonaws.com"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_name", "ec2.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_custom(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.foo"
	rName := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceCustomConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					testAccCheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
				),
			},
		},
	})
}

const testAccDataSourceAwsVpcEndpointServiceGatewayConfig = `
provider "aws" {
  region = "us-west-2"
}

data "aws_vpc_endpoint_service" "s3" {
  service = "s3"
}
`

const testAccDataSourceAwsVpcEndpointServiceInterfaceConfig = `
provider "aws" {
  region = "us-west-2"
}

data "aws_vpc_endpoint_service" "ec2" {
  service = "ec2"
}
`

func testAccDataSourceAwsVpcEndpointServiceCustomConfig(rName string) string {
	return fmt.Sprintf(
		`
resource "aws_vpc" "nlb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "nlb_test_1" {
  name = %[1]q

  subnets = [
    "${aws_subnet.nlb_test_1.id}",
    "${aws_subnet.nlb_test_2.id}",
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "nlb_test_1" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "nlb_test_2" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "foo" {
  acceptance_required = true

  network_load_balancer_arns = [
    "${aws_lb.nlb_test_1.id}",
  ]
}

data "aws_vpc_endpoint_service" "foo" {
  service_name = "${aws_vpc_endpoint_service.foo.service_name}"
}
`, rName)
}
