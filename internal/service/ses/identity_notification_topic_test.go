// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESIdentityNotificationTopic_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	topicName := sdkacctest.RandomWithPrefix("test-topic")
	resourceName := "aws_ses_identity_notification_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityNotificationTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccIdentityNotificationTopicConfig_basic, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityNotificationTopicExists(ctx, resourceName),
				),
			},
			{
				Config: fmt.Sprintf(testAccIdentityNotificationTopicConfig_update, domain, topicName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityNotificationTopicExists(ctx, resourceName),
				),
			},
			{
				Config: fmt.Sprintf(testAccIdentityNotificationTopicConfig_headers, domain, topicName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityNotificationTopicExists(ctx, resourceName),
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

func testAccCheckIdentityNotificationTopicDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_identity_notification_topic" {
				continue
			}

			identity := rs.Primary.Attributes["identity"]
			params := &ses.GetIdentityNotificationAttributesInput{
				Identities: []string{identity},
			}

			log.Printf("[DEBUG] Testing SES Identity Notification Topic Destroy: %#v", params)

			response, err := conn.GetIdentityNotificationAttributes(ctx, params)
			if err != nil {
				return err
			}

			_, exists := response.NotificationAttributes[identity]
			if exists {
				return fmt.Errorf("SES Identity Notification Topic %s still exists. Failing!", identity)
			}
		}

		return nil
	}
}

func testAccCheckIdentityNotificationTopicExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Identity Notification Topic not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Identity Notification Topic identity not set")
		}

		identity := rs.Primary.Attributes["identity"]
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		params := &ses.GetIdentityNotificationAttributesInput{
			Identities: []string{identity},
		}

		log.Printf("[DEBUG] Testing SES Identity Notification Topic Exists: %#v", params)

		response, err := conn.GetIdentityNotificationAttributes(ctx, params)
		if err != nil {
			return err
		}

		_, exists := response.NotificationAttributes[identity]
		if !exists {
			return fmt.Errorf("SES Identity Notification Topic %s not found in AWS", identity)
		}

		notificationType := rs.Primary.Attributes["notification_type"]
		headersExpected, _ := strconv.ParseBool(rs.Primary.Attributes["include_original_headers"])

		var headersIncluded bool
		switch notificationType {
		case string(awstypes.NotificationTypeBounce):
			headersIncluded = response.NotificationAttributes[identity].HeadersInBounceNotificationsEnabled
		case string(awstypes.NotificationTypeComplaint):
			headersIncluded = response.NotificationAttributes[identity].HeadersInComplaintNotificationsEnabled
		case string(awstypes.NotificationTypeDelivery):
			headersIncluded = response.NotificationAttributes[identity].HeadersInDeliveryNotificationsEnabled
		}

		if headersIncluded != headersExpected {
			return fmt.Errorf("Wrong value applied for include_original_headers for %s", identity)
		}

		return nil
	}
}

const testAccIdentityNotificationTopicConfig_basic = `
resource "aws_ses_identity_notification_topic" "test" {
  identity          = aws_ses_domain_identity.test.arn
  notification_type = "Complaint"
}

resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}
`

const testAccIdentityNotificationTopicConfig_update = `
resource "aws_ses_identity_notification_topic" "test" {
  topic_arn         = aws_sns_topic.test.arn
  identity          = aws_ses_domain_identity.test.arn
  notification_type = "Complaint"
}

resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_sns_topic" "test" {
  name = "%s"
}
`

const testAccIdentityNotificationTopicConfig_headers = `
resource "aws_ses_identity_notification_topic" "test" {
  topic_arn                = aws_sns_topic.test.arn
  identity                 = aws_ses_domain_identity.test.arn
  notification_type        = "Complaint"
  include_original_headers = true
}

resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_sns_topic" "test" {
  name = "%s"
}
`
