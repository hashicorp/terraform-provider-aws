// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	"github.com/aws/aws-sdk-go-v2/service/notifications/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsNotificationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var notificationconfiguration notifications.GetNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_notification_configuration.testAccObjectImportStateIdFromARNsFunc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfigurationConfig_basic(rName, rDescription, string(types.AggregationDurationLong), acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &notificationconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescription),
					resource.TestCheckResourceAttr(resourceName, "aggregation_duration", string(types.AggregationDurationLong)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "notifications", regexache.MustCompile(`configuration/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccObjectImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccNotificationsNotificationConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var notificationconfiguration notifications.GetNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_notification_configuration.testAccObjectImportStateIdFromARNsFunc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfigurationConfig_minimal(rName, rDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &notificationconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnotifications.ResourceNotificationConfiguration, resourceName),
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

func TestAccNotificationsNotificationConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)

	var v1, v2 notifications.GetNotificationConfigurationOutput
	rNameOne := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameTwo := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescriptionOne := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescriptionTwo := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_notifications_notification_configuration.testAccObjectImportStateIdFromARNsFunc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfigurationConfig_minimal(rNameOne, rDescriptionOne),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameOne),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescriptionOne),
					resource.TestCheckResourceAttr(resourceName, "aggregation_duration", string(types.AggregationDurationNone)),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "notifications", regexache.MustCompile(`configuration/.+$`)),
				),
			},
			{
				Config: testAccNotificationConfigurationConfig_basic(rNameTwo, rDescriptionTwo, string(types.AggregationDurationShort), acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameTwo),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescriptionTwo),
					resource.TestCheckResourceAttr(resourceName, "aggregation_duration", string(types.AggregationDurationShort)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					testAccCheckNotificationConfigurationNotRecreated(&v1, &v2),
				),
			},
			{
				Config: testAccNotificationConfigurationConfig_basic(rNameTwo, rDescriptionTwo, string(types.AggregationDurationLong), acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameTwo),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescriptionTwo),
					resource.TestCheckResourceAttr(resourceName, "aggregation_duration", string(types.AggregationDurationLong)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "notifications", regexache.MustCompile(`configuration/.+$`)),
				),
			},
		},
	})
}

func testAccCheckNotificationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_notification_configuration" {
				continue
			}

			_, err := tfnotifications.FindNotificationConfigurationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Notifications, create.ErrActionCheckingDestroyed, tfnotifications.ResNameNotificationConfiguration, rs.Primary.Attributes[names.AttrARN], err)
			}

			return create.Error(names.Notifications, create.ErrActionCheckingDestroyed, tfnotifications.ResNameNotificationConfiguration, rs.Primary.Attributes[names.AttrARN], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckNotificationConfigurationExists(ctx context.Context, name string, notificationconfiguration *notifications.GetNotificationConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Notifications, create.ErrActionCheckingExistence, tfnotifications.ResNameNotificationConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Notifications, create.ErrActionCheckingExistence, tfnotifications.ResNameNotificationConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		resp, err := tfnotifications.FindNotificationConfigurationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return create.Error(names.Notifications, create.ErrActionCheckingExistence, tfnotifications.ResNameNotificationConfiguration, rs.Primary.Attributes[names.AttrARN], err)
		}

		*notificationconfiguration = *resp

		return nil
	}
}

func testAccObjectImportStateIdFunc(resourceName string) func(state *terraform.State) (string, error) {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrARN], nil
	}
}

func testAccCheckNotificationConfigurationNotRecreated(before, after *notifications.GetNotificationConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Arn), aws.ToString(after.Arn); before != after {
			return create.Error(names.Notifications, create.ErrActionCheckingNotRecreated, tfnotifications.ResNameNotificationConfiguration, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccNotificationConfigurationConfig_basic(rName, description, aggregation_duration, tagKey, tagValue string) string {
	return fmt.Sprintf(`

resource "aws_notifications_notification_configuration" "testAccObjectImportStateIdFromARNsFunc" {
  name                 = %[1]q
  description          = %[2]q
  aggregation_duration = %[3]q
  tags = {
    %[4]q = %[5]q
  }
}
`, rName, description, aggregation_duration, tagKey, tagValue)
}

func testAccNotificationConfigurationConfig_minimal(rName, description string) string {
	return fmt.Sprintf(`

resource "aws_notifications_notification_configuration" "testAccObjectImportStateIdFromARNsFunc" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}
