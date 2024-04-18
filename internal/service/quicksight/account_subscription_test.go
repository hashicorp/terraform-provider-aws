// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accountsubscription quicksight.AccountInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
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
	var accountsubscription quicksight.AccountInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_account_subscription" {
				continue
			}

			output, err := tfquicksight.FindAccountSubscriptionByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			if output != nil && aws.StringValue(output.AccountSubscriptionStatus) != "UNSUBSCRIBED" {
				return fmt.Errorf("QuickSight Account Subscription (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

// Account subscription cannot be automatically deleted after creation. Termination protection
// must first be disabled which requires a separate API call.
func testAccCheckAccountSubscriptionDisableTerminationProtection(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameAccountSubscription, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameAccountSubscription, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		defaultNs := "default"
		_, err := conn.UpdateAccountSettingsWithContext(ctx, &quicksight.UpdateAccountSettingsInput{
			AwsAccountId:                 aws.String(rs.Primary.ID),
			DefaultNamespace:             aws.String(defaultNs),
			TerminationProtectionEnabled: aws.Bool(false),
		})

		if err != nil {
			return create.Error(names.QuickSight, "setting termination protection to false", tfquicksight.ResNameAccountSubscription, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckAccountSubscriptionExists(ctx context.Context, name string, accountsubscription *quicksight.AccountInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameAccountSubscription, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameAccountSubscription, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		resp, err := tfquicksight.FindAccountSubscriptionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameAccountSubscription, rs.Primary.ID, err)
		}

		*accountsubscription = *resp

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
