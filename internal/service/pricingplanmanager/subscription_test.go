// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pricingplanmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pricingplanmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pricingplanmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpricingplanmanager "github.com/hashicorp/terraform-provider-aws/internal/service/pricingplanmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
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
				Config: testAccSubscriptionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_family"), knownvalue.StringExact("CloudFront")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("FREE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_arns"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact(string(awstypes.StatusActive))),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: knownvalue.NotNull(),
					}),
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
				Config: testAccSubscriptionConfig_resourceARNsBase(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_arns"), knownvalue.SetSizeExact(2)),
				},
			},
			{
				Config: testAccSubscriptionConfig_resourceARNs(rName, domainName),
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
			// The hosted zone stays in the configuration: it is protected from
			// deletion while associated with the plan, so it can only be removed
			// after the disassociation has been applied.
			{
				Config: testAccSubscriptionConfig_resourceARNsDisassociated(rName, domainName),
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
				Config: testAccSubscriptionConfig_tier(rName, "FREE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("FREE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact(string(awstypes.StatusActive))),
				},
			},
			// Tier upgrades take effect immediately.
			{
				Config: testAccSubscriptionConfig_tier(rName, "PRO"),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact(string(awstypes.StatusActive))),
				},
			},
			// Downgrades are scheduled by AWS for the end of the current billing
			// period: the API keeps reporting the old tier with a DOWNGRADE
			// scheduled change, while plan_tier tracks the desired tier.
			{
				Config: testAccSubscriptionConfig_tier(rName, "FREE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("FREE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scheduled_change"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"change_type": knownvalue.StringExact(string(awstypes.ScheduledChangeTypeDowngrade)),
							"plan_tier":   knownvalue.StringExact("FREE"),
						}),
					})),
				},
			},
			// Raising the tier back before the downgrade takes effect reverts
			// the pending scheduled change.
			{
				Config: testAccSubscriptionConfig_tier(rName, "PRO"),
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
				Config: testAccSubscriptionConfig_approvalMode(rName, "PRO", "MANUAL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("approval_mode"), knownvalue.StringExact(string(awstypes.ApprovalModeManual))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("plan_tier"), knownvalue.StringExact("PRO")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact(string(awstypes.StatusPendingApproval))),
				},
			},
			{
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

			output, err := tfpricingplanmanager.FindSubscriptionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Cancellation of an active subscription takes effect at the end of
			// the current billing period; a pending CANCELLATION scheduled change
			// is the terminal state visible via the API after destroy.
			if sc := output.Subscription.ScheduledChange; sc != nil && sc.ChangeType == awstypes.ScheduledChangeTypeCancellation {
				continue
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

func testAccSubscriptionConfig_base(rName string) string {
	// Web ACLs for CloudFront distributions must be created in us-east-1.
	// lintignore:AWSAT003
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled    = true
  comment    = %[1]q
  web_acl_id = aws_wafv2_web_acl.test.arn

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    # Managed-CachingOptimized. Flat-rate plan eligibility requires modern
    # cache settings (a cache policy) rather than legacy forwarded_values.
    cache_policy_id = "658327ea-f89d-4fab-a63d-7e88639e58f6"
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

resource "aws_wafv2_web_acl" "test" {
  region = "us-east-1"

  name  = %[1]q
  scope = "CLOUDFRONT"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccSubscriptionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionConfig_base(rName), `
resource "aws_pricingplanmanager_subscription" "test" {
  plan_family = "CloudFront"
  plan_tier   = "FREE"

  resource_arns = [
    aws_cloudfront_distribution.test.arn,
    aws_wafv2_web_acl.test.arn,
  ]
}
`)
}

// CloudFront KeyValueStores are incompatible with the FREE tier, but Route 53
// hosted zones can be associated with it, so the resource_arns tests use a
// hosted zone to stay off paid tiers.
func testAccSubscriptionConfig_resourceARNsBase(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionConfig_base(rName), `
resource "aws_pricingplanmanager_subscription" "test" {
  plan_family = "CloudFront"
  plan_tier   = "FREE"

  resource_arns = [
    aws_cloudfront_distribution.test.arn,
    aws_wafv2_web_acl.test.arn,
  ]
}
`)
}

func testAccSubscriptionConfig_resourceARNs(rName, domainName string) string {
	return acctest.ConfigCompose(testAccSubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[2]q
}

resource "aws_pricingplanmanager_subscription" "test" {
  plan_family = "CloudFront"
  plan_tier   = "FREE"

  resource_arns = [
    aws_cloudfront_distribution.test.arn,
    aws_wafv2_web_acl.test.arn,
    aws_route53_zone.test.arn,
  ]
}
`, rName, domainName))
}

func testAccSubscriptionConfig_resourceARNsDisassociated(rName, domainName string) string {
	return acctest.ConfigCompose(testAccSubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[2]q
}

resource "aws_pricingplanmanager_subscription" "test" {
  plan_family = "CloudFront"
  plan_tier   = "FREE"

  resource_arns = [
    aws_cloudfront_distribution.test.arn,
    aws_wafv2_web_acl.test.arn,
  ]
}
`, rName, domainName))
}

func testAccSubscriptionConfig_tier(rName, tier string) string {
	return acctest.ConfigCompose(testAccSubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_pricingplanmanager_subscription" "test" {
  plan_family = "CloudFront"
  plan_tier   = %[1]q

  resource_arns = [
    aws_cloudfront_distribution.test.arn,
    aws_wafv2_web_acl.test.arn,
  ]
}
`, tier))
}

func testAccSubscriptionConfig_approvalMode(rName, tier, approvalMode string) string {
	return acctest.ConfigCompose(testAccSubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_pricingplanmanager_subscription" "test" {
  approval_mode = %[2]q
  plan_family   = "CloudFront"
  plan_tier     = %[1]q

  resource_arns = [
    aws_cloudfront_distribution.test.arn,
    aws_wafv2_web_acl.test.arn,
  ]
}
`, tier, approvalMode))
}
