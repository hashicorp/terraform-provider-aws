// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccProductSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_product_subscription.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				// We would like to use an AWS product subscription, but they are
				// all automatically subscribed when enabling Security Hub.
				// This configuration will enable Security Hub, then in a later PreConfig,
				// we will disable an AWS product subscription so we can test (re-)enabling it.
				Config: testAccProductSubscriptionConfig_accountOnly,
			},
			{
				// AWS product subscriptions happen automatically when enabling Security Hub.
				// Here we attempt to remove one so we can attempt to (re-)enable it.
				PreConfig: testAccDeleteProductSubscriptionFunc(ctx, t, "product-subscription/aws/guardduty"),
				Config:    testAccProductSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductSubscriptionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("securityhub", regexache.MustCompile(`product-subscription/.+`))),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Check Destroy - but only target the specific resource (otherwise Security Hub
				// will be disabled and the destroy check will fail).
				Config: testAccProductSubscriptionConfig_accountOnly,
				Check:  testAccCheckProductSubscriptionDestroy(ctx, t),
			},
		},
	})
}

func testAccProductSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_product_subscription.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				// We would like to use an AWS product subscription, but they are
				// all automatically subscribed when enabling Security Hub.
				// This configuration will enable Security Hub, then in a later PreConfig,
				// we will disable an AWS product subscription so we can test (re-)enabling it.
				Config: testAccProductSubscriptionConfig_accountOnly,
			},
			{
				// AWS product subscriptions happen automatically when enabling Security Hub.
				// Here we attempt to remove one so we can attempt to (re-)enable it.
				PreConfig: testAccDeleteProductSubscriptionFunc(ctx, t, "product-subscription/aws/guardduty"),
				Config:    testAccProductSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductSubscriptionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsecurityhub.ResourceProductSubscription(), resourceName),
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

func testAccCheckProductSubscriptionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindProductSubscriptionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccCheckProductSubscriptionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_product_subscription" {
				continue
			}

			_, err := tfsecurityhub.FindProductSubscriptionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Product Subscription (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDeleteProductSubscriptionFunc(ctx context.Context, t *testing.T, name string) func() {
	return testAccDeleteProductSubscriptionRegionFunc(ctx, t, name, acctest.Region())
}

func testAccDeleteProductSubscriptionRegionFunc(ctx context.Context, t *testing.T, name, region string) func() {
	return func() {
		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)
		productSubscriptionARN := arn.ARN{
			Partition: acctest.Partition(),
			Service:   "securityhub",
			Region:    region,
			AccountID: acctest.AccountID(ctx),
			Resource:  name,
		}.String()
		input := securityhub.DisableImportFindingsForProductInput{
			ProductSubscriptionArn: aws.String(productSubscriptionARN),
		}

		_, err := conn.DisableImportFindingsForProduct(ctx, &input, func(o *securityhub.Options) {
			o.Region = region
		})

		if err != nil {
			t.Fatalf("error disabling Security Hub Product Subscription for GuardDuty: %s", err)
		}
	}
}

const testAccProductSubscriptionConfig_accountOnly = `
resource "aws_securityhub_account" "test" {}
`

func testAccProductSubscriptionConfig_accountOnlyRegion(region string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {
  region = %[1]q
}
`, region)
}

var testAccProductSubscriptionConfig_basic = acctest.ConfigCompose(testAccProductSubscriptionConfig_accountOnly, `
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_securityhub_product_subscription" "test" {
  depends_on  = [aws_securityhub_account.test]
  product_arn = "arn:${data.aws_partition.current.partition}:securityhub:${data.aws_region.current.region}::product/aws/guardduty"
}
`)
