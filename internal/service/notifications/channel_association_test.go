// Copyright (c) HashiCorp, Inc.
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
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsChannelAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEmailAddress := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_notifications_channel_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelAssociationConfig_basic(rName, rEmailAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    testAccChannelAssociationImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccNotificationsChannelAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEmailAddress := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_notifications_channel_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelAssociationConfig_basic(rName, rEmailAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnotifications.ResourceChannelAssociation, resourceName),
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

func testAccCheckChannelAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_channel_association" {
				continue
			}

			err := tfnotifications.FindChannelAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["notification_configuration_arn"], rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("User Notifications Channel Association still exists")
		}

		return nil
	}
}

func testAccCheckChannelAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		err := tfnotifications.FindChannelAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["notification_configuration_arn"], rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccChannelAssociationImportStateIDFunc(n string) func(*terraform.State) (string, error) {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes["notification_configuration_arn"] + intflex.ResourceIdSeparator + rs.Primary.Attributes[names.AttrARN], nil
	}
}

func testAccChannelAssociationConfig_basic(rName, rEmailAddress string) string {
	return fmt.Sprintf(`
resource "aws_notifications_notification_configuration" "test" {
  name        = %[1]q
  description = "example"
}

resource "aws_notificationscontacts_email_contact" "test" {
  name          = %[1]q
  email_address = %[2]q
}

resource "aws_notifications_channel_association" "test" {
  arn                            = aws_notificationscontacts_email_contact.test.arn
  notification_configuration_arn = aws_notifications_notification_configuration.test.arn
}
`, rName, rEmailAddress)
}
