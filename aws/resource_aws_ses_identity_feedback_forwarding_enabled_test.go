package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccAWSSESIdentityFeedbackForwardingEnabled_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandString(10))
	resourceName := "aws_ses_identity_feedback_forwarding_enabled.test"
	forwardingEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers: testAccProviders,
		//CheckDestroy: testAccCheckSESDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESIdentityFeedbackForwardingEnabledConfig(domain, forwardingEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESIdentityFeedbackForwardingEnabledExists(resourceName),
				//resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
				//resource.TestCheckResourceAttr(resourceName, "domain", domain),
				//resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain1),
				),
			},
		},
	})
}

func testAccAwsSESIdentityFeedbackForwardingEnabledConfig(domain string, fowardingEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_ses_identity_feedback_forwarding_enabled" "test" {
  identity           = aws_ses_domain_identity.test.domain
  enabled = %v
}
`, domain, fowardingEnabled)

}

func testAccCheckAwsSESIdentityFeedbackForwardingEnabledExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Identity Feedback Forwarding Enabled not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Identity Feedback Forwarding not set")
		}

		identity := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).sesv2conn
		params := &sesv2.GetEmailIdentityInput{EmailIdentity: aws.String(identity)}

		_, err := conn.GetEmailIdentity(params)
		if err != nil {
			return err
		}

		//fmt.Printf("res: %v", res)
		// TODO check if valid identity with res.VerifiedForSendingStatus

		return nil
	}
}
