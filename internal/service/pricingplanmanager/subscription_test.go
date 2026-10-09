// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pricingplanmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/pricingplanmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pricingplanmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/acctest/config"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpricingplanmanager "github.com/hashicorp/terraform-provider-aws/internal/service/pricingplanmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkSubscriptionARN = tfknownvalue.GlobalARNRegexp("pricingplanmanager", regexache.MustCompile(`subscription:.+`))
)

func TestAccPricingPlanManagerSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v pricingplanmanager.GetSubscriptionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pricingplanmanager_subscription.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingPlanManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_mode"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkSubscriptionARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_family"), knownvalue.StringExact("CloudFront")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("FREE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_arns"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scheduled_change"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.StatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatusReason), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("usage_level"), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccPricingPlanManagerSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v pricingplanmanager.GetSubscriptionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pricingplanmanager_subscription.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingPlanManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpricingplanmanager.ResourceSubscription, resourceName),
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

func TestAccPricingPlanManagerSubscription_resourceARNs(t *testing.T) {
	ctx := acctest.Context(t)
	var v pricingplanmanager.GetSubscriptionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_pricingplanmanager_subscription.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingPlanManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/resourceARNs.2/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"domain_name":   config.StringVariable(domainName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_arns"), knownvalue.SetSizeExact(2)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/resourceARNs.3/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"domain_name":   config.StringVariable(domainName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_arns"), knownvalue.SetSizeExact(3)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/resourceARNs.2/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"domain_name":   config.StringVariable(domainName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_arns"), knownvalue.SetSizeExact(2)),
				},
			},
		},
	})
}

func TestAccPricingPlanManagerSubscription_planTier(t *testing.T) {
	// Tests that end with an ACTIVE paid subscription cannot delete their
	// CloudFront distribution until the scheduled cancellation takes effect
	// at the end of the billing cycle, so they always leave the distribution
	// and web ACL behind for later cleanup. Opt in explicitly.
	acctest.SkipIfEnvVarNotSet(t, "PRICINGPLANMANAGER_PAID_PLAN_TESTS")

	ctx := acctest.Context(t)
	var v pricingplanmanager.GetSubscriptionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pricingplanmanager_subscription.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingPlanManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/planTier/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"plan_tier":     config.StringVariable("FREE"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("FREE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.StatusActive)),
				},
			},
			// Tier upgrades take effect immediately.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/planTier/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"plan_tier":     config.StringVariable("PRO"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("PRO")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.StatusActive)),
				},
			},
			// Downgrades are scheduled by AWS for the end of the current billing
			// period: the API keeps reporting the old tier with a DOWNGRADE
			// scheduled change, while plan_tier tracks the desired tier.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/planTier/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"plan_tier":     config.StringVariable("FREE"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("FREE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scheduled_change"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"change_type": tfknownvalue.StringExact(awstypes.ScheduledChangeTypeDowngrade),
							"plan_tier":   knownvalue.StringExact("FREE"),
						}),
					})),
				},
			},
			// Raising the tier back before the downgrade takes effect reverts
			// the pending scheduled change.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/planTier/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"plan_tier":     config.StringVariable("PRO"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("PRO")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scheduled_change"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccPricingPlanManagerSubscription_approvalModeManual(t *testing.T) {
	ctx := acctest.Context(t)
	var v pricingplanmanager.GetSubscriptionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pricingplanmanager_subscription.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingPlanManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Paid-tier subscriptions created with MANUAL approval mode park in
			// PENDING_APPROVAL and do not start billing until approved.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/approvalMode/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"plan_tier":     config.StringVariable("PRO"),
					"approval_mode": tfconfig.StringVariable(awstypes.ApprovalModeManual),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_mode"), tfknownvalue.StringExact(awstypes.ApprovalModeManual)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("PRO")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.StatusPendingApproval)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Subscription/approvalMode/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"plan_tier":     config.StringVariable("PRO"),
					"approval_mode": tfconfig.StringVariable(awstypes.ApprovalModeManual),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				// approval_mode is a create-time-only argument, not returned by the API.
				ImportStateVerifyIgnore: []string{"approval_mode"},
			},
		},
	})
}

func testAccCheckSubscriptionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PricingPlanManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pricingplanmanager_subscription" {
				continue
			}

			_, err := tfpricingplanmanager.FindSubscriptionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Pricing Plan Manager Subscription %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckSubscriptionExists(ctx context.Context, t *testing.T, n string, v *pricingplanmanager.GetSubscriptionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).PricingPlanManagerClient(ctx)

		output, err := tfpricingplanmanager.FindSubscriptionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PricingPlanManagerClient(ctx)

	input := pricingplanmanager.ListSubscriptionsInput{}
	_, err := conn.ListSubscriptions(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
