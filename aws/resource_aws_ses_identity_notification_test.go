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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESIdentityNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsSESIdentityNotificationConfig, domain),
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
		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{aws.String(identity)},
		}

		response, err := conn.GetIdentityNotificationAttributes(params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[identity] != nil {
			return fmt.Errorf("SES Identity Notification %s still exists. Failing!", domain)
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
			return fmt.Errorf("SES Identity Notification name not set")
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

		if response.VerificationAttributes[identity] == nil {
			return fmt.Errorf("SES Identity Notification %s not found in AWS", identity)
		}

		return nil
	}
}

const testAccAwsSESIdentityNotificationConfig = `
resource "aws_ses_identity_notification" "test" {
	identity = "%s"
	notification_type = "Complaint"
}
`
