package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsVpcIDs_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcIDsConfig(),
				Check: resource.ComposeTestCheckFunc(
					//datasource aws_vpc_ids will list all the vpcs in a region, so it also includes the Default vpc.
					//https://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/default-vpc.html
					resource.TestCheckResourceAttr("data.aws_vpc_ids.all", "ids.#", "3"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcIDs_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcIDsConfig_tags(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_ids.selected", "ids.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceAwsVpcIDsConfig() string {
	return fmt.Sprintf(`
	provider "aws" {
  		region = "us-west-2"
	}

	resource "aws_vpc" "test-vpc" {
		count = 2
  		cidr_block = "10.0.0.0/24"

	}

	data "aws_vpc_ids" "all" {
		filter {
			name = "vpcs"
			values = ["${aws_vpc.test-vpc.*.id}"]
		}
	}
	`)
}

func testAccDataSourceAwsVpcIDsConfig_tags() string {
	return fmt.Sprintf(`
	provider "aws" {
  		region = "us-west-2"
	}

	resource "aws_vpc" "test-vpc" {
		count = 2
  		cidr_block = "10.0.0.0/24"

  		tags {
  			Name = "testacc-vpc"
  			Service = "test"
  		}
	}

	data "aws_vpc_ids" "selected" {
		filter {
			name = "vpcs"
			values = ["${aws_vpc.test-vpc.*.id}"]
		}

		tags {
			Service = "test"
		}
	}
	`)
}
