package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsNetworkAcls_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkAclsConfig(rName),
			},
			{
				Config: testAccDataSourceAwsNetworkAclsConfigWithDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_acls.all", "ids.#", "3"),
					resource.TestCheckResourceAttr("data.aws_network_acls.with_tags", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_network_acls.with_filter", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkAclsConfigWithDataSource(rName string) string {
	return fmt.Sprintf(`
	resource "aws_vpc" "test-vpc" {
  		cidr_block = "10.0.0.0/16"
	}

	resource "aws_network_acl" "acl1" {
  		vpc_id = "${aws_vpc.test-vpc.id}"

  		tags {
    		Name = "testacc-acl-%s"
  		}
	}

	resource "aws_subnet" "test" {
  		vpc_id            = "${aws_vpc.test-vpc.id}"
  		cidr_block        = "10.0.0.0/24"
  		availability_zone = "us-west-2a"

  		tags {
    		Name = "tf-acc-subnet"
  		}
	}

	resource "aws_network_acl" "acl2" {
  		vpc_id = "${aws_vpc.test-vpc.id}"
  		subnet_ids = ["${aws_subnet.test.id}"]

  		tags {
    		Name = "testacc-acl-%s"
  		}
	}

	data "aws_network_acls" "all" {
		vpc_id = "${aws_vpc.test-vpc.id}"
	}

	data "aws_network_acls" "with_tags" {
		vpc_id = "${aws_vpc.test-vpc.id}"

		tags {
			Name = "testacc-acl-%s"
		}
	}

	data "aws_network_acls" "with_filter" {
		vpc_id = "${aws_vpc.test-vpc.id}"

		filter {
			name = "association.subnet-id"
    		values = ["${aws_subnet.test.id}"]
		}
	}
	`, rName, rName, rName)
}

func testAccDataSourceAwsNetworkAclsConfig(rName string) string {
	return fmt.Sprintf(`
	resource "aws_vpc" "test-vpc" {
  		cidr_block = "10.0.0.0/16"
	}

	resource "aws_network_acl" "acl1" {
  		vpc_id = "${aws_vpc.test-vpc.id}"

  		tags {
    		Name = "testacc-acl-%s"
  		}
	}

	resource "aws_subnet" "test" {
  		vpc_id            = "${aws_vpc.test-vpc.id}"
  		cidr_block        = "10.0.0.0/24"
  		availability_zone = "us-west-2a"

  		tags {
    		Name = "tf-acc-subnet"
  		}
	}

	resource "aws_network_acl" "acl2" {
  		vpc_id = "${aws_vpc.test-vpc.id}"
  		subnet_ids = ["${aws_subnet.test.id}"]

  		tags {
    		Name = "testacc-acl-%s"
  		}
	}
	`, rName, rName)
}
