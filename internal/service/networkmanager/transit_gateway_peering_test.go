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

func TestAccNetworkManagerTransitGatewayPeering_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPeering
	resourceName := "aws_networkmanager_transit_gateway_peering.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPeeringDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPeeringExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`peering/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, tgwResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_arn", tgwResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_peering_attachment_id"),
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

func TestAccNetworkManagerTransitGatewayPeering_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPeering
	resourceName := "aws_networkmanager_transit_gateway_peering.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPeeringDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkmanager.ResourceTransitGatewayPeering(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTransitGatewayPeeringExists(ctx context.Context, t *testing.T, n string, v *awstypes.TransitGatewayPeering) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Transit Gateway Peering ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		output, err := tfnetworkmanager.FindTransitGatewayPeeringByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayPeeringDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_transit_gateway_peering" {
				continue
			}

			_, err := tfnetworkmanager.FindTransitGatewayPeeringByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Transit Gateway Peering %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayPeeringConfig_base(rName string) string {
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
}
`, rName)
}

func testAccTransitGatewayPeeringConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringConfig_base(rName), `
resource "aws_networkmanager_transit_gateway_peering" "test" {
  core_network_id     = aws_networkmanager_core_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn

  depends_on = [
    aws_ec2_transit_gateway_policy_table.test,
    aws_networkmanager_core_network_policy_attachment.test,
  ]
}
`)
}
