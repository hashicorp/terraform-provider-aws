package networkmanager_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNetworkManagerTransitGatewayRouteTableAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkmanager.TransitGatewayRouteTableAttachment
	resourceName := "aws_networkmanager_transit_gateway_route_table_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "networkmanager", regexp.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "0"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "TRANSIT_GATEWAY_ROUTE_TABLE"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "segment_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccNetworkManagerTransitGatewayRouteTableAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkmanager.TransitGatewayRouteTableAttachment
	resourceName := "aws_networkmanager_transit_gateway_route_table_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceTransitGatewayRouteTableAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerTransitGatewayRouteTableAttachment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkmanager.TransitGatewayRouteTableAttachment
	resourceName := "aws_networkmanager_transit_gateway_route_table_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayRouteTableAttachmentExists(ctx context.Context, n string, v *networkmanager.TransitGatewayRouteTableAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Transit Gateway Route Table Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn()

		output, err := tfnetworkmanager.FindTransitGatewayRouteTableAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_transit_gateway_route_table_attachment" {
				continue
			}

			_, err := tfnetworkmanager.FindTransitGatewayRouteTableAttachmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Transit Gateway Route Table Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayRouteTableAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmanager_transit_gateway_peering" "test" {
  core_network_id     = aws_networkmanager_core_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_transit_gateway_policy_table.test]
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_policy_table_association" "test" {
  transit_gateway_attachment_id   = aws_networkmanager_transit_gateway_peering.test.transit_gateway_peering_attachment_id
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
}
`, rName))
}

func testAccTransitGatewayRouteTableAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAttachmentConfig_base(rName), `
resource "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.test.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.test.arn

  depends_on = [aws_ec2_transit_gateway_policy_table_association.test]
}
`)
}

func testAccTransitGatewayRouteTableAttachmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.test.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.test.arn

  tags = {
    %[1]q = %[2]q
  }

  depends_on = [aws_ec2_transit_gateway_policy_table_association.test]
}
`, tagKey1, tagValue1))
}

func testAccTransitGatewayRouteTableAttachmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.test.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  depends_on = [aws_ec2_transit_gateway_policy_table_association.test]
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
