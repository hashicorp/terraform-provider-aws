package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsEip_classic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsEipClassicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsEipCheck("data.aws_eip.test_classic", "aws_eip.test_classic"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEip_vpc(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsEipVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsEipCheck("data.aws_eip.test_vpc_by_id", "aws_eip.test_vpc"),
					testAccDataSourceAwsEipCheck("data.aws_eip.test_vpc_by_public_ip", "aws_eip.test_vpc"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEip_filter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsEipFilterConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsEipCheck("data.aws_eip.by_filter", "aws_eip.test"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEipCheck(data_path string, resource_path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[data_path]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", data_path)
		}

		eipRs, ok := s.RootModule().Resources[resource_path]
		if !ok {
			return fmt.Errorf("can't find %s in state", resource_path)
		}

		attr := rs.Primary.Attributes

		if attr["id"] != eipRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"id is %s; want %s",
				attr["id"],
				eipRs.Primary.Attributes["id"],
			)
		}

		if attr["public_ip"] != eipRs.Primary.Attributes["public_ip"] {
			return fmt.Errorf(
				"public_ip is %s; want %s",
				attr["public_ip"],
				eipRs.Primary.Attributes["public_ip"],
			)
		}

		return nil
	}
}

const testAccDataSourceAwsEipClassicConfig = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_eip" "test_classic" {}

data "aws_eip" "test_classic" {
  public_ip = "${aws_eip.test_classic.public_ip}"
}

`

const testAccDataSourceAwsEipVPCConfig = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_eip" "test_vpc" {
	vpc = true
}

data "aws_eip" "test_vpc_by_id" {
  id = "${aws_eip.test_vpc.id}"
}

data "aws_eip" "test_vpc_by_public_ip" {
  public_ip = "${aws_eip.test_vpc.public_ip}"
}
`

const testAccDataSourceAwsEipFilterConfig = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_eip" "test" {
	vpc = true

	tags {
    	Name = "testeip"
  	}
}

data "aws_eip" "by_filter" {
  filter {
    name   = "tag:Name"
    values = ["${aws_eip.test.tags.Name}"]
  }
}
`
