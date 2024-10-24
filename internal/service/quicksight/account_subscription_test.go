// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accountsubscription awstypes.AccountInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSubscriptionDisableTerminationProtection(ctx, resourceName), // Workaround to remove termination protection
					testAccCheckAccountSubscriptionExists(ctx, resourceName, &accountsubscription),
					resource.TestCheckResourceAttr(resourceName, "account_name", rName),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  false,
				RefreshState: true,
			},
		},
	})
}

func testAccAccountSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var accountsubscription awstypes.AccountInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSubscriptionDisableTerminationProtection(ctx, resourceName), // Workaround to remove termination protection
					testAccCheckAccountSubscriptionExists(ctx, resourceName, &accountsubscription),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceAccountSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_account_subscription" {
				continue
			}

			_, err := tfquicksight.FindAccountSubscriptionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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
func testAccCheckAccountSubscriptionDisableTerminationProtection(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		defaultNs := "default"
		input := &quicksight.UpdateAccountSettingsInput{
			AwsAccountId:                 aws.String(rs.Primary.ID),
			DefaultNamespace:             aws.String(defaultNs),
			TerminationProtectionEnabled: false,
		}

		_, err := conn.UpdateAccountSettings(ctx, input)

		return err
	}
}

func testAccCheckAccountSubscriptionExists(ctx context.Context, n string, v *awstypes.AccountInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

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
