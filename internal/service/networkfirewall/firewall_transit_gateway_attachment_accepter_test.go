// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallFirewallTransitGatewayAttachmentAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_transit_gateway_attachment_accepter.test"
	var v ec2types.TransitGatewayAttachment
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckFirewallTransitGatewayAttachmentAccepterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallTransitGatewayAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFirewallTransitGatewayAttachmentAccepterExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrTransitGatewayAttachmentID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrTransitGatewayAttachmentID),
				ImportStateVerifyIdentifierAttribute: names.AttrTransitGatewayAttachmentID,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallTransitGatewayAttachmentAccepter_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var v ec2types.TransitGatewayAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_transit_gateway_attachment_accepter.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckFirewallTransitGatewayAttachmentAccepterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallTransitGatewayAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFirewallTransitGatewayAttachmentAccepterExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkfirewall.ResourceFirewallTransitGatewayAttachmentAccepter, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccCheckFirewallTransitGatewayAttachmentAccepterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_firewall_transit_gateway_attachment_accepter" {
				continue
			}

			output, err := tfec2.FindTransitGatewayAttachmentByID(ctx, conn, rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if output.State == ec2types.TransitGatewayAttachmentStateDeleted {
				continue
			}

			return fmt.Errorf("NetworkFirewall Firewall Transit Gateway Attachment %s still exists", rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])
		}

		return nil
	}
}

func testAccCheckFirewallTransitGatewayAttachmentAccepterExists(ctx context.Context, t *testing.T, n string, v *ec2types.TransitGatewayAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayAttachmentByID(ctx, conn, rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])

		if err != nil {
			return err
		}

		v = output

		return nil
	}
}

func testAccFirewallTransitGatewayAttachmentAccepterConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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

  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  transit_gateway_id  = aws_ec2_transit_gateway.test.id

  availability_zone_mapping {
    availability_zone_id = data.aws_availability_zones.available.zone_ids[0]
  }

  depends_on = [
    aws_ram_resource_association.test,
    aws_ram_principal_association.test,
  ]
}

resource "aws_networkfirewall_firewall_transit_gateway_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_networkfirewall_firewall.test.firewall_status[0].transit_gateway_attachment_sync_states[0].attachment_id
}
`, rName))
}
