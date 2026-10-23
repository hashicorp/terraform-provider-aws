// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfresiliencehub "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v resiliencehub.DescribeAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "assessment_schedule", string(awstypes.AppAssessmentScheduleTypeDisabled)),
					resource.TestCheckResourceAttr(resourceName, "drift_status", string(awstypes.AppDriftStatusTypeNotChecked)),
					resource.TestCheckNoResourceAttr(resourceName, "resiliency_policy_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubApp_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 resiliencehub.DescribeAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app.test"

	const (
		initial = "initial description"
		updated = "updated description"
	)

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_description(rName, initial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, initial),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccAppConfig_description(rName, updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updated),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubApp_assessmentSchedule(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 resiliencehub.DescribeAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app.test"

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_assessmentSchedule(rName, string(awstypes.AppAssessmentScheduleTypeDaily)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "assessment_schedule", string(awstypes.AppAssessmentScheduleTypeDaily)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccAppConfig_assessmentSchedule(rName, string(awstypes.AppAssessmentScheduleTypeDisabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "assessment_schedule", string(awstypes.AppAssessmentScheduleTypeDisabled)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubApp_resiliencyPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 resiliencehub.DescribeAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app.test"
	policyResourceName := "aws_resiliencehub_resiliency_policy.test"
	policy2ResourceName := "aws_resiliencehub_resiliency_policy.test2"

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_resiliencyPolicy(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "resiliency_policy_arn", policyResourceName, names.AttrARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccAppConfig_resiliencyPolicy(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "resiliency_policy_arn", policy2ResourceName, names.AttrARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccCheckAppDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehub_app" {
				continue
			}

			_, err := tfresiliencehub.FindAppByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if errs.IsA[*sdkretry.NotFoundError](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("ResilienceHub App %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckAppExists(ctx context.Context, t *testing.T, n string, v *resiliencehub.DescribeAppOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubClient(ctx)

		output, err := tfresiliencehub.FindAppByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = resiliencehub.DescribeAppOutput{App: output}

		return nil
	}
}

func testAccAppConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_app" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAppConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_app" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccAppConfig_assessmentSchedule(rName, schedule string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_app" "test" {
  name                = %[1]q
  assessment_schedule = %[2]q
}
`, rName, schedule)
}

func testAccAppConfig_resiliencyPolicy(rName, policyRef string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  name = %[1]q

  tier = "NotApplicable"

  policy {
    az {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    hardware {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    software {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
  }
}

resource "aws_resiliencehub_resiliency_policy" "test2" {
  name = "%[1]s-2"

  tier = "NotApplicable"

  policy {
    az {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    hardware {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    software {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
  }
}

resource "aws_resiliencehub_app" "test" {
  name                  = %[1]q
  resiliency_policy_arn = aws_resiliencehub_resiliency_policy.%[2]s.arn
}
`, rName, policyRef)
}
