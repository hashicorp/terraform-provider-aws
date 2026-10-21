// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamUsageReportSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_usage_report_subscription.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageReportSubscriptionDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageReportSubscriptionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageReportSubscriptionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "s3_bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "DAILY"),
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

func TestAccAppStreamUsageReportSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_usage_report_subscription.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageReportSubscriptionDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageReportSubscriptionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageReportSubscriptionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappstream.ResourceUsageReportSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUsageReportSubscriptionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)
		_, err := tfappstream.FindUsageReportSubscription(ctx, conn)

		return err
	}
}

func testAccCheckUsageReportSubscriptionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_usage_report_subscription" {
				continue
			}

			_, err := tfappstream.FindUsageReportSubscription(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream Usage Report Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUsageReportSubscriptionConfig_basic() string {
	return `
resource "aws_appstream_usage_report_subscription" "test" {}
`
}
