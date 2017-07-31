package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsNatGateway_unattached(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsNatGatewayUnattachedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_id", "id",
						"aws_nat_gateway.unattached", "id"),
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_tags", "id",
						"aws_nat_gateway.unattached", "id"),
					resource.TestCheckResourceAttrSet("data.aws_nat_gateway.test_by_id", "state"),
					resource.TestCheckResourceAttr("data.aws_nat_gateway.test_by_tags", "tags.%", "3"),
					resource.TestCheckNoResourceAttr("data.aws_nat_gateway.test_by_id", "attached_vpc_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNatGateway_attached(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsNatGatewayAttachedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_attached_vpc_id", "id",
						"aws_nat_gateway.attached", "id"),
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_attached_vpc_id", "attached_vpc_id",
						"aws_vpc.foo", "id"),
					resource.TestMatchResourceAttr("data.aws_nat_gateway.test_by_attached_vpc_id", "state", regexp.MustCompile("(?i)available")),
				),
			},
		},
	})
}

func testAccDataSourceAwsNatGatewayUnattachedConfig(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_nat_gateway" "unattached" {
    tags {
		Name = "terraform-testacc-nat-gateway-data-source-unattached-%d"
      	ABC  = "testacc-%d"
		XYZ  = "testacc-%d"
    }
}

data "aws_nat_gateway" "test_by_id" {
	id = "${aws_nat_gateway.unattached.id}"
}

data "aws_nat_gateway" "test_by_tags" {
	tags = "${aws_nat_gateway.unattached.tags}"
}
`, rInt, rInt+1, rInt-1)
}

func testAccDataSourceAwsNatGatewayAttachedConfig(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

  	tags {
    	Name = "terraform-testacc-nat-gateway-data-source-foo-%d"
  	}
}

resource "aws_nat_gateway" "attached" {
    tags {
		Name = "terraform-testacc-nat-gateway-data-source-attached-%d"
    }
}

resource "aws_nat_gateway_attachment" "nat_attachment" {
  vpc_id = "${aws_vpc.foo.id}"
  nat_gateway_id = "${aws_nat_gateway.attached.id}"
}

data "aws_nat_gateway" "test_by_attached_vpc_id" {
	attached_vpc_id = "${aws_nat_gateway_attachment.nat_attachment.vpc_id}"
}
`, rInt, rInt)
}
