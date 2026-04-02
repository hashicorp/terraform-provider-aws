// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerTransitGatewayRouteTableAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTableAttachment
	resourceName := "aws_networkmanager_transit_gateway_route_table_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "0"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "TRANSIT_GATEWAY_ROUTE_TABLE"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttr(resourceName, "segment_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	var v awstypes.TransitGatewayRouteTableAttachment
	resourceName := "aws_networkmanager_transit_gateway_route_table_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkmanager.ResourceTransitGatewayRouteTableAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerTransitGatewayRouteTableAttachment_routingPolicyLabel(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTableAttachment
	resourceName := "aws_networkmanager_transit_gateway_route_table_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	label := "testlabel"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_routingPolicyLabel(rName, label),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label),
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

func TestAccNetworkManagerTransitGatewayRouteTableAttachment_routingPolicyLabelUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTableAttachment
	resourceName := "aws_networkmanager_transit_gateway_route_table_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	label1 := "testlabel1"
	label2 := "testlabel2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_routingPolicyLabel(rName, label1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label1),
				),
			},
			{
				Config: testAccTransitGatewayRouteTableAttachmentConfig_routingPolicyLabel(rName, label2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label2),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayRouteTableAttachmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.TransitGatewayRouteTableAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		output, err := tfnetworkmanager.FindTransitGatewayRouteTableAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_transit_gateway_route_table_attachment" {
				continue
			}

			_, err := tfnetworkmanager.FindTransitGatewayRouteTableAttachmentByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

  depends_on = [aws_ec2_transit_gateway_policy_table.test, aws_networkmanager_core_network_policy_attachment.test]
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

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_transit_gateway_route_table_attachment.test.id
  attachment_type = aws_networkmanager_transit_gateway_route_table_attachment.test.attachment_type
}
`)
}

func testAccTransitGatewayRouteTableAttachmentConfig_baseWithRoutingPolicy(rName, label string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_policy_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    # Don't overlap with default TGW ASN: 64512.
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = data.aws_region.current.region
    }
  }

  segments {
    name = "test"
  }

  routing_policies {
    routing_policy_name      = "policy1"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        match_conditions {
          type  = "prefix-in-cidr"
          value = "10.0.0.0/8"
        }

        action {
          type = "allow"
        }
      }
    }
  }

  attachment_routing_policy_rules {
    rule_number = 1

    conditions {
      type  = "routing-policy-label"
      value = %[2]q
    }

    action {
      associate_routing_policies = ["policy1"]
    }
  }
}

resource "aws_networkmanager_transit_gateway_peering" "test" {
  core_network_id     = aws_networkmanager_core_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_transit_gateway_policy_table.test, aws_networkmanager_core_network_policy_attachment.test]
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
`, rName, label)
}

func testAccTransitGatewayRouteTableAttachmentConfig_routingPolicyLabel(rName, label string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAttachmentConfig_baseWithRoutingPolicy(rName, label), fmt.Sprintf(`
resource "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.test.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.test.arn
  routing_policy_label            = %[1]q

  depends_on = [aws_ec2_transit_gateway_policy_table_association.test]
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_transit_gateway_route_table_attachment.test.id
  attachment_type = aws_networkmanager_transit_gateway_route_table_attachment.test.attachment_type
}
`, label))
}
