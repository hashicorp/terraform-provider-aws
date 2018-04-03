package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsSESIdentityNotification_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	topicName := fmt.Sprintf("test-topic-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESIdentityNotificationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccAwsSESIdentityNotificationConfig_basic, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESIdentityNotificationExists("aws_ses_identity_notification.test"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccAwsSESIdentityNotificationConfig_update, domain, topicName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESIdentityNotificationExists("aws_ses_identity_notification.test"),
				),
			},
		},
	})
}

func testAccCheckAwsSESIdentityNotificationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_identity_notification" {
			continue
		}

		identity := rs.Primary.ID
		params := &ses.GetIdentityNotificationAttributesInput{
			Identities: []*string{aws.String(identity)},
		}

		response, err := conn.GetIdentityNotificationAttributes(params)
		if err != nil {
			return err
		}

		if response.NotificationAttributes[identity] != nil {
			return fmt.Errorf("SES Identity Notification %s still exists. Failing!", identity)
		}
	}

	return nil
}

func testAccCheckAwsSESIdentityNotificationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Identity Notification not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Identity Notification identity not set")
		}

		identity := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).sesConn

		params := &ses.GetIdentityNotificationAttributesInput{
			Identities: []*string{aws.String(identity)},
		}

		response, err := conn.GetIdentityNotificationAttributes(params)
		if err != nil {
			return err
		}

		if response.NotificationAttributes[identity] == nil {
			return fmt.Errorf("SES Identity Notification %s not found in AWS", identity)
		}

		return nil
	}
}

const testAccAwsSESIdentityNotificationConfig_basic = `
resource "aws_ses_identity_notification" "test" {
	identity = "${aws_ses_domain_identity.test.arn}"
	notification_type = "Complaint"
}

resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}
`
const testAccAwsSESIdentityNotificationConfig_update = `
resource "aws_ses_identity_notification" "test" {
	topic_arn = "${aws_sns_topic.test.arn}"
	identity = "${aws_ses_domain_identity.test.arn}"
	notification_type = "Complaint"
}

resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_sns_topic" "test" {
  name = "%s"
}
`
