// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccStandardsSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var standardsSubscription types.StandardsSubscription
	resourceName := "aws_securityhub_standards_subscription.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStandardsSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStandardsSubscriptionExists(ctx, t, resourceName, &standardsSubscription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStandardsSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var standardsSubscription types.StandardsSubscription
	resourceName := "aws_securityhub_standards_subscription.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStandardsSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStandardsSubscriptionExists(ctx, t, resourceName, &standardsSubscription),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsecurityhub.ResourceStandardsSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStandardsSubscriptionExists(ctx context.Context, t *testing.T, n string, v *types.StandardsSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		output, err := tfsecurityhub.FindStandardsSubscriptionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStandardsSubscriptionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_standards_subscription" {
				continue
			}

			output, err := tfsecurityhub.FindStandardsSubscriptionByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// INCOMPLETE subscription status => deleted.
			if output.StandardsStatus == types.StandardsStatusIncomplete {
				continue
			}

			return fmt.Errorf("Security Hub Standards Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccStandardsSubscriptionConfig_basic = `
resource "aws_securityhub_account" "test" {}

data "aws_partition" "current" {}

resource "aws_securityhub_standards_subscription" "test" {
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]

  timeouts {
    create = "3m"
    delete = "3m"
  }
}
`
