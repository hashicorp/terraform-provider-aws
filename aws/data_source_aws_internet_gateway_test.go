package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsInternetGateway_typical(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsInternetGatewayCheck("data.aws_internet_gateway.by_id"),
					testAccDataSourceAwsInternetGatewayCheck("data.aws_internet_gateway.by_filter"),
					testAccDataSourceAwsInternetGatewayCheck("data.aws_internet_gateway.by_tags"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSourceAwsInternetGatewayCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		igwRs, ok := s.RootModule().Resources["aws_internet_gateway.test"]
		if !ok {
			return fmt.Errorf("can't find aws_internet_gateway.test in state")
		}
		vpcRs, ok := s.RootModule().Resources["aws_vpc.test"]
		if !ok {
			return fmt.Errorf("can't find aws_vpc.test in state")
		}

		attr := rs.Primary.Attributes

		if attr["internet_gateway_id"] != igwRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"internet_gateway_id is %s; want %s",
				attr["internet_gateway_id"],
				igwRs.Primary.Attributes["id"],
			)
		}

		if attr["attachments.0.vpc_id"] != vpcRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"vpc_id is %s; want %s",
				attr["attachments.0.vpc_id"],
				vpcRs.Primary.Attributes["id"],
			)
		}
		return nil
	}
}

const testAccDataSourceAwsInternetGatewayConfig = `
provider "aws" {
  region = "eu-central-1"
}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags {
    Name = "terraform-testacc-data-source-igw-vpc"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags {
    Name = "terraform-testacc-data-source-igw"
  }
}

data "aws_internet_gateway" "by_id" {
  internet_gateway_id = "${aws_internet_gateway.test.id}"
}

data "aws_internet_gateway" "by_tags" {
  tags {
    Name = "${aws_internet_gateway.test.tags["Name"]}"
  }
}

data "aws_internet_gateway" "by_filter" {
  filter {
    name = "attachment.vpc-id"
    values = ["${aws_vpc.test.id}"]
  }

  depends_on = ["aws_internet_gateway.test"]
}
`
