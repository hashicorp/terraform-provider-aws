package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDxGateway_importBasic(t *testing.T) {
	resourceName := "aws_dx_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDxGatewayConfig(acctest.RandString(5), acctest.RandIntRange(64512, 65534)),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGateway_importComplex(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 3 {
			return fmt.Errorf("Got %d resources, expected 3. State: %#v", len(s), s)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDxGatewayConfig_complexImport(acctest.RandString(5), acctest.RandIntRange(64512, 65534)),
			},

			resource.TestStep{
				ResourceName:      "aws_dx_gateway.test",
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDxGatewayConfig_complexImport(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name = "tf-dxg-%s"
  amazon_side_asn = "%d"
}

resource "aws_vpc" "test1" {
  cidr_block = "10.255.255.16/28"
}

resource "aws_vpc" "test2" {
  cidr_block = "10.255.255.32/28"
}

resource "aws_vpn_gateway" "test1" {
  vpc_id = "${aws_vpc.test1.id}"
}

resource "aws_vpn_gateway" "test2" {
  vpc_id = "${aws_vpc.test2.id}"
}

resource "aws_dx_gateway_association" "test1" {
  dx_gateway_id = "${aws_dx_gateway.test.id}"
  virtual_gateway_id = "${aws_vpn_gateway.test1.id}"
}

resource "aws_dx_gateway_association" "test2" {
  dx_gateway_id = "${aws_dx_gateway.test.id}"
  virtual_gateway_id = "${aws_vpn_gateway.test2.id}"
}
`, rName, rBgpAsn)
}
