// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIPRestriction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_quicksight_ip_restriction.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRestrictionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRestrictionConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRestrictionExists(ctx, resourceName),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRestrictionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRestrictionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPRestrictionExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceIPRestriction, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIPRestrictionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_ip_restriction" {
				continue
			}

			_, err := tfquicksight.FindIPRestrictionByID(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID])

			if tfresource.NotFound(err) {
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

func testAccCheckIPRestrictionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		_, err := tfquicksight.FindIPRestrictionByID(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID])

		return err
	}
}

const testAccIPRestrictionConfig_basic = `
resource "aws_quicksight_ip_restriction" "test" {
  enabled = true
}
`
