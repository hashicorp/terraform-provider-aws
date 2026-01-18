// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsManagedNotificationAccountContactAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_notifications_managed_notification_account_contact_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedNotificationAccountContactAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedNotificationAccountContactAssociationConfig_basic(string(awstypes.AccountContactTypeAccountPrimary)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedNotificationAccountContactAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "contact_identifier",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "managed_notification_configuration_arn", "contact_identifier"),
			},
		},
	})
}

func TestAccNotificationsManagedNotificationAccountContactAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_notifications_managed_notification_account_contact_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedNotificationAccountContactAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedNotificationAccountContactAssociationConfig_basic(string(awstypes.AccountContactTypeAccountAlternateSecurity)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedNotificationAccountContactAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnotifications.ResourceManagedNotificationAccountContactAssociation, resourceName),
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

func TestAccNotificationsManagedNotificationAccountContactAssociation_alternateBilling(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_notifications_managed_notification_account_contact_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedNotificationAccountContactAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedNotificationAccountContactAssociationConfig_basic(string(awstypes.AccountContactTypeAccountAlternateBilling)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedNotificationAccountContactAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "contact_identifier",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "managed_notification_configuration_arn", "contact_identifier"),
			},
		},
	})
}

func TestAccNotificationsManagedNotificationAccountContactAssociation_alternateOperations(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_notifications_managed_notification_account_contact_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedNotificationAccountContactAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedNotificationAccountContactAssociationConfig_basic(string(awstypes.AccountContactTypeAccountAlternateOperations)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedNotificationAccountContactAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "contact_identifier",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "managed_notification_configuration_arn", "contact_identifier"),
			},
		},
	})
}

func testAccCheckManagedNotificationAccountContactAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_managed_notification_account_contact_association" {
				continue
			}

			_, err := tfnotifications.FindManagedNotificationAccountContactAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["managed_notification_configuration_arn"], rs.Primary.Attributes["contact_identifier"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("User Notifications Managed Notification Account Contact Association still exists")
		}

		return nil
	}
}

func testAccCheckManagedNotificationAccountContactAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		_, err := tfnotifications.FindManagedNotificationAccountContactAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["managed_notification_configuration_arn"], rs.Primary.Attributes["contact_identifier"])

		return err
	}
}

func testAccManagedNotificationAccountContactAssociationConfig_basic(contactIdentifier string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_notifications_managed_notification_account_contact_association" "test" {
  contact_identifier                     = %[1]q
  managed_notification_configuration_arn = "arn:${data.aws_partition.current.partition}:notifications::${data.aws_caller_identity.current.account_id}:managed-notification-configuration/category/AWS-Health/sub-category/Security"
}
`, contactIdentifier)
}
