package aws

import (
	"fmt"
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
	forwardingEnabled := true

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
  domain           = aws_ses_domain_identity.test.domain
  enabled = %v
}
`, domain, fowardingEnabled)

}

func testAccCheckAwsSESIdentityFeedbackForwardingEnabledExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Identity Feedback Forwarding Enabled not found: %s", n)
		}

		//
		//if rs.Primary.ID == "" {
		//	return fmt.Errorf("SES Email Identity name not set")
		//}
		//
		//email := rs.Primary.ID
		//conn := testAccProvider.Meta().(*AWSClient).sesconn
		//params := &ses.GetIdentityVerificationAttributesInput{
		//	Identities: []*string{
		//		aws.String(email),
		//	},
		//}
		//response, err := conn.GetIdentityVerificationAttributes(params)
		//if err != nil {
		//	return err
		//}
		//
		//if response.VerificationAttributes[email] == nil {
		//	return fmt.Errorf("SES Email Identity %s not found in AWS", email)
		//}

		return nil
	}
}
