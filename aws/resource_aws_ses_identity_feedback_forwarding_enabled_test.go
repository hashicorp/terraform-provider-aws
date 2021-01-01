package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSESIdentityFeedbackForwardingEnabled_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandString(10))
	resourceName := "aws_ses_identity_feedback_forwarding_enabled.test"
	forwardingEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESIdentityFeedbackForwardingEnabledDestroy,
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
	config := fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%v"
}

resource "aws_ses_identity_feedback_forwarding_enabled" "test" {
  identity = aws_ses_domain_identity.test.domain
  enabled  = "%v"
}

`, domain, fowardingEnabled)
	fmt.Printf("config: %v", config)
	return config

}

func testAccCheckAwsSESIdentityFeedbackForwardingEnabledExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Println("testAccCheckAwsSESIdentityFeedbackForwardingEnabledExists")
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
		return nil
	}
}

func testAccCheckAwsSESIdentityFeedbackForwardingEnabledDestroy(s *terraform.State) error {
	fmt.Println("testAccCheckAwsSESIdentityFeedbackForwardingEnabledDestroy")
	fmt.Printf("s: %v", s)
	conn := testAccProvider.Meta().(*AWSClient).sesv2conn
	list, err := conn.ListEmailIdentities(&sesv2.ListEmailIdentitiesInput{})
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_identity_feedback_forwarding_enabled" {
			continue
		}

		identity := rs.Primary.ID
		for _, item := range list.EmailIdentities {
			if identity == *item.IdentityName {
				return fmt.Errorf("SES Email identity still exists")
			}
		}
	}

	return nil
}
