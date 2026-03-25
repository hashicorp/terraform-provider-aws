// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIPRestriction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_quicksight_ip_restriction.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRestrictionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRestrictionConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRestrictionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnabled), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ip_restriction_rule_map"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_id_restriction_rule_map"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_id_restriction_rule_map"), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAWSAccountID),
				ImportStateVerifyIdentifierAttribute: names.AttrAWSAccountID,
			},
		},
	})
}

func testAccIPRestriction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_quicksight_ip_restriction.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRestrictionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRestrictionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPRestrictionExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceIPRestriction, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIPRestriction_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_ip_restriction.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRestrictionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRestrictionConfig_permissions1(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRestrictionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnabled), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ip_restriction_rule_map"), knownvalue.MapExact(map[string]knownvalue.Check{
						"108.56.166.202/32": knownvalue.StringExact("Allow self"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_id_restriction_rule_map"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_id_restriction_rule_map"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAWSAccountID),
				ImportStateVerifyIdentifierAttribute: names.AttrAWSAccountID,
			},
			{
				Config: testAccIPRestrictionConfig_permissions2(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRestrictionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnabled), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ip_restriction_rule_map"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_id_restriction_rule_map"), knownvalue.MapSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_id_restriction_rule_map"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func testAccCheckIPRestrictionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_ip_restriction" {
				continue
			}

			_, err := tfquicksight.FindIPRestrictionByID(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight IP Restriction (%s) still exists", rs.Primary.Attributes[names.AttrAWSAccountID])
		}

		return nil
	}
}

func testAccCheckIPRestrictionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		_, err := tfquicksight.FindIPRestrictionByID(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID])

		return err
	}
}

const testAccIPRestrictionConfig_basic = `
resource "aws_quicksight_ip_restriction" "test" {
  enabled = true
}
`

func testAccIPRestrictionConfig_permissions1(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  count = 3

  cidr_block = "10.${count.index}.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test[1].id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.quicksight-website"
  vpc_endpoint_type = "Interface"
}

resource "aws_quicksight_ip_restriction" "test" {
  enabled = %[2]t

  ip_restriction_rule_map = {
    "108.56.166.202/32" = "Allow self"
  }

  vpc_id_restriction_rule_map = {
    (aws_vpc.test[0].id) = "Main VPC"
  }
}
`, rName, enabled)
}

func testAccIPRestrictionConfig_permissions2(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  count = 3

  cidr_block = "10.${count.index}.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test[1].id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.quicksight-website"
  vpc_endpoint_type = "Interface"
}

resource "aws_quicksight_ip_restriction" "test" {
  enabled = %[2]t

  vpc_id_restriction_rule_map = {
    (aws_vpc.test[0].id) = "Main VPC"
    (aws_vpc.test[2].id) = ""
  }

  vpc_endpoint_id_restriction_rule_map = {
    (aws_vpc_endpoint.test.id) = "EP"
  }
}
`, rName, enabled)
}
