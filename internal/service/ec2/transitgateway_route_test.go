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

func testAccTransitGatewayRoute_basic(t *testing.T) {
	var v ec2.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteDestinationCIDRBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func testAccTransitGatewayRoute_basic_ipv6(t *testing.T) {
	var v ec2.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test_ipv6"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteDestinationCIDRBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "2001:db8::/56"),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func testAccTransitGatewayRoute_blackhole(t *testing.T) {
	var v ec2.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test_blackhole"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteDestinationCIDRBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func testAccTransitGatewayRoute_disappears(t *testing.T) {
	var v ec2.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteDestinationCIDRBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGatewayRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayRoute_disappears_TransitGatewayAttachment(t *testing.T) {
	var v ec2.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteDestinationCIDRBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGatewayVPCAttachment(), transitGatewayVpcAttachmentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTransitGatewayRouteExists(n string, v *ec2.TransitGatewayRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Route ID is set")
		}

		transitGatewayRouteTableID, destination, err := tfec2.TransitGatewayRouteParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindTransitGatewayRoute(conn, transitGatewayRouteTableID, destination)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayRouteDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route" {
			continue
		}

		transitGatewayRouteTableID, destination, err := tfec2.TransitGatewayRouteParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindTransitGatewayRoute(conn, transitGatewayRouteTableID, destination)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Transit Gateway Route %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccTransitGatewayRouteDestinationCIDRBlockConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route" "test" {
  destination_cidr_block         = "0.0.0.0/0"
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}

resource "aws_ec2_transit_gateway_route" "test_ipv6" {
  destination_cidr_block         = "2001:db8::/56"
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}

resource "aws_ec2_transit_gateway_route" "test_blackhole" {
  destination_cidr_block         = "10.1.0.0/16"
  blackhole                      = true
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}
`, rName))
}
