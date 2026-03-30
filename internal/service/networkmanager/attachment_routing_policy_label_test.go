// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerAttachmentRoutingPolicyLabel_basic(t *testing.T) {
	resourceName := "aws_networkmanager_attachment_routing_policy_label.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	label := "testlabel"

	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentRoutingPolicyLabelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentRoutingPolicyLabelConfig_basic(rName, label),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttachmentRoutingPolicyLabelExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttrSet(resourceName, "attachment_id"),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "attachment_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "core_network_id", "attachment_id"),
			},
		},
	})
}

func TestAccNetworkManagerAttachmentRoutingPolicyLabel_disappears(t *testing.T) {
	resourceName := "aws_networkmanager_attachment_routing_policy_label.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	label := "testlabel"

	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentRoutingPolicyLabelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentRoutingPolicyLabelConfig_basic(rName, label),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentRoutingPolicyLabelExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkmanager.ResourceAttachmentRoutingPolicyLabel, resourceName),
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

func TestAccNetworkManagerAttachmentRoutingPolicyLabel_update(t *testing.T) {
	resourceName := "aws_networkmanager_attachment_routing_policy_label.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	label1 := "labelv1"
	label2 := "labelv2"

	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentRoutingPolicyLabelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentRoutingPolicyLabelConfig_basic(rName, label1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttachmentRoutingPolicyLabelExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label1),
				),
			},
			{
				Config: testAccAttachmentRoutingPolicyLabelConfig_basic(rName, label2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttachmentRoutingPolicyLabelExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label2),
				),
			},
		},
	})
}

func testAccCheckAttachmentRoutingPolicyLabelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_attachment_routing_policy_label" {
				continue
			}

			_, err := tfnetworkmanager.FindAttachmentRoutingPolicyAssociationLabelByTwoPartKey(ctx, conn, rs.Primary.Attributes["core_network_id"], rs.Primary.Attributes["attachment_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Attachment Routing Policy Label %s still exists", rs.Primary.Attributes["attachment_id"])
		}

		return nil
	}
}

func testAccCheckAttachmentRoutingPolicyLabelExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		_, err := tfnetworkmanager.FindAttachmentRoutingPolicyAssociationLabelByTwoPartKey(ctx, conn, rs.Primary.Attributes["core_network_id"], rs.Primary.Attributes["attachment_id"])

		return err
	}
}

func testAccAttachmentRoutingPolicyLabelConfig_basic(rName, label string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = data.aws_region.current.region
    }
  }

  segments {
    name                          = "segment"
    require_attachment_acceptance = false
  }

  attachment_policies {
    rule_number     = 100
    condition_logic = "or"

    conditions {
      type = "tag-exists"
      key  = "segment"
    }

    action {
      association_method = "tag"
      tag_value_of_key   = "segment"
    }
  }

  routing_policies {
    routing_policy_name      = %[2]q
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
      associate_routing_policies = [%[2]q]
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    segment = "segment"
  }

  lifecycle {
    ignore_changes = [routing_policy_label]
  }
}

resource "aws_networkmanager_attachment_routing_policy_label" "test" {
  core_network_id      = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  attachment_id        = aws_networkmanager_vpc_attachment.test.id
  routing_policy_label = %[2]q
}
`, rName, label))
}
