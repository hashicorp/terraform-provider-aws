// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accountsubscription awstypes.AccountInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_subscription.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSubscriptionDisableTerminationProtection(ctx, t, resourceName), // Workaround to remove termination protection
					testAccCheckAccountSubscriptionExists(ctx, t, resourceName, &accountsubscription),
					resource.TestCheckResourceAttr(resourceName, "account_name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_method"}, // Not returned from the DescribeAccountSubscription API
			},
		},
	})
}

func testAccAccountSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var accountsubscription awstypes.AccountInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_subscription.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSubscriptionDisableTerminationProtection(ctx, t, resourceName), // Workaround to remove termination protection
					testAccCheckAccountSubscriptionExists(ctx, t, resourceName, &accountsubscription),
					acctest.CheckSDKResourceDisappears(ctx, t, tfquicksight.ResourceAccountSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAccountSubscriptionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_account_subscription" {
				continue
			}

			_, err := tfquicksight.FindAccountSubscriptionByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Account Subscription (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

// Account subscription cannot be automatically deleted after creation. Termination protection
// must first be disabled which requires a separate API call.
func testAccCheckAccountSubscriptionDisableTerminationProtection(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		input := &quicksight.UpdateAccountSettingsInput{
			AwsAccountId:                 aws.String(rs.Primary.ID),
			DefaultNamespace:             aws.String(tfquicksight.DefaultNamespace),
			TerminationProtectionEnabled: false,
		}

		_, err := conn.UpdateAccountSettings(ctx, input)

		return err
	}
}

func testAccCheckAccountSubscriptionExists(ctx context.Context, t *testing.T, n string, v *awstypes.AccountInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindAccountSubscriptionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccountSubscriptionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_account_subscription" "test" {
  account_name          = %[1]q
  authentication_method = "IAM_AND_QUICKSIGHT"
  edition               = "ENTERPRISE"
  notification_email    = %[2]q
}
`, rName, acctest.DefaultEmailAddress)
}
