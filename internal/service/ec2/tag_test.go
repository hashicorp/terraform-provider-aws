package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2Tag_basic(t *testing.T) {
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2Tag_disappears(t *testing.T) {
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2Tag_value(t *testing.T) {
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTagConfig_basic(rBgpAsn, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAccTagConfig_basic(rBgpAsn int, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_vpn_connection.test.transit_gateway_attachment_id
  key         = %[2]q
  value       = %[3]q
}
`, rBgpAsn, key, value)
}
