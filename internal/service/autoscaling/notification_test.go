// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingNotification_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_notification.test"
	groups := []string{rName}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationDestroy(ctx, groups),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationExists(ctx, resourceName, groups),
					resource.TestCheckResourceAttr(resourceName, "group_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "group_names.*", rName),
					resource.TestCheckResourceAttr(resourceName, "notifications.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "notifications.*", "autoscaling:EC2_INSTANCE_LAUNCH"),
					resource.TestCheckTypeSetElemAttr(resourceName, "notifications.*", "autoscaling:EC2_INSTANCE_TERMINATE"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrTopicARN),
				),
			},
		},
	})
}

func TestAccAutoScalingNotification_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_notification.test"
	groups := []string{rName}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationDestroy(ctx, groups),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationExists(ctx, resourceName, groups),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfautoscaling.ResourceNotification(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingNotification_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_notification.test"
	groups1 := []string{rName}
	groups2 := []string{rName, rName + "-2"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationDestroy(ctx, groups2),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationExists(ctx, resourceName, groups1),
					resource.TestCheckResourceAttr(resourceName, "group_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "group_names.*", rName),
					resource.TestCheckResourceAttr(resourceName, "notifications.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "notifications.*", "autoscaling:EC2_INSTANCE_LAUNCH"),
					resource.TestCheckTypeSetElemAttr(resourceName, "notifications.*", "autoscaling:EC2_INSTANCE_TERMINATE"),
				),
			},

			{
				Config: testAccNotificationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationExists(ctx, resourceName, groups2),
					resource.TestCheckResourceAttr(resourceName, "group_names.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "group_names.*", rName),
					resource.TestCheckTypeSetElemAttr(resourceName, "group_names.*", rName+"-2"),
					resource.TestCheckResourceAttr(resourceName, "notifications.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "notifications.*", "autoscaling:EC2_INSTANCE_LAUNCH"),
					resource.TestCheckTypeSetElemAttr(resourceName, "notifications.*", "autoscaling:EC2_INSTANCE_TERMINATE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "notifications.*", "autoscaling:EC2_INSTANCE_LAUNCH_ERROR"),
				),
			},
		},
	})
}

func TestAccAutoScalingNotification_paginated(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_notification.test"
	var groups []string
	for i := 0; i < 20; i++ {
		groups = append(groups, fmt.Sprintf("%s-%d", rName, i))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationDestroy(ctx, groups),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfig_paginated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationExists(ctx, resourceName, groups),
					resource.TestCheckResourceAttr(resourceName, "group_names.#", "20"),
					resource.TestCheckResourceAttr(resourceName, "notifications.#", acctest.Ct3),
				),
			},
		},
	})
}

func testAccCheckNotificationExists(ctx context.Context, n string, groups []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		output, err := tfautoscaling.FindNotificationsByTwoPartKey(ctx, conn, groups, rs.Primary.ID)

		if err == nil && len(output) == 0 {
			err = tfresource.NewEmptyResultError(nil)
		}

		return err
	}
}

func testAccCheckNotificationDestroy(ctx context.Context, groups []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_notification" {
				continue
			}

			output, err := tfautoscaling.FindNotificationsByTwoPartKey(ctx, conn, groups, rs.Primary.ID)

			if err == nil && len(output) == 0 {
				err = tfresource.NewEmptyResultError(nil)
			}

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Auto Scaling Notification %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNotificationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_basic(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_autoscaling_notification" "test" {
  group_names = [aws_autoscaling_group.test.name]

  notifications = [
    "autoscaling:EC2_INSTANCE_LAUNCH",
    "autoscaling:EC2_INSTANCE_TERMINATE",
  ]

  topic_arn = aws_sns_topic.test.arn
}
`, rName))
}

func testAccNotificationConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_basic(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_autoscaling_group" "test2" {
  availability_zones   = [data.aws_availability_zones.available.names[1]]
  max_size             = 0
  min_size             = 0
  name                 = "%[1]s-2"
  launch_configuration = aws_launch_configuration.test.name
}

resource "aws_autoscaling_notification" "test" {
  group_names = [
    aws_autoscaling_group.test.name,
    aws_autoscaling_group.test2.name,
  ]

  notifications = [
    "autoscaling:EC2_INSTANCE_LAUNCH",
    "autoscaling:EC2_INSTANCE_TERMINATE",
    "autoscaling:EC2_INSTANCE_LAUNCH_ERROR",
  ]

  topic_arn = aws_sns_topic.test.arn
}
`, rName))
}

func testAccNotificationConfig_paginated(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_autoscaling_group" "test" {
  count = 20

  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = "%[1]s-${count.index}"
  launch_configuration = aws_launch_configuration.test.name
}

resource "aws_autoscaling_notification" "test" {
  group_names = aws_autoscaling_group.test[*].name

  notifications = [
    "autoscaling:EC2_INSTANCE_LAUNCH",
    "autoscaling:EC2_INSTANCE_TERMINATE",
    "autoscaling:TEST_NOTIFICATION"
  ]
  topic_arn = aws_sns_topic.test.arn
}`, rName))
}
