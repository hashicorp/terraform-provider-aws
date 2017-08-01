package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsNatGateway(t *testing.T) {
	// This is used as a portion of CIDR network addresses.
	rInt := acctest.RandIntRange(4, 254)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsNatGatewayConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_id", "id",
						"aws_nat_gateway", "id"),
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_tags", "id",
						"aws_nat_gateway", "id"),
					resource.TestCheckResourceAttrSet("data.aws_nat_gateway.test_by_id", "state"),
					resource.TestCheckResourceAttr("data.aws_nat_gateway.test_by_tags", "tags.%", "3"),
					resource.TestCheckNoResourceAttr("data.aws_nat_gateway.test_by_id", "attached_vpc_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNatGatewayConfig(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "test" {
  cidr_block = "172.%d.0.0/16"
  tags {
    Name = "terraform-testacc-nat-gateway-data-source"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "172.%d.123.0/24"
  availability_zone = "us-west-2a"

  tags {
    Name = "terraform-testacc-nat-gateway-data-source-%d"
  }
}

# EIPs are not taggable
resource "aws_eip" "test" {}

resource "aws_nat_gateway" "test" {
  subnet_id     = "${aws_subnet.test.id}"
  allocation_id = "${aws_eip.test.id}"
    tags {
		Name = "terraform-testacc-nat-gateway-data-source-%d"
		ABC  = "testacc-%d"
		XYZ  = "testacc-%d"
    }
}

data "aws_nat_gateway" "test_by_id" {
	id = "${aws_nat_gateway.test.id}"
}

data "aws_nat_gateway" "test_by_tags" {
	tags = "${aws_nat_gateway.test.tags}"
}
`, rInt, rInt+1, rInt-1)
}
