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

func TestAccNetworkManagerPrefixListAssociation_basic(t *testing.T) {
	resourceName := "aws_networkmanager_prefix_list_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPrefixListAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixListAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrefixListAssociationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccNetworkManagerPrefixListAssociation_disappears(t *testing.T) {
	resourceName := "aws_networkmanager_prefix_list_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPrefixListAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixListAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrefixListAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkmanager.ResourcePrefixListAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckPrefixListAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_prefix_list_association" {
				continue
			}

			_, err := tfnetworkmanager.FindPrefixListAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["core_network_id"], rs.Primary.Attributes["prefix_list_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Prefix List Association %s,%s still exists", rs.Primary.Attributes["core_network_id"], rs.Primary.Attributes["prefix_list_arn"])
		}

		return nil
	}
}

func testAccCheckPrefixListAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		_, err := tfnetworkmanager.FindPrefixListAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["core_network_id"], rs.Primary.Attributes["prefix_list_arn"])

		return err
	}
}

func testAccPrefixListAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "core_network_id", "prefix_list_arn")
}

func testAccPrefixListAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_prefix_list_association" "test" {
  core_network_id   = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  prefix_list_arn   = aws_ec2_managed_prefix_list.test.arn
  prefix_list_alias = "testprefixlist"
}

resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  entry {
    cidr        = "10.0.0.0/8"
    description = "Test CIDR"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

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
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}
`, rName)
}
