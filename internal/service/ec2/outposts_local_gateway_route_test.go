package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2OutpostsLocalGatewayRoute_basic(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 255)
	destinationCidrBlock := fmt.Sprintf("172.16.%d.0/24", rInt)
	localGatewayRouteTableDataSourceName := "data.aws_ec2_local_gateway_route_table.test"
	localGatewayVirtualInterfaceGroupDataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"
	resourceName := "aws_ec2_local_gateway_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocalGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteDestinationCIDRBlockConfig(destinationCidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidrBlock),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_route_table_id", localGatewayRouteTableDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_virtual_interface_group_id", localGatewayVirtualInterfaceGroupDataSourceName, "id"),
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

func TestAccEC2OutpostsLocalGatewayRoute_disappears(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 255)
	destinationCidrBlock := fmt.Sprintf("172.16.%d.0/24", rInt)
	resourceName := "aws_ec2_local_gateway_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocalGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteDestinationCIDRBlockConfig(destinationCidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceLocalGatewayRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLocalGatewayRouteExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Local Gateway Route ID is set")
		}

		localGatewayRouteTableID, destination, err := tfec2.DecodeLocalGatewayRouteID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		route, err := tfec2.GetLocalGatewayRoute(conn, localGatewayRouteTableID, destination)

		if err != nil {
			return err
		}

		if route == nil {
			return fmt.Errorf("EC2 Local Gateway Route (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocalGatewayRouteDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_local_gateway_route" {
			continue
		}

		localGatewayRouteTableID, destination, err := tfec2.DecodeLocalGatewayRouteID(rs.Primary.ID)

		if err != nil {
			return err
		}

		route, err := tfec2.GetLocalGatewayRoute(conn, localGatewayRouteTableID, destination)

		if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if route == nil {
			continue
		}

		return fmt.Errorf("EC2 Local Gateway Route (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccLocalGatewayRouteDestinationCIDRBlockConfig(destinationCidrBlock string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

resource "aws_ec2_local_gateway_route" "test" {
  destination_cidr_block                   = %[1]q
  local_gateway_route_table_id             = data.aws_ec2_local_gateway_route_table.test.id
  local_gateway_virtual_interface_group_id = data.aws_ec2_local_gateway_virtual_interface_group.test.id
}
`, destinationCidrBlock)
}
