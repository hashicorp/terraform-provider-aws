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

func testAccCheckAssertionImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "service_arn", "assertion_id")
}

func TestAccResilienceHubV2Assertion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assertion awstypes.Assertion
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_assertion.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssertionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Assertion/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssertionExists(ctx, t, resourceName, &assertion),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("assertion_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_arn"), checkServiceARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("text"), knownvalue.StringExact("The service must recover within 5 minutes")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Assertion/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccCheckAssertionImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "assertion_id",
			},
		},
	})
}

func TestAccResilienceHubV2Assertion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assertion awstypes.Assertion
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_assertion.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssertionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Assertion/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssertionExists(ctx, t, resourceName, &assertion),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfresiliencehubv2.ResourceAssertion, resourceName),
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

func TestAccResilienceHubV2Assertion_text(t *testing.T) {
	ctx := acctest.Context(t)
	var assertion awstypes.Assertion
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_assertion.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssertionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Assertion/text/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"text":          config.StringVariable("Assertion #1"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssertionExists(ctx, t, resourceName, &assertion),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("text"), knownvalue.StringExact("Assertion #1")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Assertion/text/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"text":          config.StringVariable("Assertion #2"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssertionExists(ctx, t, resourceName, &assertion),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("text"), knownvalue.StringExact("Assertion #2")),
				},
			},
		},
	})
}

func testAccCheckAssertionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_assertion" {
				continue
			}

			_, err := tfresiliencehubv2.FindAssertionByTwoPartKey(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["assertion_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Resilience Hub V2 Assertion %s still exists", rs.Primary.Attributes["assertion_id"])
		}

		return nil
	}
}

func testAccCheckAssertionExists(ctx context.Context, t *testing.T, n string, v *awstypes.Assertion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindAssertionByTwoPartKey(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["assertion_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
