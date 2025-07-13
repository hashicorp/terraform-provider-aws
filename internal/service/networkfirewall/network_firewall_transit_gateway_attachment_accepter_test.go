// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	ec2awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallNetworkFirewallTransitGatewayAttachmentAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_network_firewall_transit_gateway_attachment_accepter.test"
	var v ec2awstypes.TransitGatewayAttachment
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckNetworkFirewallTransitGatewayAttachmentAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallTransitGatewayAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkFirewallTransitGatewayAttachmentAccepterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrTransitGatewayAttachmentID),
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

func TestAccNetworkFirewallNetworkFirewallTransitGatewayAttachmentAccepter_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var v ec2awstypes.TransitGatewayAttachment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_network_firewall_transit_gateway_attachment_accepter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckNetworkFirewallTransitGatewayAttachmentAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallTransitGatewayAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkFirewallTransitGatewayAttachmentAccepterExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkfirewall.ResourceNetworkFirewallTransitGatewayAttachmentAccepter, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckNetworkFirewallTransitGatewayAttachmentAccepterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_network_firewall_transit_gateway_attachment_accepter" {
				continue
			}

			// TIP: ==== FINDERS ====
			// The find function should be exported. Since it won't be used outside of the package, it can be exported
			// in the `exports_test.go` file.
			_, err := tfec2.FindTransitGatewayAttachmentByID(ctx, conn, rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.NetworkFirewall, create.ErrActionCheckingDestroyed, tfnetworkfirewall.ResNameNetworkFirewallTransitGatewayAttachmentAccepter, rs.Primary.ID, err)
			}

			return create.Error(names.NetworkFirewall, create.ErrActionCheckingDestroyed, tfnetworkfirewall.ResNameNetworkFirewallTransitGatewayAttachmentAccepter, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckNetworkFirewallTransitGatewayAttachmentAccepterExists(ctx context.Context, name string, v *ec2awstypes.TransitGatewayAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameNetworkFirewallTransitGatewayAttachmentAccepter, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameNetworkFirewallTransitGatewayAttachmentAccepter, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindTransitGatewayAttachmentByID(ctx, conn, rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])
		if err != nil {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameNetworkFirewallTransitGatewayAttachmentAccepter, rs.Primary.ID, err)
		}

		v = resp

		return nil
	}
}

func testAccNetworkFirewallTransitGatewayAttachmentAccepterConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_ec2_transit_gateway.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

# attachment creator.
data "aws_caller_identity" "creator" {
  provider = "awsalternate"
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.creator.account_id
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_networkfirewall_firewall_policy" "test" {
  provider = "awsalternate"
  
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "test" {
  provider = "awsalternate"
  
  name                                = %[1]q
  firewall_policy_arn                 = aws_networkfirewall_firewall_policy.test.arn
  transit_gateway_id                  = aws_ec2_transit_gateway.test.id

  dynamic "availability_zone_mapping" {
	for_each = data.aws_availability_zones.available.zone_ids
	content {	
	  availability_zone_id = availability_zone_mapping.value
	}
  }
  
  depends_on = [
	aws_ram_resource_association.test,
	aws_ram_principal_association.test,
  ]
}

resource "aws_networkfirewall_network_firewall_transit_gateway_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_networkfirewall_firewall.test.firewall_status[0].transit_gateway_attachment_sync_state[0].attachment_id
}
`, rName))
}
