// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESIdentityNotificationTopic_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	topicName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_identity_notification_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityNotificationTopicConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityNotificationTopicExists(ctx, resourceName),
				),
			},
			{
				Config: testAccIdentityNotificationTopicConfig_update(domain, topicName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityNotificationTopicExists(ctx, resourceName),
				),
			},
			{
				Config: testAccIdentityNotificationTopicConfig_headers(domain, topicName),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/36275.
func TestAccSESIdentityNotificationTopic_Disappears_domainIdentity(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_identity_notification_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityNotificationTopicConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityNotificationTopicExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceDomainIdentity(), "aws_ses_domain_identity.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIdentityNotificationTopicExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		_, err := tfses.FindIdentityNotificationAttributesByIdentity(ctx, conn, rs.Primary.Attributes["identity"])

		return err
	}
}

func testAccIdentityNotificationTopicConfig_basic(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_identity_notification_topic" "test" {
  identity          = aws_ses_domain_identity.test.arn
  notification_type = "Complaint"
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}
`, domain)
}

func testAccIdentityNotificationTopicConfig_update(domain, topicName string) string {
	return fmt.Sprintf(`
resource "aws_ses_identity_notification_topic" "test" {
  topic_arn         = aws_sns_topic.test.arn
  identity          = aws_ses_domain_identity.test.arn
  notification_type = "Complaint"
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[2]q
}
`, domain, topicName)
}

func testAccIdentityNotificationTopicConfig_headers(domain, topicName string) string {
	return fmt.Sprintf(`
resource "aws_ses_identity_notification_topic" "test" {
  topic_arn                = aws_sns_topic.test.arn
  identity                 = aws_ses_domain_identity.test.arn
  notification_type        = "Complaint"
  include_original_headers = true
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[2]q
}
`, domain, topicName)
}
