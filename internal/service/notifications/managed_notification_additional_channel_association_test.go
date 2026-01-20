// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsManagedNotificationAdditionalChannelAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEmailAddress := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_notifications_managed_notification_additional_channel_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedNotificationAdditionalChannelAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedNotificationAdditionalChannelAssociationConfig_basic(rName, rEmailAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedNotificationAdditionalChannelAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "channel_arn",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "managed_notification_arn", "channel_arn"),
			},
		},
	})
}

func TestAccNotificationsManagedNotificationAdditionalChannelAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEmailAddress := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_notifications_managed_notification_additional_channel_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedNotificationAdditionalChannelAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedNotificationAdditionalChannelAssociationConfig_basic(rName, rEmailAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedNotificationAdditionalChannelAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnotifications.ResourceManagedNotificationAdditionalChannelAssociation, resourceName),
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

func testAccCheckManagedNotificationAdditionalChannelAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_managed_notification_additional_channel_association" {
				continue
			}

			_, err := tfnotifications.FindManagedNotificationAdditionalChannelAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["managed_notification_arn"], rs.Primary.Attributes["channel_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("User Notifications Managed Notification Additional Channel Association still exists")
		}

		return nil
	}
}

func testAccCheckManagedNotificationAdditionalChannelAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		_, err := tfnotifications.FindManagedNotificationAdditionalChannelAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["managed_notification_arn"], rs.Primary.Attributes["channel_arn"])

		return err
	}
}

func testAccManagedNotificationAdditionalChannelAssociationConfig_basic(rName, rEmailAddress string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_notificationscontacts_email_contact" "test" {
  name          = %[1]q
  email_address = %[2]q
}

resource "aws_notifications_managed_notification_additional_channel_association" "test" {
  channel_arn              = aws_notificationscontacts_email_contact.test.arn
  managed_notification_arn = "arn:${data.aws_partition.current.partition}:notifications::${data.aws_caller_identity.current.account_id}:managed-notification-configuration/category/AWS-Health/sub-category/Security"
}
`, rName, rEmailAddress)
}
