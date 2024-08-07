// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdevopsguru "github.com/hashicorp/terraform-provider-aws/internal/service/devopsguru"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccNotificationChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var channel types.NotificationChannel
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devopsguru_notification_channel.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "sns.0.topic_arn", snsTopicResourceName, names.AttrARN),
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

func testAccNotificationChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var channel types.NotificationChannel
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devopsguru_notification_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationChannelExists(ctx, resourceName, &channel),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdevopsguru.ResourceNotificationChannel, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccNotificationChannel_filters(t *testing.T) {
	ctx := acctest.Context(t)
	var channel types.NotificationChannel
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_devopsguru_notification_channel.test"
	snsTopicResourceName := "aws_sns_topic.test"
	messageType := string(types.NotificationMessageTypeNewInsight)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_filters(rName, messageType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "sns.0.topic_arn", snsTopicResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.message_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.message_types.0", messageType),
					resource.TestCheckResourceAttr(resourceName, "filters.0.severities.#", acctest.Ct3),
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

func testAccCheckNotificationChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DevOpsGuruClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsguru_notification_channel" {
				continue
			}

			_, err := tfdevopsguru.FindNotificationChannelByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*retry.NotFoundError](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameNotificationChannel, rs.Primary.ID, err)
			}

			return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameNotificationChannel, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckNotificationChannelExists(ctx context.Context, name string, channel *types.NotificationChannel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameNotificationChannel, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameNotificationChannel, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DevOpsGuruClient(ctx)

		resp, err := tfdevopsguru.FindNotificationChannelByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameNotificationChannel, rs.Primary.ID, err)
		}

		*channel = *resp

		return nil
	}
}

func testAccNotificationChannelConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_devopsguru_notification_channel" "test" {
  sns {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccNotificationChannelConfig_filters(rName, messageType string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_devopsguru_notification_channel" "test" {
  sns {
    topic_arn = aws_sns_topic.test.arn
  }

  filters {
    message_types = [%[2]q]
    severities    = ["LOW", "MEDIUM", "HIGH"]
  }
}
`, rName, messageType)
}
