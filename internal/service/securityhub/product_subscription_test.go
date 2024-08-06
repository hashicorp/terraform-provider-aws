// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccProductSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	accountResourceName := "aws_securityhub_account.example"
	resourceName := "aws_securityhub_product_subscription.example"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// We would like to use an AWS product subscription, but they are
				// all automatically subscribed when enabling Security Hub.
				// This configuration will enable Security Hub, then in a later PreConfig,
				// we will disable an AWS product subscription so we can test (re-)enabling it.
				Config: testAccProductSubscriptionConfig_accountOnly,
				Check:  testAccCheckAccountExists(ctx, accountResourceName),
			},
			{
				// AWS product subscriptions happen automatically when enabling Security Hub.
				// Here we attempt to remove one so we can attempt to (re-)enable it.
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)
					productSubscriptionARN := arn.ARN{
						Partition: acctest.Partition(),
						Service:   "securityhub",
						Region:    acctest.Region(),
						AccountID: acctest.AccountID(),
						Resource:  "product-subscription/aws/guardduty",
					}.String()
					input := &securityhub.DisableImportFindingsForProductInput{
						ProductSubscriptionArn: aws.String(productSubscriptionARN),
					}

					_, err := conn.DisableImportFindingsForProduct(ctx, input)

					if err != nil {
						t.Fatalf("error disabling Security Hub Product Subscription for GuardDuty: %s", err)
					}
				},
				Config: testAccProductSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductSubscriptionExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Check Destroy - but only target the specific resource (otherwise Security Hub
				// will be disabled and the destroy check will fail)
				Config: testAccProductSubscriptionConfig_accountOnly,
				Check:  testAccCheckProductSubscriptionDestroy(ctx),
			},
		},
	})
}

func testAccCheckProductSubscriptionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindProductSubscriptionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccCheckProductSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_product_subscription" {
				continue
			}

			_, err := tfsecurityhub.FindProductSubscriptionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
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

const testAccProductSubscriptionConfig_accountOnly = `
resource "aws_securityhub_account" "example" {}
`

var testAccProductSubscriptionConfig_basic = acctest.ConfigCompose(testAccProductSubscriptionConfig_accountOnly, `
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_securityhub_product_subscription" "example" {
  depends_on  = [aws_securityhub_account.example]
  product_arn = "arn:${data.aws_partition.current.partition}:securityhub:${data.aws_region.current.name}::product/aws/guardduty"
}
`)
