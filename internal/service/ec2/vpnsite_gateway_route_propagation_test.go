package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSiteVPNGatewayRoutePropagation_basic(t *testing.T) {
	resourceName := "aws_vpn_gateway_route_propagation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPNGatewayRoutePropagationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNGatewayRoutePropagationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNGatewayRoutePropagationExists(resourceName),
				),
			},
		},
	})
}

func TestAccSiteVPNGatewayRoutePropagation_disappears(t *testing.T) {
	resourceName := "aws_vpn_gateway_route_propagation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPNGatewayRoutePropagationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNGatewayRoutePropagationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNGatewayRoutePropagationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPNGatewayRoutePropagation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPNGatewayRoutePropagationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route Table VPN Gateway route propagation ID is set")
		}

		routeTableID, gatewayID, err := tfec2.VPNGatewayRoutePropagationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		err = tfec2.FindVPNGatewayRoutePropagationExists(conn, routeTableID, gatewayID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckVPNGatewayRoutePropagationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_gateway_route_propagation" {
			continue
		}

		routeTableID, gatewayID, err := tfec2.VPNGatewayRoutePropagationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		err = tfec2.FindVPNGatewayRoutePropagationExists(conn, routeTableID, gatewayID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Route Table (%s) VPN Gateway (%s) route propagation still exists", routeTableID, gatewayID)
	}

	return nil
}

func testAccVPNGatewayRoutePropagationBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_route_propagation" "test" {
  vpn_gateway_id = aws_vpn_gateway.test.id
  route_table_id = aws_route_table.test.id
}
`, rName)
}
