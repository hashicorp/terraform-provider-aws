// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfresiliencehubv2 "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCheckUserJourneyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "system_arn", "user_journey_id")
}

func TestAccResilienceHubV2UserJourney_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var uj awstypes.UserJourney
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_user_journey.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserJourneyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/UserJourney/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserJourneyExists(ctx, t, resourceName, &uj),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("system_arn"), checkSystemARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("user_journey_id"), knownvalue.NotNull()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/UserJourney/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccCheckUserJourneyImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "user_journey_id",
			},
		},
	})
}

func TestAccResilienceHubV2UserJourney_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var uj awstypes.UserJourney
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_user_journey.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserJourneyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/UserJourney/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserJourneyExists(ctx, t, resourceName, &uj),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfresiliencehubv2.ResourceUserJourney, resourceName),
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

func testAccCheckUserJourneyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_user_journey" {
				continue
			}

			_, err := tfresiliencehubv2.FindUserJourneyByTwoPartKey(ctx, conn, rs.Primary.Attributes["system_arn"], rs.Primary.Attributes["user_journey_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Resilience Hub V2 User Journey %s still exists", rs.Primary.Attributes["user_journey_id"])
		}

		return nil
	}
}

func testAccCheckUserJourneyExists(ctx context.Context, t *testing.T, n string, v *awstypes.UserJourney) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindUserJourneyByTwoPartKey(ctx, conn, rs.Primary.Attributes["system_arn"], rs.Primary.Attributes["user_journey_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
