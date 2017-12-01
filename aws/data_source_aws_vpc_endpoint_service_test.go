package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsVpcEndpointService_gateway(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsVpcEndpointServiceConfig_gateway,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "service_name", "com.amazonaws.us-west-2.s3"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "service_type", "Gateway"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "owner", "amazon"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "vpc_endpoint_policy_supported", "true"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "acceptance_required", "false"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "availability_zones.#", "3"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "availability_zones.2050015877", "us-west-2c"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "private_dns_name", ""),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "base_endpoint_dns_names.3003388505", "s3.us-west-2.amazonaws.com"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_interface(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsVpcEndpointServiceConfig_interface,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "service_name", "com.amazonaws.us-west-2.ec2"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "service_type", "Interface"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "owner", "amazon"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "acceptance_required", "false"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "availability_zones.2050015877", "us-west-2c"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "private_dns_name", "ec2.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "base_endpoint_dns_names.1880016359", "ec2.us-west-2.vpce.amazonaws.com"),
				),
			},
		},
	})
}

const testAccDataSourceAwsVpcEndpointServiceConfig_gateway = `
provider "aws" {
  region = "us-west-2"
}

data "aws_vpc_endpoint_service" "s3" {
  service = "s3"
}
`

const testAccDataSourceAwsVpcEndpointServiceConfig_interface = `
provider "aws" {
  region = "us-west-2"
}

data "aws_vpc_endpoint_service" "ec2" {
  service = "ec2"
}
`
