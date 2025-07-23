// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsNotificationHub_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var notificationhub awstypes.NotificationHubOverview
	resourceName := "aws_notifications_notification_hub.test"

	//lintignore:AWSAT003
	rRegion := "us-west-2"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationHubDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHubConfig_basic(rRegion),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationHubExists(ctx, resourceName, &notificationhub),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rRegion,
				ImportStateVerifyIdentifierAttribute: "notification_hub_region",
			},
		},
	})
}

func TestAccNotificationsNotificationHub_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var notificationhub awstypes.NotificationHubOverview
	//lintignore:AWSAT003
	rRegion := "eu-west-1"
	resourceName := "aws_notifications_notification_hub.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationHubDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHubConfig_basic(rRegion),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationHubExists(ctx, resourceName, &notificationhub),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnotifications.ResourceNotificationHub, resourceName),
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

func testAccCheckNotificationHubDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_notification_hub" {
				continue
			}

			_, err := tfnotifications.FindNotificationHubByRegion(ctx, conn, rs.Primary.Attributes["notification_hub_region"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("User Notifications Notification Hub %s still exists", rs.Primary.Attributes["notification_hub_region"])
		}

		return nil
	}
}

func testAccCheckNotificationHubExists(ctx context.Context, n string, v *awstypes.NotificationHubOverview) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		output, err := tfnotifications.FindNotificationHubByRegion(ctx, conn, rs.Primary.Attributes["notification_hub_region"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

	var input notifications.ListNotificationHubsInput

	_, err := conn.ListNotificationHubs(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccNotificationHubConfig_basic(rRegion string) string {
	return fmt.Sprintf(`
resource "aws_notifications_notification_hub" "test" {
  notification_hub_region = %[1]q
}
`, rRegion)
}
