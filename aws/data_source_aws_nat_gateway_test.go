package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsNatGateway(t *testing.T) {
	rInt := acctest.RandInt()

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

resource "aws_nat_gateway" "test" {
    tags {
		Name = "terraform-testacc-nat-gateway-data-source-%d"
      	ABC  = "testacc-%d"
		XYZ  = "testacc-%d"
    }
}

data "aws_nat_gateway" "test_by_id" {
	id = "${aws_nat_gateway.id}"
}

data "aws_nat_gateway" "test_by_tags" {
	tags = "${aws_nat_gateway.tags}"
}
`, rInt, rInt+1, rInt-1)
}
